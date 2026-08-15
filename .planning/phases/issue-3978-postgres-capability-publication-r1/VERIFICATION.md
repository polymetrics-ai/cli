# Verification — Issues 3978 and 3977

## Behavior evidence

- Red: `traces/red-app-dispatch.txt` records production `App.RunETL(change_capture)` refusing the
  already-implemented PostgreSQL changefeed before warehouse dispatch.
- Green: `TestRunETLChangeCapturePublishesCommittedTransactionToConnectionWarehouse` requires two
  exact IDs in the connection-owned Parquet table and the full `lsn-1` checkpoint. A no-op cannot
  satisfy either assertion.
- Green: `TestRunETLChangeCaptureRestoresWarehouseReceiptBeforeCheckpointAfterRestart` simulates
  process loss after durable warehouse receipt, then proves restart restores the receipt, persists
  the checkpoint, and does not receive the transaction twice.
- Green: PostgreSQL transaction tests require warehouse receipt before app checkpoint before source
  LSN acknowledgement.
- Green after #4156 rebase: both the generic declared-transport `change_capture` test and the native
  PostgreSQL warehouse fallback test pass, proving the shared route retains precedence.

## Completed commands

```text
go test -timeout 20m ./internal/app -count=1
PASS (183.146s)

go test -timeout 20m ./internal/connectors ./internal/connectors/database ./internal/connectors/native/postgres -count=1
PASS

go test -timeout 20m ./internal/cli -count=1
PASS (314.840s)

go run ./cmd/connectorgen validate internal/connectors/defs
PASS: 552 connector bundles validated, 0 errors

go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
PASS

(cd website && npm run gen:website-data)
PASS: 552 connectors; write=241, query=0, cdc=1
```

Live evidence after rebasing onto #4156:

```text
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres \
  -run 'TestPostgres(PGOutputV2ContainerHarness|ManagedTargetWorksetDeliveryLive|ManagedTargetIncrementalDedupeHistoryLive)$' \
  -count=1 -v

PASS TestPostgresPGOutputV2ContainerHarness (10.22s)
PASS TestPostgresManagedTargetWorksetDeliveryLive (5.84s)
PASS TestPostgresManagedTargetIncrementalDedupeHistoryLive (5.93s)
```

The CDC live test observes committed replication events, a persisted checkpoint, and acknowledged
LSN behavior. The managed-target live tests assert actual PostgreSQL row/workset state and durable
delivery receipts. #4158's known base-only live control assertion was excluded as required.

## Remaining final gates

All final gates passed:

```text
gofmt (changed Go files); git diff --check
PASS

go test -race -timeout 20m ./internal/app -run 'TestRunETL(ChangeCapture|TransportDispatchesDeclaredChangeCapture)' -count=1
PASS

go test -race -timeout 20m ./internal/connectors/native/postgres -run 'Test(PGOutputV2StreamCommitUsesDurableTransactionReceiverBeforeCheckpoint|PostgresDeclaresProviderHTTPRateLimitsNotApplicable|NameAndMetadata)$' -count=1
PASS

go vet ./...
PASS

make tidy-check; make build
PASS

make -j2 lint docs-check-no-build smoke-no-build agent-contract-check connectorgen-validate \
  connectorgen-surface-sync connector-runtime-preflight connector-canon-check connector-boundary \
  release-workflow-check
PASS; boundary outcome=clean, checked_files=261, connectors_loaded=552
```

The built CLI returned `write=true`, `cdc=true`, `query=false` from
`pm connectors inspect postgres --json`; capability-filtered catalog queries included PostgreSQL
for write and CDC and excluded it for query. `pm help connectors`, bare `pm connectors`, and
`pm connectors --help` all exited successfully. The command/help surface did not move, so golden
transcripts were not regenerated.

## Inline code review

Review of the whole base-to-head diff found one actionable durability issue: initial PostgreSQL
delivery derived its receiver receipt from the staged content digest, while crash restoration
supplied the raw transaction-stage lookup ID. The disposition was accepted. Both boundaries now use
the stage's opaque `TransactionKey`, and the unit test requires the restored ID to equal the initial
receiver ID while differing from the raw lookup value.

Post-fix verification:

```text
go test -race -timeout 20m ./internal/connectors/native/postgres \
  -run '^TestPGOutputV2StreamCommitUsesDurableTransactionReceiverBeforeCheckpoint$' -count=1
PASS

go test -timeout 20m ./internal/connectors/native/postgres -count=1
PASS

go test -timeout 20m ./internal/app \
  -run 'TestRunETL(ChangeCapture|TransportDispatchesDeclaredChangeCapture)' -count=1
PASS

POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
go test -timeout 20m -tags=databaseintegration ./internal/connectors/native/postgres \
  -run '^TestPostgresPGOutputV2ContainerHarness$' -count=1 -v
PASS (11.15s)
```

No shared transport descriptor declaration or validation was changed. #4125 and #4158 remain out
of scope. The PR targets a non-default integration base, so review coverage is recorded as
`primary_route=claude_auto`, `coverage_route=sub_pr`, `fallback_route=parent_pr_fallback`,
`review_status=pending` until the PR workflow posts its review record.
