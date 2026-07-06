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

- Pending.

## Refactor

- Pending.
