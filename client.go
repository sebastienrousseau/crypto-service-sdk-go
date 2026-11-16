// SPDX-License-Identifier: Apache-2.0 OR MIT

package cryptoservice

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultTimeout bounds each attempt when WithTimeout is not given.
const DefaultTimeout = 30 * time.Second

// Client calls one crypto-server. It holds one pooled HTTP client and
// no per-call state, so a single Client is safe for concurrent use by
// any number of goroutines; create it once and share it.
type Client struct {
	baseURL     string
	apiKey      string
	token       string
	traceparent string
	autoTrace   bool
	httpClient  *http.Client
	ownsHTTP    bool
	timeout     time.Duration
	deadline    time.Duration
	retry       retryConfig
	clock       clock
}

// Option configures a Client in NewClient.
type Option func(*Client)

// WithAPIKey sends key in the x-api-key header on every authenticated
// call. An empty key sends no header.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithToken sends "Authorization: Bearer <token>" on every
// authenticated call, for a server that verifies JWTs (JWT_SECRET). An
// empty token sends no header. The server prefers the token when both
// it and an API key are configured.
func WithToken(token string) Option {
	return func(c *Client) { c.token = token }
}

// WithTimeout bounds each attempt, including reading the response, to
// d. Zero or a negative d turns the per-attempt timeout off; the
// caller's context and WithDeadline still apply.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithDeadline bounds a whole call, every attempt and every wait
// between retries included, to d. Zero or a negative d (the default)
// leaves the call bounded only by the per-attempt timeout, the retry
// budget and the caller's context.
func WithDeadline(d time.Duration) Option {
	return func(c *Client) { c.deadline = d }
}

// WithTraceparent sends value as the W3C traceparent header on every
// request. Surrounding space is trimmed; an empty value sends none.
func WithTraceparent(value string) Option {
	return func(c *Client) {
		c.traceparent = strings.TrimSpace(value)
		c.autoTrace = false
	}
}

// WithAutoTraceparent generates a fresh W3C traceparent (version 00,
// sampled) for every call; retries of one call share it.
func WithAutoTraceparent() Option {
	return func(c *Client) {
		c.traceparent = ""
		c.autoTrace = true
	}
}

// WithHTTPClient sends requests through a copy of hc, for a custom
// transport, proxy or test double. The copy refuses redirects whatever
// hc's CheckRedirect says, and hc itself is not modified. A nil hc is
// ignored.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			clone := *hc
			c.httpClient = &clone
			c.ownsHTTP = false
		}
	}
}

// WithTransport sends requests through rt, keeping the client's
// redirect refusal. A nil rt is ignored.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *Client) {
		if rt != nil {
			c.httpClient = &http.Client{Transport: rt}
			c.ownsHTTP = false
		}
	}
}

// errRedirect is returned for any redirect: following one would send
// the API key or token to whatever origin the Location names.
var errRedirect = errors.New("redirects are refused")

func refuseRedirect(*http.Request, []*http.Request) error { return errRedirect }

// newTransport is the pooled transport a Client uses unless one is
// injected: keep-alive on, HTTP/2 negotiated over TLS, TLS 1.2 or newer
// with certificate verification always on.
func newTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}
}

// NewClient returns a client for the server at baseURL, for example
// "http://127.0.0.1:3000". The URL must be absolute http or https,
// without credentials, query or fragment; a trailing slash is dropped.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if err := checkBaseURL(baseURL); err != nil {
		return nil, err
	}
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Transport: newTransport()},
		ownsHTTP:   true,
		timeout:    DefaultTimeout,
		retry:      resolveRetry(RetryPolicy{}),
		clock:      systemClock,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.httpClient.CheckRedirect = refuseRedirect
	if err := c.check(); err != nil {
		return nil, err
	}
	return c, nil
}

// checkBaseURL accepts an absolute http or https URL with a host and
// nothing that does not belong in a base URL.
func checkBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return errors.New("crypto-service: base URL does not parse")
	}
	switch {
	case u.Scheme != "http" && u.Scheme != "https", u.Host == "":
		return errors.New("crypto-service: base URL must be absolute http or https")
	case u.User != nil:
		return errors.New("crypto-service: base URL must not carry credentials; use WithAPIKey or WithToken")
	case u.RawQuery != "" || u.Fragment != "":
		return errors.New("crypto-service: base URL must not have a query or fragment")
	}
	return nil
}

// check refuses settings that would fail on every call or weaken TLS.
func (c *Client) check() error {
	for _, v := range []string{c.apiKey, c.token, c.traceparent} {
		if strings.ContainsFunc(v, isControl) {
			return errors.New("crypto-service: a credential or traceparent contains a control character")
		}
	}
	if t, ok := c.httpClient.Transport.(*http.Transport); ok && t.TLSClientConfig != nil && t.TLSClientConfig.InsecureSkipVerify {
		return errors.New("crypto-service: TLS certificate verification must stay on")
	}
	return nil
}

func isControl(r rune) bool { return r < 0x20 || r == 0x7f }

// Close releases the idle connections of the client's own transport. A
// client built with WithHTTPClient or WithTransport leaves the injected
// transport alone. The client must not be used after Close.
func (c *Client) Close() {
	if c.ownsHTTP {
		c.httpClient.CloseIdleConnections()
	}
}

// BaseURL returns the server URL the client calls.
func (c *Client) BaseURL() string { return c.baseURL }

// String describes the client without its credentials.
func (c *Client) String() string {
	return fmt.Sprintf("cryptoservice.Client{baseURL: %q, apiKey: %s, token: %s}",
		c.baseURL, redact(c.apiKey), redact(c.token))
}

// GoString is String, so %#v never prints a credential either.
func (c *Client) GoString() string { return c.String() }

// redact says whether a credential is set, never what it is.
func redact(v string) string {
	if v == "" {
		return "unset"
	}
	return "[REDACTED]"
}
