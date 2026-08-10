# Verification — PostgreSQL logical-replication CDC containment

## Current containment evidence

- [x] Focused native PostgreSQL tests prove `ReadCDC` fails closed before a connection, non-ASCII values round-trip byte-exact, invalid UTF-8 is rejected, and replica-identity mode guards are named.
- [x] Focused CLI capability-projection tests prove `cdc=false`, PostgreSQL is absent from the CDC catalogue, and inspect reports `planned` with a reason.
- [x] Full `go test -count=1 -timeout 20m ./internal/cli` passes after containment (**662.780s**); scoped PostgreSQL tests, vet/build, connector validation/surface sync, connector docs validation, and `agentcontractgen check` also pass.
- [x] The integration test skips before source access while CDC is planned; it is not current conformance evidence.
- [x] Scoped `go vet` and `go build ./cmd/pm`.
- [x] `make verify` non-suite gates individually: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check.
- [x] `pm help connectors`, `pm connectors`, `pm connectors inspect postgres --json`, and `pm connectors catalog --capability cdc --json` checked; website connector data regenerated.
- [x] Rebased onto `origin/main` `f96a47e80`; inspected the #3964 golden/help change and reran the real exit-code-sensitive CLI suite.
- [x] Final binary-size delta and dependency evidence recorded in `PR-BODY.md`.
- [x] Manual code-review record completed after implementation (`REVIEW.md`).

## Historical pre-containment evidence

The following records predate the captain's streamed-transaction ruling. They
are retained for provenance and dependency review only; they do not claim
current CDC, source progress, checkpoint, or slot lifecycle conformance.

## Recorded evidence

- `scripts/gsd doctor`: pass.
- Required GSD command sources: resolved.
- `go run ./cmd/agentcontractgen check`: pass.
- Rebased cleanly onto `origin/main` at `f96a47e80`; regenerated the website
  connector catalog rather than hand-merging it. Its only working-tree
  propagation is PostgreSQL CDC metadata/docs. The upstream PR #3964 golden
  transcript was inspected in the rebase, and the focused exit-code-sensitive
  CLI package suite passes: `go test -count=1 -timeout 20m ./internal/cli`
  (**599.878s**).
- Baseline `go build -trimpath -o <temporary>/pm ./cmd/pm` at `origin/main`:
  **145,079,490 bytes**. Post-change binary on the same Darwin/arm64 host and
  Go `go1.26.5`: **145,239,730 bytes** (+**160,240 bytes**).
- Historical live conformance: PostgreSQL **12.22**, isolated local data directory,
  non-default port, `wal_level=logical`, `max_replication_slots=8`, and
  `max_wal_senders=8`: pass. It proves selected-table filtering in a
  multi-table publication; insert/update/delete/truncate mapping; durable LSN
  restart without replaying prior records; missing-checkpoint rebootstrap
  refusal; active-slot teardown refusal; and inactive slot cleanup.
