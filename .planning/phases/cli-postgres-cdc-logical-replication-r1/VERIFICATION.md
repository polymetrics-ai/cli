# Verification — PostgreSQL logical-replication CDC

## Required evidence

- [x] Focused native PostgreSQL unit tests, including negative lifecycle and source-bound recovery paths.
- [x] Focused connector/engine/CLI capability-projection tests.
- [x] `POLYMETRICS_INTEGRATION=1 go test -count=1 -timeout 20m ./internal/connectors/native/postgres` against a real logical PostgreSQL source.
- [x] Integration test proves DML decode, durable LSN restart semantics, and post-teardown slot absence.
- [x] Scoped `go vet` and `go build ./cmd/pm`.
- [x] `make verify` non-suite gates individually: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check.
- [x] `pm help connectors`, `pm connectors`, `pm connectors inspect postgres --json`, and `pm connectors catalog --capability cdc --json` checked; website connector data regenerated.
- [x] Rebased onto `origin/main` `f96a47e80`; inspected the #3964 golden/help change and reran the real exit-code-sensitive CLI suite.
- [x] Final binary-size delta and dependency evidence recorded in `PR-BODY.md`.
- [x] Manual code-review record completed after implementation (`REVIEW.md`).

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
- Live conformance: PostgreSQL **12.22**, isolated local data directory,
  non-default port, `wal_level=logical`, `max_replication_slots=8`, and
  `max_wal_senders=8`: pass. It proves selected-table filtering in a
  multi-table publication; insert/update/delete/truncate mapping; durable LSN
  restart without replaying prior records; missing-checkpoint rebootstrap
  refusal; active-slot teardown refusal; and inactive slot cleanup.
