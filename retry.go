// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"math"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// The RetryPolicy defaults for a zero InitialDelay or MaxDelay.
const (
	DefaultRetryInitialDelay = 200 * time.Millisecond
	DefaultRetryMaxDelay     = 2 * time.Second
)

// RetryPolicy configures WithRetry. The zero value retries nothing.
type RetryPolicy struct {
	// MaxRetries is the number of retries after the first attempt.
	// Zero or negative means one attempt only.
	MaxRetries int
	// InitialDelay caps the first wait; each later cap doubles it, up
	// to MaxDelay, and the wait is drawn uniformly below the cap (full
	// jitter). Zero or negative means DefaultRetryInitialDelay.
	InitialDelay time.Duration
	// MaxDelay caps every wait. A Retry-After longer than MaxDelay ends
	// the retries and returns that response's error. Zero or negative
	// means DefaultRetryMaxDelay.
	MaxDelay time.Duration
}

// WithRetry retries a call that failed transiently, with exponential
// backoff and full jitter, or after the response's Retry-After (seconds
// or an HTTP date) when it sends one.
//
// A call that cannot create or delete anything twice (every GET, and
// every POST except the stateful routes) is retried after a lost
// connection, a timed-out attempt, 429, 503 or 504. A stateful call
// (keys/generate, keys/delete, every */keygen, kms/create-key and
// opaque/register/*) is retried only after 429 or 503, when the server
// certainly did not act on it; never after 504 or a lost connection.
//
// Without this option the client makes exactly one attempt per call.
func WithRetry(p RetryPolicy) Option {
	return func(c *Client) { c.retry = resolveRetry(p) }
}

// retryConfig is a resolved RetryPolicy.
type retryConfig struct {
	maxRetries int
	initial    time.Duration
	max        time.Duration
}

func resolveRetry(p RetryPolicy) retryConfig {
	return retryConfig{
		maxRetries: max(p.MaxRetries, 0),
		initial:    orDefault(p.InitialDelay, DefaultRetryInitialDelay),
		max:        orDefault(p.MaxDelay, DefaultRetryMaxDelay),
	}
}

// orDefault returns d, or def when d is not positive.
func orDefault(d, def time.Duration) time.Duration {
	if d <= 0 {
		return def
	}
	return d
}

// statefulRoute matches the routes that create or delete server state.
// After a dropped connection or a 504 the server may already have done
// the work, so these are retried only after 429 or 503. It is the
// TypeScript SDK's table (request.ts), unchanged.
var statefulRoute = regexp.MustCompile(`/(keygen|keys/generate|keys/delete|kms/create-key|opaque/register/)`)

// isIdempotent reports whether repeating method path cannot create or
// delete anything twice.
func isIdempotent(method, path string) bool {
	return method == http.MethodGet || !statefulRoute.MatchString(path)
}

// clock holds the time and randomness seams; tests replace them.
type clock struct {
	now    func() time.Time
	sleep  func(ctx context.Context, d time.Duration) error
	jitter func(n int64) int64
}

var systemClock = clock{now: time.Now, sleep: sleep, jitter: rand.Int64N}

// outcome is the result of one attempt, as the retry decision sees it.
type outcome struct {
	err        error
	status     int
	retryAfter string
	// noResponse marks an attempt that got no response while the call's
	// own context was still live: a lost connection or an attempt that
	// hit its per-attempt timeout.
	noResponse bool
}

// next reports whether attempt n, which produced o, is retried, and
// after how long.
func (c *Client) next(n int, idempotent bool, o outcome) (time.Duration, bool) {
	if n > c.retry.maxRetries || o.err == nil {
		return 0, false
	}
	if o.noResponse {
		return c.backoff(n), idempotent
	}
	if !retryableStatus(o.status, idempotent) {
		return 0, false
	}
	if d, ok := parseRetryAfter(o.retryAfter, c.clock.now()); ok {
		return d, d <= c.retry.max
	}
	return c.backoff(n), true
}

// retryableStatus: 429 and 503 always (the server did not act on the
// request), 504 only for an idempotent call (it may have).
func retryableStatus(status int, idempotent bool) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return true
	case http.StatusGatewayTimeout:
		return idempotent
	}
	return false
}

// backoff is the wait after attempt n: drawn uniformly from [0, cap),
// where cap is the initial delay doubled n-1 times, at most max.
func (c *Client) backoff(n int) time.Duration {
	limit := min(c.retry.initial, c.retry.max)
	for i := 1; i < n && limit < c.retry.max; i++ {
		limit = min(limit, c.retry.max/2) * 2
	}
	return time.Duration(c.clock.jitter(int64(limit)))
}

// maxRetryAfter bounds a Retry-After before it becomes a Duration.
const maxRetryAfter = float64(math.MaxInt64 / int64(time.Second))

// parseRetryAfter reads a Retry-After header: delay seconds (a
// non-negative number) or an HTTP date. A date in the past is a zero
// wait.
func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if secs, err := strconv.ParseFloat(value, 64); err == nil {
		if secs < 0 || secs > maxRetryAfter || math.IsNaN(secs) {
			return 0, false
		}
		return time.Duration(secs * float64(time.Second)), true
	}
	at, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	return max(at.Sub(now), 0), true
}

// sleep waits for d or until ctx ends, whichever is first.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
