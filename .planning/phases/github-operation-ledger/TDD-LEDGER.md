# TDD Ledger

## Red

- `gofmt -w internal/connectors/engine/bundle_test.go cmd/connectorgen/main_test.go cmd/connectorgen/github_api_surface_test.go`
- `go test ./internal/connectors/engine -run APISurfaceOperationLedger -count=1`
  - Fails to build because `APISurface.OperationLedgerVersion` and `SurfaceEndpoint.Operation`
    do not exist yet.
- `go test ./cmd/connectorgen -run 'APISurfaceOperationLedger|GitHubAPISurfaceOperationLedgerMetrics' -count=1`
  - Fails to build because `ruleSurfaceOperation` does not exist yet.

## Green

- Pending.

## Refactor

- Pending.
