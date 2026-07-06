# Verification

## Commands

- `go test ./internal/connectors/bundleregistry -run CLISurface`
- `pnpm --filter cli-polymetrics-ai test -- connector-data`
- `go test ./internal/connectors/engine -run CLISurface`
- `go test ./cmd/connectorgen -run CLISurface`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors -run TestEveryRegisteredConnectorHasGuideManualAndSkill`
- `pnpm --filter cli-polymetrics-ai build`
- `git diff --check`

## Result

Passed.

Note: `go run ./cmd/connectorgen validate internal/connectors/defs/github` was attempted first and
failed because this validator expects a defs root containing connector directories; targeting the
GitHub directory directly makes it inspect `schemas/` and `fixtures/` as connector directories.
The correct defs-root validation passed with 547 connectors and 0 findings.
