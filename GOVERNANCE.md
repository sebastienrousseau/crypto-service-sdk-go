<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- SPDX-FileCopyrightText: 2026 Sebastien Rousseau -->

# Governance

How crypto-service-sdk-go is run, who decides, and what happens if its
maintainer becomes unavailable. It follows the
[Crypto Service suite's governance](https://github.com/sebastienrousseau/crypto-service/blob/main/GOVERNANCE.md);
this page records what is specific to this repository.

## Scope

A thin, typed Go client for the Crypto Service REST API, held to the
server's conformance kit. Changes are weighed against that scope:
correctness, security and a small, stable API over breadth.
Cryptography belongs in
[crypto-service](https://github.com/sebastienrousseau/crypto-service),
not here.

## Roles and responsibilities

| Role | Who | Responsibilities |
| :--- | :--- | :--- |
| **Maintainer** | Sebastien Rousseau (sole maintainer) | Sets direction, reviews and merges pull requests, opens each suite release, signs tags, approves the `release` environment, triages issues and security reports |
| **Contributor** | Anyone with a merged pull request | Proposes changes with tests, signs off every commit (DCO), keeps CI green |
| **AI-assisted contributor** | Agents working under a human's direction | Follow [`AGENTS.md`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/AGENTS.md); never merge, tag or publish; the human signs off and is accountable |
| **User** | Everyone | Files issues and questions ([`SUPPORT.md`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/SUPPORT.md)) and reports vulnerabilities privately ([`SECURITY.md`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/SECURITY.md)) |

## Decisions

- **Day-to-day changes** (fixes, tests, docs, additive API within
  scope) go through a pull request into the open `feat/v<version>`
  branch; the maintainer merges once CI is green.
- **Significant changes** (a breaking API change, a new dependency, a
  change to the retry, error or concurrency model) get an ADR in
  `docs/adr/` in the same pull request.
- **Suite-wide changes** (method surface, resilience contract,
  conformance kit) are agreed in crypto-service first and land in every
  SDK in the same suite release.

## Releases

Versions are `0.0.x`, move by `0.0.1` and follow the suite; the
maintainer opens each release. A release is a signed annotated tag; the
release workflow builds and attests the artifacts, writes the GitHub
release and warms proxy.golang.org, gated by the `release` environment
(see [`DEVELOPMENT.md`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/DEVELOPMENT.md#releasing)).

## Continuity plan

The project has one maintainer. Items marked **Not yet arranged** have
no arrangement today.

| Asset | Today | Continuity |
| :--- | :--- | :--- |
| Published versions | Tags on GitHub, cached by proxy.golang.org and recorded in sum.golang.org | The proxy keeps every published version, so builds that depend on one keep working even if the repository disappears. |
| Tag-signing keys | SSH keys listed in `KEYS.asc` | A new maintainer signs with their own key, added in a reviewed change. **Not yet arranged:** a documented key backup. |
| Registry accounts | None: Go publishes from the tag, with no account or token | Nothing to hand over. |
| GitHub repository | Owned by the maintainer's account | Transfer to an organisation with at least two owners, keeping redirects. **Not yet arranged:** no second owner is named. |
| Module path | `github.com/sebastienrousseau/crypto-service-sdk-go` | A transfer keeps working through GitHub's redirect; a fork continues under its own path. |
| The code | Apache-2.0 OR MIT, public history | Anyone can fork and continue it. |

## Conduct and security

Participation follows
[`CODE_OF_CONDUCT.md`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/CODE_OF_CONDUCT.md).
Vulnerabilities go through the private process in
[`SECURITY.md`](https://github.com/sebastienrousseau/crypto-service-sdk-go/blob/main/SECURITY.md),
never a public issue.
