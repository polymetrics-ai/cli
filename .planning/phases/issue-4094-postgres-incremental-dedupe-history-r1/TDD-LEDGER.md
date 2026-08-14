# TDD ledger — Issue 4094

Manual inline GSD execution; the generated Pi workflow cannot run compatible
isolated agents in this worker. Each entry records an observable state that an
empty implementation cannot satisfy.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | PostgreSQL-only routing | `TestIncrementalDedupeHistoryRefusesEachNonPostgresRouteBeforeSessionMutation` initially did not compile because the sealed route types and gate were absent. | Each inapplicable route returns its typed source, destination, or both-legs reason while fake driver begin/batch/commit/rollback, ledger writes, and target mutation remain zero. |
| R2 | Validity windows | `TestPostgresManagedTargetIncrementalDedupeHistoryLive` had no target implementation to create/close history rows. | Queries show v1 retained and closed exactly at v2's valid-from boundary, with v2 the only current row. |
| R3 | Soft delete | The pre-change driver used physical `DELETE` for every tombstone. | The live query retains both rows, clears v2's current flag, and sets its validity end. |
| R4 | Late/replay determinism | The pre-change driver had no history-version comparison or replay behavior. | A fresh executor replays v1/v2 and the queried row set is byte-for-byte stable; its new durable receipt is read back from the ledger. |
| R5 | Recovery | The pre-change driver had no history state to recover. | The live test closes the initial native connection, builds a fresh driver/ledger/executor, then observes the same durable rows before closing v2. |
| R6 | Real PostgreSQL write | The pre-change native driver did not admit or write `incremental_dedupe_history`. | Tagged dbtest queries the actual target relation and exact history row states after every write phase. |
| R7 | Bundle-only widening | `TestBundleLoadPostgresDatabaseDefinitionWithProvenCDCCapability` failed because it encoded the pre-#4094 five-mode PostgreSQL definition. | The PostgreSQL bundle asserts the sixth history mode, while an adjacent synthetic non-PostgreSQL database definition retains exactly the original five modes. Existing R1 proves all other routes refuse before I/O. |

## Planned red commands

```sh
go test -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test.*(IncrementalDedupeHistory|History.*Route|Route.*History)' -count=1
```

## Green commands and results

```sh
go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/... -count=1
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
```

The targeted package command and focused tagged
`TestPostgresManagedTargetIncrementalDedupeHistoryLive` command passed. The
final full tagged command is recorded in verification evidence before the PR.

## CI regression green command

```sh
go test -timeout 20m ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres
```

This passed after the PostgreSQL six-mode expectation and non-PostgreSQL
five-mode guard were added. No other test or fixture changed.
