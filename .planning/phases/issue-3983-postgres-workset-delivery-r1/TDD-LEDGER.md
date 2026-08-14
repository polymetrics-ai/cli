# TDD ledger — Issue 3983

Manual inline GSD execution. Red and green command output is retained under
`traces/` before and after production changes.

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Immutable delivery binding | A workset can be delivered with a plan/approval created for another workset, owner/OID, schema, or key order. | The plan hash changes for a new OID, a changed destination is refused, and a changed-workset approval is refused before the fake session opens or writes. |
| R2 | Explicit effects only | A physically absent source row can become a delete during consumption. | Sealed delta contains no absence-derived delete; live target row remains until a supplied tombstone deletes it. |
| R3 | Receipt-before-baseline | A failed, missing, or unknown receipt advances the destination baseline. | The prior fake baseline identity/receipt remain unchanged on each failure; confirmed receipt/ledger promotes exactly one durable candidate. |
| R4 | Replay correctness | A replay can silently choose a different workset or retry unknown commit. | Replay record retains the immutable workset identity; unknown commit returns replay-required without baseline mutation. |
| R5 | Destination isolation | One owner/control can write or advance another destination's ledger/baseline. | Concurrent independently owned destinations retain distinct baseline entries; a changed workset destination is rejected while sealing the plan. |
| R6 | Real PostgreSQL outcome | Controller success can be a no-op or destructive absence deletion. | dbtest queries exact managed target rows and its durable delivery receipt after unchanged/insert/update/tombstone cases. |

## Planned red commands

```sh
go test -timeout 20m ./internal/connectors/database -run 'TestChangeDeliveryWorkset.*(Plan|Approval|Baseline|Replay|Tombstone)' -count=1
go test -timeout 20m ./internal/connectors/native/postgres -run 'TestPostgres.*WorksetDelivery' -count=1
```

## Planned green commands

```sh
go test -timeout 20m ./internal/connectors/database/... ./internal/connectors/native/postgres/... ./internal/warehouse/... -count=1
go test -race -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test.*WorksetDelivery' -count=1
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
```
