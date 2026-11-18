// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package kit

import (
	"strings"
	"testing"
)

func TestMismatches(t *testing.T) {
	captured := map[string]string{"id": "k_1"}
	template := mustDecode(t, `{"data":{"keyId":"${id}","n":32,"list":[1,{"$shape":{"type":"string"}}],"ok":true,"none":null}}`)
	good := mustDecode(t, `{"data":{"keyId":"k_1","n":32.0,"list":[1,"s"],"ok":true,"none":null}}`)
	if m := Mismatches(template, good, captured); len(m) != 0 {
		t.Errorf("%v", m)
	}
	cases := map[string]string{
		`{"data":{"keyId":"k_2","n":32,"list":[1,"s"],"ok":true,"none":null}}`:       "/data/keyId: expected",
		`{"data":{"keyId":"k_1","n":32,"list":[1],"ok":true,"none":null}}`:           "/data/list: expected 2 items",
		`{"data":{"keyId":"k_1","n":32,"list":[1,2],"ok":true,"none":null}}`:         "/data/list/1 must be string",
		`{"data":{"keyId":"k_1","n":32,"list":[1,"s"],"none":null}}`:                 "/data/ok: missing",
		`{"data":{"keyId":"k_1","n":32,"list":[1,"s"],"ok":true,"none":null,"x":1}}`: "/data/x: unexpected",
		`{"data":5}`: "/data: expected an object",
		`{"data":{"keyId":"k_1","n":"32","list":[1,"s"],"ok":true,"none":null}}`: "/data/n: expected 32",
	}
	for actual, want := range cases {
		m := Mismatches(template, mustDecode(t, actual), captured)
		if len(m) != 1 || !strings.HasPrefix(m[0], want) {
			t.Errorf("%s: %v, want %q", actual, m, want)
		}
	}
}

func TestMismatchesEdges(t *testing.T) {
	captured := map[string]string{"id": "k_1"}
	if m := Mismatches(mustDecode(t, `"${missing}"`), "x", captured); len(m) != 1 || !strings.Contains(m[0], "nothing captured") {
		t.Errorf("%v", m)
	}
	if m := Mismatches(mustDecode(t, `[1]`), "x", captured); len(m) != 1 || !strings.HasPrefix(m[0], "/: expected 1 items") {
		t.Errorf("%v", m)
	}
	if m := Mismatches(mustDecode(t, `{"$shape":{"type":"integer"}}`), mustDecode(t, `1.5`), nil); len(m) != 1 || m[0] != "/ must be integer" {
		t.Errorf("%v", m)
	}
}

func TestEqualScalar(t *testing.T) {
	if equalScalar(mustDecode(t, `1e999`), mustDecode(t, `1e999`)) {
		t.Error("unparsable numbers compared equal")
	}
}
