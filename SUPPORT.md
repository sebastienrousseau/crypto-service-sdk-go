<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Getting support

crypto-service-sdk-go is maintained by one person in a personal
capacity, with no SLA. This page says where to ask, and which versions
get fixes.

## Supported versions

The module releases in lockstep with the
[Crypto Service suite](https://github.com/sebastienrousseau/crypto-service).
While pre-`1.0`, only the latest `0.0.x` receives fixes.

| Version | Status | Receives fixes? |
| :--- | :--- | :--- |
| 0.0.40 (the first release) | Current | Yes |

The client is tested against the crypto-server commit pinned in
`ci/server.ref`.

## Read first

- [README.md](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/README.md):
  install, quick start, every option and the example index.
- The [API reference](https://pkg.go.dev/github.com/sebastienrousseau/crypto-service-sdk-go).
- [`examples/`](https://github.com/sebastienrousseau/crypto-service-sdk-go/tree/main/examples):
  runnable programs, each run against a live server in CI.

## Questions and bugs

Open an issue at
<https://github.com/sebastienrousseau/crypto-service-sdk-go/issues/new/choose>
with:

- the module version and `go version`;
- the crypto-server version or commit (`client.Root` returns it);
- a minimal program that reproduces the problem, with synthetic keys
  and data only;
- the full error text (it carries no secret by design).

A problem in what the server accepts, refuses or computes belongs in
[crypto-service](https://github.com/sebastienrousseau/crypto-service/issues);
this client only carries requests and responses.

## Security reports

Never in a public issue. Follow
[SECURITY.md](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/SECURITY.md).
