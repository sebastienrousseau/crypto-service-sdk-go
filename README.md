<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<p align="center">
  <img src="https://cloudcdn.pro/crypto-service/v1/logos/crypto-service.svg" alt="Crypto Service logo" width="128" />
</p>

<h1 align="center">Crypto Service SDK for Go</h1>

<p align="center">
  A typed Go client for the Crypto Service REST API: hashing, AEAD, server-held signing keys, post-quantum KEMs and signatures, HPKE and KMS from Go, with the cryptography kept on the server.
</p>

<p align="center">
  <a href="https://github.com/sebastienrousseau/crypto-service-sdk-go/actions"><img src="https://img.shields.io/github/actions/workflow/status/sebastienrousseau/crypto-service-sdk-go/ci.yml?style=for-the-badge&logo=github" alt="Build" height="28" /></a>
  <a href="https://pkg.go.dev/github.com/sebastienrousseau/crypto-service-sdk-go"><img src="https://img.shields.io/github/v/tag/sebastienrousseau/crypto-service-sdk-go?sort=semver&label=go%20module&style=for-the-badge&color=fc8d62&logo=go" alt="Registry" height="28" /></a>
  <a href="https://pkg.go.dev/github.com/sebastienrousseau/crypto-service-sdk-go"><img src="https://img.shields.io/badge/docs-pkg.go.dev-blue?style=for-the-badge&labelColor=555555&logo=go" alt="Docs" height="28" /></a>
  <a href="https://scorecard.dev/viewer/?uri=github.com/sebastienrousseau/crypto-service-sdk-go"><img src="https://img.shields.io/ossf-scorecard/github.com/sebastienrousseau/crypto-service-sdk-go?style=for-the-badge&label=OpenSSF%20Scorecard&logo=openssf" alt="OpenSSF Scorecard" height="28" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg?style=for-the-badge" alt="License: Apache-2.0 OR MIT" height="28" /></a>
  <a href="https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/docs/POLICIES.md"><img src="https://img.shields.io/badge/Go-1.26%2B-93450a.svg?style=for-the-badge&logo=go" alt="Go 1.26 or newer" height="28" /></a>
</p>

<p align="center">
  <img src="https://raw.githubusercontent.com/sebastienrousseau/crypto-service-sdk-go/main/.github/demo.gif" alt="A terminal running the client's examples against a local crypto-server: a SHA-256 digest, an AEAD round trip, an Ed25519 signature made with a server-held key, and an ML-KEM-768 encapsulation whose shared secrets match" width="100%" />
</p>

---

## Contents

**Getting started**

