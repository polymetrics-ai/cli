# Verification — #3977 pgoutput v2 CDC

## Pending execution

- [ ] Red focused test has failed against the planned reader for the intended reason.
- [ ] Focused PostgreSQL and transaction-stage unit suites pass.
- [ ] `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres` passes without a skip.
- [ ] `go vet` and `go build ./cmd/pm` pass.
- [ ] `connectorgen validate`, `connectorgen surface-sync --check`, docs, lint, connector boundary, release-workflow, and agent-contract gates pass individually.
- [ ] Code review findings are recorded and dispositioned.
