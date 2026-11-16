// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// server answers every request with h.
func server(t *testing.T, h http.HandlerFunc, opts ...Option) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func slow(d time.Duration, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
		case <-r.Context().Done():
		}
		_, _ = io.WriteString(w, body)
	}
}

func TestPerAttemptTimeoutIsATimedOutTransportError(t *testing.T) {
	c := server(t, slow(time.Second, `{}`), WithTimeout(20*time.Millisecond))
	_, err := c.Hash(context.Background(), HashParams{})
	var terr *TransportError
	if !errors.As(err, &terr) || !terr.TimedOut || terr.Operation() != "POST /v2/hash" {
		t.Fatalf("err %v", err)
	}
	if !strings.Contains(terr.Error(), "timed out") || terr.Unwrap() == nil {
		t.Error(terr.Error())
	}
}

func TestDeadlineBoundsTheWholeCall(t *testing.T) {
	c := server(t, slow(time.Second, `{}`), WithTimeout(0), WithDeadline(20*time.Millisecond),
		WithRetry(RetryPolicy{MaxRetries: 5}))
	start := time.Now()
	_, err := c.Hash(context.Background(), HashParams{})
	var terr *TransportError
	if !errors.As(err, &terr) || !terr.TimedOut || time.Since(start) > 500*time.Millisecond {
		t.Fatalf("err %v after %v", err, time.Since(start))
	}
}

func TestCancellationIsNotATimeout(t *testing.T) {
	c := server(t, slow(time.Second, `{}`))
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(20*time.Millisecond, cancel)
	_, err := c.Hash(ctx, HashParams{})
	var terr *TransportError
	if !errors.As(err, &terr) || terr.TimedOut || !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v", err)
	}
	if !strings.Contains(terr.Error(), "connection failed") {
		t.Error(terr.Error())
	}
}

func TestTimeoutWhileReadingTheBody(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":{"digest":"`)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}, WithTimeout(50*time.Millisecond))
	_, err := c.Hash(context.Background(), HashParams{})
	var terr *TransportError
	if !errors.As(err, &terr) || !terr.TimedOut {
		t.Fatalf("err %v", err)
	}
}

func TestBrokenBodyIsATransportError(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = io.WriteString(w, `{"data":`)
		conn, _, _ := w.(http.Hijacker).Hijack()
		_ = conn.Close()
	})
	_, err := c.Hash(context.Background(), HashParams{})
	var de *DecodeError
	var terr *TransportError
	if !errors.As(err, &terr) && !errors.As(err, &de) {
		t.Fatalf("err %v", err)
	}
}

func TestReadFailureClassifies(t *testing.T) {
	c, _ := NewClient("http://h")
	rq := &request{method: "POST", path: "/v2/hash"}
	live := context.Background()
	if _, ok := c.readFailure(rq, live, 200, io.ErrClosedPipe).(*TransportError); !ok {
		t.Error("a read error must be a transport error")
	}
	if _, ok := c.readFailure(rq, live, 200, errDecode).(*DecodeError); !ok {
		t.Error("a decode error must be a decode error")
	}
	done, cancel := context.WithCancel(live)
	cancel()
	if _, ok := c.readFailure(rq, done, 200, errDecode).(*TransportError); !ok {
		t.Error("an expired attempt must be a transport error")
	}
}

func TestDrainLeavesLongBodiesToClose(t *testing.T) {
	body := io.NopCloser(strings.NewReader(strings.Repeat("x", maxDrain*2)))
	drain(body)
}
