<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Roadmap

What may come next, in rough order. Nothing here is a commitment; each
item lands only if it serves users of the suite and keeps the client
thin. The first release, 0.0.40 (2026-11-16), shipped what this list
opened with: the tag on proxy.golang.org and a GitHub release with
provenance and an SBOM; [CHANGELOG.md](../CHANGELOG.md) records it.

1. **Published benchmark results**: generate `docs/BENCHMARKS.md` from
   `make bench` runs with their environment recorded, comparable
   across the suite's SDKs.
2. **Response-schema validation in the conformance run**: check each
   answer against the OpenAPI response schema, as the server's own
   runner does, in addition to the kit's templates.
3. **Typed helpers for multi-step flows** where the server adds them
   (for example key rotation), only once the server and the TypeScript
   SDK define them.

Out of scope: local cryptography, an offline mode, wrapping the
deprecated v1 routes.
