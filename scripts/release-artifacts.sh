#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
# SPDX-FileCopyrightText: 2026 Sebastien Rousseau
#
# Builds the files attached to a GitHub release into dist/:
#
#   crypto-service-sdk-go-<version>.tar.gz   the source at HEAD (git archive)
#   crypto-service-sdk-go-<version>.cdx.json a CycloneDX SBOM of the module
#   SHA256SUMS                        the SHA-256 of both
#
#   scripts/release-artifacts.sh            # version from version.go
#
# Every byte is a function of the commit, so anyone can rebuild a
# release and compare it with SHA256SUMS (docs/REPRODUCING.md):
#
#   - the tar stream is git archive's, whose file times come from the
#     commit; it is compressed by gzip -n (no name, no time stamp), not
#     by git's built-in compressor, whose output changes with the git
#     and zlib in use;
#   - the SBOM has no time stamp, its serial number is a name-based
#     UUID (version 5) of the module, version and the tar stream's
#     SHA-256 instead of a random one, and cyclonedx-gomod, whose own
#     hashes the SBOM records, is built by the Go release that mise.toml
#     pins (GOTOOLCHAIN), with -trimpath, without cgo, for linux/amd64.
#
# Needs git, gzip, python3 (for the UUID) and Go; the pinned Go release
# is downloaded by the go command when it is not the one installed.
set -eu

cd "$(dirname "$0")/.."

CYCLONEDX="github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.12.0"

module="$(sed -n 's/^module //p' go.mod)"
version="$(sed -n 's/^const Version = "\(.*\)"$/\1/p' version.go)"
sbom_go="$(sed -n 's/^go = "\(.*\)"$/\1/p' mise.toml)"
test -n "$sbom_go" || { echo "release-artifacts: no go = \"x.y.z\" in mise.toml" >&2; exit 1; }
name="crypto-service-sdk-go-$version"
rm -rf dist
mkdir -p dist

git archive --format=tar --prefix="$name/" -o "dist/$name.tar" HEAD
tar_sha="$(sha256sum "dist/$name.tar" 2>/dev/null || shasum -a 256 "dist/$name.tar")"
tar_sha="${tar_sha%% *}"
gzip -n -9 "dist/$name.tar"

serial="urn:uuid:$(python3 -I -c 'import sys, uuid; print(uuid.uuid5(uuid.NAMESPACE_URL, sys.argv[1]))' \
  "pkg:golang/$module@v$version?tar-sha256=$tar_sha")"
GOTOOLCHAIN="go$sbom_go" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
  go run -trimpath "$CYCLONEDX" mod -json -licenses -notimestamp -serial "$serial" \
  -output "dist/$name.cdx.json" .

(
  cd dist
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$name.tar.gz" "$name.cdx.json" >SHA256SUMS
  else
    shasum -a 256 "$name.tar.gz" "$name.cdx.json" >SHA256SUMS
  fi
)
echo "tar stream SHA-256: $tar_sha"
cat dist/SHA256SUMS
