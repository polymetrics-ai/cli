# Verification checklist — Issue 4094

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| PostgreSQL → PostgreSQL history state | pending live | Query each history row's key, validity window, current/closed state, and soft-delete outcome. |
| Keyed versions and deterministic late/replay behavior | pending unit + live | Assert exact queried row set and stable receipt/ledger identity after replay. |
| Restart recovery | pending unit + live | Recreate the controller and assert the durable receipt/ledger and target rows are unchanged. |
| PostgreSQL-only admission | pending fake | Assert each other route's typed reason plus zero provider/driver/session/ledger/database counters. |
| Required live harness | pending live | Run the explicit Colima Docker command and record its exact passed output. |

## Required commands

- `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/...`
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
- Relevant non-test gates from `make verify`, run individually under the
  per-command timeout policy.
