// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
)

// request is one call as the transport sees it.
type request struct {
	method string
	path   string
	body   []byte
	header http.Header
	// public sends no credentials: health, liveness and readiness.
	public bool
	// accept is a non-2xx status decoded as a result instead of an
	// error (503 from /ready, whose body is the same document).
	accept int
	// raw passes every response to the handler (Raw).
	raw bool
}

// handler reads a response the attempt did not turn into an error. It
// runs while the attempt's context is live, so a streamed body is
// bounded by the same timeout.
type handler func(*http.Response) error

// execute runs a call: one attempt, then retries while the policy
// allows, all bounded by the deadline when one is set.
func (c *Client) execute(ctx context.Context, rq *request, h handler) error {
	if c.deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.deadline)
		defer cancel()
	}
	trace := c.traceHeader()
	idempotent := isIdempotent(rq.method, rq.path)
	for n := 1; ; n++ {
		o := c.attempt(ctx, rq, trace, h)
		d, again := c.next(n, idempotent, o)
		if !again || c.clock.sleep(ctx, d) != nil {
			return o.err
		}
	}
}

// attempt performs one HTTP exchange under the per-attempt timeout.
func (c *Client) attempt(ctx context.Context, rq *request, trace string, h handler) outcome {
	actx := ctx
	if c.timeout > 0 {
		var cancel context.CancelFunc
		actx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	req, err := c.newRequest(actx, rq, trace)
	if err != nil {
		return outcome{err: err}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return outcome{err: newTransportError(rq, ctx, err), noResponse: ctx.Err() == nil}
	}
	defer drain(resp.Body)
	ok := isSuccess(resp.StatusCode) || resp.StatusCode == rq.accept
	if !ok && !rq.raw {
		return outcome{err: readAPIError(rq, resp), status: resp.StatusCode, retryAfter: resp.Header.Get("Retry-After")}
	}
	if err := h(resp); err != nil {
		return outcome{err: c.readFailure(rq, actx, resp.StatusCode, err)}
	}
	if ok {
		return outcome{}
	}
	return outcome{err: errRawStatus, status: resp.StatusCode, retryAfter: resp.Header.Get("Retry-After")}
}

// errRawStatus marks a non-2xx response Raw has read: an error for the
// retry decision, a result for Raw's caller.
var errRawStatus = errors.New("non-2xx status")

// readFailure is the error for a response whose body could not be read
// or decoded: a transport error when the attempt ran out of time or the
// connection broke, a decode error otherwise.
func (c *Client) readFailure(rq *request, actx context.Context, status int, err error) error {
	if actx.Err() != nil || !errors.Is(err, errDecode) {
		return newTransportError(rq, actx, err)
	}
	return &DecodeError{Method: rq.method, Path: rq.path, Status: status, Err: err}
}

func isSuccess(status int) bool { return status >= 200 && status <= 299 }

// maxDrain is how much of an unread body is discarded so the
// connection can go back to the pool; anything longer closes it.
const maxDrain = 64 << 10

func drain(body io.ReadCloser) {
	_, _ = io.CopyN(io.Discard, body, maxDrain)
	_ = body.Close()
}

// newRequest builds the HTTP request. Only a Raw path that is not a
// valid URL path makes it fail.
func (c *Client) newRequest(ctx context.Context, rq *request, trace string) (*http.Request, error) {
	var body io.Reader
	if rq.body != nil {
		body = bytes.NewReader(rq.body)
	}
	req, err := http.NewRequestWithContext(ctx, rq.method, c.baseURL+rq.path, body)
	if err != nil {
		return nil, fmt.Errorf("crypto-service: %s %s: invalid request path", rq.method, rq.path)
	}
	for k, vs := range rq.header {
		req.Header[k] = vs
	}
	req.Header.Set("Accept", "application/json, application/problem+json")
	req.Header.Set("User-Agent", userAgent)
	if rq.body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if trace != "" {
		req.Header.Set("Traceparent", trace)
	}
	if !rq.public {
		c.setCredentials(req.Header)
	}
	return req, nil
}

func (c *Client) setCredentials(h http.Header) {
	if c.apiKey != "" {
		h.Set("X-Api-Key", c.apiKey)
	}
	if c.token != "" {
		h.Set("Authorization", "Bearer "+c.token)
	}
}

// traceHeader is the traceparent for one call: the configured value, a
// generated one, or none.
func (c *Client) traceHeader() string {
	if c.autoTrace {
		return newTraceparent(rand.Uint64)
	}
	return c.traceparent
}

// newTraceparent builds a W3C Trace Context version 00 header with a
// random trace ID and parent ID, flagged sampled. Neither ID may be all
// zeros, so each first byte is forced non-zero. The IDs correlate logs;
// they are not secrets, so a fast non-cryptographic source is enough.
func newTraceparent(next func() uint64) string {
	var id [24]byte
	for i := 0; i < len(id); i += 8 {
		v := next()
		for j := range 8 {
			id[i+j] = byte(v >> (8 * j))
		}
	}
	id[0] = id[0]%255 + 1
	id[16] = id[16]%255 + 1
	return "00-" + hex.EncodeToString(id[:16]) + "-" + hex.EncodeToString(id[16:]) + "-01"
}
