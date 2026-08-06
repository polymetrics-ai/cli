# Verification checklist — provider-cited rate-limit declarations R1

## Planned commands

- `jq empty internal/connectors/defs/*/rate_limits.json`
- `go test ./internal/connectors/engine -run TestEveryProductionRateLimitDeclarationIsEmbedded`
- `go test ./internal/connectors/engine`
- `go build ./cmd/pm`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`
- `make tidy-check`
- `make lint`
- `make docs-check`
- `make smoke-no-build`
- `make agent-contract-check`
- `make connector-boundary`
- `make release-workflow-check`

## Results

Pending implementation.
