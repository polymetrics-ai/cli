# TDD ledger — Issue #3982 PostgreSQL managed-table write driver

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Contract fidelity | Descriptor-only driver does not satisfy provisioning/ledger ports or require a pinned connection. | The red compiler trace precedes `NewDatabaseDriver`, and current compile-time assertions prove both native ports while preserving `write=false`. |
| R2 | Owner-safe observation | A named existing relation/namespace can be adopted, replaced, or changed by a failed admission. | dbtest seeds independent private control rows; exact reassertion returns the live OID while every currently testable refusal re-queries OIDs/control values/counts and observes no driver mutation. |
| R3 | Durable control ledger | A receipt-store port can return a value without persisting it at the asserted relation key. | dbtest reads the empty ledger, stores one opaque delivery identifier, re-reads it, and asserts its row count increased by exactly one. |
| R4 | Durability preflight | A driver can accept `fsync=off` or `synchronous_commit=off`. | dbtest observes an unsafe session GUC, requires preflight refusal, restores the setting, and requires success. |
| R5 | Mapping-dependent DDL and write session | A PostgreSQL-local map can create a placeholder target or accept a record without `MappingContractV1`. | Held for #3973: no DDL that claims a business schema, write-mode admission, batch DML, or tombstone path exists in this slice. |
| R6 | Bounded modes, rollback, and commit certainty | A write can use multiple transactions, mutate after an error, or retry an unknown commit. | Held for #3973's mapping/receipt contract; later tests must use row/count and receipt-ledger assertions, with an explicit commit-boundary fake only where live injection is impossible. |
| R7 | Capability fence | Native driver availability can change public capability before certification. | Existing PostgreSQL metadata assertion observes `write=false` and `cdc=false`. |

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

Every live proof reads target rows/counts/OIDs/control records after the operation or refusal. A successful command without an observable assertion is not evidence. The explicit direct-local Docker dbtest proof passed; `traces/provisioning-green.md` records the observable state assertions and the live permission-path defect it found.
