<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

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
| **Transport** | TLS 1.2 or newer with certificate verification; there is no option to turn verification off, and an injected transport that does is refused. Plain HTTP is accepted for loopback development; use HTTPS elsewhere. |
| **Duplicate state changes** | Key generation, deletion, KMS key creation and OPAQUE registration are never retried after a lost connection or a 504, when the server may already have acted. |
| **Resource exhaustion** | Success bodies are capped at 64 MiB and error bodies at 1 MiB; every attempt has a timeout (30 s by default), and `WithDeadline` bounds a whole call. |
| **Dependency CVEs** | No dependencies outside the Go standard library; govulncheck runs in CI, Dependabot watches `go.mod` and the CI actions. |

### Cryptography status

The module implements no cryptography: every operation runs on
crypto-server. TLS is Go's `crypto/tls` through `net/http`. Trace IDs
for `WithAutoTraceparent` come from a non-cryptographic generator; they
correlate logs and are not secrets.
