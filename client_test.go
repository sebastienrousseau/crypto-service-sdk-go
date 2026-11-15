// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClientRefusesBadBaseURLs(t *testing.T) {
	for _, u := range []string{
		"://bad", "ftp://h", "http://", "/relative",
		"http://user:pass@h", "http://h/?q=1", "http://h/#f",
	} {
		c, err := NewClient(u)
		if err == nil || c != nil {
			t.Errorf("%q: accepted", u)
			continue
		}
		if strings.Contains(err.Error(), "pass") {
			t.Errorf("%q: error %q repeats the credential", u, err)
		}
	}
}

func TestNewClientRefusesControlCharactersAndInsecureTLS(t *testing.T) {
	for _, opt := range []Option{WithAPIKey("a\nb"), WithToken("t\r"), WithTraceparent("00-\x7f")} {
		if _, err := NewClient("http://h", opt); err == nil {
			t.Error("accepted a control character")
		} else if strings.Contains(err.Error(), "a\nb") {
			t.Error("error repeats the credential")
		}
	}
	insecure := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if _, err := NewClient("https://h", WithTransport(insecure)); err == nil {
		t.Error("accepted a transport that skips TLS verification")
	}
	if _, err := NewClient("https://h", WithTransport(&http.Transport{TLSClientConfig: &tls.Config{}})); err != nil {
		t.Error(err)
	}
}

func TestClientDefaultsAndOptions(t *testing.T) {
	c, err := NewClient("http://h:1/", WithTimeout(time.Second), WithDeadline(2*time.Second),
		WithHTTPClient(nil), WithTransport(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.BaseURL() != "http://h:1" || c.timeout != time.Second || c.deadline != 2*time.Second || !c.ownsHTTP {
		t.Errorf("%+v", c)
	}
	tr := c.httpClient.Transport.(*http.Transport)
	if !tr.ForceAttemptHTTP2 || tr.MaxIdleConnsPerHost < 2 || tr.TLSClientConfig.MinVersion < tls.VersionTLS12 {
		t.Error("the default transport is not tuned")
	}
	c.Close()
}

func TestClientDefaults(t *testing.T) {
	d, _ := NewClient("http://h")
	if d.timeout != DefaultTimeout || d.retry.maxRetries != 0 {
		t.Errorf("defaults %+v", d.retry)
	}
}

func TestWithHTTPClientCopiesAndRefusesRedirects(t *testing.T) {
	hc := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return nil }}
	c, err := NewClient("http://h", WithHTTPClient(hc))
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient == hc || c.ownsHTTP || c.httpClient.CheckRedirect(nil, nil) == nil {
		t.Error("the injected client was used as is")
	}
	if hc.CheckRedirect(nil, nil) != nil {
		t.Error("the caller's client was modified")
	}
	c.Close() // leaves the injected transport alone
}

func TestRedirectsAreRefused(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("followed a redirect with x-api-key %q", r.Header.Get("X-Api-Key"))
	}))
	t.Cleanup(target.Close)
	srv := httptest.NewServer(http.RedirectHandler(target.URL, http.StatusTemporaryRedirect))
	t.Cleanup(srv.Close)
	c, _ := NewClient(srv.URL, WithAPIKey("secret-api-key"))
	_, err := c.Hash(context.Background(), HashParams{})
	if err == nil || !strings.Contains(err.Error(), "redirects are refused") {
		t.Fatalf("err %v", err)
	}
}

func TestStringNeverPrintsCredentials(t *testing.T) {
	c, _ := NewClient("http://h", WithAPIKey("secret-api-key"), WithToken("secret-token"))
	for _, s := range []string{c.String(), c.GoString(), fmt.Sprintf("%v %+v %#v %s", c, c, c, c)} {
		if strings.Contains(s, "secret") || !strings.Contains(s, "[REDACTED]") {
			t.Errorf("%q", s)
		}
	}
	bare, _ := NewClient("http://h")
	if !strings.Contains(bare.String(), "apiKey: unset") {
		t.Error(bare.String())
	}
}

func TestCredentialsAndHeaders(t *testing.T) {
	c, seen := fakeServer(t, 200, `{"data":{}}`)
	c.token = "tok"
	if _, err := c.Hash(context.Background(), HashParams{}); err != nil {
		t.Fatal(err)
	}
	h := seen.header
	if h.Get("X-Api-Key") != "test-key" || h.Get("Authorization") != "Bearer tok" ||
		h.Get("User-Agent") != "crypto-service-sdk-go/"+Version || h.Get("Traceparent") != "" {
		t.Errorf("headers %v", h)
	}
	if _, err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if seen.header.Get("X-Api-Key") != "" || seen.header.Get("Authorization") != "" {
		t.Error("credentials sent to a public route")
	}
}

func TestTraceparent(t *testing.T) {
	c, seen := fakeServer(t, 200, `{"data":{}}`)
	WithTraceparent("  00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01 ")(c)
	_, _ = c.Hash(context.Background(), HashParams{})
	if got := seen.header.Get("Traceparent"); got != "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01" {
		t.Errorf("static traceparent %q", got)
	}
	WithAutoTraceparent()(c)
	_, _ = c.Hash(context.Background(), HashParams{})
	first := seen.header.Get("Traceparent")
	_, _ = c.Hash(context.Background(), HashParams{})
	if first == seen.header.Get("Traceparent") || len(first) != 55 || !strings.HasPrefix(first, "00-") {
		t.Errorf("generated %q then %q", first, seen.header.Get("Traceparent"))
	}
}

func TestNewTraceparentNeverAllZero(t *testing.T) {
	got := newTraceparent(func() uint64 { return 0 })
	if got != "00-01000000000000000000000000000000-0100000000000000-01" {
		t.Errorf("%q", got)
	}
}
