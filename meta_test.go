// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestProbesAndMetadata(t *testing.T) {
	ctx := context.Background()
	c, seen := fakeServer(t, 200, `{"status":"ok","statusCode":200,"uptime":1.5,"bomFormat":"CycloneDX","title":"T"}`)
	if h, err := c.Health(ctx); err != nil || h.Uptime != 1.5 || seen.path != "/health" {
		t.Errorf("%+v %v", h, err)
	}
	if r, err := c.Root(ctx); err != nil || r.Title != "T" || seen.method != http.MethodGet {
		t.Errorf("%+v %v", r, err)
	}
	if b, err := c.GetCbom(ctx); err != nil || b.BomFormat != "CycloneDX" || seen.path != "/v2/compliance/cbom" {
		t.Errorf("%+v %v", b, err)
	}
}

func TestLiveAndAlgorithms(t *testing.T) {
	ctx := context.Background()
	live, seen := fakeServer(t, 200, `{"status":"alive","uptime":7}`)
	if l, err := live.Live(ctx); err != nil || l.Uptime != 7 || seen.path != "/live" {
		t.Errorf("%+v %v", l, err)
	}
	a, _ := fakeServer(t, 200, `{"data":{"hashing":["sha256"]}}`)
	if algs, err := a.Algorithms(ctx); err != nil || algs["hashing"][0] != "sha256" {
		t.Errorf("%v %v", algs, err)
	}
}

func TestReadyReturnsNotReadyWithoutRetrying(t *testing.T) {
	calls := 0
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"status":"not_ready","checks":{"accepting":false,"keyStore":true}}`)
	}, WithRetry(RetryPolicy{MaxRetries: 3}))
	r, err := c.Ready(context.Background())
	if err != nil || r.Status != "not_ready" || r.Checks["accepting"] || calls != 1 {
		t.Errorf("%+v %v after %d calls", r, err, calls)
	}
}

func TestGetReturnsTheAPIError(t *testing.T) {
	c, _ := fakeServer(t, 401, `{"type":"urn:crypto-service:problem:unauthorized","title":"Unauthorized","status":401,"detail":"no"}`)
	_, err := c.Root(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Problem.Slug() != "unauthorized" {
		t.Errorf("%v", err)
	}
	if _, err := c.Algorithms(context.Background()); err == nil {
		t.Error("no error")
	}
}
