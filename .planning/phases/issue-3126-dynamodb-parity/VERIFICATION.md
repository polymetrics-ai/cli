# Verification — issue 3126 DynamoDB parity

## Required local gates

- [x] focused connectorgen validation for DynamoDB via temp root: `tmp=$(mktemp -d); mkdir -p "$tmp/dynamodb"; cp -R internal/connectors/defs/dynamodb/. "$tmp/dynamodb/"; go run ./cmd/connectorgen validate "$tmp"` → `1 connector(s) checked, 0 findings`.
- [x] all-defs validation: `go run ./cmd/connectorgen validate internal/connectors/defs` → `549 connector(s) checked, 0 findings`.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/dynamodb' -count=1` → pass.
- [x] `go test ./internal/connectors/native/dynamodb -count=1` → pass.
- [x] focused CLI tests: `go test ./internal/cli -run 'TestConnector(Inspect|Catalog)|TestDynamicConnector|TestGoldenTranscripts|TestGoldenDocsGenerate|TestCobraRouterShellPreservesDynamicConnector|TestRootHelpListsDynamicConnectorCommands' -count=1` → pass.
- [x] `go build ./cmd/pm` → pass.
- [x] `make connector-boundary` → clean boundary report, no findings/warnings (existing documented exceptions only).
- [ ] `make verify` → attempted; `fmt`, `tidy-check`, `vet`, and many packages passed, but the global `go test -timeout 20m ./...` gate timed out in pre-existing non-DynamoDB slow tests (`internal/cli` Bahmni command matrix and `internal/connectors/certify` write preview/certify harness). A HEAD-vs-current `engine.LoadAll(os.DirFS(...))` comparison showed current DynamoDB defs are not slower than HEAD defs (~2.8–3.3s vs ~3.1–3.6s per load). No DynamoDB package failure occurred.
- [x] `git diff --check` → pass.

## Manual checks / evidence captured

- [x] No live provider calls were made; all HTTP execution uses `httptest.Server` or static fixtures.
- [x] No secret-shaped fixture values are committed; synthetic access keys only appear in tests and are never real credentials.
- [x] `api_surface.json` ledger count: total 61, implemented/fixture-tested 56, blocked 2, excluded 3, certified 0.
- [x] Binary import/export rows (`ExportTableToPointInTime`, `ImportTable`) are blocked with shared-runtime dependency, not omitted.
- [x] PartiQL rows (`BatchExecuteStatement`, `ExecuteStatement`, `ExecuteTransaction`) are excluded/disallowed; no raw PartiQL/query command exists.
- [x] Batch/transaction write rows use typed builders; no raw `RequestItems`/`TransactItems` body passthrough remains.
- [x] Destructive/admin actions have typed schemas, risk text, confirmation, preview redaction/idempotency notes where applicable.
- [x] Appended idempotent captain-policy addendum to #3126-#3133 using `gh-axi issue comment`; marker verification shows exactly one `cli-connector-captain-policy-addendum:dynamodb:wave04-r1` marker on each issue.
