# Docker Hub parity corrective slice — code review

## Method

Standard inline review after resolving scripts/gsd sources code-review and reading the generated GSD prompt. The canonical parent-worker contract forbids role spawning in this worktree, so this is the documented manual fallback.

## Reviewed scope

- internal/safety/safety.go and safety_test.go
- internal/connectors/engine/direct_read.go and direct_read_test.go
- internal/connectors/defs/dockerhub/scim_schema_urn_test.go
- Docker Hub reconciliation evidence under .planning/phases/dockerhub-parity-sweep-r1

## Review evidence

- Read the full corrective diff from the red SCIM test through the final reconciliation commit.
- Re-ran focused safety, engine, Docker Hub definition, fleet command-preflight, race, generator, lint, docs, smoke, release, vet, build, and CLI regression gates.
- Mechanically checked the final verification ledger against cli_surface.json: all 54 implemented commands occur once, with no missing, duplicate, or extra row.
- Ran the rebuilt binary through all 54 Docker Hub help routes, bare namespace help, and the SCIM schema command help.

## Findings

No Critical, Warning, or Info findings.

The fix correctly separates opaque provider path values from local connector and credential identifiers. It allows only the documented colon, plus, and at-sign cases before url.PathEscape while retaining rejection of separators, traversal, raw percent escapes, whitespace, controls, dangerous Unicode, query/fragment delimiters, and generic identifier loosening.

The Docker Hub regression test uses an isolated local server with the bundle's auth hook removed only in test memory. It proves the canonical SCIM URN reaches the declared path without a credential or live Docker Hub request. The email regression independently proves the shared direct-read route transports a documented opaque email segment.

The live ledger makes no certification claim. Explicit HTTP 403s remain provider-permission dependencies, documented SCIM remains Enterprise-only, and the three intentionally non-dispatched writes name their exact provider dependency.

## Verdict

PASS — no source, security, test-quality, or evidence-integrity issue found in the corrective slice. The PR must remain explicitly **not certified** because live provider acceptance is account-tier and permission dependent.
