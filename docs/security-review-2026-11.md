<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Security review, 2026-11

| | |
| :--- | :--- |
| Date | 2026-11-18 |
| Reviewer | Sebastien Rousseau, with AI-assisted analysis (Claude) |
| Baseline | `main` at `58d9fea` (after v0.0.40), on branch `feat/v0.0.41` |
| Reference | The [assurance case](ASSURANCE.md), written in the same change |

## Scope

- **Library code**: every non-test Go file of the module root
  (`client.go`, `transport.go`, `retry.go`, `codec.go`, `errors.go`,
  `problem.go`, `raw.go`, `encoding.go`, `meta.go` and the per-operation
  files) and `internal/kit`.
- **Transport and TLS**: the default transport, the injection options
  (`WithHTTPClient`, `WithTransport`), redirects and proxies.
- **Credential handling**: where the API key and token are stored,
  sent, printed and redacted.
- **Input parsing**: the base URL, `Raw` requests, success bodies,
  problem bodies, `Retry-After`.
- **Error paths**: the messages of `*APIError`, `*TransportError`,
  `*DecodeError` and the construction errors.
- **CI workflows and their permissions**: all eight files in
  `.github/workflows/`.
- **Release pipeline**: `release.yml`, `scripts/release-artifacts.sh`,
  `scripts/release-notes.sh`, `KEYS.asc`, the `release` environment.
- **Dependencies**: `go.mod`, the pinned analysers, the hash-pinned
  Python tools, the devcontainer.

