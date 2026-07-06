# TDD Ledger

## Red

- `go test ./internal/connectors/engine -run QueryOverride` fails because
  `connectors.ReadRequest` has no `Query` field.
- `go test ./internal/connectors/commandrunner` fails because the command runner package/API does
  not exist and `ReadRequest.Query` is missing.
- `go test ./internal/cli -run GitHubCommandSurface` fails because `pm github ...` is currently an
  unknown command.

## Green

- `go test ./internal/connectors/engine -run QueryOverride` passes after adding
  `ReadRequest.Query` overrides.
- `go test ./internal/connectors/commandrunner` passes after adding the stream-only command runner.
- `go test ./internal/cli -run GitHubCommandSurface` passes after adding the provider command route.
- `go test ./internal/app` passes after adding the narrow connector credential resolver.

## Refactor

- Added coverage for `pr list`, `release list`, and `workflow list` stream mappings.
- Broader local `go test ./internal/cli -timeout 60s` timed out in pre-existing
  `TestScheduleCLI_Remove` while invoking local `crontab`; targeted GitHub command tests pass.
