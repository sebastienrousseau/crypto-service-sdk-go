// SPDX-License-Identifier: Apache-2.0 OR MIT

// Package kit reads the conformance kit vendored in conformance/ and
// checks answers against it: substitution of captured values, JSON
// pointers, template matching and the JSON Schema subset its $shape
// markers use. It sends nothing itself; the conformance test does.
package kit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// Auth is how a vector authenticates.
type Auth struct {
	// None sends no credentials.
	None bool
	// Scopes is the token's scopes; ["crypto:admin"] for "admin".
	Scopes []string
}

// UnmarshalJSON reads "admin", "none" or {"scopes": [...]}.
func (a *Auth) UnmarshalJSON(raw []byte) error {
	var name string
	if json.Unmarshal(raw, &name) == nil {
		switch name {
		case "admin":
			*a = Auth{Scopes: []string{"crypto:admin"}}
			return nil
		case "none":
			*a = Auth{None: true}
			return nil
		}
		return fmt.Errorf("unknown auth %q", name)
	}
	var scoped struct {
		Scopes []string `json:"scopes"`
	}
	if err := json.Unmarshal(raw, &scoped); err != nil {
		return err
	}
	*a = Auth{Scopes: append([]string{}, scoped.Scopes...)}
	return nil
}

// Request is what a vector sends.
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
	RawBody *string           `json:"rawBody"`
}

// Expect is what a vector must get back.
type Expect struct {
	Status    int               `json:"status"`
	MediaType string            `json:"mediaType"`
	Headers   map[string]string `json:"headers"`
	Body      json.RawMessage   `json:"body"`
	Text      json.RawMessage   `json:"text"`
}

// Vector is one request and its expected response.
type Vector struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Auth        *Auth             `json:"auth"`
	Request     Request           `json:"request"`
	Expect      Expect            `json:"expect"`
	Capture     map[string]string `json:"capture"`
}

// Credentials returns the vector's auth, "admin" when it names none.
func (v Vector) Credentials() Auth {
	if v.Auth == nil {
		return Auth{Scopes: []string{"crypto:admin"}}
	}
	return *v.Auth
}

// File is one vector file of the kit.
type File struct {
	Name        string
	Group       string   `json:"group"`
	Description string   `json:"description"`
	Vectors     []Vector `json:"vectors"`
}

// Load reads every *.json file in dir, in name order.
func Load(dir string) ([]File, error) {
	names, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	slices.Sort(names)
	files := make([]File, 0, len(names))
	for _, name := range names {
		f, err := loadFile(name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(name), err)
		}
		files = append(files, f)
	}
	return files, nil
}

func loadFile(name string) (File, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return File{}, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	f.Name = filepath.Base(name)
	return f, nil
}

// Decode parses JSON with numbers kept as json.Number.
func Decode(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, errors.New("data after the JSON value")
	}
	return v, nil
}

var placeholder = regexp.MustCompile(`\$\{(\w+)\}`)

// Substitute replaces each ${name} in every string of v with the
// captured value; a name nothing captured is an error.
func Substitute(v any, captured map[string]string) (any, error) {
	switch x := v.(type) {
	case string:
		return substituteString(x, captured)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			s, err := Substitute(e, captured)
			if err != nil {
				return nil, err
			}
			out[i] = s
		}
		return out, nil
	case map[string]any:
		return substituteObject(x, captured)
	}
	return v, nil
}

func substituteString(s string, captured map[string]string) (string, error) {
	var missing error
	out := placeholder.ReplaceAllStringFunc(s, func(m string) string {
		name := m[2 : len(m)-1]
		v, ok := captured[name]
		if !ok {
			missing = fmt.Errorf("nothing captured as %s", name)
		}
		return v
	})
	return out, missing
}

func substituteObject(x map[string]any, captured map[string]string) (map[string]any, error) {
	out := make(map[string]any, len(x))
	for k, e := range x {
		s, err := Substitute(e, captured)
		if err != nil {
			return nil, err
		}
		out[k] = s
	}
	return out, nil
}

// Pointer returns the value at an RFC 6901 JSON pointer.
func Pointer(v any, at string) (any, bool) {
	if at == "" {
		return v, true
	}
	for _, part := range strings.Split(at, "/")[1:] {
		obj, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		if v, ok = obj[part]; !ok {
			return nil, false
		}
	}
	return v, true
}

// Capture renders a captured value as the kit keeps it: strings as
// they are, anything else as its JSON text.
func Capture(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	data, _ := json.Marshal(v)
	return string(data)
}
