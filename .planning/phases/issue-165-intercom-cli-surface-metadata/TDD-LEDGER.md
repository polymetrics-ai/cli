# TDD Ledger: Intercom CLI Surface Metadata

## Red Test

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1`
  - Initial result after adding `cmd/connectorgen/intercom_api_surface_test.go` and before refreshing production metadata: failed.
  - Failure: `operation_ledger_version = 0, want 1` because current `internal/connectors/defs/intercom/api_surface.json` had 10 legacy entries and no operation ledger mode.

## Green Tests

- `gofmt -w cmd/connectorgen/intercom_api_surface_test.go`
  - Passed.
- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1`
  - Passed after refreshing Intercom `api_surface.json` to 149 official operations.
- `jq empty internal/connectors/defs/intercom/api_surface.json internal/connectors/defs/intercom/cli_surface.json .planning/phases/issue-165-intercom-cli-surface-metadata/RUN-STATE.json`
  - Passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs`
  - Passed: 547 connector(s) checked, 0 findings.
- `go test ./internal/connectors/engine -run 'Intercom|CLISurface' -count=1`
  - Passed.
- `go test ./cmd/connectorgen ./internal/connectors/engine`
  - Passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1`
  - Passed.
- `go test ./cmd/connectorgen -run CLISurface -count=1`
  - Passed.
- `go test ./internal/connectors/certify -run TestWriteCreateFailureRecordsNoLeak -count=1 -timeout=20m`
  - Passed after the default full-suite `go test ./...` hit the package timeout in this existing long-running certify test.
- `go test ./... -timeout=20m`
  - Passed.
- `go build ./cmd/pm`
  - Passed.
- `make verify`
  - Passed.

## Refactor Notes

- `api_surface.json` uses `operation_ledger_version: 1`; all non-executable rows are blocked-by-default `operation` metadata instead of legacy `excluded` rows.
- Existing executable coverage remains the five legacy streams only: `admins`, `contacts`, `companies`, `conversations`, and `tags`.
- `cli_surface.json` is metadata only. Planned direct-read, binary/export, and reverse-ETL commands intentionally avoid executable mappings until follow-up issues implement schemas/policies.
