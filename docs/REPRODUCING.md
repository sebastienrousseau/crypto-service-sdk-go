<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Reproducing a release

A release of crypto-service-sdk-go is three things, and each can be
rebuilt from the signed tag and compared byte for byte:

| Artefact | Where it is published | What pins it |
| :--- | :--- | :--- |
| The Go module zip (what `go get` downloads) | proxy.golang.org | Its `h1:` hash, recorded in the public checksum database sum.golang.org and checked by every `go` command |
| `crypto-service-sdk-go-<version>.tar.gz`, the source archive | The GitHub release | `SHA256SUMS` on the release, and a Sigstore build-provenance attestation |
| `crypto-service-sdk-go-<version>.cdx.json`, the CycloneDX SBOM | The GitHub release | The same |

Check the tag's signature first ([SECURITY.md, "Verifying a
release"](project/security.md#verifying-a-release)); everything below is
rebuilt from the tagged commit.

## The module zip

The module zip is not built by this project: proxy.golang.org builds it
from the tagged tree with the Go team's deterministic zip format
(`golang.org/x/mod/zip`), and sum.golang.org records its hash in a
public, append-only log. Rebuild it from the git repository itself,
bypassing the proxy and the checksum database, and compare:

```sh
tag=v0.0.40  # the release to check
mkdir -p /tmp/repro-mod && cd /tmp/repro-mod
GOPROXY=direct GOSUMDB=off GOMODCACHE="$PWD/cache" GOFLAGS=-modcacherw \
  go mod download -json "github.com/sebastienrousseau/crypto-service-sdk-go@$tag"
curl -s "https://sum.golang.org/lookup/github.com/sebastienrousseau/crypto-service-sdk-go@$tag"
```

The `Sum` the first command prints must equal the `h1:` hash the
checksum database returns. For v0.0.40 both are
`h1:Lq10Lqf3dW5FpzUfXzOVVgWjD7LvxI/jt/rhfLlVC28=` (checked on
2026-11-18 with Go 1.27.4).

## The source archive and the SBOM

`scripts/release-artifacts.sh` builds both into `dist/`, and the release
workflow runs exactly that script. Every byte is a function of the
tagged commit:

- the tar stream is `git archive`'s, whose file times come from the
  commit;
- it is compressed with `gzip -n -9`, which writes no file name and no
  time stamp, instead of git's built-in compressor, whose output
  changes with the git and zlib versions in use;
- the SBOM carries no time stamp, and its serial number is a version 5
  (name-based) UUID of the module, the version and the tar stream's
  SHA-256, not a random one;
- the SBOM records the hashes of the cyclonedx-gomod binary that wrote
  it, so that binary is built reproducibly: pinned version
  (`v1.12.0`), the Go release `mise.toml` pins (the script sets
  `GOTOOLCHAIN`, and the `go` command downloads that release, verified
  against sum.golang.org, when it is not the one installed),
  `-trimpath`, `CGO_ENABLED=0`, `linux/amd64`.

To reproduce a release from v0.0.41 on, you need git, GNU gzip,
python3 (for the UUID) and any Go 1.21 or newer:

```sh
git clone https://github.com/sebastienrousseau/crypto-service-sdk-go
cd crypto-service-sdk-go
git checkout v0.0.41
scripts/release-artifacts.sh
gh release download v0.0.41 -p SHA256SUMS -O /tmp/SHA256SUMS
diff /tmp/SHA256SUMS dist/SHA256SUMS && echo reproduced
```

Use a fresh clone at the tag, not a work tree with other changes: the
SBOM's module version is read from git.

The procedure was checked across machines before v0.0.41 was tagged:
for commit `e3c4d71` of `feat/v0.0.41`, the release workflow's dry run
on GitHub's `ubuntu-latest` runner
([run 44263946227](https://github.com/sebastienrousseau/crypto-service-sdk-go/actions/runs/44263946227))
and a local build on Arch Linux (git 2.56.1, gzip 1.15, Go 1.27.4)
produced the same `SHA256SUMS`.

### v0.0.40

v0.0.40 was released before this procedure existed: its archive was
compressed by git's built-in gzip and its SBOM has a random serial
number and a time stamp. Its content still reproduces, and the
volatile parts can be set aside:

```sh
git checkout v0.0.40
gh release download v0.0.40 -D /tmp/v0.0.40
# The tar stream inside the archive, byte for byte:
git archive --format=tar --prefix=crypto-service-sdk-go-0.0.40/ v0.0.40 | sha256sum
gunzip -c /tmp/v0.0.40/crypto-service-sdk-go-0.0.40.tar.gz | sha256sum
# The SBOM without its serial number, time stamp and tool hashes,
# rebuilt with the command v0.0.40's release script ran:
go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.12.0 \
  mod -json -licenses -output /tmp/v0.0.40/rebuilt.cdx.json .
filter='del(.serialNumber, .metadata.timestamp, .metadata.tools[].hashes)'
jq -S "$filter" /tmp/v0.0.40/crypto-service-sdk-go-0.0.40.cdx.json | sha256sum
jq -S "$filter" /tmp/v0.0.40/rebuilt.cdx.json | sha256sum
```

Expected values for v0.0.40:

| What | SHA-256 |
| :--- | :--- |
| `crypto-service-sdk-go-0.0.40.tar.gz` (released, `SHA256SUMS`) | `b4990e8ad8d51d0039a61c37a5c181689a36aa4d7c580e5c158bd0db1a2db905` |
| `crypto-service-sdk-go-0.0.40.cdx.json` (released, `SHA256SUMS`) | `7fd4b33131c8ab10275e3435fccd216b8cd8be2cefe481755b2f74c545b82437` |
| The tar stream inside the archive | `0c577ddd896b42ec10dcbb1143b04e7b795c6bf25e6bc085517534474d871abb` |
| The SBOM through the `jq` filter | `5ffa5dd47a09c78a6a4a103a0efd0f8f7b7b09faf3945cb00a592fde8bf0b278` |

Both pairs matched when checked on 2026-11-18 (git 2.56.1, gzip 1.15,
Go 1.27.4, linux/amd64), and two runs of the current
`scripts/release-artifacts.sh` at v0.0.40, from two clones with
separate module caches, gave byte-identical `dist/` files.
