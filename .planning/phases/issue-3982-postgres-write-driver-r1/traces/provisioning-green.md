# PostgreSQL provisioning partial green evidence

## Green commands

```text
go test -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database
ok   polymetrics.ai/internal/connectors/native/postgres
ok   polymetrics.ai/internal/connectors/database

go test -timeout 20m -count=1 -tags=databaseintegration ./internal/connectors/native/postgres -run TestPostgresManagedTargetDriverLiveControlAssertions
ok   polymetrics.ai/internal/connectors/native/postgres
```

The first command verifies the production PostgreSQL driver implements the
managed-target provisioning and ledger ports while preserving the existing
`write=false` capability test. The tagged command compiles the dbtest-backed
live proof.

## Live-test availability

The opt-in test correctly skipped because this worktree has no explicit
`POLYMETRICS_DATABASE_INTEGRATION=1` and direct local container endpoint. It
did not inspect or use a global/named container runtime connection. When the
documented explicit endpoint is supplied, the test independently seeds the
private control layout, then asserts live PostgreSQL OIDs/rows for exact
reassertion, durable ledger persistence, foreign/colliding control refusal,
permission denial, OID replacement, schema drift, and durability settings.

The mapping-dependent target create and all record writes remain intentionally
unimplemented until #3973's shared mapping contract lands.
