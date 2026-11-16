// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
	"github.com/sebastienrousseau/crypto-service-sdk-go/internal/kit"
)

const kitDir = "conformance"

// runner runs one kit file against the live server, as its own
// principal (sub "conformance-<group>"), keeping its captures.
type runner struct {
	live     live
	rec      *recorder
	hc       *http.Client
	sub      string
	captured map[string]string
	typed    *int
}

// TestConformance runs every vector of the vendored kit, in file order,
// through the public method for its route, or through Raw when no
// method can send its exact request (v1, /metrics, malformed input,
// unknown members, missing required members).
func TestConformance(t *testing.T) {
	l := liveServer(t)
	files, err := kit.Load(kitDir)
	if err != nil {
		t.Fatal(err)
	}
	rec := &recorder{next: http.DefaultTransport}
	hc := &http.Client{Transport: rec}
	total, typed := 0, 0
	for _, f := range files {
		r := &runner{live: l, rec: rec, hc: hc, sub: "conformance-" + f.Group, captured: map[string]string{}, typed: &typed}
		for _, v := range f.Vectors {
			total++
			t.Run(f.Group+"/"+v.ID, func(t *testing.T) { r.run(t, v) })
		}
	}
	t.Logf("conformance: %d vectors, %d through typed methods, %d through Raw", total, typed, total-typed)
}

// client returns a client with the vector's credentials.
func (r *runner) client(t *testing.T, auth kit.Auth) *cs.Client {
	opts := []cs.Option{cs.WithHTTPClient(r.hc)}
	if !auth.None {
		opts = append(opts, cs.WithToken(r.live.mint(r.sub, auth.Scopes)))
	}
	c, err := cs.NewClient(r.live.url, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func (r *runner) run(t *testing.T, v kit.Vector) {
	body, err := r.requestBody(v)
	if err != nil {
		t.Fatal(err)
	}
	c := r.client(t, v.Credentials())
	actual, err := r.send(t, c, v, body)
	if err != nil {
		t.Fatalf("%s %s: %v", v.Request.Method, v.Request.Path, err)
	}
	status, header := r.rec.last()
	r.check(t, v, status, header, actual)
	for name, at := range v.Capture {
		val, ok := kit.Pointer(actual, at)
		if !ok {
			t.Fatalf("capture %s: nothing at %s", name, at)
		}
		r.captured[name] = kit.Capture(val)
	}
}

// requestBody is the vector's body after substitution, as JSON text.
func (r *runner) requestBody(v kit.Vector) ([]byte, error) {
	if v.Request.RawBody != nil {
		return []byte(*v.Request.RawBody), nil
	}
	if v.Request.Body == nil {
		return nil, nil
	}
	decoded, err := kit.Decode(v.Request.Body)
	if err != nil {
		return nil, err
	}
	sub, err := kit.Substitute(decoded, r.captured)
	if err != nil {
		return nil, err
	}
	return json.Marshal(sub)
}

// send runs the vector and returns the response body as decoded JSON
// (or text, for a non-JSON body).
func (r *runner) send(t *testing.T, c *cs.Client, v kit.Vector, body []byte) (any, error) {
	if o, ok := ops[v.Request.Method+" "+v.Request.Path]; ok && v.Request.RawBody == nil && len(v.Request.Headers) == 0 {
		result, handled, err := o(context.Background(), c, body)
		if handled {
			*r.typed++
			return answer(result, err)
		}
	}
	t.Logf("%s %s via Raw", v.Request.Method, v.Request.Path)
	return r.raw(c, v, body)
}

// answer turns a typed method's result or *APIError into the JSON the
// server sent, re-encoded from the SDK's types.
func answer(result any, err error) (any, error) {
	var apiErr *cs.APIError
	if errors.As(err, &apiErr) {
		result = apiErr.Problem
	} else if err != nil {
		return nil, err
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return kit.Decode(data)
}

func (r *runner) raw(c *cs.Client, v kit.Vector, body []byte) (any, error) {
	header := http.Header{}
	for k, val := range v.Request.Headers {
		header.Set(k, val)
	}
	res, err := c.Raw(context.Background(), cs.RawRequest{
		Method: v.Request.Method, Path: v.Request.Path, Header: header, Body: body,
		Public: v.Credentials().None,
	})
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(res.MediaType(), "json") {
		return string(res.Body), nil
	}
	return kit.Decode(res.Body)
}

// check compares the answer with the vector's expectations.
func (r *runner) check(t *testing.T, v kit.Vector, status int, header http.Header, actual any) {
	t.Helper()
	if status != v.Expect.Status {
		t.Errorf("status %d, want %d", status, v.Expect.Status)
	}
	if mt, _, _ := strings.Cut(header.Get("Content-Type"), ";"); mt != v.Expect.MediaType {
		t.Errorf("media type %q, want %q", mt, v.Expect.MediaType)
	}
	for name, want := range v.Expect.Headers {
		if got := header.Get(name); got != want {
			t.Errorf("header %s %q, want %q", name, got, want)
		}
	}
	tmpl := v.Expect.Body
	if tmpl == nil {
		tmpl = v.Expect.Text
	}
	template, err := kit.Decode(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range kit.Mismatches(template, actual, r.captured) {
		t.Error(m)
	}
}
