# TDD ledger — #3859 native database apply strategies

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Closed registered apply boundary | An unregistered or descriptor-incompatible apply can receive a page. | Refusal occurs before target/session/ledger counters change. |
| R2 | Bounded page execution | Oversized record/byte input can reach a native batch. | Each observed fake/native batch remains within descriptor bounds; over-limit input has zero writes. |
| R3 | Durable acknowledgement | A partial batch, failed rollback, unknown commit, or absent receipt can acknowledge a source page. | Only confirmed commit plus durable ledger record produces acknowledgement; all other paths have none. |
| R4 | Source-order fence | A replayed/older source tuple can overwrite the retained newer keyed value. | The PostgreSQL query returns the newer row after the late old event. |
| R5 | Explicit deletion | An omitted record deletes or closes a target row. | The absent row remains; only a typed tombstone changes it. |
| R6 | Atomic overwrite | A failed replacement can leave partial replacement rows. | Live re-read exactly equals pre-write state after failure/cancellation. |
| R7 | History correctness | Update/delete physically removes history or leaves two current rows. | Re-read history shows one closed prior interval, one current successor, then a tombstone-closed final interval. |

## Red commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run 'TestPostgres.*Apply' -v ./internal/connectors/native/postgres
```

The actual failing command output will be retained under `traces/` before any
production implementation for the added contracts.

## Green commands

Use the scoped and live commands in `PLAN.md`; record the exact outcome here
and in `VERIFICATION.md` after implementation.
