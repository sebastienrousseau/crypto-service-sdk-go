// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// maxResponseBytes caps how much of a success body the client reads.
// The server accepts 10 MiB plaintexts; their base64 or hex encodings,
// and a CBOM, stay well below 64 MiB. A variable so tests can lower it.
var maxResponseBytes int64 = 64 << 20

// ErrResponseTooLarge reports a success body over the client's cap.
var ErrResponseTooLarge = errors.New("response body exceeds the client's 64 MiB cap")

// errDecode marks a body that was read but is not the documented JSON.
var errDecode = errors.New("undecodable response body")

// errNotProblem reports JSON that is not an RFC 9457 problem.
var errNotProblem = errors.New("not a problem document: no string detail")

// errNoData reports a wrapped response without its "data" member.
var errNoData = errors.New(`response has no "data" member`)

// encode writes v as compact JSON, without HTML escaping, so what is
// sent is exactly the caller's text.
func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// capReader fails with ErrResponseTooLarge past n bytes instead of
// truncating silently.
type capReader struct {
	r io.Reader
	n int64
}

func (c *capReader) Read(p []byte) (int, error) {
	if c.n <= 0 {
		var probe [1]byte
		n, err := c.r.Read(probe[:])
		if n > 0 {
			return 0, ErrResponseTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > c.n {
		p = p[:c.n]
	}
	n, err := c.r.Read(p)
	c.n -= int64(n)
	return n, err
}

// decode streams one JSON value from body into out, then requires
// nothing but whitespace after it. A body that is not the documented
// JSON wraps errDecode; a failed read is returned as it is.
func decode(body io.Reader, out any) error {
	dec := json.NewDecoder(&capReader{r: body, n: maxResponseBytes})
	err := dec.Decode(out)
	if err == nil {
		err = expectEOF(dec)
	}
	if err == nil || !isDecodeFailure(err) {
		return err
	}
	return fmt.Errorf("%w: %w", errDecode, err)
}

// expectEOF fails when anything but whitespace follows the value.
func expectEOF(dec *json.Decoder) error {
	if _, err := dec.Token(); err != io.EOF {
		return errTrailingData
	}
	return nil
}

var errTrailingData = errors.New("data after the JSON value")

// isDecodeFailure separates a body that is wrong from a read that
// failed (a broken connection or an expired context).
func isDecodeFailure(err error) bool {
	var syntax *json.SyntaxError
	var typ *json.UnmarshalTypeError
	return errors.As(err, &syntax) || errors.As(err, &typ) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, ErrResponseTooLarge) || errors.Is(err, errTrailingData) ||
		errors.Is(err, errNoData) || errors.Is(err, errNotProblem)
}

// envelope is the {"data": ...} wrapper of most v2 responses. Data
// stays nil when the member is absent or null.
type envelope[T any] struct {
	Data *T `json:"data"`
}

// call sends method path with in as its JSON body (none when in is
// nil) and decodes a 2xx body into out.
func (c *Client) call(ctx context.Context, rq *request, in, out any) error {
	if in != nil {
		body, err := encode(in)
		if err != nil {
			return fmt.Errorf("crypto-service: %s %s: encoding the request: %w", rq.method, rq.path, err)
		}
		rq.body = body
	}
	return c.execute(ctx, rq, func(resp *http.Response) error { return decode(resp.Body, out) })
}

// wrapped posts in to path and returns the response's data member.
func wrapped[T any](ctx context.Context, c *Client, method, path string, in any) (*T, error) {
	var env envelope[T]
	if err := c.call(ctx, &request{method: method, path: path}, in, &env); err != nil {
		return nil, err
	}
	if env.Data == nil {
		return nil, &DecodeError{Method: method, Path: path, Status: http.StatusOK, Err: errNoData}
	}
	return env.Data, nil
}

// post is wrapped for the POST routes, which all take a JSON body.
func post[T any](ctx context.Context, c *Client, path string, in any) (*T, error) {
	return wrapped[T](ctx, c, http.MethodPost, path, in)
}

// postString is post for the routes whose data is a bare string.
func postString(ctx context.Context, c *Client, path string, in any) (string, error) {
	s, err := post[string](ctx, c, path, in)
	if err != nil {
		return "", err
	}
	return *s, nil
}

// empty is the body of a POST route that takes no parameters.
var empty = struct{}{}
