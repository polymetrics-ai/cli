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

## Red / green log

Pending. No production edits before this ledger and the verification checklist are committed or staged.

## Red evidence: 2026-07-10

- `go test ./cmd/connectorgen -run TestIntercomAPISurfaceFullCoverage -count=1` failed as expected: first Intercom endpoint was metadata-only/blocked rather than covered.
- `go test ./internal/connectors/commandrunner -run TestRunImplementedIntercomJSONDirectReadCommand -count=1` failed as expected: `json_response` output policy is not yet supported.
- `go test ./internal/cli -run TestIntercomConnectorCommand -count=1` failed as expected: bare `pm intercom` reports missing command path and `contact view --help` is still planned/blocked.
- `go test ./internal/connectors/engine -run 'TestDirectRead(JSONResponse|TextResponse|BinaryMetadata)' -count=1` initially caught missing test hook args; fixed test harness and remains expected to fail until policies are implemented.


## Green evidence: 2026-07-10

- Covered by combined #166-#171 red/green loop. Relevant targeted tests and broad gates passed; see `.planning/phases/issue-166-171-intercom-complete-implementation/TDD-LEDGER.md`.
