<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Assurance case

This is the argument that crypto-service-sdk-go meets its security
requirements, and the evidence for it. Every claim names the code or
the test that makes it true; if a claim and the code disagree, the code
is the defect. [SECURITY.md](project/security.md) has the policy and
the short form; the [2026-11 security review](security-review-2026-11.md)
checked this document against the code.

## The claim

Used as documented, the client:

1. sends the caller's credentials only to the server the caller named,
   only in headers, and never into a log, an error message or debug
   output;
2. does not weaken transport security: TLS certificate verification is
   always on, and TLS is 1.2 or newer on the client's own transport;
3. never sends a request that creates or deletes server state twice
   unless the server certainly did not act on the first;
4. treats every response as untrusted input: bounded in size and
   time, decoded strictly, and never echoed into an error message;
5. adds no third-party code to the caller's build, and ships releases
   that can be verified and rebuilt.

It does not claim to protect a caller that sends credentials over plain
`http://` across a network it does not trust (see
[Residual risks](#residual-risks)), nor anything about the cryptography,
which runs on crypto-server.

## Threat model

### Assets

| Asset | Why it matters |
| :--- | :--- |
| API key (`WithAPIKey`) and JWT (`WithToken`) | Grant use of the server's keys and quota |
| Request data: plaintexts, passwords, key IDs, wrapped keys | Confidential to the caller |
| Response data: plaintexts, exported private keys (`ExportKey`), KMS data keys | Confidential to the caller |
| Server-side state: generated keys, KMS keys, OPAQUE registrations | A duplicate or lost operation costs the caller |
| The caller's process: memory, goroutines, time | A hostile response must not exhaust them |
| The released module, archive and SBOM | Users build on them |

### Actors

| Actor | Trust |
| :--- | :--- |
| The caller (application code using the module) | Trusted: it chooses the server, the credentials and the options |
| crypto-server, as configured by the caller | Trusted to do the cryptography and to authorise; its responses are still parsed as untrusted input |
| A network attacker (passive or on-path) | Untrusted |
| A proxy, load balancer or impostor answering for the server | Untrusted: may redirect, send HTML, stall or send huge bodies |
| A pull-request author | Untrusted until review; CI runs their code with a read-only token |
| An upstream (Go toolchain, GitHub Actions, PyPI tools, the server repository) | Pinned and monitored, not trusted blindly |
| The maintainer | Trusted; the only one today (see [GOVERNANCE.md](project/governance.md)) |

### Entry points

- **The Go API**: `NewClient` and its options, the params structs of
  the typed methods, and `Raw` (any method path, headers and body the
  caller chooses).
- **The environment**: `HTTPS_PROXY`, `HTTP_PROXY` and `NO_PROXY`, read
  by `http.ProxyFromEnvironment` in `newTransport` (`client.go`).
- **The network**: every response's status, headers (`Retry-After`,
  `Content-Type`, `Location`) and body.
- **The repository and CI**: pull requests, `ci/server.ref` (which
  crypto-server commit CI builds), Dependabot updates, and release tags.

## Trust boundaries

```text
 caller code ──(1)── Client ──(2)── network ──(3)── crypto-server
                       │
                      (4) build: Go standard library only;
                          proxy.golang.org / sum.golang.org pin the module
 contributors ──(5)── CI (GitHub Actions) ──(6)── release (tag, GitHub, proxy)
```

1. **Caller to client.** In-process. The caller is trusted, but its
   configuration is checked once, in `NewClient` (`checkBaseURL` and
   `check` in `client.go`): an absolute `http`/`https` base URL with no
   user info, query or fragment, no control character in a credential
   or `traceparent`, and no transport that skips TLS verification.
   Request parameters are not validated client-side; the server
   validates them and answers with a problem.
2. **Client to network.** Everything leaving the process is a request
   the client built in one place (`newRequest`, `transport.go`), with
   credentials added by `setCredentials` unless the route is public.
   With `https://`, TLS protects it; with `http://` nothing does.
3. **Network to client.** Everything arriving is untrusted. `attempt`
   bounds it in time, `capReader` and `maxProblemBytes` in size,
   `decode` requires exactly one JSON value of the documented shape,
   and error messages carry no response text except
   `safeLabel`-filtered identifiers.
4. **Module to registries.** `go.mod` has no `require`: the caller's
   build gains no third-party code. The published module is the zip
   proxy.golang.org builds from the tag, pinned by sum.golang.org.
5. **Contributors to CI.** Workflows run on `pull_request`, never
   `pull_request_target`; the default token is `contents: read`
   (`read-all` for the Scorecard workflow, which never runs on a pull
   request);
   checkouts set `persist-credentials: false`; every action is pinned
   by commit SHA and every Python tool by hash
   (`ci/*-requirements.txt`). `ci/server.ref` is checked to be a full
   commit SHA before it is fetched (`ci.yml`, conformance job).
6. **CI to release.** The release workflow runs on a tag; it checks the
   tag is annotated, named and signed by a key in `KEYS.asc`, that every
   version restatement agrees, and re-runs the tests. Publishing runs in
   the `release` environment, which requires the maintainer's approval,
   and attests the build provenance of every asset
   (`.github/workflows/release.yml`).

## Secure design principles

How the Saltzer and Schroeder principles are applied:

| Principle | Where |
| :--- | :--- |
| **Economy of mechanism** | A thin client with no local cryptography ([ADR 0001](adr/0001-thin-client-over-the-server.md)); the standard library only ([ADR 0002](adr/0002-http-stack.md)); one request path for all 59 operations (`execute` and `attempt`, `transport.go`). |
| **Fail-safe defaults** | No retries unless `WithRetry` is given (`resolveRetry`, `retry.go`); a 30 s per-attempt timeout (`DefaultTimeout`); TLS 1.2 minimum (`newTransport`); redirects refused (`refuseRedirect`); a body that is too large, malformed or followed by trailing data is an error, not a partial result (`decode`, `codec.go`). |
| **Complete mediation** | Every request is built by `newRequest` and every response read by `attempt`; credentials are set in one function (`setCredentials`); `CheckRedirect` is assigned after the options run, so no option can re-enable redirects (`NewClient`); `check` inspects the transport requests actually use, `http.DefaultTransport` included when an injected client has none (`roundTripper`). |
| **Open design** | Everything is public: the code, the conformance kit the CI runs against a live server, the ADRs and this document. Nothing depends on a secret other than the caller's credentials. |
| **Separation of privilege** | A release needs both a tag signed by a key in `KEYS.asc` and an approval in the `release` environment. With one maintainer both are the same person; see [Residual risks](#residual-risks). |
| **Least privilege** | Public probes (`Health`, `Live`, `Ready`) send no credentials (`request.public`, `meta.go`); workflows run with `contents: read` and raise a permission only in the job that needs it (`security-events`, `id-token`, `contents: write` in `release`). |
| **Least common mechanism** | No shared mutable state between clients: each `Client` owns its transport, `WithHTTPClient` copies the caller's client rather than modifying it, and a `Client` is immutable after `NewClient` ([ADR 0004](adr/0004-concurrency-model.md)). |
| **Psychological acceptability** | The secure behaviour needs no configuration; `String` and `GoString` redact credentials so that printing a client is safe. |

## Weaknesses countered

The CWE Top 25 and OWASP Top 10 items that apply to an HTTP client
library, and what counters each.

| Weakness | Counter | Evidence |
| :--- | :--- | :--- |
| CWE-295 Improper certificate validation (OWASP A02, A07) | Verification cannot be turned off; a `*http.Transport` with `InsecureSkipVerify`, injected directly or reached through `http.DefaultTransport`, is refused | `check`, `roundTripper` (`client.go`); `TestNewClientRefusesControlCharactersAndInsecureTLS`, `TestNewClientRefusesAnInsecureDefaultTransport` |
| CWE-319 Cleartext transmission of sensitive information (A02) | HTTPS with TLS 1.2+ on the default transport; plain HTTP only when the caller passes an `http://` URL | `newTransport`; `TestClientDefaultsAndOptions`. Plain HTTP is a residual risk, below |
| CWE-532 Sensitive information in a log file; CWE-209 Error message with sensitive information (A09) | The module logs nothing; error messages carry the operation, status, problem type and code only; `Problem.Detail` and bodies stay out | `Error` methods in `errors.go`; `TestAPIErrorMessageLeavesOutTheDetail`, `TestDecodeErrorMessage`, `TestStringNeverPrintsCredentials`, `FuzzDecodeProblem`, `FuzzDecodeResponse` |
| CWE-117 Improper output neutralisation for logs | Server-supplied identifiers quoted in a message are cut to 64 printable ASCII characters | `safeLabel`; `TestSafeLabel` |
| CWE-601 Open redirect, leading to credential forwarding | Every redirect is refused, so `x-api-key` and `Authorization` never reach the `Location` origin | `refuseRedirect`; `TestRedirectsAreRefused`, `TestWithHTTPClientCopiesAndRefusesRedirects` |
| CWE-522 Insufficiently protected credentials; CWE-598 credentials in a URL | Credentials travel in headers only; a base URL with user info is refused | `setCredentials`, `checkBaseURL`; `TestCredentialsAndHeaders`, `TestNewClientRefusesBadBaseURLs` |
| CWE-113 / CWE-93 CRLF injection into headers | Control characters in a credential or `traceparent` are refused at construction; `net/http` rejects invalid header values from `Raw` | `check`; `TestNewClientRefusesControlCharactersAndInsecureTLS` |
| CWE-20 Improper input validation | The base URL and `Raw` requests are checked before use; typed methods use constant paths, so no caller data reaches a URL | `checkBaseURL`, `checkRaw`; `TestNewClientRefusesBadBaseURLs`, `TestRawRefusesBadRequests` |
| CWE-400 / CWE-770 Uncontrolled resource consumption | Success bodies capped at 64 MiB, error bodies at 1 MiB, unread bodies drained up to 64 KiB; per-attempt timeout and optional whole-call deadline; a `Retry-After` beyond `MaxDelay` ends the retries | `capReader`, `maxProblemBytes`, `drain`, `next`; `TestResponsesOverTheCapAreRefused`, `TestDrainLeavesLongBodiesToClose`, `TestPerAttemptTimeoutIsATimedOutTransportError`, `TestDeadlineBoundsTheWholeCall`, `TestRetryAfterIsHonouredOrEndsTheRetries` |
| CWE-502 Deserialisation of untrusted data (A08) | JSON only, into fixed Go types; exactly one value, nothing after it; problems without a string `detail` become `about:blank` | `decode`, `parseProblem`; `TestDecode`, `TestNonProblemBodiesBecomeAboutBlank`, both fuzz targets |
| CWE-837 Improper enforcement of a single, unique action | Stateful routes are retried only after 429 or 503, never after a lost connection or 504 | `statefulRoute`, `retryableStatus`; `TestStatefulRoutesAreNotRetriedAfter504`, `TestConnectionFailuresRetryOnlyIdempotentCalls`, `TestIsIdempotentFollowsTheTable` |
| CWE-362 Race condition | A `Client` holds no per-call state; tests run with `-race` | ADR 0004; `scripts/coverage.sh`; `TestLiveConcurrentCallsShareOneClient` |
| CWE-327 / CWE-330 Broken cryptography or weak randomness (A02) | No cryptography in the module. The only randomness is retry jitter and generated trace IDs, which are not secrets | ADR 0001; `newTraceparent` comment (`transport.go`) |
| CWE-1104 Unmaintained third-party components; supply chain (A06, A08) | No third-party module; govulncheck on every change; Dependabot for Go, actions, pip and the devcontainer; actions pinned by SHA; tools hash-pinned; signed tags; provenance attestations; reproducible archive and SBOM | `go.mod`; `ci.yml` (`vulncheck`); `.github/dependabot.yml`; [REPRODUCING.md](REPRODUCING.md) |

## Residual risks

Accepted, with the reason:

- **Plain HTTP to any host.** `NewClient` accepts `http://` for every
  host, as its documentation has said since the first release, and the
  API key or token then crosses the network in clear. Refusing it for
  non-loopback hosts would break callers that terminate TLS in a local
  sidecar or service mesh and reach the server by a cluster name, so it
  is left to the caller: use `https://` for anything that leaves the
  machine or a network you control. A future opt-in guard (refuse
  non-loopback `http://` unless an option allows it) would be a
  breaking change and needs a suite-wide decision first.
- **Proxies from the environment.** `HTTPS_PROXY` and `HTTP_PROXY` are
  honoured, as in every `net/http` program. Over `https://` the proxy
  sees only the host; over `http://` it sees everything, credentials
  included.
- **Transports the client cannot inspect.** A `RoundTripper` that wraps
  another (a tracing or retry middleware) is opaque, so a wrapped
  transport that skips verification is not detected. The check covers
  `*http.Transport`, direct or default.
- **Secrets in caller-held values.** Results such as `ExportedKey` and
  params such as `DecryptParams` are plain structs: printing one with
  `%v` prints its secret fields. The client never prints them; callers
  must not either. Go strings cannot be reliably zeroed, so secrets
  stay in memory until collected.
- **Trust in the server.** A compromised crypto-server can return
  wrong results; the client checks shape, not cryptographic truth.
  Verify signatures or MACs locally when that matters.
- **One maintainer.** Separation of privilege in the release path and
  two-person review are not available until a second maintainer joins
  ([GOVERNANCE.md](project/governance.md)).

## Evidence

| Activity | Where |
| :--- | :--- |
| 100% statement and 100% branch coverage, race-enabled | `scripts/coverage.sh`, `scripts/branch-coverage.sh` |
| Fuzzing of both decoders on every change and weekly | `fuzz_test.go`, `.github/workflows/fuzz.yml` |
| The conformance kit against a live server built from a pinned commit | `conformance_test.go`, `scripts/regression.sh` |
| Static analysis: vet, staticcheck, CodeQL (Go and workflows), complexity ceilings | `Makefile`, `.github/workflows/codeql.yml` |
| Known vulnerabilities in the standard library in use | `make vulncheck` (govulncheck) |
| Supply-chain posture | OpenSSF Scorecard (`.github/workflows/scorecard.yml`) |
| Release integrity | Signed tags, `KEYS.asc`, provenance attestations, [REPRODUCING.md](REPRODUCING.md) |
| Manual review against this document | [security-review-2026-11.md](security-review-2026-11.md) |
