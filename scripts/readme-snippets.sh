#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Extracts every fenced code block of README.md and runs it, so the
# README cannot drift from the module:
#
#   ```go   a complete program; built in a scratch module that requires
#           this module at its version (replaced by this checkout) and
#           run against CRYPTO_SERVICE_URL with CRYPTO_SERVICE_API_KEY
#   ```sh   a shell script, run with `sh -eu`; a block that runs
#           `go get` runs in that scratch module (offline, through the
#           replace), any other block in the repository root
#
# Any other info string, or a fence without one, fails: every block in
# the README must be runnable. Needs a running server, like
# `make conformance`; fails when CRYPTO_SERVICE_URL is unset.
#
#   scripts/readme-snippets.sh [README.md]
set -euo pipefail

cd "$(dirname "$0")/.."
root="$(pwd)"
readme="${1:-README.md}"
: "${CRYPTO_SERVICE_URL:?set CRYPTO_SERVICE_URL to a running crypto-server}"

module="$(sed -n 's/^module //p' go.mod)"
version="$(sed -n 's/^const Version = "\(.*\)"$/\1/p' version.go)"
work="$(mktemp -d "${TMPDIR:-/tmp}/crypto-service-sdk-go-readme.XXXXXX")"
trap 'rm -rf -- "$work"' EXIT

# Split the README into numbered block files: <n>.<lang>.
awk -v dir="$work" '
  /^```/ {
    if (inside) { inside = 0; close(file); next }
    inside = 1; n++
    lang = substr($0, 4); if (lang == "") lang = "none"
    file = sprintf("%s/%02d.%s", dir, n, lang)
    printf "" > file
    next
  }
  inside { print > file }
' "$readme"

scratch_module() {
  local dir="$1"
  mkdir -p "$dir"
  cat >"$dir/go.mod" <<EOF
module example.com/readme

go 1.26

require $module v$version

replace $module => $root
EOF
}

export GOFLAGS=-mod=mod GOPROXY=off GONOSUMDB="$module" GOSUMDB=off

count=0
for block in "$work"/*.*; do
  [ -e "$block" ] || continue
  name="$(basename "$block")"
  count=$((count + 1))
  case "$name" in
    *.go)
      dir="$work/${name%.go}"
      scratch_module "$dir"
      cp "$block" "$dir/main.go"
      echo "readme: $name (go run)"
      (cd "$dir" && go run .)
      ;;
    *.sh)
      if grep -q 'go get' "$block"; then
        dir="$work/${name%.sh}"
        scratch_module "$dir"
        echo "readme: $name (sh, scratch module)"
        (cd "$dir" && sh -eu "$block")
      else
        echo "readme: $name (sh, repository root)"
        (cd "$root" && sh -eu "$block")
      fi
      ;;
    *)
      echo "readme: $name: every README block must be go or sh" >&2
      exit 1
      ;;
  esac
done
test "$count" -gt 0 || { echo "readme: no code blocks in $readme" >&2; exit 1; }
echo "readme: $count blocks ran"
