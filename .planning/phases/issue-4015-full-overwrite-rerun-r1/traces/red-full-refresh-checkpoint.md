# Red evidence: full-refresh checkpoint reuse

Date: 2026-08-18

No production code had changed when these commands ran.

## Six-mode source boundary

Command:

```text
go test -count=1 -timeout 20m -run '^TestOrchestratorSourceCheckpointFollowsRefreshSemantics$' -v ./internal/synctransport
```

Observed:

```text
--- FAIL: TestOrchestratorSourceCheckpointFollowsRefreshSemantics
    --- FAIL: .../full_append
        source checkpoint = <prior checkpoint>, want nil for full-refresh mode
    --- FAIL: .../full_overwrite
        source checkpoint = <prior checkpoint>, want nil for full-refresh mode
    --- PASS: .../incremental_append
    --- PASS: .../incremental_dedupe
    --- PASS: .../incremental_dedupe_history
    --- PASS: .../incremental_upsert
```

This confirms that the orchestration boundary conflated source refresh semantics with saved-position resumption. It also proves the red test distinguishes the four incremental contracts from the two full-refresh contracts.

## Live binary rerun

Command:

```text
POLYMETRICS_DATABASE_INTEGRATION=1 \
POLYMETRICS_CONTAINER_RUNTIME=docker \
POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
go test -tags=databaseintegration -count=1 -timeout 20m \
  -run '^TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage$' -v ./internal/cli
```

Observed from the first run and independent target query:

```text
records_read=3 records_loaded=3 target_rows=3 target_ids=[1 2 3]
target_sample=id=2 label="page-one-b"
```

The source was then changed independently to three rows: ID 1 deleted, ID 2 renamed to `replacement-two`, and ID 4 inserted. Its queried state was `source_ids=[2 3 4]`.

Observed from the second run and independent target query:

```text
status="completed" records_read=0 records_loaded=0
target_rows=0 target_ids=[]
target_sample=id=2 label="<missing>" removed_id_1_rows=0
```

The firstmate hypothesis about the source checkpoint is confirmed. One symptom detail differs on this PostgreSQL-to-PostgreSQL binary route: the run publishes an empty replacement instead of retaining stale rows. That is still silent data loss and not a valid full overwrite. The shared checkpoint boundary explains both the reported `0/0` and the absence of replacement data; the green contract is an unconditional full source reread for both full-refresh modes.
