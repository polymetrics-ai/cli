# Verification — PostgreSQL transport surface truthfulness

## Checklist

- [x] Every declared PostgreSQL destination mode preflights through `app.Open` production composition.
- [x] `incremental_dedupe_history` refuses as an undeclared source mode before executor I/O.
- [x] Source and destination declaration sets exactly match and retain `full_overwrite`.
- [x] Scoped PostgreSQL certification generation is run; `certification-matrix --check` is current.
- [x] `go run ./cmd/pm connectors inspect postgres --json` shows the same truthful source/destination mode sets.
- [x] Focused tests, build, and applicable repository gates pass.
- [x] Inline verify-work and code review have no unresolved findings.

## Executed evidence

```text
go test -count=1 -timeout 20m ./internal/app -run '^TestOpen(PreflightsEveryDeclaredPostgresDestinationMode|RefusesPostgresUnpairedHistoryModeBeforeExecutorIO|PostgresTransportDeclarationsAreExactModeIntersection)$'
go test -count=1 -timeout 20m ./internal/connectors/native/postgres ./internal/synctransport ./internal/cli ./cmd/connectorgen
go test -race -count=1 -timeout 20m ./internal/app -run '^TestOpen(PreflightsEveryDeclaredPostgresDestinationMode|RefusesPostgresUnpairedHistoryModeBeforeExecutorIO|PostgresTransportDeclarationsAreExactModeIntersection)$'
go vet ./internal/app ./internal/connectors/native/postgres ./internal/synctransport ./internal/cli ./cmd/connectorgen
go build ./cmd/pm
go run ./cmd/connectorgen certification-matrix --connector postgres
go run ./cmd/connectorgen certification-matrix --check
./pm help connectors
./pm connectors
./pm connectors inspect postgres --json
./pm docs generate --dir docs/cli --connectors-dir docs/connectors
./pm docs validate --connectors-dir docs/connectors
make tidy-check && make lint && make docs-check && make smoke-no-build && make agent-contract-check
make connectorgen-validate && make connectorgen-surface-sync && make connector-boundary && make release-workflow-check && make connector-canon-check && make github-parity-artifacts-check
```

The first `make docs-check` correctly failed while the generated connector catalog still carried the removed mode. `pm docs generate --dir docs/cli --connectors-dir docs/connectors` updated only the PostgreSQL catalog entry; the rerun passed.
