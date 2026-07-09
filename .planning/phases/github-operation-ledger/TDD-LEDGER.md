# TDD Ledger

## Red

- `gofmt -w internal/connectors/engine/bundle_test.go cmd/connectorgen/main_test.go cmd/connectorgen/github_api_surface_test.go`
- `go test ./internal/connectors/engine -run APISurfaceOperationLedger -count=1`
  - Fails to build because `APISurface.OperationLedgerVersion` and `SurfaceEndpoint.Operation`
    do not exist yet.
- `go test ./cmd/connectorgen -run 'APISurfaceOperationLedger|GitHubAPISurfaceOperationLedgerMetrics' -count=1`
  - Fails to build because `ruleSurfaceOperation` does not exist yet.

## Green

- `gofmt -w internal/connectors/engine/bundle.go cmd/connectorgen/validate.go internal/connectors/conformance/static.go`
- `go test ./internal/connectors/engine -run APISurfaceOperationLedger -count=1`
  - Passes.
- `go test ./cmd/connectorgen -run 'APISurfaceOperationLedger|GitHubAPISurfaceOperationLedgerMetrics' -count=1`
  - Passes.
- `go test ./internal/connectors/conformance -run Static -count=1`
  - Passes.
- `go run ./cmd/connectorgen validate internal/connectors/defs`
  - Passes with `547 connector(s) checked, 0 findings`.
- `go test ./cmd/connectorgen -count=1`
  - Passes.
- `go test ./internal/connectors/engine -count=1`
  - Passes.
- `go test ./internal/connectors/conformance -run 'TestConformance/github|Static' -count=1`
  - Passes.
- `go test ./internal/cli -run GitHubCommandSurface -count=1`
  - Passes.

## Refactor

- No dependency or runtime-dispatch changes.
- Operation-ledger rows remain validation metadata only; executable command routing still requires
  existing stream/write mappings.
