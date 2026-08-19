# Green evidence: source refresh semantics restored

Date: 2026-08-18

## Shared six-mode boundary

Command:

```text
go test -count=1 -timeout 20m -run '^TestOrchestratorSourceCheckpointFollowsRefreshSemantics$' -v ./internal/synctransport
```

Result: PASS. `full_append` and `full_overwrite` received no prior source checkpoint. `incremental_append`, `incremental_dedupe`, `incremental_dedupe_history`, and `incremental_upsert` each received the prior checkpoint unchanged.

The full `./internal/synctransport` package also passed. The shared selector is used by the generic path, the run-scoped full-overwrite path, the serial Arrow controller, and the ordered Arrow producer.

## Live full-overwrite replacement

Command:

```text
POLYMETRICS_DATABASE_INTEGRATION=1 \
POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
go test -tags=databaseintegration -count=1 -timeout 20m \
  -run '^TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage$' -v ./internal/cli
```

Result: PASS.

```text
first:  records_read=3 records_loaded=3 target_rows=3 target_ids=[1 2 3]
        target_sample=id=2 label="page-one-b"
source: source_rows=3 source_ids=[2 3 4]
        source_sample=id=2 label="replacement-two"
rerun:  records_read=3 records_loaded=3 target_rows=3 target_ids=[2 3 4]
        target_sample=id=2 label="replacement-two" removed_id_1_rows=0
```

The target facts came from independent PostgreSQL queries after each run. They prove replacement rather than command success alone.

## Live incremental no-regression

Command:

```text
POLYMETRICS_DATABASE_INTEGRATION=1 \
POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
go test -tags=databaseintegration -count=1 -timeout 20m \
  -run '^TestPMBinaryPostgresIncrementalUpsertStillSkipsUnchangedSource$' -v ./internal/cli
```

Result: PASS.

```text
first_records_read=3 first_records_loaded=3
rerun_records_read=0 rerun_records_loaded=0
target_rows_before=3 target_rows_after=3
target_sample_before=id=2 label="event-two"
target_sample_after=id=2 label="event-two"
```

This independently re-proves that an unchanged incremental source keeps its durable skip behavior.
