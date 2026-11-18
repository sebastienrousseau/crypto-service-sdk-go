<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Examples

Each example is a program under `examples/` in the repository. Run one
against a server with
`CRYPTO_SERVICE_URL=http://127.0.0.1:3000 CRYPTO_SERVICE_API_KEY=... go run ./examples/<name>`,
or all of them with `make examples`. CI runs every one against the
pinned crypto-server; each exits non-zero when the server does not
answer as it expects.

| Example | Shows |
| :--- | :--- |
| `hash` | `Hash` with the SHA-256 test vector |
| `aead` | `Encrypt` and `Decrypt` round trip |
| `sign` | A server-held Ed25519 key: generate, sign by keyId, verify, delete |
| `mlkem` | ML-KEM-768 encapsulation to a server-held key and decapsulation by keyId |
| `errors` | `*APIError`, its field errors, `*TransportError` and the `Error` interface |
| `concurrent` | Goroutines over one client with a context deadline |
