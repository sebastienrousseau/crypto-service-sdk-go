<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Development

How to build, test and release crypto-service-sdk-go, and which local
command matches each CI gate.

## Toolchain

- Go 1.26 or newer (`go.mod`); `mise.toml` pins the local toolchain
  (`mise install`).
- For the live gates: Node.js 24 and pnpm 11.9.0 to build the pinned
  crypto-server, as CI does.
- Analysers (staticcheck, gocyclo, gocognit, govulncheck, gorelease,
  cyclonedx-gomod) run through `go run <module>@<version>` at the
  versions the Makefile and scripts pin; nothing is installed globally.

## Gates

| CI job | Local command | What it holds |
| :--- | :--- | :--- |
| `gofmt, vet, staticcheck, complexity` | `make versions fmt-check vet staticcheck complexity` | Version restatements, formatting, vet, staticcheck, cyclomatic 10, cognitive 15 |
| `govulncheck` | `make vulncheck` | No reachable known vulnerability in the standard library in use |
| `build and unit tests (Go 1.26, 1.27)` | `make build test branch-coverage` | Race-enabled tests at 100.0% statement coverage and 100% branch coverage (gobco), 60 lines per function, 500 per file, Halstead difficulty 30 |
| `fuzz` (`fuzz.yml`) | `make fuzz FUZZTIME=60s` | Problem-body parser and success decoder: no panic, no body in a message |
| `API-breakage check (gorelease)` | `make api-check` | No incompatible exported change without a `Breaking:` CHANGELOG line |
| `conformance` | `CRYPTO_SERVICE_SERVER_DIR=... make regression` | Every kit vector, the live tests, every example, every README block and a benchmark smoke run against the pinned server |
| Docs lint | `make docs-lint` | markdownlint, codespell, REUSE |

Coverage is measured twice, on the module root and `internal/kit`.
`go test -cover` measures statements (`scripts/coverage.sh`, floor
100.0%). Branches are measured by
[gobco](https://github.com/rillig/gobco) (BSD-2-Clause, run through
`go run` at a pinned version) in its `-branch` mode, which instruments
every `if`, `else`, `switch` case and loop condition and fails unless
the unit tests take each one both ways (`scripts/branch-coverage.sh`,
floor 100%: 160 of 160 branches in the root package and 120 of 120 in
`internal/kit` when it was introduced). gobco does not instrument
`select` statements; the one in the library (`sleep` in `retry.go`) is
held by statement coverage and the retry tests. Without `-branch`,
gobco reports the stricter per-condition coverage of `&&` and `||`
operands; that report is informational.

## The live gates

`make conformance` needs a crypto-server and three variables:

- `CRYPTO_SERVICE_URL`: its base URL;
- `CRYPTO_SERVICE_API_KEY`: its `CRYPTO_API_KEY`;
- `CRYPTO_SERVICE_JWT_SECRET`: its `JWT_SECRET`. The kit runs each file
  as its own principal (`conformance-<group>`), as its README requires,
  because the server meters 60 requests a minute per principal; only a
  JWT can name the principal.

It fails, not skips, when any is missing. Start the server with
`RATE_LIMIT_MAX=100000` (the per-address default is 10 a minute). The
password routes keep their own limit of 5 a minute per address, so
running the kit more than twice a minute against one server trips it.

`make regression` starts the server itself from
`CRYPTO_SERVICE_SERVER_DIR`, a checkout of
[crypto-service](https://github.com/sebastienrousseau/crypto-service) at
the commit in `ci/server.ref`, built with
`pnpm install --frozen-lockfile && pnpm -r run build`. It refuses a
checkout at another commit unless `ALLOW_OTHER_SERVER=1`.

## Moving the server pin

1. Write the new commit SHA to `ci/server.ref`.
2. Replace `conformance/` with that commit's `conformance/`, byte for
   byte, including `SHA256SUMS`.
3. Run `make regression` against a checkout of that commit, and fix the
   client in the same change.

## Benchmarks

`make bench` runs `BenchmarkSDK`: `encode_request`, `decode_response`
and `decode_problem` without a network, and `round_trip_hash` against
`CRYPTO_SERVICE_URL` (it authenticates with JWTs for rotating
principals, so the per-principal quota does not throttle it). Output
goes to `bin/bench.txt` in benchstat format. Results are not published
until they are generated into `docs/BENCHMARKS.md` with their
environment.

## Releasing

1. On the `feat/v<version>` branch: set `Version` in `version.go`, the
   `CHANGELOG.md` heading to `## [<version>] - <date>`, `CITATION.cff`,
   the support tables in `SECURITY.md` and `SUPPORT.md`, the install
   snippet in `README.md`, and `.github/release-highlights.md`. Run
   `scripts/verify-release-versions --release`.
2. Merge the release pull request once CI is green.
3. Tag the merge commit with a signed annotated tag whose message is
   `crypto-service-sdk-go v<version>`:
   `git tag -s v<version> -m "crypto-service-sdk-go v<version>"`, then
   push the tag.
4. The release workflow verifies the tag, versions, tests and API,
   builds the source archive, the CycloneDX SBOM (`.cdx.json`) and
   `SHA256SUMS`, attests their build provenance, writes the GitHub
   release titled `crypto-service-sdk-go <version>` (sections
   `## Highlights ⭐️`, `## What's Changed`, `## Checksums`,
   `**Full Changelog**`), and asks proxy.golang.org for the version,
   which is how a Go module is published. The `release` environment
   gates the last two jobs; no registry secret is involved.

Run the workflow by hand (`workflow_dispatch`, `dry_run` on) to see the
release it would make without publishing anything.

The archive and SBOM are reproducible: `scripts/release-artifacts.sh`
gives byte-identical files for the same commit on any machine, and
[docs/REPRODUCING.md](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/docs/REPRODUCING.md) gives the procedure and the
expected hashes. Keep it so: no time stamps, no random identifiers,
and every tool at a pinned version. The tag-verification
step checks the tag against the maintainer's allowed signers in
`KEYS.asc`.
