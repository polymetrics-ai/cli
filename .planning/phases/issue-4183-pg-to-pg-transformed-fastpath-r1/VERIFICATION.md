# Verification — issue #4183

## Planned local gates

- `go test -count=1 -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/warehouse`
- `go test -count=1 -timeout 20m ./internal/app ./internal/cli`
- `go vet ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/warehouse ./internal/app ./internal/cli`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`
- With explicit opt-in and a direct local Podman endpoint only: `go test -tags=databaseintegration -count=1 -timeout 20m` for the named live PostgreSQL test.
- External qualified host only: `TestPostgresToPostgresTransformed3GBCLI` with both database and performance opt-ins.

## Acceptance record

Pending implementation. The final record will state Red and Green command output, the binary versus component coverage distinction, output measurements and their logical-byte definition, the hardware/opt-in status of the 3 GB gate, and the code-review disposition.

