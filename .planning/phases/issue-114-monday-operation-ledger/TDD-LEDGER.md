# TDD Ledger — issue #114 Monday operation ledger

Manual GSD programming-loop fallback is active.

| Step | Evidence | Result |
| --- | --- | --- |
| Red | `go test ./cmd/connectorgen -run 'TestMondayOperationLedger' -count=1` fails on legacy Monday `api_surface.json` (`operation_ledger_version = 0, want 1`). | Captured |
| Green | Generated Monday operation-ledger `api_surface.json` and `operations.json`: 367 operations, 87 semantic GET query rows, 280 POST mutation rows; implemented streams remain the only executable coverage. | Captured |
| Refactor | `go test ./cmd/connectorgen -run 'TestMondayOperationLedger' -count=1`, `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayOperationLedger' -count=1`, and `go run ./cmd/connectorgen validate internal/connectors/defs --json` pass (547 connectors, 0 findings, 0 warnings). | Captured |
