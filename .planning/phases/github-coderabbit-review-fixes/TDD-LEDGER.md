# TDD Ledger

## Review-Fix Slice

This slice fixes review findings against existing behavior. Red evidence came from local validation:

- `go test ./internal/cli -run 'TestGitHubCommandWriteUsesReversePlanApproval|TestConnectorCommand'`
  initially failed after adding unsupported `allOf` schema constraints because the embedded
  connector meta-schema compiler reported `unknown keyword "allOf"`.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` failed for the same
  unsupported schema keyword.

Resolution:

- Reverted the unsupported schema-keyword change and kept operation-ledger constraints enforced by
  `cmd/connectorgen` Go validation.
- Implemented the remaining still-valid review fixes.

Green evidence:

- `go test ./internal/cli -run 'TestGitHubCommandWriteUsesReversePlanApproval|TestConnectorCommand'`
- `jq empty internal/connectors/defs/github/operations.json website/package.json`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `node website/scripts/gen-connector-bundles.mjs && node website/scripts/gen-connector-catalog.mjs`
