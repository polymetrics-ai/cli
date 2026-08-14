# TDD ledger — Issue #3982 PostgreSQL managed-table write driver

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Contract fidelity | Descriptor-only driver does not satisfy provisioning/ledger ports or require a pinned connection. | Red compiler trace proves the missing `DatabaseWriteDriver`; compile-time assertions now prove native provisioning, ledger, and write-session ports while preserving public `write=false`. |
| R2 | Owner-safe observation | A named existing relation/namespace can be adopted, replaced, or changed by a failed admission. | dbtest first creates from the shared mapping, exact-reasserts its OID, and re-queries state after ownerless, foreign, collision, replacement, drift and permission refusals. |
| R3 | Durable control ledger | A receipt-store port can return a value without persisting it at the asserted relation key. | dbtest reads the empty ledger, stores one opaque delivery identifier, re-reads it, and asserts its row count increased by exactly one; session receipts then pass through the same ledger authority. |
| R4 | Durability preflight | A driver can accept `fsync=off` or `synchronous_commit=off`. | dbtest observes an unsafe session GUC, requires preflight refusal, restores the setting, and requires success. |
| R5 | Mapping-dependent DDL and write session | A PostgreSQL-local map can create a placeholder target or accept a record without `MappingContractV1`. | #4144's shared mapping is attached to the existing typed provisioning plan; live dbtest requires first-create DDL and write plans to use that same mapping, and observes no namespace/relation/control rows after unsupported mapping/value refusal. |
| R6 | Bounded modes, rollback, and commit certainty | A write can use multiple transactions, mutate after an error, or retry an unknown commit. | Live dbtest reads target rows/counts after five modes, rollback/cancellation and explicit tombstones; the session returns a shared `DeliveryReceiptV1` only after confirmed commit. An injected commit-boundary fake proves unknown cannot be retried where a real network disconnect cannot be deterministically timed. |
| R7 | Capability fence | Native driver availability can change public capability before certification. | PostgreSQL `database.json` admits exactly the five internal managed modes while existing metadata and `Connector.Write` assertions observe `write=false` and `ErrUnsupportedOperation`. |

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
