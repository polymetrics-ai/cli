# Verification — cli-engine-secret-input-r1

Planned focused gates, to be recorded with real output after each green slice:

- focused package tests for typed input, command runner, engine bundle loading, and CLI behavior;
- `go test ./internal/cli/` in a separate command;
- `go vet ./...`, `go build ./cmd/pm`, and individual `make verify` component gates;
- `go run ./cmd/connectorgen surface-sync --check` after Zendesk surface integration;
- CLI parity checks: namespace help, command help, and generated docs/website checks as applicable;
- explicit sentinel-leak test plus a mutation check that proves the assertion fails when redaction is
  bypassed.

The full suite remains CI-owned because the repository's per-command timeout makes one `go test
./...` invocation unreliable.

## Completed slice: source-reference parser

| Command | Result |
| --- | --- |
| `go test ./internal/connectors/commandrunner -run 'TestResolveSecretInputs' -count=1` | pass |
| `go test ./internal/connectors/commandrunner -count=1` | pass |

The leak assertion was mutation-checked by returning a downstream error directly; the focused test
failed without printing the sentinel value, then passed after the fixed error boundary was restored.
