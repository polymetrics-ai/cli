# TDD Ledger

## Red

- `gofmt -w internal/connectors/commandrunner/runner_test.go internal/connectors/engine/direct_read_test.go internal/cli/cli_test.go`
- `go test ./internal/connectors/commandrunner -run DirectRead -count=1`
  - Fails to build because `connectors.DirectReadRequest`, `connectors.DirectReadResult`,
    command `APISurface` refs, and `Result.DirectRead` do not exist yet.
- `go test ./internal/connectors/engine -run DirectRead -count=1`
  - Fails to build because `engine.DirectRead` and `connectors.DirectReadRequest` do not exist yet.
- `go test ./internal/cli -run DirectReadFile -count=1`
  - Fails because `repo read-file` is currently `availability=planned` and blocked by policy.

## Green

- `gofmt -w internal/connectors/connectors.go internal/connectors/command_surface.go internal/connectors/engine/connector.go internal/connectors/engine/direct_read.go internal/connectors/commandrunner/runner.go internal/cli/cli.go cmd/connectorgen/validate.go`
- `go test ./internal/connectors/commandrunner -run DirectRead -count=1`
  - Passes.
- `go test ./internal/connectors/engine -run DirectRead -count=1`
  - Passes.
- `go test ./internal/cli -run DirectReadFile -count=1`
  - Passes.
- `go test ./cmd/connectorgen -run CLISurface -count=1`
  - Passes.
- `go run ./cmd/connectorgen validate internal/connectors/defs`
  - Passes with `547 connector(s) checked, 0 findings`.
- `go test ./internal/cli -run 'GitHubCommandSurface|DirectReadFile' -count=1`
  - Passes.
- `go test ./cmd/connectorgen -count=1`
  - Passes.
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner -count=1`
  - Passes.
- `go vet ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/cli ./cmd/connectorgen`
  - Passes.
- `go build ./cmd/pm`
  - Passes.

## Refactor

- No raw API, direct write, mutation execution, or generic HTTP write path was added.
- Runtime direct reads rely on validated `cli_surface.json` endpoint refs because production embeds
  intentionally omit `api_surface.json`.
- A broad `go test ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/cli ./cmd/connectorgen -count=1`
  run was stopped after `internal/cli` entered the known slow schedule-test path; targeted CLI
  GitHub command-surface coverage passed.
