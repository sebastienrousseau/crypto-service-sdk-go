<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Benchmarks

No results are published yet. This page will be generated from
`make bench` output, with the machine, Go version, server commit and
load recorded next to every number.

## Scenarios

The scenario names are shared by every Crypto Service SDK, so results
can be compared across languages on one machine.

| Scenario | What it measures | Network |
| :--- | :--- | :--- |
| `encode_request` | Encoding a `HashParams` with 1 KiB of data to JSON | None |
| `decode_response` | Decoding a recorded `/v2/hash` answer into `HashResult` | None |
| `decode_problem` | Parsing a recorded validation problem into an `*APIError` | None |
| `round_trip_hash` | One `POST /v2/hash` to a loopback crypto-server, client and server work together | Loopback; needs `CRYPTO_SERVICE_URL` |

## Running them

```sh
make bench                                  # 6 samples of 1 s each
CRYPTO_SERVICE_URL=http://127.0.0.1:3000 \
  CRYPTO_SERVICE_JWT_SECRET=... make bench  # adds round_trip_hash
```

The output in `bin/bench.txt` is benchstat input. `round_trip_hash`
authenticates with JWTs for principals that change every 50 requests, so
the server's quota of 60 requests a minute per principal does not
throttle it. CI smoke-runs every scenario once in the conformance job.
