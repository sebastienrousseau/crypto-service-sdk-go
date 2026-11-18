<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Ecosystem comparison

How this client relates to the other ways of calling crypto-server.
Claims about other projects are limited to what they state; see each
repository for its own evidence.

| | crypto-service-sdk-go | TypeScript SDK (`@sebastienrousseau/crypto-sdk`) | Other suite SDKs (Rust, Java, .NET, Python, C++) | Raw HTTP |
| :--- | :--- | :--- | :--- | :--- |
| v2 operations | 59 of 59, one method each | The methods in its `index.ts` | See each repository | Any, by hand |
| Conformance kit | Every vector, through the public methods, against the pinned server in CI | Run by crypto-server's own test suite | See each repository | None |
| Resilience contract | Per-attempt timeout, deadline, retry table with full jitter and `Retry-After` | The reference for the table | Same contract by design | By hand |
| Redirects | Refused | Refused (`redirect: "error"`) | Same contract by design | Client-dependent |
| Error model | `*APIError` (RFC 9457), `*TransportError`, `*DecodeError` | `CryptoApiError`, `CryptoTransportError` | Same model by design | By hand |
| Runtime dependencies | Go standard library | None beyond `fetch` | See each repository | Client-dependent |
| Concurrency | Blocking calls with a context; goroutines | Promises | Each language's idiom | Client-dependent |

The TypeScript SDK also has `negotiateAlgorithm`, a local helper that
picks an algorithm from network and policy constraints. It calls no
route, so it is outside this client's scope (no local logic beyond
encodings).
