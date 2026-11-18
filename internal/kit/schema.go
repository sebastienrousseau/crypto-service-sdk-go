// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package kit

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Validate checks v against the JSON Schema (2020-12) subset the kit's
// $shape markers use: type, const, pattern, format (date-time, uuid),
// required, properties, additionalProperties and items. It returns one
// line per failure; an unknown keyword is a failure, so the subset
// cannot silently fall behind the kit.
func Validate(schema map[string]any, v any) []string {
	var out []string
	for k, arg := range schema {
		check, ok := keywords[k]
		if !ok {
			out = append(out, "unsupported schema keyword "+k)
			continue
		}
		out = append(out, check(arg, v)...)
	}
	return out
}

var keywords map[string]func(arg, v any) []string

func init() {
	keywords = map[string]func(arg, v any) []string{
		"type":                 checkType,
		"const":                checkConst,
		"pattern":              checkPattern,
		"format":               checkFormat,
		"required":             checkRequired,
		"properties":           checkProperties,
		"additionalProperties": checkAdditional,
		"items":                checkItems,
	}
}

func checkType(arg, v any) []string {
	want, _ := arg.(string)
	if typeOf(v, want) {
		return nil
	}
	return []string{"must be " + want}
}

func typeOf(v any, want string) bool {
	switch x := v.(type) {
	case string:
		return want == "string"
	case bool:
		return want == "boolean"
	case nil:
		return want == "null"
	case []any:
		return want == "array"
	case map[string]any:
		return want == "object"
	case json.Number:
		return want == "number" || (want == "integer" && isInteger(x))
	}
	return false
}

func isInteger(n json.Number) bool {
	f, err := n.Float64()
	return err == nil && f == math.Trunc(f)
}

func checkConst(arg, v any) []string {
	if equalScalar(arg, v) {
		return nil
	}
	return []string{"must equal " + Capture(arg)}
}

func checkPattern(arg, v any) []string {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	p, _ := arg.(string)
	m, err := compilePattern(p)
	if err != nil {
		return []string{fmt.Sprintf("pattern %q: %v", p, err)}
	}
	if !m.match(s) {
		return []string{fmt.Sprintf("must match %q", p)}
	}
	return nil
}

// longRepeat is a character class repeated more times than RE2's limit
// of 1000, as in ^[0-9a-f]{4800}$.
var longRepeat = regexp.MustCompile(`(\[[^\]]+\])\{([0-9]{4,})\}`)

// pattern is a compiled kit pattern: an RE2 expression in which each
// over-long repeat became a named starred group, plus the exact length
// each such group must have.
type pattern struct {
	re      *regexp.Regexp
	lengths map[string]int
}

func compilePattern(p string) (*pattern, error) {
	lengths := map[string]int{}
	i := 0
	rewritten := longRepeat.ReplaceAllStringFunc(p, func(m string) string {
		sub := longRepeat.FindStringSubmatch(m)
		n, _ := strconv.Atoi(sub[2])
		name := fmt.Sprintf("kitlen%d", i)
		i++
		lengths[name] = n
		return "(?P<" + name + ">" + sub[1] + "*)"
	})
	re, err := regexp.Compile(rewritten)
	if err != nil {
		return nil, err
	}
	return &pattern{re: re, lengths: lengths}, nil
}

func (p *pattern) match(s string) bool {
	m := p.re.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	for i, name := range p.re.SubexpNames() {
		if n, ok := p.lengths[name]; ok && utf8.RuneCountInString(m[i]) != n {
			return false
		}
	}
	return true
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func checkFormat(arg, v any) []string {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	switch f, _ := arg.(string); f {
	case "date-time":
		if _, err := time.Parse(time.RFC3339Nano, strings.ToUpper(s)); err != nil {
			return []string{"must be a date-time"}
		}
	case "uuid":
		if !uuidPattern.MatchString(s) {
			return []string{"must be a uuid"}
		}
	default:
		return []string{"unsupported format " + f}
	}
	return nil
}

func checkRequired(arg, v any) []string {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	names, _ := arg.([]any)
	var out []string
	for _, n := range names {
		name, _ := n.(string)
		if _, ok := obj[name]; !ok {
			out = append(out, "missing required "+name)
		}
	}
	return out
}

func checkProperties(arg, v any) []string {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	props, _ := arg.(map[string]any)
	var out []string
	for name, sub := range props {
		if val, ok := obj[name]; ok {
			s, _ := sub.(map[string]any)
			out = append(out, prefixed("/"+name, Validate(s, val))...)
		}
	}
	return out
}

// checkAdditional applies its schema to every member; the kit uses it
// only on objects without properties.
func checkAdditional(arg, v any) []string {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	s, _ := arg.(map[string]any)
	var out []string
	for name, val := range obj {
		out = append(out, prefixed("/"+name, Validate(s, val))...)
	}
	return out
}

func checkItems(arg, v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	s, _ := arg.(map[string]any)
	var out []string
	for i, val := range arr {
		out = append(out, prefixed(fmt.Sprintf("/%d", i), Validate(s, val))...)
	}
	return out
}

func prefixed(at string, errs []string) []string {
	for i, e := range errs {
		errs[i] = at + " " + e
	}
	return errs
}
