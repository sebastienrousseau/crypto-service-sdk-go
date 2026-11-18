#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0 OR MIT
# SPDX-FileCopyrightText: 2026 Sebastien Rousseau
#
# The full regression run, as CI's conformance job runs it:
#
#   1. uses the built crypto-service checkout in CRYPTO_SERVICE_SERVER_DIR
#      (CI checks out the commit in ci/server.ref and builds it with
#      `pnpm install --frozen-lockfile && pnpm -r run build`); it refuses
#      a checkout at any other commit unless ALLOW_OTHER_SERVER=1;
#   2. starts `node packages/crypto-server/dist/index.js` on a free
#      loopback port with a random CRYPTO_API_KEY and JWT_SECRET and a
#      raised RATE_LIMIT_MAX, and waits for /health;
#   3. runs `make check` and `make conformance` against it: every static
#      gate, the tests at 100% coverage, the conformance kit, the typed
#      live tests, every example and every README code block;
#   4. smoke-runs every benchmark once (`make bench`);
#   5. stops the server it started (by PID), also on failure.
#
#   CRYPTO_SERVICE_SERVER_DIR=/path/to/crypto-service make regression
#
# Set REGRESSION_PORT to pick the port, and VULNDB to point govulncheck
# at a local database copy.
set -euo pipefail

cd "$(dirname "$0")/.."

server_dir="${CRYPTO_SERVICE_SERVER_DIR:?set CRYPTO_SERVICE_SERVER_DIR to a built crypto-service checkout}"
entry="$server_dir/packages/crypto-server/dist/index.js"
test -f "$entry" || { echo "regression: $entry not found; build the checkout first" >&2; exit 1; }

pin="$(tr -d '[:space:]' < ci/server.ref)"
head="$(git -C "$server_dir" rev-parse HEAD 2>/dev/null || true)"
if [ "$head" != "$pin" ] && [ "${ALLOW_OTHER_SERVER:-}" != "1" ]; then
  echo "regression: $server_dir is at '${head:-unknown}', ci/server.ref pins $pin" >&2
  exit 1
fi

log="$(mktemp "${TMPDIR:-/tmp}/crypto-service-sdk-go-server.XXXXXX")"
server_pid=""

cleanup() {
  status=$?
  if [ -n "$server_pid" ] && kill -0 "$server_pid" 2>/dev/null; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  if [ "$status" -ne 0 ]; then
    echo "regression: failed; server log follows" >&2
    tail -n 50 "$log" >&2 || true
  fi
  rm -f -- "$log"
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

free_port() {
  node -e 'const s=require("net").createServer();s.listen(0,"127.0.0.1",()=>{console.log(s.address().port);s.close()})'
}

random_hex() {
  od -An -N"$1" -tx1 /dev/urandom | tr -d ' \n'
}

port="${REGRESSION_PORT:-$(free_port)}"
CRYPTO_SERVICE_API_KEY="regression-$(random_hex 16)"
CRYPTO_SERVICE_JWT_SECRET="$(random_hex 32)"
export CRYPTO_SERVICE_API_KEY CRYPTO_SERVICE_JWT_SECRET
export CRYPTO_SERVICE_URL="http://127.0.0.1:$port"

echo "regression: starting crypto-server $pin on 127.0.0.1:$port"
(
  cd "$server_dir/packages/crypto-server"
  PORT="$port" HOST=127.0.0.1 NODE_ENV=development RATE_LIMIT_MAX=100000 \
    CRYPTO_API_KEY="$CRYPTO_SERVICE_API_KEY" JWT_SECRET="$CRYPTO_SERVICE_JWT_SECRET" \
    exec node dist/index.js
) >"$log" 2>&1 &
server_pid=$!

for _ in $(seq 1 60); do
  if curl -fsS "$CRYPTO_SERVICE_URL/health" >/dev/null 2>&1; then
    break
  fi
  kill -0 "$server_pid" 2>/dev/null || { echo "regression: the server exited during start-up" >&2; exit 1; }
  sleep 1
done
curl -fsS "$CRYPTO_SERVICE_URL/health" >/dev/null

echo "regression: make check"
make check
echo "regression: make conformance against $CRYPTO_SERVICE_URL"
make conformance
echo "regression: benchmark smoke run"
make bench BENCHCOUNT=1 BENCHTIME=1x
echo "regression: passed"
