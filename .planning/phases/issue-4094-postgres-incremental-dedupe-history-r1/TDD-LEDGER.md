# TDD ledger — Issue 4094

Manual inline GSD execution; the generated Pi workflow cannot run compatible
isolated agents in this worker. Each entry records an observable state that an
empty implementation cannot satisfy.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | PostgreSQL-only routing | A non-PostgreSQL source or destination enters any source/target/provider/database operation. | Each inapplicable route returns its typed route reason and fake source, driver, session, ledger, and mutation counters are zero. |
| R2 | Validity windows | A changed keyed record leaves more than one current version or fails to close the prior version. | Queried history rows show one current version and the previous row closed at the change boundary. |
| R3 | Soft delete | A tombstone leaves a current version or removes history improperly. | Queries show a closed row with the required soft-delete state and no current version. |
| R4 | Late/replay determinism | A late/replayed input produces duplicate/conflicting history rows or a non-deterministic receipt. | Repeating the same sealed input yields the prescribed stable rows and receipt/ledger state. |
| R5 | Recovery | Restart loses the durable history/receipt recovery state or reopens history. | A newly constructed controller recovers the same receipt/ledger and queries unchanged history state. |
| R6 | Real PostgreSQL write | Success can be a no-op while reports pass. | Tagged dbtest queries the actual target relation and exact history row states after each write phase. |

## Planned red commands

```sh
go test -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test.*(IncrementalDedupeHistory|History.*Route|Route.*History)' -count=1
```

## Planned green commands

```sh
go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/... -count=1
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
```
