# Verification — Issue 3985 connector canon

## Planned checks

- `make connector-runtime-preflight`
- `make connector-canon-check`
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'`
- `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/commandrunner`
- `make tidy-check`
- `make lint`
- `make docs-check`
- `make agent-contract-check`
- `make connectorgen-validate`
- `make connectorgen-surface-sync`
- `make connector-boundary`
- `make release-workflow-check`
- `git diff --check`

## Full-suite note

Per repository guidance, this task runner will not run `go test ./...` or `make verify` as a
single local command because the suite commonly exceeds command timeouts. CI/no-mistakes owns the
aggregate run. Each required non-suite verification gate is run separately and reported here.
