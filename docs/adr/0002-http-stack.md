<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0002: net/http with a tuned transport, standard library only

- Status: accepted
- Date: 2026-11-15

## Context

The client needs connection reuse, HTTP/2 where the server offers it,
TLS with verification, timeouts, cancellation and safe concurrent use.
Every dependency it adds becomes a dependency of every service that
imports it, and a supply-chain risk for a security product.

## Decision

Use the Go standard library only: `net/http`, `encoding/json`,
`crypto/tls` (through `net/http`). `go.mod` has no requirements.

Each `Client` owns one `*http.Client` with its own `*http.Transport`:
keep-alive on, `ForceAttemptHTTP2`, `MaxIdleConns` 100 and
`MaxIdleConnsPerHost` 32, dial timeout 10 s, TLS handshake timeout
10 s, idle timeout 90 s, proxy from the environment, and TLS 1.2 or
newer with certificate verification. `CheckRedirect` returns an error
for every redirect. Per-attempt and whole-call bounds come from
contexts (`WithTimeout`, `WithDeadline`, the caller's context), not
from `http.Client.Timeout`, so a retry loop and a streamed body share
one model.

`WithHTTPClient` copies the injected client and forces the redirect
refusal on the copy; `WithTransport` wraps a transport the same way. An
injected `*http.Transport` with `InsecureSkipVerify` is refused at
construction.

Tools (staticcheck, gocyclo, gocognit, govulncheck, gorelease,
cyclonedx-gomod) run through `go run <module>@<version>`; they never
enter `go.mod`.

## Consequences

- `go get` adds this module and nothing else; govulncheck covers the
  standard library only.
- No HTTP/3; no client-side load balancing. Both can come from an
  injected transport.
- JSON decoding streams from the response body, so a large ciphertext
  is decoded without first being copied into a separate buffer.
