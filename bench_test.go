// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"bytes"
	"strings"
	"testing"
)

// Recorded inputs of the client-side scenarios.
var (
	benchParams   = HashParams{Algorithm: SHA256, Data: strings.Repeat("a", 1024)}
	benchResponse = []byte(`{"data":{"digest":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","algorithm":"sha256","length":32}}`)
	benchProblem  = []byte(`{"type":"urn:crypto-service:problem:validation-failed","title":"Validation failed","status":400,"detail":"body must have required property 'data'","instance":"/v2/hash","errors":[{"field":"/data","message":"must have required property 'data'"}]}`)
)

// BenchmarkSDK measures the client's own work with no network, under
// the scenario names every Crypto Service SDK uses. round_trip_hash, a
// loopback call to a live server, is in bench_live_test.go.
func BenchmarkSDK(b *testing.B) {
	b.Run("encode_request", benchEncodeRequest)
	b.Run("decode_response", benchDecodeResponse)
	b.Run("decode_problem", benchDecodeProblem)
}

func benchEncodeRequest(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := encode(benchParams); err != nil {
			b.Fatal(err)
		}
	}
}

func benchDecodeResponse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var env envelope[HashResult]
		if err := decode(bytes.NewReader(benchResponse), &env); err != nil || env.Data == nil {
			b.Fatal(err)
		}
	}
}

func benchDecodeProblem(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if e := newAPIError("POST", "/v2/hash", 400, "application/problem+json", benchProblem); e.Problem.Slug() != "validation-failed" {
			b.Fatal(e)
		}
	}
}
