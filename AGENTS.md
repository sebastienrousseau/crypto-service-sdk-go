<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# AGENTS.md

`crypto-service-sdk-go` is the Go client for the Crypto Service REST API
(crypto-server). The server does the cryptography; this module calls it
and returns typed results. It exists so Go services use the same engine
as every other Crypto Service client, held to the same conformance kit.
Read [CONTRIBUTING.md](CONTRIBUTING.md) and
[DEVELOPMENT.md](DEVELOPMENT.md) as well; this file lists what must
never drift.

## Rules

- **Versions** are `0.0.x` and move by `0.0.1` only, in lockstep with
  the Crypto Service suite; `0.1.0` comes after `0.0.999`, never as a
  jump. `version.go` is the single source of truth and
  `scripts/verify-release-versions` checks every restatement. Never bump
  one as a side effect; the maintainer opens each release on a
  `feat/v<next-version>` branch.
- **Thin client, no cryptography.** Every operation is an HTTP call to
  crypto-server (`docs/adr/0001-thin-client-over-the-server.md`). Never
  add a local algorithm, key handling or a crypto dependency; encoding
  helpers (`Hex`, `Base64`) are the limit.
- **`conformance/` is vendored, never edited.** Replace it whole, byte
  for byte, from crypto-service at the commit in `ci/server.ref`;
  `kit_test.go` checks it against its `SHA256SUMS`.
- **Standard library only.** No third-party module in `go.mod`;
  analysers run through `go run <module>@<pinned version>`.
- **Every method takes a `context.Context` first**; the client stays
  safe for concurrent use (no per-call state on `Client`).
- **Never log or expose secrets.** No logging at all. Error messages,
  `String`, `GoString` and debug output never carry a credential, a key,
  a plaintext or a request or response body. `fuzz_test.go` holds this.
- **Retry table.** Stateful routes (keys/generate, keys/delete, every
  `*/keygen`, kms/create-key, opaque/register/*) retry only after 429 or
  503, as the TypeScript SDK's `request.ts` does.
- **Redirects refused, TLS verification always on.** No option may
  weaken either.
- **Complexity**: cyclomatic 10 and cognitive 15 per function (gocyclo,
  gocognit), Halstead difficulty 30 (`halstead_test.go`), 60 lines per
  function and 500 per file (`size_test.go`).
- **Coverage**: 100.0% of statements (`scripts/coverage.sh`) and 100%
  of branches (`scripts/branch-coverage.sh`, gobco); never lower
  either.
- **Data**: synthetic only. Test keys are fixed test vectors; no real
  key, credential or personal data anywhere.
- **Headers**: every source file opens with the SPDX identifier
  `Apache-2.0 OR MIT`; Markdown carries it as an HTML comment on line 1.
- **Commits**: conventional, signed off (`git commit -s`) and signed.
  Never rewrite published history.

## The gate

```sh
make check                      # without a server
make conformance                # against a running server (fails without one)
CRYPTO_SERVICE_SERVER_DIR=/path/to/crypto-service make regression
```

A change is done when these pass and add no warnings.

## Leave alone

`conformance/` (vendored), `LICENSE`, `LICENSE-APACHE`, `LICENSE-MIT`,
`LICENSES/` and `DCO.txt` (legal texts), and the server pin in
`ci/server.ref` outside a release.
