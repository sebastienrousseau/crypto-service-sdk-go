// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClock records waits and never sleeps.
type fakeClock struct {
	waits []time.Duration
	err   error
}

func (f *fakeClock) install(c *Client) {
	c.clock = clock{
		now: func() time.Time { return time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC) },
		sleep: func(_ context.Context, d time.Duration) error {
			f.waits = append(f.waits, d)
			return f.err
		},
		jitter: func(n int64) int64 { return n - 1 },
	}
}

// scripted serves the given statuses in order, then 200.
func scripted(t *testing.T, statuses []int, retryAfter string) (*Client, *atomic.Int32, *fakeClock) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1)) - 1
		if i < len(statuses) {
			if retryAfter != "" {
				w.Header().Set("Retry-After", retryAfter)
			}
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(statuses[i])
			_, _ = io.WriteString(w, `{"type":"urn:crypto-service:problem:rate-limited","title":"t","status":`+strconv.Itoa(statuses[i])+`,"detail":"d"}`)
			return
		}
		_, _ = io.WriteString(w, `{"data":{"digest":"00","algorithm":"sha256","length":1}}`)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, WithRetry(RetryPolicy{MaxRetries: 3}))
	if err != nil {
		t.Fatal(err)
	}
	fc := &fakeClock{}
	fc.install(c)
	return c, &n, fc
}

func TestRetriesTransientStatusesWithBackoff(t *testing.T) {
	c, n, fc := scripted(t, []int{429, 503, 504}, "")
	if _, err := c.Hash(context.Background(), HashParams{}); err != nil {
		t.Fatal(err)
	}
	want := []time.Duration{200*time.Millisecond - 1, 400*time.Millisecond - 1, 800*time.Millisecond - 1}
	if n.Load() != 4 || len(fc.waits) != 3 || fc.waits[0] != want[0] || fc.waits[2] != want[2] {
		t.Errorf("attempts %d, waits %v", n.Load(), fc.waits)
	}
}

func TestGivesUpAfterMaxRetries(t *testing.T) {
	c, n, _ := scripted(t, []int{503, 503, 503, 503, 503}, "")
	_, err := c.Hash(context.Background(), HashParams{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 503 || n.Load() != 4 {
		t.Errorf("err %v after %d attempts", err, n.Load())
	}
}

func TestStatefulRoutesAreNotRetriedAfter504(t *testing.T) {
	c, n, _ := scripted(t, []int{504}, "")
	if _, err := c.GenerateKeyPair(context.Background(), GenerateKeyPairParams{}); err == nil || n.Load() != 1 {
		t.Errorf("err %v after %d attempts", err, n.Load())
	}
	c2, n2, _ := scripted(t, []int{429, 503}, "")
	if _, err := c2.KmsCreateKey(context.Background(), KmsCreateKeyParams{}); err != nil || n2.Load() != 3 {
		t.Errorf("err %v after %d attempts", err, n2.Load())
	}
}

func TestOtherStatusesAreNotRetried(t *testing.T) {
	c, n, _ := scripted(t, []int{500}, "")
	if _, err := c.Hash(context.Background(), HashParams{}); err == nil || n.Load() != 1 {
		t.Errorf("err %v after %d attempts", err, n.Load())
	}
}

func TestRetryAfterIsHonouredOrEndsTheRetries(t *testing.T) {
	c, _, fc := scripted(t, []int{429}, "1")
	if _, err := c.Hash(context.Background(), HashParams{}); err != nil || fc.waits[0] != time.Second {
		t.Errorf("err %v, waits %v", err, fc.waits)
	}
	c2, n2, _ := scripted(t, []int{429}, "60")
	if _, err := c2.Hash(context.Background(), HashParams{}); err == nil || n2.Load() != 1 {
		t.Errorf("a 60 s Retry-After was waited out: %v", err)
	}
}

func TestSleepFailureEndsTheRetries(t *testing.T) {
	c, n, fc := scripted(t, []int{503}, "")
	fc.err = context.Canceled
	if _, err := c.Hash(context.Background(), HashParams{}); err == nil || n.Load() != 1 {
		t.Errorf("err %v after %d attempts", err, n.Load())
	}
}

func TestConnectionFailuresRetryOnlyIdempotentCalls(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	c, _ := NewClient("http://"+addr, WithRetry(RetryPolicy{MaxRetries: 2}))
	fc := &fakeClock{}
	fc.install(c)
	_, err = c.Hash(context.Background(), HashParams{})
	var terr *TransportError
	if !errors.As(err, &terr) || terr.TimedOut || len(fc.waits) != 2 {
		t.Errorf("err %v, waits %v", err, fc.waits)
	}
	fc.waits = nil
	if _, err := c.GenerateKeyPair(context.Background(), GenerateKeyPairParams{}); err == nil || len(fc.waits) != 0 {
		t.Errorf("a stateful call was retried after a lost connection: %v", fc.waits)
	}
}

func TestIsIdempotentFollowsTheTable(t *testing.T) {
	for path, want := range map[string]bool{
		"/v2/hash": true, "/v2/keys/generate": false, "/v2/keys/delete": false,
		"/v2/pq/keygen": false, "/v2/hpke/keygen": false, "/v2/pq/dsa/keygen": false,
		"/v2/kms/create-key": false, "/v2/opaque/register/init": false,
		"/v2/opaque/login/init": true, "/v2/keys/export": true,
	} {
		if got := isIdempotent(http.MethodPost, path); got != want {
			t.Errorf("POST %s: %v", path, got)
		}
	}
	if !isIdempotent(http.MethodGet, "/v2/keys/generate") {
		t.Error("GET must be idempotent")
	}
}

func TestBackoffIsCappedAndPolicyResolves(t *testing.T) {
	c := &Client{retry: resolveRetry(RetryPolicy{MaxRetries: -1, InitialDelay: time.Second, MaxDelay: 3 * time.Second})}
	c.clock.jitter = func(n int64) int64 { return n }
	for n, want := range map[int]time.Duration{1: time.Second, 2: 2 * time.Second, 3: 3 * time.Second, 9: 3 * time.Second} {
		if got := c.backoff(n); got != want {
			t.Errorf("backoff(%d) = %v, want %v", n, got, want)
		}
	}
	if c.retry.maxRetries != 0 {
		t.Error("negative MaxRetries must mean none")
	}
	d := resolveRetry(RetryPolicy{})
	if d.initial != DefaultRetryInitialDelay || d.max != DefaultRetryMaxDelay {
		t.Errorf("%+v", d)
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	cases := map[string]struct {
		d  time.Duration
		ok bool
	}{
		"":                              {0, false},
		" 2 ":                           {2 * time.Second, true},
		"0.5":                           {500 * time.Millisecond, true},
		"-1":                            {0, false},
		"NaN":                           {0, false},
		"1e300":                         {0, false},
		"Sun, 15 Nov 2026 12:00:03 GMT": {3 * time.Second, true},
		"Sun, 15 Nov 2026 11:00:00 GMT": {0, true},
		"soon":                          {0, false},
	}
	for in, want := range cases {
		d, ok := parseRetryAfter(in, now)
		if d != want.d || ok != want.ok {
			t.Errorf("%q: %v %v, want %v %v", in, d, ok, want.d, want.ok)
		}
	}
	if maxRetryAfter > math.MaxInt64 {
		t.Error("bound overflows")
	}
}

func TestSleep(t *testing.T) {
	if err := sleep(context.Background(), time.Millisecond); err != nil {
		t.Error(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
