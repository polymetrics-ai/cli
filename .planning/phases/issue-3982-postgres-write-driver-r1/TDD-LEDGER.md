# TDD ledger — Issue #3982 PostgreSQL managed-table write driver

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Contract fidelity | Descriptor-only driver does not satisfy provisioning/write-session ports. | Compile-time assertions and behavioural tests exercise only #3973/#3981 interfaces. |
| R2 | Owner-safe provisioning | A named existing relation/namespace can be adopted, replaced, or changed by a failed admission. | Live tests query rows, OIDs, and control records after every refusal and observe no mutation. |
| R3 | Exact encoding | A value/type outside the sealed PostgreSQL mapping can reach DML. | Live typed row reads show exact persisted values; rejected values leave target/control/ledger unchanged. |
| R4 | Bounded transaction | A write can use multiple transactions or exceed the approved batch bound. | Instrumented/live session tests show one transaction and batches no larger than plan batch size. |
| R5 | Mode correctness | Overwrite can expose an empty/partial relation, and keyed modes can retain tombstones. | Live row/count assertions prove atomic overwrite, append, upsert, dedupe, composite-key handling, and explicit deletes. |
| R6 | Durability/rollback | Unsafe GUCs, errors, or cancellation can leave mutations/receipt. | Live state checks prove refusal/rollback leaves pre-state intact; successful commit creates a durable receipt. |
| R7 | Commit certainty | A disconnect during commit can be retried or reported as confirmed/rolled back. | Deterministic commit fault seam proves unknown outcome, one commit attempt, no retry/rollback/receipt. |
| R8 | Concurrency and fence | Writers can cross owner scope or driver availability changes public capability. | Live concurrent queries prove isolated control/data; capability tests retain `write=false`. |

## Red command

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database -run 'TestPostgres.*(ManagedTarget|DatabaseWrite|WriteSession)'
```

The Red trace must demonstrate that the descriptor-only PostgreSQL driver lacks the required concrete typed behavior before production implementation. It will be stored under `traces/`.

## Green commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres/... ./internal/connectors/database/...
go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=<docker|podman> POLYMETRICS_CONTAINER_ENDPOINT=<direct-unix-socket> go test -tags=databaseintegration -timeout 20m -count=1 -v ./internal/connectors/native/postgres
```

Every live proof reads target rows/counts/OIDs/control records after the operation or refusal. A successful command without an observable assertion is not evidence.
