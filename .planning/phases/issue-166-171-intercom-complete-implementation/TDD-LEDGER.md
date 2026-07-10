# TDD Ledger: Intercom Complete CLI Implementation (#166-#171)

## Planned red tests

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceFullCoverage -count=1`
  - Expect failure before ledger rewrite because only 5/149 endpoints are covered.
- `go test ./internal/connectors/commandrunner -run 'TestIntercomDirectRead|TestIntercomWriteCommand' -count=1`
  - Expect failure before generic JSON direct-read policy / representative write actions.
- `go test ./internal/cli -run 'TestConnectorCommandHelp|TestIntercomConnectorCommandHelp' -count=1`
  - Expect failure before connector namespace help rendering.
- `go run ./cmd/connectorgen validate internal/connectors/defs/intercom`
  - Must pass after each bundle-data rewrite.

## Red evidence: 2026-07-10

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceFullCoverage -count=1` failed as expected: first Intercom endpoint was metadata-only/blocked rather than covered.
- `go test ./internal/connectors/commandrunner -run TestRunImplementedIntercomJSONDirectReadCommand -count=1` failed as expected: `json_response` output policy is not yet supported.
- `go test ./internal/cli -run TestIntercomConnectorCommand -count=1` failed as expected: bare `pm intercom` reported missing command path and `contact view --help` was still planned/blocked.
- `go test ./internal/connectors/engine -run 'TestDirectRead(JSONResponse|TextResponse|BinaryMetadata)' -count=1` initially caught missing test hook args; fixed test harness and kept the behavioral failures until policy implementation.

## Green evidence: 2026-07-10

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceFullCoverage -count=1` passed after converting all 149 Intercom operations to covered stream/direct-read/write/binary policies.
- `go test ./internal/connectors/engine -run 'TestDirectRead(JSONResponse|TextResponse|BinaryMetadata)|TestWrite' -count=1` passed after adding bounded output policies, query defaults, and write query/body handling.
- `go test ./internal/connectors/commandrunner -run 'TestRunImplementedIntercomJSONDirectReadCommand|TestBuildWriteCommand' -count=1` passed after commandrunner dispatch supported direct reads and reverse-ETL write plans.
- `go test ./internal/cli -run 'TestIntercom(CommandSurface|ConnectorCommand)' -count=1` passed after connector namespace/help rendering.
- `go test ./internal/cli -run 'TestIntercom|TestHelp' -count=1` passed after adding `pm help intercom` coverage.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed with 547 connectors and 0 findings.
- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli ./internal/connectors/conformance -count=1` passed.
- Full gates passed: `go vet ./...`, `go test ./... -timeout=20m`, `go build ./cmd/pm`, `make verify`.
- Implementation checkpoint commit: `bb382e48 feat(intercom): implement complete CLI command parity`.
- Stacked PR opened: https://github.com/polymetrics-ai/cli/pull/257; CI passed and the PR was squash-merged into the parent branch at `8362291f`.

## Refactor / fix evidence

- Conformance initially failed Intercom because generated stream query params referenced optional `query.*` values with defaults/omit semantics; `resolveQueryParams` now treats absent query namespace values like absent config/secrets/incremental values for object-form defaults/omits.
- Intercom docs initially tripped the secret-shaped literal guard due an example env-var name adjacent to `access_token=`; the docs now use `<env-var-name>`.
- Existing Intercom admin/conversation fixtures were updated to match the declared records paths.
