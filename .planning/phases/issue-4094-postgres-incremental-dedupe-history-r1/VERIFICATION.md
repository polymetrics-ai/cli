# Verification checklist — Issue 4094

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| PostgreSQL → PostgreSQL history state | passed focused live | Query v1/v2 key, exact close boundary, current state, and the retained soft-deleted v2 row. |
| Keyed versions and deterministic late/replay behavior | passed focused live | A newly constructed executor replays v1/v2; the query is unchanged and the replay receipt is durably readable. |
| Restart recovery | passed focused live | Close the first connection, construct a fresh driver/ledger/executor, and query/replay the durable history before close. |
| PostgreSQL-only admission | passed fake | Three non-PostgreSQL route cells return their typed reason while fake begin/batch/commit/rollback/ledger/mutation counters remain zero. |
| Required live harness | passed full live | The explicit Colima Docker command passed across the complete PostgreSQL integration package. |

## Completed local gates

- `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/...`
- `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres` after CI exposed the stale five-mode expectation
- tagged focused and full PostgreSQL dbtest commands recorded under `traces/`
- `go vet ./...`
- `make tidy-check`, `make build`, `make docs-check-no-build`, `make smoke-no-build`, and `make lint`
- `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connectorgen-certification-matrix`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check`

## Required commands

- `go test -timeout 20m ./internal/connectors/native/postgres/... ./internal/connectors/database/...`
- `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
- Relevant non-test gates from `make verify`, run individually under the
  per-command timeout policy.
