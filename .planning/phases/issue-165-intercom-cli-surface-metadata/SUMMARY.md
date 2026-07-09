# Summary: Intercom CLI Surface Metadata

Status: provisionally integrated into the parent branch; CodeRabbit coverage pending through parent PR fallback.

## Completed

- Loaded official Intercom OpenAPI 2.14 source and confirmed baseline: 149 operations across 105 paths; GET 67, PUT 16, POST 47, DELETE 19.
- Created GSD plan, TDD ledger, and verification checklist before production edits.
- Recorded manual GSD fallback because `scripts/gsd` does not expose `programming-loop` in this checkout.
- Added `cmd/connectorgen/intercom_api_surface_test.go` as a data-contract test for the official Intercom operation ledger metrics.
- Refreshed `internal/connectors/defs/intercom/api_surface.json` into operation-ledger mode with all 149 operations enumerated.
- Added `internal/connectors/defs/intercom/cli_surface.json` with implemented five-stream commands and planned safe app intents for direct reads, binary/export, and typed reverse ETL.
- Updated `metadata.json` and `docs.md` to document full-surface metadata without claiming executable writes.

## Focused Verification

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1` passed.
- `jq empty ...` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 547 connector(s), 0 findings.
- `go test ./internal/connectors/engine -run 'Intercom|CLISurface' -count=1` passed.
- `go test ./cmd/connectorgen ./internal/connectors/engine` passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1` passed.
- `go test ./cmd/connectorgen -run CLISurface -count=1` passed.
- `go test ./...` with default timeout hit `internal/connectors/certify` package timeout; focused retry with `-timeout=20m` passed.
- `go test ./... -timeout=20m` passed.
- `go build ./cmd/pm` passed.
- `make verify` passed.

## Next

- Stacked #165 PR opened against the parent branch: https://github.com/polymetrics-ai/cli/pull/234.
- CodeRabbit auto-review skipped PR #234 because reviews are disabled on non-default base branches; this is not review completion and requires parent PR #220 fallback coverage after integration.
- PR #234 CI passed and was squash-merged into the parent branch at `fded1e72`.
- Parent PR #220 CodeRabbit skipped because the PR is draft; #165 remains pending review coverage.
- Leave #168-#171 to refine operation classifications and implement streams/direct reads/binary/write policies.
