// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RawRequest is a request Raw sends as given.
type RawRequest struct {
	// Method is GET or POST.
	Method string
	// Path starts with "/", for example "/v2/hash". It must not carry
	// a query string or fragment.
	Path string
	// Header holds extra headers, such as a Content-Type for a body
	// that is not JSON. The client's own headers (credentials,
	// User-Agent, Accept, traceparent) take precedence.
	Header http.Header
	// Body is sent as it is; with no Content-Type in Header it is sent
	// as application/json. Nil sends no body.
	Body []byte
	// Public sends no credentials.
	Public bool
}

// RawResponse is the server's answer to a RawRequest, whatever its
// status.
type RawResponse struct {
	Status int
	Header http.Header
	Body   []byte
}

// MediaType returns the response's Content-Type without parameters.
func (r *RawResponse) MediaType() string { return mediaType(r.Header.Get("Content-Type")) }

// Raw sends one request with the client's credentials, timeouts,
// retries and redirect refusal, and returns the response whatever its
// status: a non-2xx status is not an error here. It is the escape hatch
// for routes the typed methods leave out (the deprecated v1 routes,
// /metrics) and the conformance runner's way to send malformed input.
// The body read is capped like any other response.
func (c *Client) Raw(ctx context.Context, r RawRequest) (*RawResponse, error) {
	if err := checkRaw(r); err != nil {
		return nil, err
	}
	rq := &request{method: r.Method, path: r.Path, body: r.Body, header: r.Header.Clone(), public: r.Public, raw: true}
	var res *RawResponse
	err := c.execute(ctx, rq, func(resp *http.Response) error {
		body, err := readAll(resp.Body)
		res = &RawResponse{Status: resp.StatusCode, Header: resp.Header, Body: body}
		return err
	})
	if err != nil && !errors.Is(err, errRawStatus) {
		return nil, err
	}
	return res, nil
}

// readAll reads a whole body up to the cap; an oversized body is a
// decode failure, not a transport one.
func readAll(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(&capReader{r: body, n: maxResponseBytes})
	if errors.Is(err, ErrResponseTooLarge) {
		return nil, fmt.Errorf("%w: %w", errDecode, err)
	}
	return data, err
}

// checkRaw refuses a request the transport would mangle.
func checkRaw(r RawRequest) error {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		return errors.New("crypto-service: Raw sends GET or POST only")
	}
	if !strings.HasPrefix(r.Path, "/") || strings.ContainsAny(r.Path, "?#") {
		return errors.New("crypto-service: Raw path must start with / and carry no query or fragment")
	}
	return nil
}
