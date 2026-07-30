# Summary — Connector Guard Issue C Certification Migration

## What changed

- Added optional typed `certification.json` bundle metadata and embedded schema validation.
- Added `internal/connectors/defs/github/certification.json` for GitHub certification contracts: source defaults/default stream/live-unavailable classifiers, direct-read candidates, binary candidate, and write pairings.
- Updated certification runtime to load pairings, candidates, source defaults/classifiers, and write record schemas from connector definitions instead of shared GitHub-specific Go tables.
- Removed the six obsolete GitHub `provider_certify_contract` Connector Boundary Guard exceptions without changing scanner rules.
- Added regression tests for parsing/malformed metadata, GitHub behavior, record schema sourcing, unknown connector no-op behavior, and sweeper safety.

## Safety and scope

- No live or credentialed connector checks were run.
- No secrets were requested or stored.
- Reverse ETL gates remain plan → preview → approval → execute.
- No PM Broker, release/Homebrew, native-hook rollout, generated public docs, or website work was included.

## Verification

- Focused Go tests passed: `go test ./internal/connectors/engine ./internal/connectors/certify ./cmd/connectorgen ./internal/connectors/boundary`.
- Full defs validation passed: 548 connectors, 0 findings, 0 warnings.
- Boundary guard passed: 0 findings, 0 warnings, 6 remaining non-certify exceptions.
- `make connector-boundary`, `make verify`, and `git diff --check` passed.
