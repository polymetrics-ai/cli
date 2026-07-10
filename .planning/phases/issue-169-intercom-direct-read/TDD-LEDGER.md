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
