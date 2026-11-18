<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Contributing

Thanks for looking. This module is a thin client on purpose: crypto-server
does the cryptography, and the client's job is to call it correctly and
return typed results.

## Before you open a pull request

Everything CI checks without a server, you can run locally with Go 1.26
or newer:

```sh
make check
```

That runs the version check, `gofmt -l`, `go vet`, staticcheck, the
gocyclo and gocognit ceilings, govulncheck, `go build` and
`go test -race` at 100% statement coverage, and gobco at 100% branch
coverage. To add the conformance kit, the live tests, the examples and
the README blocks, start a server from a built crypto-service checkout
and point the tests at it:

```sh
cd /path/to/crypto-service/packages/crypto-server
PORT=3000 HOST=127.0.0.1 CRYPTO_API_KEY=local-test-key-0123456789 \
  JWT_SECRET=local-test-secret-0123456789abcdef RATE_LIMIT_MAX=100000 \
  node dist/index.js &
cd -
CRYPTO_SERVICE_URL=http://127.0.0.1:3000 \
  CRYPTO_SERVICE_API_KEY=local-test-key-0123456789 \
  CRYPTO_SERVICE_JWT_SECRET=local-test-secret-0123456789abcdef \
  make conformance
```

`make regression` does the same with a server it starts itself from
`CRYPTO_SERVICE_SERVER_DIR`, as CI does. Use synthetic data only; never
add a real key or credential to a test.

## The README demo

`.github/demo.gif` is rendered from `.github/demo.tape` by
[vhs](https://github.com/charmbracelet/vhs) with `make demo`, against a
server you started on loopback (`CRYPTO_SERVICE_URL`). Re-render it in
the same change whenever the examples it shows change.

## What the tests are for

- `client_test.go`, `transport_test.go`, `errors_test.go`,
  `codec_test.go`, `meta_test.go`, `raw_test.go`, `version_test.go`:
  options, headers, timeouts, redaction, decoding and error mapping
  against stub servers; no crypto-server needed.
- `methods_test.go`: every typed method calls its route and returns an
  `*APIError` for a problem.
- `retry_test.go`: every branch of the retry table, backoff and
  `Retry-After`.
- `internal/kit/`: the conformance kit reader (substitution, JSON
  pointers, template matching, the `$shape` schema subset), unit-tested
  without a server.
- `conformance_test.go`: every vector of `conformance/` through the
  public method for its route (Raw only where no method can send the
  vector's exact request); `live_typed_test.go`: round trips through
  the typed API. Both skip without a server, and fail without one under
  `make conformance`.
- `kit_test.go`: the vendored kit matches `conformance/SHA256SUMS`;
  `ci/server.ref` is a commit SHA.
- `size_test.go`, `halstead_test.go`: the per-function and per-file
  ceilings.
- `fuzz_test.go`: the problem-body parser and the success-response
  decoder against arbitrary input: no panic, and no message that quotes
  a body or a detail. `make fuzz` fuzzes each target; commit any
  crashing input it writes to `testdata/fuzz` with the fix.
- `bench_test.go`, `bench_live_test.go`: the shared benchmark scenarios
  (`make bench`).

## The conformance kit

`conformance/` is vendored from crypto-service byte for byte. **Do not
edit it here.** A change to a vector belongs in crypto-service; this
repository takes the new kit by copying its `conformance/` at the
commit `ci/server.ref` pins, in the same change that moves the pin.

## Commits

Conventional commits, each signed off (`git commit -s`) under the
[Developer Certificate of Origin](DCO.txt); the DCO workflow checks
every commit in a pull request.

## Versioning

Versions increment by 0.0.1 in lockstep with the Crypto Service suite;
`0.1.0` follows `0.0.999`. The maintainer decides when a release opens.

## Licence

Apache-2.0 OR MIT, at your option. By contributing you agree your work is
released under the same dual grant.