Out of scope: crypto-server and its cryptography (reviewed in
[crypto-service](https://github.com/sebastienrousseau/crypto-service)),
and the vendored `conformance/` kit, which is test data.

## Method

1. Read every file in scope against each claim and boundary of the
   assurance case, and against the CWE and OWASP items it lists.
2. Probed the suspicious paths with small programs against
   `httptest` servers: base URLs with an empty query or fragment,
   `net/url` parsing of unusual base URLs, redirect error text, and an
   injected client without a transport.
3. Checked every workflow for `pull_request_target`, expression
   injection into `run:` scripts, token permissions, unpinned actions
   and credential persistence; read the repository's environment and
   branch-protection settings through the GitHub API (read only).
4. Rebuilt the v0.0.40 release assets and the module zip and compared
   them with what was published ([REPRODUCING.md](REPRODUCING.md)).
5. Ran the full local gate: `make check` (now including branch
   coverage), `make docs-lint`, `make docs`, and `make fuzz` is run by CI
   on the pull request.

Each finding that is a clear bug was fixed with a test that failed
before the fix, in the same pull request.

## Findings

| ID | Severity | Finding | Disposition |
| :--- | :--- | :--- | :--- |
| SR-1 | Low | A base URL ending in an empty query or fragment (`http://h/p?`, `http://h#`) passed `checkBaseURL`, because `net/url` reports an empty `RawQuery` and `Fragment` for it. Every path the client appended then landed in the query or fragment, so each call reached the wrong route (CWE-20). Credentials still went only to the configured host. | **Fixed**: `checkBaseURL` refuses any `?` or `#`. `TestNewClientRefusesBadBaseURLs` gained the three cases and failed before the fix. |
| SR-2 | Low | `check` refused an injected `*http.Transport` with `InsecureSkipVerify`, but not an injected `*http.Client` whose `Transport` is nil, which sends through `http.DefaultTransport`. A process that had turned verification off on the default transport got past the "verification must stay on" guarantee (CWE-295). | **Fixed**: `check` inspects `roundTripper()`, the transport requests actually use. `TestNewClientRefusesAnInsecureDefaultTransport` failed before the fix. |
| SR-3 | Medium | `NewClient` accepts `http://` for any host, so credentials and data cross the network in clear when a caller uses it beyond loopback (CWE-319). SECURITY.md said plain HTTP was "accepted for loopback development", which read as a restriction the code does not make. | **Accepted, documented**: restricting it would break documented behaviour (callers behind a TLS sidecar or mesh). SECURITY.md now states the actual behaviour; the assurance case records it as a residual risk and the shape of a future opt-in guard. |
| SR-4 | Low | The release workflow verifies the tag against `KEYS.asc` from the tagged tree itself, and does not check that the tagged commit is on `main`. Someone with push access could tag a commit on another branch that adds their own key. | **Accepted, mitigated**: pushing a tag already needs write access, and the publishing jobs need an approval in the `release` environment (required reviewer: the maintainer). Recorded as a hardening task: verify against `KEYS.asc` from `main` and require the tag to be reachable from `main`. |
| SR-5 | Low | The docs-lint job runs `npx --yes markdownlint-cli2@0.23.3`: the package version is pinned, its dependency tree is resolved at run time without a lockfile or hashes. | **Accepted**: the job has a read-only token, no secrets and produces no artefact; a compromised dependency could only fail or pass a lint. Recorded as a hardening task. |
| SR-6 | Informational | The release archive and SBOM were not byte-reproducible: git's built-in gzip output varies with the git version, and the SBOM had a random serial number, a time stamp and hashes of a tool binary built with machine-specific paths. | **Fixed**: `scripts/release-artifacts.sh` compresses with `gzip -n`, derives the serial from the content, drops the time stamp and builds the SBOM tool reproducibly; verified by two independent rebuilds ([REPRODUCING.md](REPRODUCING.md)). |
| SR-7 | Informational | A refused redirect's `*TransportError` message quotes the server's `Location` value (through `net/url`'s `url.Error`), for example `Post "https://other.example/...": redirects are refused`: server-chosen text in an error message. It is `%q`-quoted, so it cannot break a log line, and it carries no caller secret. | **Accepted**: no credential or body is exposed. Recorded as a small task to report the request URL instead. |
| SR-8 | Informational | The check in SR-2 cannot see through a `RoundTripper` that wraps another (tracing or metrics middleware), so a wrapped transport that skips verification is not detected. | **Accepted**: the wrapped transport is opaque by design; documented in the assurance case. |
| SR-9 | Informational | Params and results that carry secrets (`ExportedKey`, `DecryptParams`, `KmsDataKey`, ...) are plain structs and print their secret fields with `%v`. | **Accepted**: the client never prints them; documented in the assurance case as the caller's responsibility. |
| SR-10 | Informational | `HTTPS_PROXY` and `HTTP_PROXY` are honoured (`http.ProxyFromEnvironment`). Over `http://` a proxy sees credentials. | **Accepted**: standard Go behaviour that operators rely on; documented. |
| SR-11 | Informational | The devcontainer image is pinned by digest, but its Node.js feature (`ghcr.io/devcontainers/features/node:2`) by major version only. | **Accepted**: development environment only, not part of CI or the release; Dependabot watches the devcontainer. |

No finding exposed a credential, a key, a plaintext or a body, and no
finding was of high or critical severity.

## Checked and found sound

- Credentials are set in one place (`setCredentials`), only in
  headers, never on public routes, and never printed (`String`,
  `GoString`, every `Error` method).
- Redirects are refused even when the caller's client allows them;
  `CheckRedirect` is assigned after the options run.
- TLS 1.2 minimum on the default transport; Go's verification defaults
  are unchanged.
- Response bodies are capped, drained and closed on every path; the
  decoder rejects trailing data; both decoders survive fuzzing.
- The retry table never resends a stateful call after a lost
  connection or a 504; `Retry-After` is bounded by `MaxDelay`.
- No `pull_request_target`, no expression injection into a `run:`
  script, `persist-credentials: false` on every checkout, every action
  pinned by commit SHA, Python tools installed with `--require-hashes`.
- `ci/server.ref` is validated as a 40-character SHA before CI fetches
  it, and the fetched commit is checked to be that SHA.
- Main is protected: required status checks, signed commits, no force
  push, administrators included.
- `go.mod` has no `require`; govulncheck found no reachable
  vulnerability in the standard library in use.

## Next review

Before v0.1.0, or after any change to the transport, the credential
handling, the retry table or the release pipeline, whichever comes
first.
