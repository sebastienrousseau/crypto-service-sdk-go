// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// Error is implemented by every error a call returns: *APIError when
// the server answered with an error status, *TransportError when no
// response arrived, *DecodeError when a success response could not be
// read as documented. Use errors.As with a *Error target to catch all
// three, or with the concrete type to tell them apart.
//
// No Error message carries a credential, a key, a plaintext or a
// request or response body.
type Error interface {
	error
	// Operation returns the method and path of the failed call, for
	// example "POST /v2/sign".
	Operation() string
}

// APIError is a non-2xx response. Problem is the RFC 9457 problem the
// server sent; a body that is not a problem (an HTML page from a
// proxy, say) becomes an "about:blank" problem for the status.
type APIError struct {
	Method string
	Path   string
	// Status is the HTTP status code.
	Status int
	// ContentType is the response's media type, without parameters.
	ContentType string
	Problem     Problem
}

// Error gives the operation, the status, the problem type and its code
// (for example KEY_NOT_FOUND). It leaves out the detail: read
// Problem.Detail for it.
func (e *APIError) Error() string {
	msg := fmt.Sprintf("crypto-service: %s %s: HTTP %d %s", e.Method, e.Path, e.Status, safeLabel(e.Problem.Slug()))
	if e.Problem.Code != "" {
		msg += " (" + safeLabel(e.Problem.Code) + ")"
	}
	return msg
}

// Operation returns "<method> <path>".
func (e *APIError) Operation() string { return e.Method + " " + e.Path }

// maxLabel caps a server-supplied identifier quoted in a message.
const maxLabel = 64

// safeLabel keeps a server-supplied identifier printable and short, so
// a hostile server cannot inject text into a log line through it.
func safeLabel(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x21 || r > 0x7e {
			return '?'
		}
		return r
	}, s)
	if len(s) > maxLabel {
		s = s[:maxLabel] + "..."
	}
	return s
}

// TransportError reports a call that got no usable response: the
// connection failed, a redirect was refused, the attempt timed out, or
// the caller's context ended.
type TransportError struct {
	Method string
	Path   string
	// TimedOut is true when a timeout or deadline ended the call
	// (per-attempt timeout, WithDeadline or the context's deadline),
	// false for a cancellation or a connection failure.
	TimedOut bool
	// Err is the cause; errors.Is(err, context.Canceled) and
	// errors.Is(err, context.DeadlineExceeded) see through it.
	Err error
}

// Error describes the failure and its cause.
func (e *TransportError) Error() string {
	what := "connection failed"
	if e.TimedOut {
		what = "timed out"
	}
	return fmt.Sprintf("crypto-service: %s %s: %s: %v", e.Method, e.Path, what, e.Err)
}

// Unwrap returns the cause.
func (e *TransportError) Unwrap() error { return e.Err }

// Operation returns "<method> <path>".
func (e *TransportError) Operation() string { return e.Method + " " + e.Path }

// newTransportError wraps a failed exchange. ctx is the call's
// context; the attempt timed out when its error says so.
func newTransportError(rq *request, ctx context.Context, err error) *TransportError {
	return &TransportError{Method: rq.method, Path: rq.path, TimedOut: timedOut(ctx, err), Err: err}
}

func timedOut(ctx context.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// DecodeError reports a 2xx response whose body is not the documented
// JSON, or is larger than the client reads. The message names neither
// the body nor the decoder's complaint about it; Err has the cause.
type DecodeError struct {
	Method string
	Path   string
	Status int
	Err    error
}

// Error gives the operation and the status only.
func (e *DecodeError) Error() string {
	return fmt.Sprintf("crypto-service: %s %s: HTTP %d response does not decode as documented", e.Method, e.Path, e.Status)
}

// Unwrap returns the cause.
func (e *DecodeError) Unwrap() error { return e.Err }

// Operation returns "<method> <path>".
func (e *DecodeError) Operation() string { return e.Method + " " + e.Path }

// maxProblemBytes caps how much of an error body is read.
const maxProblemBytes = 1 << 20

// readAPIError reads an error response into an *APIError.
func readAPIError(rq *request, resp *http.Response) *APIError {
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxProblemBytes))
	return newAPIError(rq.method, rq.path, resp.StatusCode, mediaType(resp.Header.Get("Content-Type")), raw)
}

// newAPIError builds the error for status from its raw body.
func newAPIError(method, path string, status int, contentType string, raw []byte) *APIError {
	e := &APIError{Method: method, Path: path, Status: status, ContentType: contentType}
	if p, ok := parseProblem(raw); ok {
		e.Problem = p
	} else {
		e.Problem = blankProblem(status)
	}
	return e
}

// mediaType strips the parameters from a Content-Type value.
func mediaType(v string) string {
	t, _, _ := strings.Cut(v, ";")
	return strings.ToLower(strings.TrimSpace(t))
}
