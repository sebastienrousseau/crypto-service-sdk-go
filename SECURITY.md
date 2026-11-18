<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Security Policy

## Supported versions

This module follows the
[Crypto Service](https://github.com/sebastienrousseau/crypto-service)
suite cadence. While pre-`1.0`, the latest released `0.0.x` receives
security fixes; older `0.0.x` versions do not.

| Version | Status | Receives security fixes? |
| :--- | :--- | :--- |
| 0.0.40 (the first release) | Current | Yes |

## Reporting a vulnerability

**Do not open a public issue for security reports.**

Report privately through GitHub Private Vulnerability Reporting:
<https://github.com/sebastienrousseau/crypto-service-sdk-go/security/advisories/new>.
The maintainer is notified automatically.

Please include:

- a description of the vulnerability and its potential impact;
- a minimal reproducer, preferably a failing test, with synthetic keys
  and data only;
- the module version and the crypto-server commit or version;
- any proposed remediation, if known.

A vulnerability in the server or its cryptography belongs in
[crypto-service's advisories](https://github.com/sebastienrousseau/crypto-service/security/advisories/new).

These are goals, not guarantees: the project is maintained in a personal
capacity, with no SLA. **Acknowledgement**: the goal is within 5
working days. **Triage**: a severity assessment and remediation plan
after acknowledgement.

## Security posture

### Scope

The module is an HTTP client. It sends JSON to the server the caller
configures and decodes the reply. It implements no cryptography, reads
and writes no files and runs no code from the server.

### Threat model

| Surface | How it is handled |
| :--- | :--- |
| **Secrets in logs and errors** | The client logs nothing. Error messages carry the operation, status, problem type and code only, never a credential, key, plaintext, detail or body; two fuzz targets hold this. `Client.String` and `GoString` redact the API key and token. |
| **Credential leakage** | The API key travels only in `x-api-key`, the token only in `Authorization`, never in a URL; a base URL with user info is refused. Every redirect is refused, so neither reaches another origin. |
| **Transport** | TLS 1.2 or newer with certificate verification; there is no option to turn verification off, and an injected transport that does, or an injected client whose default transport does, is refused. `NewClient` accepts plain `http://` for any host, so credentials then travel in clear: use it for loopback development or behind a local TLS sidecar, and HTTPS for anything else ([assurance case, residual risks](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/docs/ASSURANCE.md#residual-risks)). |
| **Duplicate state changes** | Key generation, deletion, KMS key creation and OPAQUE registration are never retried after a lost connection or a 504, when the server may already have acted. |
| **Resource exhaustion** | Success bodies are capped at 64 MiB and error bodies at 1 MiB; every attempt has a timeout (30 s by default), and `WithDeadline` bounds a whole call. |
| **Dependency CVEs** | No dependencies outside the Go standard library; govulncheck runs in CI, Dependabot watches `go.mod` and the CI actions. |

The full argument, with trust boundaries, the secure design
principles applied and the CWE items countered, each traced to code and
tests, is the
[assurance case](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/docs/ASSURANCE.md).
The latest review of the code against it is the
[2026-11 security review](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/docs/security-review-2026-11.md).

### Cryptography status

The module implements no cryptography: every operation runs on
crypto-server. TLS is Go's `crypto/tls` through `net/http`. Trace IDs
for `WithAutoTraceparent` come from a non-cryptographic generator; they
correlate logs and are not secrets.

## Verifying a release

Every release is a signed annotated tag `v<version>`, and its GitHub
release carries the source archive, a CycloneDX SBOM, `SHA256SUMS` and
a Sigstore build-provenance attestation for each asset. To check one
(v0.0.40 here):

1. **The tag.** Tags are signed with the SSH keys listed in
   [`KEYS.asc`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/KEYS.asc),
   an OpenSSH allowed-signers file whose git history is the trust log.
   In a clone, with git 2.34 or newer:
   `git -c gpg.ssh.allowedSignersFile=KEYS.asc verify-tag v0.0.40`
   must print `Good "git" signature` for one of those keys.
2. **The assets.** Download them with
   `gh release download v0.0.40 -R sebastienrousseau/crypto-service-sdk-go`,
   then `sha256sum -c SHA256SUMS` must print `OK` for each file, and
   `gh attestation verify <file> -R sebastienrousseau/crypto-service-sdk-go --signer-workflow sebastienrousseau/crypto-service-sdk-go/.github/workflows/release.yml`
   must succeed for each, which proves the file was built by this
   repository's release workflow from the tag.
3. **The module.** `go get` checks the module zip against the public
   checksum database sum.golang.org on every download, and
   `go mod verify` re-checks the module cache; no extra step is needed.
4. **Rebuild it.** The archive and SBOM are reproducible from the
   tagged commit, and the module zip from the repository:
   [docs/REPRODUCING.md](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/docs/REPRODUCING.md)
   has the commands and the expected hashes.
