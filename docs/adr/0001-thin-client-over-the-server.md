<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0001: A thin client over the server

- Status: accepted
- Date: 2026-11-15

## Context

Crypto Service is one engine, crypto-server, with clients in several
languages. The server holds the keys, enforces the security floors
(OWASP KDF parameters, Argon2id costs, key quotas, scopes) and is the
code that is reviewed, fuzzed and conformance-tested as cryptography.
A client that also implemented algorithms would be a second
implementation per language, each with its own bugs, side channels and
upgrade cadence, and its results could differ from the server's.

## Decision

The Go client implements no cryptography. Every operation is one HTTP
call to crypto-server; the client encodes the request, sends it with
the caller's credentials, and decodes the answer into Go types.

- Private keys never cross the API: key-generation methods return a
  `keyId` and the public key, and signing, decapsulation and opening
  take the `keyId`. Only `ExportKey`, behind its own scope, returns
  private material.
- The client's only data transformations are the encodings the API
  documents: `Hex` and `Base64` convert between strings and bytes. They
  are encodings, not cryptography.
- No algorithm, key derivation, random-number generation for keys, or
  third-party crypto module is added to this repository. W3C trace IDs
  use a non-cryptographic generator because they are not secrets.

## Consequences

- One behaviour across languages: the conformance kit checks that this
  client sends what the server expects and reads what it answers.
- The client needs a reachable server; there is no offline mode (the
  README's "When not to use" section says so).
- Every operation costs a round trip. The client keeps that cost to the
  network and the server's work: pooled keep-alive connections, HTTP/2
  over TLS, streaming JSON decoding.
- Security fixes to algorithms ship once, in the server.
