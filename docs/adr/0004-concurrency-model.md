<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# 0004: Concurrency model: blocking calls with a context

- Status: accepted
- Date: 2026-11-15

## Context

The suite's SDKs offer an asynchronous API as the primary surface and a
blocking one as secondary, in each language's idiom. In Go the idiom is
different: a function blocks and takes a `context.Context`, and
concurrency comes from the caller starting goroutines. A futures or
callback API would be foreign to Go code and add nothing the runtime
does not already do.

## Decision

- Every method blocks until the server answers, the context ends, or
  the client's per-attempt timeout or deadline passes. That
  context-aware blocking call is the asynchronous API in Go terms;
  there is no second variant.
- Cancellation and deadlines come from the context the caller passes
  first, combined with `WithTimeout` (per attempt) and `WithDeadline`
  (whole call). A wait between retries ends as soon as the context
  does.
- **Guarantee:** a `*Client` is safe for concurrent use by any number
  of goroutines once `NewClient` has returned. It holds no per-call
  state; its fields are set at construction and never written again,
  and its transport's connection pool is safe for concurrent use. The
  only exception is `Close`, which must not race with calls in flight.
- Fan-out uses goroutines and `sync.WaitGroup` from the standard
  library (`examples/concurrent`, `Example_concurrent`); no errgroup
  dependency.

## Consequences

- Tests run with `-race`, and a live test fans eight calls out over one
  client.
- Callers bound concurrency themselves (a semaphore channel, a worker
  pool); the transport keeps up to 32 idle connections per host.
