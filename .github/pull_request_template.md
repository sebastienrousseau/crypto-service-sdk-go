<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

## What this changes

<!-- What behaviour changes, and why. -->

## How it was verified

<!-- What you ran, and what you observed. Name the test and the result
     it produced; say whether you ran `make conformance` against a live
     server (CRYPTO_SERVICE_URL, CRYPTO_SERVICE_API_KEY and
     CRYPTO_SERVICE_JWT_SECRET set) or `make regression`. -->

## Checklist

- [ ] Every commit is signed off (`git commit -s`), see [DCO.txt](../DCO.txt)
- [ ] `make check` passes, and `make conformance` (or `make regression`) against a live server
- [ ] Tests cover the change
- [ ] `conformance/` is untouched, or replaced whole from crypto-service at the commit in `ci/server.ref`, `SHA256SUMS` included
- [ ] Documentation updated if behaviour or interfaces changed
