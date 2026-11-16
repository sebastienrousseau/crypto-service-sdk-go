// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

// live holds the server the live tests and the conformance kit run
// against, from the environment:
//
//	CRYPTO_SERVICE_URL         base URL, e.g. http://127.0.0.1:3103
//	CRYPTO_SERVICE_API_KEY     the server's CRYPTO_API_KEY
//	CRYPTO_SERVICE_JWT_SECRET  the server's JWT_SECRET (the kit runs each
//	                           file as its own principal, as its README
//	                           requires, which an API key cannot do)
//
// Without them the live tests skip, unless CRYPTO_SERVICE_REQUIRE_LIVE
// is set (make conformance sets it): then they fail.
type live struct {
	url, apiKey, secret string
}

func liveServer(t *testing.T) live {
	t.Helper()
	l := live{
		url:    os.Getenv("CRYPTO_SERVICE_URL"),
		apiKey: os.Getenv("CRYPTO_SERVICE_API_KEY"),
		secret: os.Getenv("CRYPTO_SERVICE_JWT_SECRET"),
	}
	if l.url != "" && l.apiKey != "" && l.secret != "" {
		return l
	}
	const msg = "CRYPTO_SERVICE_URL, CRYPTO_SERVICE_API_KEY and CRYPTO_SERVICE_JWT_SECRET must all be set to run the live tests"
	if os.Getenv("CRYPTO_SERVICE_REQUIRE_LIVE") != "" {
		t.Fatal(msg)
	}
	t.Skip(msg)
	return l
}

// mint returns an HS256 JWT for sub with scopes, valid for five
// minutes. Test-only: the SDK itself never signs anything.
func (l live) mint(sub string, scopes []string) string {
	now := time.Now().Unix()
	claims := map[string]any{"sub": sub, "scopes": scopes, "iat": now, "exp": now + 300}
	if iss := os.Getenv("CRYPTO_SERVICE_JWT_ISSUER"); iss != "" {
		claims["iss"] = iss
	}
	if aud := os.Getenv("CRYPTO_SERVICE_JWT_AUDIENCE"); aud != "" {
		claims["aud"] = aud
	}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	enc := base64.RawURLEncoding
	unsigned := enc.EncodeToString(header) + "." + enc.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(l.secret))
	mac.Write([]byte(unsigned))
	return unsigned + "." + enc.EncodeToString(mac.Sum(nil))
}

// apiKeyClient is a client authenticated with the API key.
func (l live) apiKeyClient(t *testing.T, opts ...cs.Option) *cs.Client {
	t.Helper()
	c, err := cs.NewClient(l.url, append([]cs.Option{cs.WithAPIKey(l.apiKey)}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

// tokenClient is a client authenticated with a JWT for its own
// principal, so the live tests do not spend the API key's quota of 60
// requests a minute, which the examples and README snippets share.
func (l live) tokenClient(t *testing.T, sub string, opts ...cs.Option) *cs.Client {
	t.Helper()
	token := l.mint(sub, []string{"crypto:admin"})
	c, err := cs.NewClient(l.url, append([]cs.Option{cs.WithToken(token)}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

// recorder is a RoundTripper that remembers the last response's status
// and headers, so the kit can check what typed methods do not return.
type recorder struct {
	mu     sync.Mutex
	next   http.RoundTripper
	status int
	header http.Header
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.next.RoundTrip(req)
	if err == nil {
		r.mu.Lock()
		r.status, r.header = resp.StatusCode, resp.Header
		r.mu.Unlock()
	}
	return resp, err
}

func (r *recorder) last() (int, http.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status, r.header
}
