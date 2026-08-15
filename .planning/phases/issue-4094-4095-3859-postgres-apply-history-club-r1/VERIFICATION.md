# Verification — PostgreSQL apply/history club

Status: passed on 2026-08-15.

## Acceptance evidence

| Item | Acceptance criterion | Evidence | Observable assertion, or why this individual fake is necessary |
| --- | --- | --- | --- |
| #4094 | PostgreSQL → PostgreSQL maintains keyed versions and correct validity windows | `TestPostgresManagedTargetIncrementalDedupeHistoryLive` against the Docker/Colima PostgreSQL | Adapter-applied v1 produced one open current row. Adapter-applied v2 produced two rows, closed v1 exactly at v2's `_valid_from`, and left only v2 current. |
| #4094 | A soft delete closes history without physically deleting it | Same live test, using a tombstone created by `CDCDeleteTombstone` and mapped by `MappingContractV1` | Both history versions remained queryable; v2 gained a non-null `_valid_to` and `_is_current=false`. The durable delivery ID changed after the close. |
| #4094 | Late/replayed history input is deterministic | Same live test after constructing a fresh native connection, ledger, write executor, and polling adapter | Replaying v1 and v2 returned a durable acknowledgement and new receipt while the queried history rows remained byte-for-value equivalent to the pre-restart rows. |
| #4094 | History receipts and restart recovery are durable | Same live test plus the managed-target delivery ledger | Every apply required a non-empty sink/time and a non-empty, changed delivery ID read from PostgreSQL; the new ledger/executor recovered and continued applying a close. |
| #4094 | The three non-PostgreSQL route cells refuse with typed source/destination reasons before I/O | `TestIncrementalDedupeHistoryRefusesEachNonPostgresRouteBeforeSessionMutation` | Individual fakes are necessary to prove an unsupported engine is not contacted. Each cell asserted `DatabaseWriteHistoryRouteError.Reason`, zero begin/batch/commit/rollback/ledger counters, and preservation of a sentinel target row. |
| #4095 | R3/R4 insert, update, and delete reach keyed apply/history close | Full tagged PostgreSQL package, `TestPostgresPGOutputV2ContainerHarness`, `TestPostgresManagedTargetWorksetDeliveryLive`, and the history live test | The live decoder/read paths observed committed insert/update/delete events; workset read-back showed retained omission then explicit keyed delete; the CDC-derived delete closed the live history row. |
| #4095 | Receipt precedes source acknowledgement/LSN advancement and remains replay-safe across restart | `TestPGOutputV2StreamCommitReceiptsBeforeCheckpointAndAcknowledgement`, the live PGOutput harness, and the restarted history live test | The deterministic probe asserted exact `emit → receipt → checkpoint → ack` order; the live slot advanced only after the durable checkpoint; restarted apply wrote a receipt without changing history state. |
| #4095 | R1/R2 and destination CDC declarations refuse before I/O | Existing preflight/CDC closed-contract tests exercised by the focused default packages and full tagged PostgreSQL package | These fakes are necessary because forbidden routes must not make provider calls. They assert typed admission/refusal results and zero downstream activity; abort/mismatch tests additionally assert no event, receipt, checkpoint, or acknowledgement. |
| #3859 residual | The database polling adapter can plan the required history strategy | Red/green execution of `TestPostgresManagedTargetIncrementalDedupeHistoryLive` | Red returned a zero acknowledgement with the missing-route plan error. Green passed the same page through `DatabasePollingApplyExecutor`, then queried the new row and durable receipt from PostgreSQL. |

## TDD evidence

**Red:** Before the production edit, the focused live history test failed at
the adapter plan boundary with a zero acknowledgement and `database history
route source and destination do not match the declared managed-target driver`.
No adapter-driven history row or receipt could exist. Exact output is retained
in `traces/red-adapter-history-plan.txt`.

**Green:** The same live test passed in 6.24s. It observed four distinct durable
delivery IDs across v1, v2, restart replay, and the CDC-derived close, and
queried each resulting validity state. Exact output is retained in
`traces/green-adapter-history-live.txt`; route/CDC focused output is retained in
`traces/green-route-cdc-focused.txt`.

## Verification commands

- PASS — `go test -timeout 20m -count=1 ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres`
- PASS — `go test -race -timeout 20m -count=1 ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres`
- PASS — focused live history, non-PostgreSQL route matrix, CDC tombstone mapping, and live workset commands recorded under `traces/`.
- PASS — full live PostgreSQL package: `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres` (47.189s).
- PASS — `go vet ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres`.
- PASS — `go build ./cmd/pm`.
- PASS — `make tidy-check`, `make lint` (zero issues), `make docs-check-no-build`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate` (552 connectors, zero findings), `make connectorgen-surface-sync` (552 connectors, zero changes), `make connector-boundary`, and `make release-workflow-check`.

The requested tagged `go test ... ./...` completed all packages except one
parallel container-harness teardown/LSN-inspection failure in the untouched
`TestPostgresPGOutputV2ContainerHarness`. Its exact isolated rerun passed in
10.22s, and the entire tagged PostgreSQL package then passed in 47.189s. Per the
high-load direction, the broad suite was not repeated. This is recorded as a
transient full-suite observation, not hidden as a clean broad run.

## Scope and baseline notes

- No file under `internal/synctransport`, descriptor validation, or the
  transport registry changed; #4093 owns that shared work.
- #4125 was not touched.
- #4158 was supplied as a pre-existing base failure and was not taken as a
  separate fix. The current full tagged PostgreSQL package, including
  `TestPostgresManagedTargetDriverLiveControlAssertions`, passes after the
  adapter residual fixed here.
- CLI/help/manual/docs/website parity is not applicable because no public
  command surface changed.
