// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package kit

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

// Mismatches lists where actual departs from template, as the kit's
// README defines matching: exact JSON values (strings after
// substitution), arrays item by item, objects with exactly the
// template's members, and {"$shape": schema} for values only a schema
// can describe. Both sides are decoded JSON (numbers as json.Number).
func Mismatches(template, actual any, captured map[string]string) []string {
	return mismatches(template, actual, captured, "")
}

func mismatches(template, actual any, captured map[string]string, at string) []string {
	if schema, ok := shapeOf(template); ok {
		return prefix(at, Validate(schema, actual))
	}
	switch t := template.(type) {
	case []any:
		return arrayMismatches(t, actual, captured, at)
	case map[string]any:
		return objectMismatches(t, actual, captured, at)
	}
	want, err := Substitute(template, captured)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", where(at), err)}
	}
	if !equalScalar(want, actual) {
		return []string{fmt.Sprintf("%s: expected %s", where(at), Capture(want))}
	}
	return nil
}

// shapeOf returns the schema of a {"$shape": schema} marker.
func shapeOf(t any) (map[string]any, bool) {
	obj, ok := t.(map[string]any)
	if !ok || len(obj) != 1 {
		return nil, false
	}
	schema, ok := obj["$shape"].(map[string]any)
	return schema, ok
}

func arrayMismatches(t []any, actual any, captured map[string]string, at string) []string {
	a, ok := actual.([]any)
	if !ok || len(a) != len(t) {
		return []string{fmt.Sprintf("%s: expected %d items", where(at), len(t))}
	}
	var out []string
	for i := range t {
		out = append(out, mismatches(t[i], a[i], captured, fmt.Sprintf("%s/%d", at, i))...)
	}
	return out
}

func objectMismatches(t map[string]any, actual any, captured map[string]string, at string) []string {
	a, ok := actual.(map[string]any)
	if !ok {
		return []string{where(at) + ": expected an object"}
	}
	var out []string
	for _, k := range unionKeys(t, a) {
		out = append(out, memberMismatches(t, a, k, captured, at+"/"+k)...)
	}
	return out
}

// unionKeys lists the template's members, sorted, then the extra ones.
func unionKeys(t, a map[string]any) []string {
	keys := slices.Sorted(maps.Keys(t))
	for k := range a {
		if _, ok := t[k]; !ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// memberMismatches compares one member present in either object.
func memberMismatches(t, a map[string]any, k string, captured map[string]string, at string) []string {
	tv, inT := t[k]
	av, inA := a[k]
	switch {
	case !inA:
		return []string{at + ": missing"}
	case !inT:
		return []string{at + ": unexpected"}
	}
	return mismatches(tv, av, captured, at)
}

// equalScalar compares two JSON scalars; numbers by value.
func equalScalar(want, got any) bool {
	wn, wok := want.(json.Number)
	gn, gok := got.(json.Number)
	if wok && gok {
		wf, err1 := wn.Float64()
		gf, err2 := gn.Float64()
		return err1 == nil && err2 == nil && wf == gf
	}
	return want == got
}

func where(at string) string {
	if at == "" {
		return "/"
	}
	return at
}

func prefix(at string, errs []string) []string {
	for i, e := range errs {
		errs[i] = where(at) + " " + e
	}
	return errs
}
