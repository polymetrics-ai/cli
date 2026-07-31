# TDD ledger — Mixpanel parity (#3158)

## Red / expectations before production edits

- Current Mixpanel bundle had 10 read streams, no writes, no operations/CLI surface, and only a partial `api_surface.json`.
- Required target was an exact 105-row official operation ledger with fixture-backed executable coverage and truthful blocked/planned rows.

## Green ledger

Credential-free, provider-free gates run locally:

- Official Mixpanel OpenAPI inventory comparison: 105 official operations, 105 local operations, 105 `api_surface.json` rows, missing 0, extra 0, duplicate IDs 0.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, 549 connectors, 0 findings.
- Temp defs-root Mixpanel-only `connectorgen validate`: pass, 1 connector, 0 findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/mixpanel|TestMixpanelOperationDirectReadsReplay' -count=1`: pass.
- Focused CLI/golden tests: `go test ./internal/cli -run 'Test(CobraRouterShellPreservesDynamicConnector|ConnectorInspect|ConnectorCatalog|DocsGenerateIncludesConnectorCatalog|Golden)' -count=1`: pass.
- `go build ./cmd/pm`: pass.
- `./pm docs validate --connectors-dir docs/connectors`: pass.
- `make connector-boundary`: pass, outcome clean.
- `git diff --check`: pass.
- Help smoke: `pm help mixpanel`, bare `pm mixpanel`, `pm mixpanel lookup-tables list-lookup-tables --help`, and `pm mixpanel insights insights-query --help`: pass.
- GitHub issue addendum marker verification: #3158-#3165 each contain exactly one `captain-policy-addendum-mixpanel-destructive-confirmation-r1` marker.

## Full make verify status

`make verify` was attempted three times. The non-test steps (`gofmt`, `go mod tidy`, `go.mod/go.sum` diff, and `go vet ./...`) completed. The full `go test -timeout 20m ./...` phase hit the repository-wide 20-minute package timeout in unrelated long-running `internal/cli` and `internal/connectors/certify` tests on this machine (examples: `TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked`, `TestReverseETLCLIWorkflowIsScriptableAndApprovalBounded`, `TestWriteStagesSkipWhenDisabled`). Focused Mixpanel gates, focused CLI/golden gates, `internal/cli` with a longer timeout (`go test ./internal/cli -count=1 -timeout 35m`), and targeted certify skip test passed separately. No live provider calls or credentials were used.
