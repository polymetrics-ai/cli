# Live PostgreSQL evidence for #3982

Command:

    POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run TestPostgresManagedTargetDriverLiveControlAssertions -v ./internal/connectors/native/postgres

Observed output:

    === RUN   TestPostgresManagedTargetDriverLiveControlAssertions
    --- PASS: TestPostgresManagedTargetDriverLiveControlAssertions (6.98s)
    PASS
    ok      polymetrics.ai/internal/connectors/native/postgres    7.744s

The tagged live test reads PostgreSQL state after every operation and refusal:
the mapped target relation, namespace/relation OIDs, private owner/control and
per-session delivery-ledger rows, all five phase-one modes, composite-key
tombstones, rollback retention, and the shared unknown-commit outcome. It
verifies an ordinary batch omission leaves the existing row and an explicit
tombstone deletes it.
