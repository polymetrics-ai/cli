# Verification — PostgreSQL logical-replication CDC containment

## Current containment evidence

- [x] Focused PostgreSQL tests confirm `ReadCDC` returns the fail-closed
  unsupported result before opening a source connection.
- [x] The PostgreSQL bundle and capability projection describe change capture
  as planned and do not advertise CDC.
- [x] `TestHistoricalLogicalReplicationResumesAndCleansSlot` is skipped before
  reading integration configuration, so an enabled integration environment
  cannot create a source slot while change capture is planned.
- [x] Dependency and binary-delta evidence remain recorded in `PR-BODY.md`.

## Historical pre-containment evidence

The following evidence records the withdrawn implementation work. It is kept
for dependency and protocol provenance only; it is not current end-to-end CDC,
source-LSN progress, checkpoint, or slot-lifecycle verification.

- Focused native PostgreSQL unit tests, including negative lifecycle and
  source-bound recovery paths.
- Focused connector/engine/CLI capability-projection tests.
- `POLYMETRICS_INTEGRATION=1 go test ./internal/connectors/native/postgres`
  against a real logical PostgreSQL source.
- Integration proof of DML decode, durable LSN restart semantics, and
  post-teardown slot absence.
- Scoped `go vet` and `go build ./cmd/pm`.
- `make verify` non-suite gates individually: tidy-check, lint, docs-check,
  smoke-no-build, agent-contract-check, connectorgen-validate,
  connectorgen-surface-sync, connector-boundary, release-workflow-check.
- `pm help connectors`, `pm connectors`, `pm connectors inspect postgres --json`,
  and `pm connectors catalog --capability cdc --json`; website connector data
  regenerated.
- Manual code-review record completed after implementation (`REVIEW.md`).

## Historical implementation evidence

- `scripts/gsd doctor`: pass.
- Required GSD command sources: resolved.
- `go run ./cmd/agentcontractgen check`: pass.
- Baseline `go build -trimpath -o <temporary>/pm ./cmd/pm`: **93,577,602 bytes**.
- Post-change binary: **93,740,082 bytes** (+**162,480 bytes**).
- PostgreSQL **12.22** live conformance, on an isolated local data directory
  and non-default port with `wal_level=logical`, `max_replication_slots=4`, and
  `max_wal_senders=4`: historical pass.
