# PostgreSQL provisioning partial green evidence

## Green commands

```text
go test -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database
ok   polymetrics.ai/internal/connectors/native/postgres
ok   polymetrics.ai/internal/connectors/database

POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres
ok   polymetrics.ai/internal/connectors/native/postgres
```

The first command verifies the production PostgreSQL driver implements the
managed-target provisioning and ledger ports while preserving the existing
`write=false` capability test. The tagged command passed through dbtest's
explicit direct-local Docker endpoint; its managed-target control test passed
in 11.17 seconds and the complete tagged package passed in 22.719 seconds.

## Observable live evidence

The dbtest case independently seeds private control state, then proves the
driver observes and reasserts the exact live OID; its ledger write increases
the private ledger count by one and reads back the same delivery identifier.
It changes a session durability setting and proves refusal before restoring it.
For foreign/tampered owner, collision, permission denial, relation OID
replacement, and schema fingerprint drift, it captures namespace/relation
OIDs plus owner/control/ledger values before the driver call and requires an
exactly unchanged snapshot afterward. The permission proof found and fixed a
real delayed-row-error path: PostgreSQL reports its table permission failure
while iterating rows, so the driver now classifies it as unreadable rather than
generic/unverifiable.

The mapping-dependent target create and all record writes remain intentionally
unimplemented until #3973's shared mapping contract lands.
