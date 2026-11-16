// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice_test

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestVendoredKitMatchesItsSums fails when a kit file changes, appears
// or disappears: conformance/ is replaced only by re-vendoring it,
// byte for byte, from crypto-service at the commit in ci/server.ref.
func TestVendoredKitMatchesItsSums(t *testing.T) {
	sums := readSums(t)
	entries, err := os.ReadDir(kitDir)
	if err != nil {
		t.Fatal(err)
	}
	var onDisk []string
	for _, e := range entries {
		if e.Name() != "SHA256SUMS" {
			onDisk = append(onDisk, e.Name())
		}
	}
	listed := make([]string, 0, len(sums))
	for name, want := range sums {
		listed = append(listed, name)
		data, err := os.ReadFile(filepath.Join(kitDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if got := sha256.Sum256(data); hex.EncodeToString(got[:]) != want {
			t.Errorf("%s does not match SHA256SUMS", name)
		}
	}
	slices.Sort(listed)
	if !slices.Equal(listed, onDisk) {
		t.Errorf("kit files %q, SHA256SUMS lists %q", onDisk, listed)
	}
}

func readSums(t *testing.T) map[string]string {
	t.Helper()
	f, err := os.Open(filepath.Join(kitDir, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sums := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sum, name, ok := strings.Cut(scanner.Text(), "  ")
		if !ok || len(sum) != 64 {
			t.Fatalf("malformed SHA256SUMS line %q", scanner.Text())
		}
		sums[name] = sum
	}
	return sums
}

// TestServerRefIsACommit keeps ci/server.ref a full commit SHA, which
// the conformance job checks out.
func TestServerRefIsACommit(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("ci", "server.ref"))
	if err != nil {
		t.Fatal(err)
	}
	ref := strings.TrimSpace(string(data))
	if _, err := hex.DecodeString(ref); err != nil || len(ref) != 40 {
		t.Errorf("ci/server.ref %q is not a 40-character commit SHA", ref)
	}
}
