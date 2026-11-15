// SPDX-License-Identifier: Apache-2.0 OR MIT

package kit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func mustDecode(t *testing.T, s string) any {
	t.Helper()
	v, err := Decode([]byte(s))
	if err != nil {
		t.Fatalf("%s: %v", s, err)
	}
	return v
}

func TestLoadReadsTheVendoredKit(t *testing.T) {
	files, err := Load(filepath.Join("..", "..", "conformance"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 10 || files[0].Name != "asymmetric.json" || len(files[0].Vectors) == 0 {
		t.Fatalf("%d files", len(files))
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}
	if !slices.IsSorted(names) {
		t.Error("files are not in name order")
	}
}

func TestLoadReportsBadFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"vectors":5}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "bad.json") {
		t.Errorf("%v", err)
	}
	if _, err := Load("[["); err == nil {
		t.Error("a bad glob loaded")
	}
	if err := os.Mkdir(filepath.Join(dir, "dir.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := loadFile(filepath.Join(dir, "dir.json")); err == nil {
		t.Error("read a directory")
	}
}

func TestAuth(t *testing.T) {
	var v Vector
	if got := v.Credentials(); got.None || !slices.Equal(got.Scopes, []string{"crypto:admin"}) {
		t.Errorf("default %+v", got)
	}
	for in, want := range map[string]Auth{
		`"admin"`:                    {Scopes: []string{"crypto:admin"}},
		`"none"`:                     {None: true},
		`{"scopes":[]}`:              {Scopes: []string{}},
		`{"scopes":["crypto:hash"]}`: {Scopes: []string{"crypto:hash"}},
	} {
		var a Auth
		if err := json.Unmarshal([]byte(in), &a); err != nil || a.None != want.None || !slices.Equal(a.Scopes, want.Scopes) {
			t.Errorf("%s: %+v %v", in, a, err)
		}
		v.Auth = &a
		if v.Credentials().None != want.None {
			t.Error("Credentials ignores Auth")
		}
	}
	for _, in := range []string{`"root"`, `5`} {
		var a Auth
		if err := json.Unmarshal([]byte(in), &a); err == nil {
			t.Errorf("%s accepted", in)
		}
	}
}

func TestDecode(t *testing.T) {
	if v := mustDecode(t, `{"n":1.50}`); v.(map[string]any)["n"].(json.Number) != "1.50" {
		t.Error("numbers are not kept")
	}
	for _, in := range []string{`{`, `1 2`} {
		if _, err := Decode([]byte(in)); err == nil {
			t.Errorf("%q decoded", in)
		}
	}
}

func TestSubstitute(t *testing.T) {
	captured := map[string]string{"a": "1", "b": "two"}
	got, err := Substitute(mustDecode(t, `{"x":["${a}-${b}",3,true],"y":"$a"}`), captured)
	if err != nil {
		t.Fatal(err)
	}
	if out, _ := json.Marshal(got); string(out) != `{"x":["1-two",3,true],"y":"$a"}` {
		t.Errorf("%s", out)
	}
	for _, in := range []string{`"${c}"`, `["${c}"]`, `{"k":"${c}"}`} {
		if _, err := Substitute(mustDecode(t, in), captured); err == nil {
			t.Errorf("%s: no error for a missing capture", in)
		}
	}
}

func TestPointerAndCapture(t *testing.T) {
	v := mustDecode(t, `{"data":{"a/b":{"c~d":5},"s":"x"}}`)
	if got, ok := Pointer(v, "/data/a~1b/c~0d"); !ok || Capture(got) != "5" {
		t.Errorf("%v %v", got, ok)
	}
	if got, ok := Pointer(v, "/data/s"); !ok || Capture(got) != "x" {
		t.Errorf("%v", got)
	}
	if got, ok := Pointer(v, ""); !ok || got == nil {
		t.Error("root pointer")
	}
	for _, at := range []string{"/nope", "/data/s/x"} {
		if _, ok := Pointer(v, at); ok {
			t.Errorf("%s resolved", at)
		}
	}
}
