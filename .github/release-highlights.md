<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

<!--
The hand-written Highlights of the next GitHub release, read by
scripts/release-notes.sh. Two to four bullets, each
"* **<Headline>**: <one or two plain sentences>". No em dashes.
Rewrite it in the release-preparation commit of every version.
-->

* **First release, in lockstep with the Crypto Service suite 0.0.40**: a typed Go client for every v2 operation of crypto-server, standard library only, held to the server's conformance kit in CI against a live server.
* **One resilience contract across the SDKs**: per-attempt timeouts and an overall deadline, opt-in retries with full jitter and `Retry-After` that never resend a stateful call the server may have acted on, refused redirects and TLS verification that cannot be turned off.
* **Errors without secrets**: RFC 9457 problems as `*APIError`, transport failures with `TimedOut`, and messages that never carry a credential, key, plaintext or body, held by two fuzz targets.
