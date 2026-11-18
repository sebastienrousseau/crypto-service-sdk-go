// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ProblemTypePrefix starts every problem type crypto-server sends.
const ProblemTypePrefix = "urn:crypto-service:problem:"

// Problem is an RFC 9457 problem details object
// (application/problem+json), as crypto-server sends for every error.
type Problem struct {
	// Type is "urn:crypto-service:problem:<slug>", or "about:blank".
	Type string
	// Title summarises the problem type.
	Title string
	// Status repeats the HTTP status code.
	Status int
	// Detail explains this occurrence. The server keeps it safe to
	// show a client; the client still leaves it out of Error().
	Detail string
	// Instance is the request path, when the server sends it.
	Instance string
	// Code is the crypto-lib, key-store or KMS error code, when the
	// server sends one (for example INVALID_HEX or KEY_NOT_FOUND).
	Code string
	// Errors lists each field of a validation-failed problem.
	Errors []ValidationError
	// Extensions holds every other member as raw JSON (for example
	// limit and resetSeconds of a tenant-quota 429), and any standard
	// member whose JSON type was not the documented one.
	Extensions map[string]json.RawMessage
}

// ValidationError is one failed field of a validation-failed problem.
type ValidationError struct {
	// Field is a JSON pointer or a body field name, e.g. "/algorithm".
	Field string `json:"field"`
	// Message says why the field was refused.
	Message string `json:"message"`
}

// Slug returns the part of Type after ProblemTypePrefix, or Type itself
// when it has no such prefix ("about:blank").
func (p Problem) Slug() string {
	return strings.TrimPrefix(p.Type, ProblemTypePrefix)
}

// parseProblem decodes raw as a problem. Like the TypeScript SDK, it
// accepts a JSON object whose detail is a string; anything else is not
// a problem.
func parseProblem(raw []byte) (Problem, bool) {
	var members map[string]json.RawMessage
	if json.Unmarshal(raw, &members) != nil || members == nil {
		return Problem{}, false
	}
	var p Problem
	if !take(members, "detail", &p.Detail) {
		return Problem{}, false
	}
	take(members, "type", &p.Type)
	take(members, "title", &p.Title)
	take(members, "status", &p.Status)
	take(members, "instance", &p.Instance)
	take(members, "code", &p.Code)
	take(members, "errors", &p.Errors)
	if len(members) > 0 {
		p.Extensions = members
	}
	return p, true
}

// take decodes members[name] into dst and removes it, reporting
// success. A member of the wrong JSON type stays in members, so it
// lands in Extensions instead of being lost.
func take(members map[string]json.RawMessage, name string, dst any) bool {
	raw, ok := members[name]
	if !ok || bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, dst) != nil {
		return false
	}
	delete(members, name)
	return true
}

// blankProblem stands in for an error body that is not a problem.
func blankProblem(status int) Problem {
	title := http.StatusText(status)
	if title == "" {
		title = "HTTP error"
	}
	return Problem{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: fmt.Sprintf("The server answered %d without a problem body", status),
	}
}

// MarshalJSON writes the problem back in its wire form: the four
// required members, the optional ones that are set, and the extensions.
func (p Problem) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, 7+len(p.Extensions))
	for k, v := range p.Extensions {
		out[k] = v
	}
	out["type"], out["title"], out["status"], out["detail"] = p.Type, p.Title, p.Status, p.Detail
	if p.Instance != "" {
		out["instance"] = p.Instance
	}
	if p.Code != "" {
		out["code"] = p.Code
	}
	if p.Errors != nil {
		out["errors"] = p.Errors
	}
	return json.Marshal(out)
}

// UnmarshalJSON reads a problem; a body without a string detail is an
// error.
func (p *Problem) UnmarshalJSON(raw []byte) error {
	parsed, ok := parseProblem(raw)
	if !ok {
		return errNotProblem
	}
	*p = parsed
	return nil
}
