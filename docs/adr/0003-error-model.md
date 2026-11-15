<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0003: Error model

- Status: accepted
- Date: 2026-11-15

## Context

crypto-server answers every error with an RFC 9457 problem
(`application/problem+json`): `type`
(`urn:crypto-service:problem:<slug>`), `title`, `status`, `detail`,
`instance`, and extension members such as `code` and `errors[]`. A
proxy in front of it may answer with HTML or nothing. Callers need to
tell a refused request from a lost connection, and a timeout from a
cancellation, without parsing strings. Error messages end up in logs,
so they must never carry secrets.

## Decision

Three concrete types, one interface:

- `*APIError`: an error status. `Problem` holds the problem with typed
  `Type`, `Title`, `Status`, `Detail`, `Instance`, `Code` and `Errors`
  (`[]ValidationError{Field, Message}`), and `Extensions` for every
  other member as raw JSON. A body that is not a JSON object with a
  string `detail` becomes an `about:blank` problem for the status, as
  the TypeScript SDK does. `ContentType` keeps the media type.
- `*TransportError`: no usable response. `TimedOut` is true for a
  per-attempt timeout, `WithDeadline` or a context deadline, false for a
  cancellation or a connection failure; `Err` is the cause, so
  `errors.Is(err, context.Canceled)` works.
- `*DecodeError`: a 2xx body that is not the documented JSON, or larger
  than the 64 MiB cap.
- `Error`, an interface with `Operation()`, implemented by all three:
  `errors.As(err, &base)` catches any of them.

`Error()` strings carry the operation, the status, the problem slug
and the code, passed through a filter that keeps them printable and
short. They never carry the detail, a body, a credential, a key or a
plaintext; the decoder's own complaint stays in `Unwrap()`. Two fuzz
targets hold this.

## Consequences

- `errors.As` is the whole matching story; no sentinel strings.
- A caller who wants the detail reads `Problem.Detail` deliberately.
- The `Ready` probe is the one exception to "error status is an error":
  `/ready` answers 503 with its normal document, which `Ready` returns
  with `Status` "not_ready".
