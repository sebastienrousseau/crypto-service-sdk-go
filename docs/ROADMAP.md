<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Roadmap

What may come next, in rough order. Nothing here is a commitment; each
item lands only if it serves users of the suite and keeps the client
thin.

1. **First release, 0.0.40**: tag, proxy.golang.org, a GitHub release
   with provenance and an SBOM.
2. **Published benchmark results**: generate `docs/BENCHMARKS.md` from
   `make bench` runs with their environment recorded, comparable
   across the suite's SDKs.
3. **Response-schema validation in the conformance run**: check each
   answer against the OpenAPI response schema, as the server's own
   runner does, in addition to the kit's templates.
4. **Typed helpers for multi-step flows** where the server adds them
   (for example key rotation), only once the server and the TypeScript
   SDK define them.

Out of scope: local cryptography, an offline mode, wrapping the
deprecated v1 routes.
