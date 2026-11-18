<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Architecture

## Layers

| File | Role |
| :--- | :--- |
| `client.go` | `Client`, `NewClient`, options, the pooled transport, redaction |
| `transport.go` | One call: deadline, attempts, request building, credentials, `traceparent` |
| `retry.go` | `RetryPolicy`, the stateful-route table, backoff with full jitter, `Retry-After` |
| `codec.go` | JSON encoding, streaming decoding with a size cap, the `{data}` envelope |
| `errors.go`, `problem.go` | `Error`, `*APIError`, `*TransportError`, `*DecodeError`, `Problem` |
| `encoding.go` | `Hex` and `Base64` |
| `symmetric.go`, `hashing.go`, `keys.go`, `pq.go`, `boxes.go`, `protocols.go`, `kms.go`, `meta.go` | One method per operation, with its request and response types |
| `raw.go` | `Raw`, the escape hatch |
| `internal/kit` | Reads the conformance kit and matches answers against it (tests only) |

## One call

1. The method encodes its params struct to JSON (`encode`, no HTML
   escaping).
2. `execute` applies `WithDeadline` to the caller's context, picks the
   `traceparent` for the call, and classifies the route as idempotent
   or stateful.
3. `attempt` applies `WithTimeout`, builds the request (credentials
   unless the route is public, `User-Agent`, `Accept`), and sends it
   through the client's `*http.Client`, which refuses redirects.
4. A 2xx body is decoded as a stream into the result type, under the
   attempt's timeout and a 64 MiB cap. Any other status is read (up to
   1 MiB) into an `*APIError`.
5. `next` decides on a retry from the attempt's outcome, the retry
   table and `Retry-After`; `sleep` waits unless the context ends.

## Conformance

`conformance_test.go` reads every vector file in order, as its own
principal, and sends each vector through the public method for its
route. A vector whose exact body no method can send (an unknown member,
a missing required member, malformed JSON, a non-JSON media type) and
the routes the client does not wrap (v1, `/metrics`) go through `Raw`.
The typed result or `Problem` is encoded back to JSON and matched
against the vector's template, so the check covers the request the
method builds and the types it decodes into.
