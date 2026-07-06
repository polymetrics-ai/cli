# TDD Ledger

## Red

- `go test ./internal/connectors/engine -run QueryOverride` fails because
  `connectors.ReadRequest` has no `Query` field.
- `go test ./internal/connectors/commandrunner` fails because the command runner package/API does
  not exist and `ReadRequest.Query` is missing.
- `go test ./internal/cli -run GitHubCommandSurface` fails because `pm github ...` is currently an
  unknown command.

## Green

- Pending.

## Refactor

- Pending.
