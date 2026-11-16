// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestAPIErrorMessageLeavesOutTheDetail(t *testing.T) {
	raw := `{"type":"urn:crypto-service:problem:key-not-found","title":"Key not found","status":404,` +
		`"detail":"secret plaintext here","instance":"/v2/sign","code":"KEY_NOT_FOUND","resetSeconds":3}`
	e := newAPIError("POST", "/v2/sign", 404, "application/problem+json", []byte(raw))
	if got := e.Error(); got != "crypto-service: POST /v2/sign: HTTP 404 key-not-found (KEY_NOT_FOUND)" {
		t.Errorf("%q", got)
	}
	if e.Operation() != "POST /v2/sign" || e.Problem.Detail != "secret plaintext here" || e.Problem.Instance != "/v2/sign" {
		t.Errorf("%+v", e.Problem)
	}
	if string(e.Problem.Extensions["resetSeconds"]) != "3" {
		t.Errorf("extensions %v", e.Problem.Extensions)
	}
	var base Error
	if !errors.As(error(e), &base) {
		t.Error("APIError is not an Error")
	}
}

func TestNonProblemBodiesBecomeAboutBlank(t *testing.T) {
	for _, raw := range []string{"<html>proxy</html>", `[1]`, `{"detail":5}`, `null`, ``} {
		e := newAPIError("GET", "/", 502, "text/html", []byte(raw))
		p := e.Problem
		if p.Type != "about:blank" || p.Title != "Bad Gateway" || p.Status != 502 || !strings.Contains(p.Detail, "502") {
			t.Errorf("%q: %+v", raw, p)
		}
		if strings.Contains(e.Error(), "proxy") {
			t.Error("message repeats the body")
		}
	}
	if blankProblem(599).Title != "HTTP error" {
		t.Error("unknown status title")
	}
}

func TestProblemKeepsWrongTypedMembersAsExtensions(t *testing.T) {
	p, ok := parseProblem([]byte(`{"detail":"d","status":"x","errors":[{"field":"/a","message":"m"}],"title":null}`))
	if !ok || p.Status != 0 || string(p.Extensions["status"]) != `"x"` || len(p.Errors) != 1 || p.Errors[0].Field != "/a" {
		t.Errorf("%+v", p)
	}
	if _, ok := p.Extensions["title"]; !ok {
		t.Error("a null member was lost")
	}
}

func TestProblemRoundTrips(t *testing.T) {
	in := `{"code":"C","detail":"d","errors":[{"field":"/f","message":"m"}],"instance":"/i","limit":60,"status":400,"title":"T","type":"urn:crypto-service:problem:validation-failed"}`
	var p Problem
	if err := json.Unmarshal([]byte(in), &p); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(p)
	if err != nil || string(out) != in {
		t.Errorf("%s, %v", out, err)
	}
	if p.Slug() != "validation-failed" {
		t.Error(p.Slug())
	}
	minimal, _ := json.Marshal(Problem{Type: "about:blank"})
	if string(minimal) != `{"detail":"","status":0,"title":"","type":"about:blank"}` {
		t.Errorf("%s", minimal)
	}
	if err := json.Unmarshal([]byte(`{"title":"x"}`), &p); !errors.Is(err, errNotProblem) {
		t.Error(err)
	}
}

func TestSafeLabel(t *testing.T) {
	if got := safeLabel("a b\ncé"); got != "a?b?c?" {
		t.Errorf("%q", got)
	}
	long := safeLabel(strings.Repeat("x", 100))
	if len(long) != maxLabel+3 {
		t.Errorf("%d", len(long))
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

var _ net.Error = timeoutErr{}

func TestTimedOut(t *testing.T) {
	bg := context.Background()
	if !timedOut(bg, timeoutErr{}) || !timedOut(bg, context.DeadlineExceeded) || timedOut(bg, errors.New("x")) {
		t.Error("timedOut misclassifies")
	}
	ctx, cancel := context.WithTimeout(bg, 0)
	defer cancel()
	<-ctx.Done()
	if !timedOut(ctx, errors.New("x")) {
		t.Error("an expired context is a timeout")
	}
}

func TestDecodeErrorMessage(t *testing.T) {
	e := &DecodeError{Method: "POST", Path: "/v2/hash", Status: 200, Err: errors.New("invalid character 's' secret")}
	if strings.Contains(e.Error(), "secret") || e.Unwrap() == nil || e.Operation() != "POST /v2/hash" {
		t.Error(e.Error())
	}
}

func TestMediaType(t *testing.T) {
	if mediaType("Application/Problem+JSON; charset=utf-8") != "application/problem+json" || mediaType("") != "" {
		t.Error("mediaType")
	}
}
