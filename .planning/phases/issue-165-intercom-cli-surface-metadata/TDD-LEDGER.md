# TDD Ledger: Intercom CLI Surface Metadata

## Planned Red Test

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1`
  - Expected initial failure after adding the test: current `internal/connectors/defs/intercom/api_surface.json` has 10 entries and no `operation_ledger_version`; target is 149 official operations with method split GET 67, PUT 16, POST 47, DELETE 19.

## Green Tests

Pending implementation.

## Refactor Notes

Pending.
