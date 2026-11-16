<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Policies

The rules this module keeps about versions, toolchains, dependencies
and supported servers, and where each is enforced.

## Versioning

- Versions are `0.0.x` and move by `0.0.1` only, in lockstep with the
  Crypto Service suite; `0.1.0` follows `0.0.999`. The maintainer opens
  each release on a `feat/v<version>` branch; nothing bumps a version as
  a side effect.
- `version.go` is the source of truth. `scripts/verify-release-versions`
  fails CI when `CITATION.cff`, the `SECURITY.md` and `SUPPORT.md`
  support tables, an install snippet, a quoted User-Agent or the
  CHANGELOG heading disagree with it.
- The breaking axis is the exported Go API (functions, methods, types,
  struct fields, options, constants) and the requests the client sends.
  While pre-1.0, a release may break either, and the CHANGELOG calls the
  break out with a `Breaking:` line; `scripts/api-check.sh` (gorelease)
  fails CI on an incompatible change without one.

## Minimum Go toolchain

- **The floor is the older of the two Go releases the Go team
  supports**: Go 1.26 today, in the `go` directive of `go.mod`.
- **It moves only when that release leaves support** (Go 1.28 retires
  1.26), as its own change with a CHANGELOG line.
- **CI enforces it**: the test job runs on 1.26 and 1.27.

## Dependencies

- Runtime: the Go standard library only
  ([ADR 0002](adr/0002-http-stack.md)). `go.mod` has no requirements
  and the module has no `go.sum`.
- Tools run through `go run <module>@<version>` at pinned versions;
  Python tools install from hash-pinned lockfiles in `ci/`; CI actions
  are pinned by commit SHA.

## Supported servers

- CI tests against the crypto-service commit pinned in `ci/server.ref`.
- The vendored conformance kit is byte-identical to that commit's and
  pinned by `conformance/SHA256SUMS`.
- The client covers the v2 API. The v1 OpenPGP routes are deprecated by
  the server and sunset on 2027-01-01; they are not wrapped.

## Security fixes

Only the latest `0.0.x` receives fixes
([SECURITY.md](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/SECURITY.md)).
