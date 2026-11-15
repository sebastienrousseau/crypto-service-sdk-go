// SPDX-License-Identifier: Apache-2.0 OR MIT

package kit

import (
	"strings"
	"testing"
)

func schema(t *testing.T, s string) map[string]any {
	t.Helper()
	return mustDecode(t, s).(map[string]any)
}

func TestValidateTypes(t *testing.T) {
	for typ, value := range map[string]string{
		"string": `"s"`, "boolean": `true`, "null": `null`, "array": `[]`,
		"object": `{}`, "number": `1.5`, "integer": `2`,
	} {
		s := schema(t, `{"type":"`+typ+`"}`)
		if errs := Validate(s, mustDecode(t, value)); len(errs) != 0 {
			t.Errorf("%s %s: %v", typ, value, errs)
		}
		if typ != "number" {
			if errs := Validate(s, mustDecode(t, `1.5`)); len(errs) != 1 {
				t.Errorf("1.5 passed as %s", typ)
			}
		} else if errs := Validate(s, "x"); len(errs) != 1 {
			t.Error("a string passed as number")
		}
	}
	if errs := Validate(schema(t, `{"type":"string"}`), struct{}{}); len(errs) != 1 {
		t.Error("an unknown Go value passed")
	}
}

func TestValidatePatterns(t *testing.T) {
	hex := schema(t, `{"type":"string","pattern":"^[0-9a-f]{2368}$"}`)
	if errs := Validate(hex, strings.Repeat("a", 2368)); len(errs) != 0 {
		t.Error(errs)
	}
	for _, bad := range []string{strings.Repeat("a", 2367), strings.Repeat("a", 2369), strings.Repeat("g", 2368)} {
		if errs := Validate(hex, bad); len(errs) != 1 {
			t.Errorf("len %d passed", len(bad))
		}
	}
	two := schema(t, `{"pattern":"^[a]{1001}-[b]{1002}$"}`)
	if errs := Validate(two, strings.Repeat("a", 1001)+"-"+strings.Repeat("b", 1002)); len(errs) != 0 {
		t.Error(errs)
	}
	if errs := Validate(schema(t, `{"pattern":"^k_"}`), 5); len(errs) != 0 {
		t.Error("pattern applied to a number")
	}
	if errs := Validate(schema(t, `{"pattern":"("}`), "x"); len(errs) != 1 || !strings.Contains(errs[0], "pattern") {
		t.Error(errs)
	}
	if errs := Validate(schema(t, `{"pattern":"^\"[0-9]+\""}`), `"1"`); len(errs) != 0 {
		t.Error(errs)
	}
}

func TestValidateFormatsAndConst(t *testing.T) {
	dt := schema(t, `{"format":"date-time"}`)
	if errs := Validate(dt, "2026-11-15T11:03:17.270Z"); len(errs) != 0 {
		t.Error(errs)
	}
	uuid := schema(t, `{"format":"uuid"}`)
	if errs := Validate(uuid, "0af76519-16cd-43dd-8448-eb211c80319c"); len(errs) != 0 {
		t.Error(errs)
	}
	for s, v := range map[string]any{`{"format":"date-time"}`: "yesterday", `{"format":"uuid"}`: "x", `{"format":"email"}`: "a@b", `{"const":"A"}`: "B", `{"minimum":1}`: 1} {
		if errs := Validate(schema(t, s), v); len(errs) != 1 {
			t.Errorf("%s accepted %v", s, v)
		}
	}
	if errs := Validate(dt, 5); len(errs) != 0 {
		t.Error("format applied to a number")
	}
}

func TestValidateObjectsAndArrays(t *testing.T) {
	s := schema(t, `{"type":"object","required":["a","b"],"properties":{"a":{"const":"x"},"z":{"type":"string"}}}`)
	if errs := Validate(s, mustDecode(t, `{"a":"x","b":1}`)); len(errs) != 0 {
		t.Error(errs)
	}
	if errs := Validate(s, mustDecode(t, `{"a":"y"}`)); len(errs) != 2 {
		t.Error(errs)
	}
	add := schema(t, `{"additionalProperties":{"type":"array","items":{"type":"string"}}}`)
	if errs := Validate(add, mustDecode(t, `{"a":["x"],"b":[]}`)); len(errs) != 0 {
		t.Error(errs)
	}
	if errs := Validate(add, mustDecode(t, `{"a":["x",1]}`)); len(errs) != 1 || !strings.HasPrefix(errs[0], "/a /1 ") {
		t.Error(errs)
	}
	for _, kw := range []string{"required", "properties", "additionalProperties", "items"} {
		if errs := Validate(schema(t, `{"`+kw+`":{}}`), "scalar"); len(errs) != 0 {
			t.Errorf("%s applied to a scalar", kw)
		}
	}
}
