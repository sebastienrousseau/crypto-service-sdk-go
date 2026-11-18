// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc serves canned responses without a network.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func cannedClient(tb testing.TB, status int, contentType string, body []byte) *Client {
	tb.Helper()
	c, err := NewClient("http://fuzz.invalid", WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": {contentType}},
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    r,
		}, nil
	})))
	if err != nil {
		tb.Fatal(err)
	}
	return c
}

// FuzzDecodeProblem feeds arbitrary error bodies to the problem
// parser: it must not panic, must always produce a problem with a
// status, must re-encode, and its message must quote neither the body
// nor the detail.
func FuzzDecodeProblem(f *testing.F) {
	f.Add(400, []byte(`{"type":"urn:crypto-service:problem:validation-failed","title":"Validation failed","status":400,"detail":"body must have required property 'data'","errors":[{"field":"/data","message":"is required"}]}`))
	f.Add(404, []byte(`{"type":"urn:crypto-service:problem:key-not-found","title":"Key not found","status":404,"detail":"Key not found","code":"KEY_NOT_FOUND"}`))
	f.Add(429, []byte(`{"type":"urn:crypto-service:problem:rate-limited","title":"Too Many Requests","status":429,"detail":"quota","limit":60,"resetSeconds":12}`))
	f.Add(502, []byte(`<html>bad gateway</html>`))
	f.Add(500, []byte(`{"detail":"x","status":"500","errors":"none"}`))
	f.Fuzz(func(t *testing.T, status int, body []byte) {
		status = 400 + abs(status)%200
		e := newAPIError("POST", "/v2/hash", status, "application/problem+json", body)
		if e.Problem.Type == "" && e.Problem.Detail == "" && !json.Valid(body) {
			t.Fatal("a non-JSON body did not become about:blank")
		}
		if _, err := json.Marshal(e.Problem); err != nil {
			t.Fatalf("problem does not re-encode: %v", err)
		}
		assertQuietMessage(t, e.Error(), body, e.Problem)
	})
}

// FuzzDecodeResponse feeds arbitrary success bodies to the response
// decoder through the whole client: no panic, and an error message that
// never quotes the body.
func FuzzDecodeResponse(f *testing.F) {
	f.Add([]byte(`{"data":{"digest":"ba7816bf","algorithm":"sha256","length":32}}`))
	f.Add([]byte(`{"data":"plaintext"}`))
	f.Add([]byte(`{"data":null}`))
	f.Add([]byte(`{"data":{"hashing":["sha256"]}} trailing`))
	f.Add([]byte(`{"bomFormat":"CycloneDX","components":[{"name":"x"}]}`))
	f.Fuzz(func(t *testing.T, body []byte) {
		c := cannedClient(t, 200, "application/json", body)
		ctx := context.Background()
		_, err1 := c.Hash(ctx, HashParams{})
		_, err2 := c.SecretboxOpen(ctx, SecretboxOpenParams{})
		_, err3 := c.Algorithms(ctx)
		_, err4 := c.GetCbom(ctx)
		for _, err := range []error{err1, err2, err3, err4} {
			if err != nil {
				assertQuietMessage(t, err.Error(), body, Problem{})
			}
		}
	})
}

// assertQuietMessage fails when msg quotes a long run of body or the
// problem's detail (short runs can match the fixed text by chance).
func assertQuietMessage(t *testing.T, msg string, body []byte, p Problem) {
	t.Helper()
	if len(msg) > 300 {
		t.Fatalf("message of %d bytes", len(msg))
	}
	fixed := "crypto-service: POST /v2/hash GET /v2/algorithms /v2/secretbox/open /v2/compliance/cbom: HTTP response does not decode as documented"
	for _, s := range []string{string(body), p.Detail} {
		if len(s) >= 24 && strings.Contains(msg, s) && !strings.Contains(fixed, s) &&
			!strings.Contains(p.Type+p.Code, s) {
			t.Fatalf("message %q quotes %q", msg, s)
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -(n + 1)
	}
	return n
}
