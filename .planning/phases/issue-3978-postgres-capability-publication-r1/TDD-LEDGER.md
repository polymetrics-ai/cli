# TDD ledger — Issues 3978 and 3977

| ID | Red guarantee | Green implementation | Status |
| --- | --- | --- | --- |
| PGCAP-1 | `App.RunETL(change_capture)` advertises PostgreSQL CDC but returns a typed no-dispatch refusal. | A matching implemented changefeed dispatches into the connection-owned warehouse. | Red: `traces/red-app-dispatch.txt` |
| PGCAP-2 | A per-event callback can be mistaken for a durable whole-transaction sink. | A streaming committed-transaction receiver returns the warehouse receipt only after atomic WAL publication, Parquet materialization, and directory sync. | Planned |
| PGCAP-3 | Source LSN could advance without application checkpoint persistence. | The app consumes the warehouse acknowledgement to persist the full checkpoint before the native reader sends standby status. | Planned |
| PGCAP-4 | Capability publication can drift from production behavior. | Exact CLI/bundle/native tests require `write=true`, `cdc=true`, `query=false` and observable write/CDC behavior. | Planned |
| PGCAP-5 | Generated docs/catalog/website artifacts can retain old flags. | Repository generators produce a clean parity diff and checks pass. | Planned |
| PGCAP-6 | Unit-only claims can hide live no-ops. | PostgreSQL dbtest observes CDC rows/checkpoint/LSN and managed-target row/receipt changes. | Planned |

## Planned commands

```sh
go test -timeout 20m ./internal/app -run 'TestRunETL.*ChangeCapture' -count=1
go test -timeout 20m ./internal/connectors ./internal/connectors/native/postgres ./internal/cli -count=1
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
```
