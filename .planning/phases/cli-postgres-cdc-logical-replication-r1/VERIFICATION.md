# Verification — PostgreSQL logical-replication CDC

## Required evidence

- [x] Focused native PostgreSQL unit tests, including negative lifecycle and source-bound recovery paths.
- [x] Focused connector/engine/CLI capability-projection tests.
- [x] `POLYMETRICS_INTEGRATION=1 go test ./internal/connectors/native/postgres` against a real logical PostgreSQL source.
- [x] Integration test proves DML decode, durable LSN restart semantics, and post-teardown slot absence.
- [x] Scoped `go vet` and `go build ./cmd/pm`.
- [x] `make verify` non-suite gates individually: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check.
- [x] `pm help connectors`, `pm connectors`, `pm connectors inspect postgres --json`, and `pm connectors catalog --capability cdc --json` checked; website connector data regenerated.
- [x] Final binary-size delta and dependency evidence recorded in `PR-BODY.md`.
- [x] Manual code-review record completed after implementation (`REVIEW.md`).

## Initial evidence

- `scripts/gsd doctor`: pass.
- Required GSD command sources: resolved.
- `go run ./cmd/agentcontractgen check`: pass.
- Baseline `go build -trimpath -o <temporary>/pm ./cmd/pm`: **93,577,602 bytes**.
- Post-change binary: **93,740,082 bytes** (+**162,480 bytes**).
- Live conformance: PostgreSQL **12.22**, isolated local data directory,
  non-default port, `wal_level=logical`, `max_replication_slots=4`, and
  `max_wal_senders=4`: pass.
