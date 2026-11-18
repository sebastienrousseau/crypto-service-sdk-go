// SPDX-License-Identifier: Apache-2.0 OR MIT
// SPDX-FileCopyrightText: 2026 Sebastien Rousseau

package cryptoservice_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"testing"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

// rotatingToken authenticates each request with a JWT for a principal
// that changes every 50 requests, so a benchmark is not throttled by
// the server's quota of 60 requests a minute per principal.
type rotatingToken struct {
	live  live
	next  http.RoundTripper
	n     atomic.Int64
	token atomic.Value
}

func (r *rotatingToken) RoundTrip(req *http.Request) (*http.Response, error) {
	n := r.n.Add(1) - 1
	if n%50 == 0 {
		r.token.Store(r.live.mint(fmt.Sprintf("bench-%d-%d", os.Getpid(), n/50), []string{"crypto:hash"}))
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+r.token.Load().(string))
	return r.next.RoundTrip(req)
}

// BenchmarkSDK/round_trip_hash is one loopback POST /v2/hash to
// CRYPTO_SERVICE_URL: the client's work plus the server's. It is
// skipped when CRYPTO_SERVICE_URL is unset.
func BenchmarkSDK(b *testing.B) {
	b.Run("round_trip_hash", func(b *testing.B) {
		if os.Getenv("CRYPTO_SERVICE_URL") == "" {
			b.Skip("CRYPTO_SERVICE_URL is unset")
		}
		l := live{url: os.Getenv("CRYPTO_SERVICE_URL"), secret: os.Getenv("CRYPTO_SERVICE_JWT_SECRET")}
		rt := &rotatingToken{live: l, next: http.DefaultTransport}
		c, err := cs.NewClient(l.url, cs.WithTransport(rt))
		if err != nil {
			b.Fatal(err)
		}
		ctx := context.Background()
		params := cs.HashParams{Algorithm: cs.SHA256, Data: "abc"}
		b.ReportAllocs()
		for b.Loop() {
			if _, err := c.Hash(ctx, params); err != nil {
				b.Fatal(err)
			}
		}
	})
}
