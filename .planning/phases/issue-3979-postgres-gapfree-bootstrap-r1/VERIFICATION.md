# Verification — #3979 PostgreSQL gap-free bootstrap

## Planned checks

1. `go test -timeout 20m ./internal/connectors/native/postgres/...`
2. `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres`
3. Focused non-test gates: `gofmt -w`, `go vet ./internal/connectors/native/postgres/...`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, `go run ./cmd/connectorgen validate`, and `go run ./cmd/connectorgen surface-sync --check`.
4. Manual inline GSD verification: inspect the changed PostgreSQL coordinator and tests against `CONTEXT.md`, then record code-review findings/disposition in `REVIEW.md`.

## Acceptance checklist

- [ ] Live concurrent before/during/after mutations prove final keyed records exactly match PostgreSQL.
- [ ] Multi-row atomicity and explicit delete tombstone are observed.
- [ ] Barrier/snapshot/receipt/checkpoint/acknowledgement failure paths advance nothing undurable.
- [ ] Invalid slot, retention, system/timeline, publication, and schema drift return actionable rebootstrap failures.
- [ ] Live dbtest observes source rows, bootstrap checkpoint/LSN, and slot teardown.
- [ ] GSD red/green traces and review record are complete.
