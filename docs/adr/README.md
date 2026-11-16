<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Architecture decision records

Each record states a decision, why it was made and what follows from
it. A significant change (a breaking API change, a new dependency, a
change to the error, retry or concurrency model) adds or supersedes a
record in the same pull request.

| Record | Decision |
| :--- | :--- |
| [0001](0001-thin-client-over-the-server.md) | A thin client: no cryptography in the SDK, every operation on the server |
| [0002](0002-http-stack.md) | `net/http` with a tuned transport, standard library only |
| [0003](0003-error-model.md) | `*APIError`, `*TransportError`, `*DecodeError` and the `Error` interface |
| [0004](0004-concurrency-model.md) | Blocking calls with a context; goroutines for concurrency |