- [Install](#install): `go get`, standard library only
- [Requirements](#requirements): toolchain floor, server, platforms
- [Quick Start](#quick-start): one call, its errors, and concurrent calls

**The Crypto Service ecosystem**

- [The Crypto Service ecosystem](#the-crypto-service-ecosystem): the server and its packages, the six language SDKs and the website

**Library reference**

- [Capabilities at a glance](#capabilities-at-a-glance): the current surface by theme
- [Ecosystem comparison](#ecosystem-comparison): short matrix; full table at [`docs/COMPARISON.md`](docs/COMPARISON.md)
- [Benchmarks](#benchmarks): the scenarios; results at [`docs/BENCHMARKS.md`](docs/BENCHMARKS.md)
- [Features](#features): module-level capability list
- [Configuration](#configuration): client options
- [Examples](#examples): runnable example index

**Operational**

- [When not to use Crypto Service](#when-not-to-use-crypto-service): limitations
- [Development](#development): make targets, fuzzing, CI
- [Security](#security): guarantees and compliance
- [Documentation](#documentation): all reference docs
- [Stability guarantees](#stability-guarantees): SemVer axis, deprecated routes, minimum toolchain
- [License](#license)

---

## Install

### As a Go library

```sh
go get github.com/sebastienrousseau/crypto-service-sdk-go@v0.0.40
```

The module uses the Go standard library only: `go get` adds one line to
your `go.mod` and nothing to your `go.sum` beyond this module. It needs a
running crypto-server to talk to; from a checkout of
[crypto-service](https://github.com/sebastienrousseau/crypto-service),
run `pnpm install && pnpm -r run build`, then start
`node packages/crypto-server/dist/index.js` with `CRYPTO_API_KEY` set.

---

## Requirements

- **Go 1.26 or newer.** CI tests Go 1.26 and 1.27, the two releases the
  Go team supports; the floor moves when 1.26 leaves support
  ([policy](docs/POLICIES.md#minimum-go-toolchain)).
- **A crypto-server** reachable over HTTP or HTTPS, authenticated with an
  API key (`x-api-key`) or a JWT (`Authorization: Bearer`). CI tests
  against the commit pinned in [`ci/server.ref`](ci/server.ref).
- **Any platform Go builds for.** Pure Go, no cgo; CI runs on Linux.

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

func main() {
    client, err := cs.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
        cs.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    res, err := client.Hash(ctx, cs.HashParams{Algorithm: cs.SHA256, Data: "abc"})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(res.Digest) // ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}
```

`Hash` posts to `/v2/hash` and returns the `data` member of the answer.
Every method takes a `context.Context` first and blocks until the server
answers; that context-aware call is the whole API.

A refused request is an `*APIError` carrying the server's RFC 9457
problem; no response at all is a `*TransportError`:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"

    cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

func main() {
    client, err := cs.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
        cs.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
    if err != nil {
        log.Fatal(err)
    }
    _, err = client.Sign(context.Background(), cs.SignParams{KeyID: "k_AAAAAAAAAAAAAAAAAAAAAA", Message: "m"})
    var apiErr *cs.APIError
    if !errors.As(err, &apiErr) {
        log.Fatalf("want a problem, got %v", err)
    }
    fmt.Println(apiErr.Status, apiErr.Problem.Slug(), apiErr.Problem.Code) // 404 key-not-found KEY_NOT_FOUND
}
```

Concurrency is goroutines over one shared client. The `Client` is safe
for concurrent use and pools its keep-alive connections:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    "sync"

    cs "github.com/sebastienrousseau/crypto-service-sdk-go"
)

func main() {
    client, err := cs.NewClient(os.Getenv("CRYPTO_SERVICE_URL"),
        cs.WithAPIKey(os.Getenv("CRYPTO_SERVICE_API_KEY")))
    if err != nil {
        log.Fatal(err)
    }
    inputs := []string{"a", "b", "c"}
    digests := make([]cs.Hex, len(inputs))
    errs := make([]error, len(inputs))
    var wg sync.WaitGroup
    for i, in := range inputs {
        wg.Add(1)
        go func() {
            defer wg.Done()
            res, err := client.Hash(context.Background(), cs.HashParams{Algorithm: cs.BLAKE3, Data: in})
            if err != nil {
                errs[i] = err
                return
            }
            digests[i] = res.Digest
        }()
    }
    wg.Wait()
    if err := errors.Join(errs...); err != nil {
        log.Fatal(err)
    }
    fmt.Println(len(digests), "digests")
}
```

---

## The Crypto Service ecosystem

This module is one client of the Crypto Service suite, which releases at
one version, server first. The design is one engine (crypto-server) and
thin typed clients per language, all held to one conformance kit that
the server repository publishes and every client vendors byte for byte.

| Component | Purpose | Use case |
| :--- | :--- | :--- |
| [`crypto-service`][core] | The server and 18 npm packages (`@sebastienrousseau/crypto-*`: lib, server, CLI, TypeScript SDK, middleware, KMS, WASM, MCP, LSP, CBOM and more) | The engine behind every client |
| **`crypto-service-sdk-go`** | Typed Go client for the REST API | Go services |
| [`crypto-service-sdk-rust`][rust] | Typed Rust client for the REST API | Rust services |
| [`crypto-service-sdk-java`][java] | Typed Java client for the REST API | JVM services |
| [`crypto-service-sdk-dotnet`][dotnet] | Typed .NET client for the REST API | .NET services |
| [`crypto-service-sdk-python`][python] | Typed Python client for the REST API | Python services and notebooks |
| [`crypto-service-sdk-cpp`][cpp] | Typed C++ client for the REST API | C and C++ systems |
| [crypto-service.co][site] | Project website | Overview and documentation entry point |

---

## Capabilities at a glance

| Area | Capability | Status |
| :--- | :--- | :--- |
| Hashing and MACs | `Hash` (SHA-2, SHA-3, BLAKE2b, BLAKE3), `Mac`, `MacVerify` | Supported |
| Key derivation | `Kdf` (scrypt, HKDF, PBKDF2), `PasswordHash`, `PasswordVerify` (Argon2id) | Supported |
| Symmetric | `Encrypt`, `Decrypt`, `SecretboxSeal`, `SecretboxOpen`, `PasswordEncrypt`, `PasswordDecrypt`, `KeyWrap`, `KeyUnwrap` | Supported |
| Server-held keys | `GenerateKeyPair`, `ExportKey`, `KeysDelete`, `Sign`, `Verify`, `StreamSign`, `StreamVerify` | Supported |
| Post-quantum | ML-KEM (`MlKem*`), hybrid X25519 + ML-KEM (`Pq*`), ML-DSA (`PqSign*`, `PqVerify`), SLH-DSA (`PqHash*`) | Supported |
| Public-key encryption | Sealed boxes (classical and hybrid), `MultiRecipientEncrypt`, STREAM (`StreamPq*`, `StreamMultiPq*`), HPKE (`Hpke*`) | Supported |
| Payments | `StreamIso20022`: dual-signed ISO 20022 envelope verification | Supported |
| KMS | `KmsCreateKey`, `KmsWrap`, `KmsUnwrap`, `KmsGenerateDataKey`, `KmsEncrypt`, `KmsDecrypt` | Supported |
| OPAQUE | `OpaqueRegisterInit`, `OpaqueRegisterFinish`, `OpaqueLoginInit`, `OpaqueLoginFinish` (transport only; the protocol runs on the caller's side) | Supported |
| Metadata | `Health`, `Live`, `Ready`, `Root`, `Algorithms`, `GetDoraCompliance`, `GetCbom` | Supported |
| Escape hatch | `Raw`: any GET or POST, status and body, whatever the status | Supported |
| v1 OpenPGP routes, `/metrics` | Deprecated (sunset 2027-01-01) and operational | Not wrapped; reachable through `Raw` |

---

## Ecosystem comparison

The suite's SDKs share one method surface, one option set and one
resilience contract, and are held to the conformance kit that
crypto-server runs in its own test suite.

| Project | Language | Conformance kit | Types |
| :--- | :---: | :---: | :---: |
| **crypto-service-sdk-go** | **Go 1.26+** | **Every vector, in CI, against the pinned server** | **Hand-written** |
| `@sebastienrousseau/crypto-sdk` | TypeScript | The server's own suite runs the kit | Hand-written |
| crypto-service-sdk-rust, -java, -dotnet, -python, -cpp | Their languages | See each repository | Hand-written |
| Raw HTTP | Any | No | None |

See [`docs/COMPARISON.md`](docs/COMPARISON.md) for the complete matrix.

---

## Benchmarks

`testing.B` benchmarks under the scenario names every Crypto Service SDK
uses: `encode_request`, `decode_response` and `decode_problem` measure
the client's own work with no network; `round_trip_hash` is one loopback
`POST /v2/hash` to a local crypto-server and runs only when
`CRYPTO_SERVICE_URL` is set. No numbers are published yet: results will
be generated into [`docs/BENCHMARKS.md`](docs/BENCHMARKS.md) with their
environment, and CI smoke-runs every scenario once.

---

## Features

- **One method per v2 operation.** All 59 operations of the OpenAPI
  document, named after the TypeScript SDK (`Hash`, `Sign`,
  `MlKemEncapsulate`, ...) or, where it has none, after the
  `operationId`, plus the probes and the compliance documents.
- **Context first, goroutine safe.** Every method takes a
  `context.Context`; cancellation and deadlines work the usual way. One
  `*Client` serves any number of goroutines over one connection pool.
- **Encodings as the spec gives them.** `Hex` for keys, signatures,
  digests and KEM output; `Base64` for AEAD, box, stream and key-wrap
  output; `string` for text and opaque KMS values. `Bytes()`, `HexOf`
  and `Base64Of` convert; nothing is guessed. `HpkeOpen` and
  `KmsDecrypt` return the plaintext as text and as hex.
- **Typed errors.** `*APIError` (status and RFC 9457 problem: type,
  title, detail, instance, code, per-field errors), `*TransportError`
  (with `TimedOut`) and `*DecodeError`, all implementing `Error`. A
  non-problem error body becomes an `about:blank` problem.
- **Retries that respect state.** Opt-in `WithRetry`: exponential
  backoff with full jitter and `Retry-After`. Calls that create or delete
  server state are retried only after `429` or `503`.
- **Timeouts at two levels.** `WithTimeout` per attempt (30 s default),
  `WithDeadline` for the whole call including retries.
- **Tracing.** `WithTraceparent` or `WithAutoTraceparent` sends a W3C
  `traceparent`.
- **Identifies itself.** Every request carries
  `User-Agent: crypto-service-sdk-go/0.0.40`.

---

## Configuration

`NewClient(baseURL, opts...)` takes an absolute `http` or `https` base
URL without credentials, and these options:

| Option | Effect | Default |
| :--- | :--- | :--- |
| `WithAPIKey(key)` | Sends `x-api-key` on authenticated calls | No header |
| `WithToken(jwt)` | Sends `Authorization: Bearer <jwt>` | No header |
| `WithTimeout(d)` | Bounds each attempt, body included; zero turns it off | `DefaultTimeout`, 30 s |
| `WithDeadline(d)` | Bounds the whole call, retries and waits included | None |
| `WithRetry(RetryPolicy{MaxRetries, InitialDelay, MaxDelay})` | Retries transient failures as described above | No retries; 200 ms and 2 s when zero |
| `WithTraceparent(v)` / `WithAutoTraceparent()` | Sends a fixed or a per-call generated `traceparent` | None |
| `WithHTTPClient(hc)` / `WithTransport(rt)` | Sends through an injected client or transport (proxy, test double); redirects stay refused | A pooled transport: keep-alive, HTTP/2, TLS 1.2+ |

---

## Examples

Every example is a runnable program under [`examples/`](examples/). Run
one with
`CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/<name>`,
or all of them with `make examples`. CI builds each and runs it against
the pinned crypto-server.

| Example | Shows |
| :--- | :--- |
| [`hash`](examples/hash/main.go) | `Hash` with the SHA-256 test vector |
| [`aead`](examples/aead/main.go) | `Encrypt` and `Decrypt` round trip, `Base64.Bytes` |
| [`sign`](examples/sign/main.go) | A server-held Ed25519 key: `GenerateKeyPair`, `Sign` by keyId, `Verify`, `KeysDelete` |
| [`mlkem`](examples/mlkem/main.go) | ML-KEM-768: `MlKemGenerateKeyPair`, `MlKemEncapsulate`, `MlKemDecapsulate` by keyId |
| [`errors`](examples/errors/main.go) | `*APIError` with its problem and field errors, `*TransportError`, the `Error` base type |
| [`concurrent`](examples/concurrent/main.go) | Goroutines over one client, a context deadline, `WithRetry` |
| [`example_test.go`](example_test.go) | The pkg.go.dev examples |

---

## When not to use Crypto Service

- **You cannot run crypto-server.** This client does nothing without it;
  there is no offline mode and no local cryptography.
- **You need cryptography inside your process**, for example on a
  device with no network, or where a round trip per operation is too
  slow. Use Go's `crypto/...` packages or a vetted library directly.
- **You need the v1 OpenPGP routes.** They are deprecated and sunset on
  2027-01-01; the client does not wrap them (only `Raw` reaches them).
  `/metrics` is for scrapers and is not wrapped either.
- **You need stateful calls retried after a lost connection.** Key
  generation, deletion and registration are not resent when the server
  may already have acted; deciding that is left to the caller.

---

## Development

```sh
make versions fmt-check vet build
```

`make check` runs the version check, `gofmt -l`, `go vet`, staticcheck,
the gocyclo (cyclomatic 10) and gocognit (cognitive 15) ceilings,
govulncheck, `go build` and `go test -race`, which fails below 100.0%
statement coverage and holds every function to 60 lines, every file to
500 and a Halstead difficulty of 30. `make conformance` runs the
conformance kit, the live tests, every example and every code block of
this README against a running server; `make regression` starts the
pinned server and runs both. `make fuzz` runs the two fuzz targets and
`make bench` the benchmarks. CI adds a Go 1.26 and 1.27 matrix, fuzzing,
the gorelease API check, markdownlint, codespell, REUSE, CodeQL, OpenSSF
Scorecard and the DCO check. See [`DEVELOPMENT.md`](DEVELOPMENT.md).

---

## Security

The client sends JSON to the server you configure and decodes the
reply. It implements no cryptography, writes nothing to disk and logs
nothing. Error messages carry the operation, status, problem type and
code, never a credential, key, plaintext or body; `String` and
`GoString` redact the credentials; two fuzz targets hold the message
rule. Redirects are refused, so credentials never reach another origin,
and TLS certificate verification cannot be turned off (an injected
transport that skips it is refused). Response bodies are capped at
64 MiB and every attempt has a timeout. The module has no dependencies
outside the Go standard library; govulncheck, CodeQL and OpenSSF
Scorecard run in CI, actions are pinned by commit SHA, and every file is
REUSE-compliant.

Report vulnerabilities according to [`SECURITY.md`](SECURITY.md).

---

## Documentation

[API reference](https://pkg.go.dev/github.com/sebastienrousseau/crypto-service-sdk-go) ·
[Developer docs](DEVELOPMENT.md) ·
[Family map](#the-crypto-service-ecosystem)

Also: [Architecture](docs/ARCHITECTURE.md) ·
[Decisions](docs/adr/README.md) · [Policies](docs/POLICIES.md) ·
[Roadmap](docs/ROADMAP.md) · [Conformance kit](conformance/README.md) ·
[Changelog](CHANGELOG.md) · [Support](SUPPORT.md) ·
[Governance](GOVERNANCE.md) · [Security](SECURITY.md) ·
[Agent rules](AGENTS.md)

---

## Stability guarantees

The project is pre-1.0 and versions advance in `0.0.1` steps, in
lockstep with the Crypto Service suite. The breaking axis is the
exported Go API (function signatures, struct fields, option and
constant names) and the requests it sends; a release may break either
while pre-1.0, and the CHANGELOG calls out every break with a
`Breaking:` line, which the gorelease check in CI requires. The client
covers the v2 API only: the v1 routes are deprecated by the server and
sunset on 2027-01-01. The Go floor is the older of the two supported Go
releases and moves only when it leaves support
([policy](docs/POLICIES.md#minimum-go-toolchain)).

---

## License

This project is dual-licensed under either:

- Apache License, Version 2.0 ([`LICENSE-APACHE`](LICENSE-APACHE) or <https://www.apache.org/licenses/LICENSE-2.0>)
- MIT License ([`LICENSE-MIT`](LICENSE-MIT) or <https://opensource.org/licenses/MIT>)

at your option.

[core]: https://github.com/sebastienrousseau/crypto-service
[rust]: https://github.com/sebastienrousseau/crypto-service-sdk-rust
[java]: https://github.com/sebastienrousseau/crypto-service-sdk-java
[dotnet]: https://github.com/sebastienrousseau/crypto-service-sdk-dotnet
[python]: https://github.com/sebastienrousseau/crypto-service-sdk-python
[cpp]: https://github.com/sebastienrousseau/crypto-service-sdk-cpp
[site]: https://crypto-service.co
