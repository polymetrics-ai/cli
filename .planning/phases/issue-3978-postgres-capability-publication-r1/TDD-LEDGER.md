# TDD ledger — Issues 3978 and 3977

| ID | Red guarantee | Green implementation | Status |
| --- | --- | --- | --- |
| PGCAP-1 | `App.RunETL(change_capture)` advertises PostgreSQL CDC but returns a typed no-dispatch refusal. | A matching implemented changefeed dispatches into the connection-owned warehouse. | Green; Red: `traces/red-app-dispatch.txt`; Green: exact Parquet IDs plus checkpoint in `TestRunETLChangeCapturePublishesCommittedTransactionToConnectionWarehouse` |
| PGCAP-2 | A per-event callback can be mistaken for a durable whole-transaction sink. | A streaming committed-transaction receiver returns the warehouse receipt only after atomic WAL publication, Parquet materialization, and directory sync. | Green; app transaction publication and PostgreSQL ordering tests pass |
| PGCAP-3 | Source LSN could advance without application checkpoint persistence. | The app consumes the warehouse acknowledgement to persist the full checkpoint before the native reader sends standby status. | Green; receipt/checkpoint/ack ordering and restart restoration tests pass |
| PGCAP-4 | Capability publication can drift from production behavior. | Exact CLI/bundle/native tests require `write=true`, `cdc=true`, `query=false` and observable write/CDC behavior. | Green; exact projection plus behavior suites pass |
| PGCAP-5 | Generated docs/catalog/website artifacts can retain old flags. | Repository generators produce a clean parity diff and checks pass. | Green; connector docs/catalog and website generators rerun |
| PGCAP-6 | Unit-only claims can hide live no-ops. | PostgreSQL dbtest observes CDC rows/checkpoint/LSN and managed-target row/receipt changes. | Green; three selected live tests passed after #4156 rebase |
| PGCAP-7 | PostgreSQL had no explicit rate-limit artifact, so the required no-provider-HTTP state was absent. | `rate_limits.json` declares `not_applicable` with an exact native-wire-protocol reason, covered through the real bundle loader. | Green; red observed `bundle.RateLimits == nil`, focused test and bundle validation pass |
| PGCAP-8 | A recovered stage receipt could use the raw stage lookup ID while its warehouse receipt was derived from the stage's opaque transaction key, making restart reject its own receipt. | Both initial delivery and receipt restoration use `CommittedTransaction.TransactionKey`; the test requires exact initial/restored identity equality and rejects the raw lookup ID. | Green; code-review assertion and race test pass; live CDC rerun passes |
| PGCAP-9 | A fresh `pm etl run` selected PostgreSQL's snapshot-only generic transport and failed with `source transport does not support sync mode "change_capture"` before creating a replication slot. | Exact implemented changefeeds dispatch source-only CDC to the local warehouse before descriptor-presence routing; sources without that exact executor retain #4156 generic preflight. | Green; Red: `traces/red-binary-dispatch.txt`; fresh-binary live test observes row/checkpoint/receipt/LSN |
| PGCAP-10 | Once binary dispatch reached `ReadCDC`, PostgreSQL created the checkpoint observation after the warehouse receipt, so the durability contract refused the inverted timestamp order. | The commit candidate is observed before downstream delivery; receipt recovery uses the persisted receipt time as a conservative recovered observation bound. | Green; Red: `traces/red-binary-dispatch.txt`; unit race and fresh-binary live test pass |

## Planned commands

```sh
go test -timeout 20m ./internal/app -run 'TestRunETL.*ChangeCapture' -count=1
go test -timeout 20m ./internal/connectors ./internal/connectors/native/postgres ./internal/cli -count=1
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
```

The live post-rebase selection was deliberately scoped around the two published true flags and
excluded the known base-only #4158 test:

```sh
go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres \
  -run 'TestPostgres(PGOutputV2ContainerHarness|ManagedTargetWorksetDeliveryLive|ManagedTargetIncrementalDedupeHistoryLive)$' -count=1 -v
```
