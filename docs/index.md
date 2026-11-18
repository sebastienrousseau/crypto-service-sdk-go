<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# crypto-service-sdk-go

A typed Go client for the Crypto Service REST API (crypto-server):
hashing, AEAD, server-held signing keys, post-quantum KEMs and
signatures, HPKE and KMS from Go, with the cryptography kept on the
server.

- **What it is**: one method per v2 operation, each taking a
  `context.Context`, decoding into Go types, with typed RFC 9457 errors.
- **What it is not**: a cryptography library. It implements no
  algorithm (see [ADR 0001](adr/0001-thin-client-over-the-server.md)).
- **How it is checked**: every vector of the server's conformance kit
  runs through the public methods in CI, against the pinned server.

Start with [Using the client](guide.md), then the
[examples](examples.md). The
[API reference](https://pkg.go.dev/github.com/sebastienrousseau/crypto-service-sdk-go)
lists every type and method.
