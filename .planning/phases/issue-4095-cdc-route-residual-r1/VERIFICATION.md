# Verification — Issue 4095 residual route

| Acceptance criterion | Result | Evidence |
| --- | --- | --- |
| Live PostgreSQL CDC → PostgreSQL R4 route | pass | `TestPostgresCDCToManagedTargetHistoryRouteLive` ran against PostgreSQL 16 using pgoutput v2. Its real insert/update/delete transaction passed the existing durable stage, receiver Parquet receipt, workset, `MappingContractV1` tombstone mapping, and a live keyed history target; an independent normal PostgreSQL connection read the resulting history rows. |
| Receipt precedes LSN acknowledgement; restart and replay are covered | pass | The committer reads `pg_replication_slots.confirmed_flush_lsn` before it applies the received workset, then acknowledges only after downstream durable acknowledgement. The live test restarts source and target, observes an existing active slot whose `confirmed_flush_lsn` remains at the first durable checkpoint, and replays through a third target connection without changing independent SQL read-back. |
| Enumerated R1/R2/destination `change_capture` pre-I/O matrix | pass | `TestDeclaredChangeCaptureRoutesRefuseBeforeIO` names R1 GitHub→GitHub, R2 GitHub→PostgreSQL, R3 PostgreSQL→GitHub, and R4 PostgreSQL→PostgreSQL. Each gets `*synccontract.ModeNotExecutableError` with the concrete local-warehouse reason before zero probe `Check`, `Catalog`, `Read`, `ReadCDC`, or `Write` calls. |
| R3 PostgreSQL CDC → API | non_executable | GitHub destination `sync_transport.json` has no CDC source binding and declares deletes unavailable; `change_capture` is source-only to the connection-owned local warehouse. This task intentionally adds no API writer/action. |
| No surface overstates R3 support | pass | The 2026-08-17 declaration inspection found GitHub's destination binds only the declarative GitHub stream and `postgres_polling_watermark`, never the PostgreSQL logical-replication CDC executor, and says `deletes: not_available`. GitHub's metadata keeps `cdc: false`; PostgreSQL's polling transport omits `change_capture`. No inspected declaration or capability surface advertises R3 as executable. |

## Counterfactual and target-limit finding

The initial route test looked stalled after the first checkpoint because its test committer waited only for a callback and did not observe that `ReadCDC` had already returned an error. The requested counterfactuals disproved a restart regression after the fixture was corrected:

1. no restart — pass;
2. source-only restart — pass, with `pg_replication_slots` reporting `exists=true active=true confirmed_flush_lsn=0/1975618 restart_lsn=0/1975420` after re-entry;
3. target-only restart — pass, preserving the same independently observed slot position.

The first counterfactual revealed the specific fixture fault: its one three-event committed pgoutput transaction was sent to a test-only target descriptor with `MaxBatchRecords: 2`. The engine already refuses that unsafe page before target mutation — `TestApplyPollingPageRefusesUnsafeInputBeforeTargetMutation/record_limit` passes with zero target calls. Its error is a generic validation error rather than a mode-refusal type; adding a new public typed page-limit error is broader than this residual and was not absorbed. The live test now observes a reader error immediately rather than masking it as a long wait.

## Required gates

- Targeted Go test commands for every changed package and `./internal/cli`.
- `go test -timeout 20m ./cmd/connectorgen` (consumer-package rule).
- Tagged PostgreSQL live test with the shared Docker endpoint; any unavailable endpoint will be recorded as `not_run` with the exact reason.
- `gofmt`, `go vet`, `go build ./cmd/pm`, individual `make verify` non-test gates, generator byte-stability checks, connector boundary, and website docs generation twice if repository policy requires it.

## Execution record

Executed as of 2026-08-17:

```text
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres -run '^TestPostgresCDCToManagedTargetHistoryRouteLive$' -count=1 -v
PASS (PostgreSQL 16 live pgoutput route; source/target restart and replay)

POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres -run '^TestPostgresCDCToManagedTargetHistoryRouteRestartCounterfactualsLive$' -count=1 -v
PASS (no restart, source-only restart, target-only restart)

go test -timeout 20m ./internal/app -run '^TestDeclaredChangeCaptureRoutesRefuseBeforeIO$' -count=1 -v
PASS (four named typed zero-I/O rows)

go test -timeout 20m ./internal/connectors/engine -run '^TestApplyPollingPageRefusesUnsafeInputBeforeTargetMutation$/^record_limit$' -count=1 -v
PASS (declared page limit refuses before target mutation)
```

The remaining repository gates below are pending at this planning checkpoint and will be marked with their actual results before handoff; no unrun gate is a pass.
