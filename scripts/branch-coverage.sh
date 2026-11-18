#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
# SPDX-FileCopyrightText: 2026 Sebastien Rousseau
#
# Branch coverage of the library, measured with gobco
# (https://github.com/rillig/gobco, BSD-2-Clause) in its -branch mode:
# every if, else, switch case and for/range condition must have been
# taken both ways by the unit tests. Fails below 100% of branches, the
# same floor scripts/coverage.sh holds for statements. The measured
# packages are the ones coverage.sh measures: the module root and
# internal/kit. The live tests skip here, as in `make test`.
#
#   scripts/branch-coverage.sh
#
# gobco runs through `go run` at a pinned version, like the other
# analysers. It instruments a copy of the module in a temporary
# directory; the work tree is not touched.
set -eu

cd "$(dirname "$0")/.."

GOBCO="github.com/rillig/gobco@v1.3.4"

failures=0
for pkg in . ./internal/kit; do
  out="$(go run "$GOBCO" -branch "$pkg" 2>&1)" || {
    echo "$out" >&2
    echo "branch-coverage: gobco failed on $pkg" >&2
    exit 1
  }
  line="$(printf '%s\n' "$out" | grep '^Branch coverage: ' || true)"
  covered="$(printf '%s\n' "$line" | sed -n 's|^Branch coverage: \([0-9]*\)/\([0-9]*\)$|\1|p')"
  total="$(printf '%s\n' "$line" | sed -n 's|^Branch coverage: \([0-9]*\)/\([0-9]*\)$|\2|p')"
  if [ -z "$covered" ] || [ -z "$total" ] || [ "$total" -eq 0 ]; then
    echo "$out" >&2
    echo "branch-coverage: no 'Branch coverage: n/m' line for $pkg" >&2
    exit 1
  fi
  if [ "$covered" -ne "$total" ]; then
    printf '%s\n' "$out" >&2
    echo "branch-coverage: $pkg covers $covered of $total branches, the floor is 100%" >&2
    failures=$((failures + 1))
  else
    echo "branch-coverage: $pkg covers $covered of $total branches (100%)"
  fi
done
test "$failures" -eq 0
