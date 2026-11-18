<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Changelog

All notable changes to this project are documented in this file. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and versions advance in `0.0.1` steps in lockstep with the Crypto
Service suite.

## [0.0.41] - Unreleased

### Added

- `docs/ASSURANCE.md`, the assurance case: claim, assets, six trust
  boundaries, the secure design principles applied and the CWE/OWASP
  items countered, each traced to the code and test that does it.
- `docs/security-review-2026-11.md`, the first recorded security
  review, with its findings and their disposition.
- `docs/REPRODUCING.md`: verify a release (tag signature,
  `SHA256SUMS`, provenance, sum.golang.org) and rebuild its archive,
  SBOM and module zip bit for bit, with the expected v0.0.40 hashes.
  SECURITY.md gains a "Verifying a release" section.
- Branch coverage held at 100% with gobco (`scripts/branch-coverage.sh`),
  in `make check` and in CI on both Go versions.

### Changed

- `scripts/release-artifacts.sh` builds reproducible release assets:
  the archive is compressed with `gzip -n -9`, and the SBOM has a
  derived serial number, no time stamp and a pinned cyclonedx-gomod
  build.
- Every source file states `SPDX-FileCopyrightText` next to its
  licence identifier.

### Fixed

- `NewClient` refuses a base URL with an empty query or fragment
  (`http://h/p?`, `http://h#`); every call then went to the wrong
  route.
- `NewClient` inspects the transport requests actually use: an
  injected `*http.Client` without a Transport sends through
  `http.DefaultTransport`, which was not checked, so disabled TLS
  verification there got past the refusal.
- The live tests no longer run under gobco, which ran them a third
  time and tripped the server's password-route rate limit in CI.

## [0.0.40] - 2026-11-16

The first release, in lockstep with the Crypto Service suite 0.0.40.
CI tests against crypto-service at commit
`3e8ecfd757e3577011953d8b47677f4c003426fb` (`ci/server.ref`).

### Added

- `NewClient` with `WithAPIKey` (`x-api-key`), `WithToken` (Bearer),
  `WithTimeout` (per attempt, 30 s by default), `WithDeadline` (whole
  call), `WithRetry`, `WithTraceparent`, `WithAutoTraceparent`,
  `WithHTTPClient` and `WithTransport`. One pooled transport per
  client (keep-alive, HTTP/2 over TLS, TLS 1.2 or newer); the client is
  safe for concurrent use.
- One method per v2 operation of the OpenAPI document, 59 in all, each
  taking a `context.Context` first, plus `Health`, `Live`, `Ready` and
  `Root`. Names follow the TypeScript SDK, or the `operationId` where it
  has none (`KeysDelete`, `MultiRecipientEncrypt`, `StreamSign`,
  `StreamVerify`, `StreamIso20022`, the `Opaque*` methods).
- `Hex` and `Base64` string types for the encodings the API documents,
  with `Bytes`, `HexOf` and `Base64Of`.
- `*APIError` with the RFC 9457 `Problem` (type, title, status, detail,
  instance, code, per-field errors, extensions), `*TransportError` with
  `TimedOut`, `*DecodeError`, and the `Error` interface all three
  implement. A non-problem error body becomes an `about:blank` problem.
  No message carries a credential, key, plaintext or body.
- Retries with exponential backoff, full jitter and `Retry-After`,
  using the TypeScript SDK's table: stateful routes retry only after
  429 or 503.
- Redirects refused; TLS verification cannot be turned off.
- `Raw`, for the routes the typed methods leave out (v1, `/metrics`).
- The conformance kit, vendored byte for byte from crypto-service at
  the pinned commit, run vector by vector through the public methods in
  CI; examples and README blocks run against the same server.
- Fuzz targets for the problem parser and the success decoder, and
  benchmarks under the suite's scenario names.

### Not included

- The deprecated v1 OpenPGP routes (sunset 2027-01-01) and `/metrics`
  are not wrapped; `Raw` reaches them.
