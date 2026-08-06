# Verification checklist — provider-cited rate-limit declarations R1

## Planned commands

- `jq empty internal/connectors/defs/*/rate_limits.json`
- `go test ./internal/connectors/engine -run TestProductionDefinitionsEmbedEveryRateLimitDeclaration`
- `go test ./internal/connectors/engine`
- `go build ./cmd/pm`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`

## Recorded evidence

- 2026-08-06: `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` failed as expected when
  `harvest/rate_limits.json` existed before the `defs.FS` wildcard. It passed after the wildcard
  was added.
- `make tidy-check`
- `make lint`
- `make docs-check`
- `make smoke-no-build`
- `make agent-contract-check`
- `make connector-boundary`
- `make release-workflow-check`

## Results

Pending implementation.
