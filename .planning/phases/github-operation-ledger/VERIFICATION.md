# Verification

## Passed

- `jq empty .planning/phases/github-operation-ledger/RUN-STATE.json`
- `jq empty internal/connectors/defs/github/api_surface.json`
- `go test ./internal/connectors/engine -run APISurfaceOperationLedger -count=1`
- `go test ./cmd/connectorgen -run 'APISurfaceOperationLedger|GitHubAPISurfaceOperationLedgerMetrics' -count=1`
- `go test ./internal/connectors/conformance -run Static -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
  - `547 connector(s) checked, 0 findings`
- `go test ./cmd/connectorgen -count=1`
- `go test ./internal/connectors/engine -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/github|Static' -count=1`
- `go test ./internal/cli -run GitHubCommandSurface -count=1`
- `go vet ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/conformance ./internal/cli`
- `go build ./cmd/pm`

## Notes

- `go run ./cmd/connectorgen validate internal/connectors/defs/github` was run first and failed
  because the command expects the parent directory of connector bundles, not a connector directory.
  The correct parent-root validation passed.
