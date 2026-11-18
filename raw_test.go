// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestRawReturnsAnyStatus(t *testing.T) {
	c, seen := fakeServer(t, 415, `{"type":"x","title":"t","status":415,"detail":"d"}`)
	res, err := c.Raw(context.Background(), RawRequest{
		Method: http.MethodPost, Path: "/v2/hash", Body: []byte("<a/>"),
		Header: http.Header{"Content-Type": {"application/xml"}},
	})
	if err != nil || res.Status != 415 || res.MediaType() != "application/json" || len(res.Body) == 0 {
		t.Fatalf("%+v %v", res, err)
	}
	if seen.header.Get("Content-Type") != "application/xml" || seen.header.Get("X-Api-Key") != "test-key" {
		t.Errorf("headers %v", seen.header)
	}
}

func TestRawPublicAndSuccess(t *testing.T) {
	c, seen := fakeServer(t, 200, `{}`)
	if _, err := c.Raw(context.Background(), RawRequest{Method: http.MethodGet, Path: "/health", Public: true}); err != nil {
		t.Fatal(err)
	}
	if seen.header.Get("X-Api-Key") != "" || seen.header.Get("Content-Type") != "" {
		t.Errorf("public raw headers %v", seen.header)
	}
	if res, err := c.Raw(context.Background(), RawRequest{Method: http.MethodPost, Path: "/v2/hash", Body: []byte(`{}`)}); err != nil || res.Status != 200 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestRawRetriesTransientStatuses(t *testing.T) {
	c, n, fc := scripted(t, []int{503}, "")
	res, err := c.Raw(context.Background(), RawRequest{Method: http.MethodGet, Path: "/v2/algorithms"})
	if err != nil || res.Status != 200 || n.Load() != 2 || len(fc.waits) != 1 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestRawRefusesBadRequests(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1")
	for _, r := range []RawRequest{
		{Method: http.MethodPut, Path: "/"},
		{Method: http.MethodGet, Path: "v2"},
		{Method: http.MethodGet, Path: "/v2?x=1"},
	} {
		if _, err := c.Raw(context.Background(), r); err == nil {
			t.Errorf("%+v accepted", r)
		}
	}
	if _, err := c.Raw(context.Background(), RawRequest{Method: http.MethodGet, Path: "/%zz"}); err == nil {
		t.Error("an invalid path was sent")
	}
	_, err := c.Raw(context.Background(), RawRequest{Method: http.MethodGet, Path: "/health"})
	var terr *TransportError
	if !errors.As(err, &terr) {
		t.Errorf("%v", err)
	}
}
