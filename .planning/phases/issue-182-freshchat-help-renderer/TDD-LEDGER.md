# TDD Ledger — Issue #182 Freshchat help renderer

## Red target

Add CLI tests proving connector namespace help works without credentials:

- `pm freshchat` exits 0 and prints Freshchat command-surface help.
- `pm freshchat --help` exits 0 and prints Freshchat command-surface help.
- Output includes `pm freshchat <command> [flags]`, `user list`, `conversation update`, and reverse-ETL approval guidance.

Observed red failure:

```bash
gofmt -w internal/cli/cli_test.go
go test ./internal/cli -run TestFreshchatCommandSurfaceHelp
```

Result:

```text
--- FAIL: TestFreshchatCommandSurfaceHelp (0.52s)
    --- FAIL: TestFreshchatCommandSurfaceHelp/freshchat (0.52s)
        cli_test.go:328: Run([freshchat]) code = 2 stderr=error: missing connector command path
             stdout=
    --- FAIL: TestFreshchatCommandSurfaceHelp/freshchat_--help (0.00s)
        cli_test.go:328: Run([freshchat --help]) code = 1 stderr=error: help topic "freshchat" not found
             stdout=
FAIL
FAIL	polymetrics.ai/internal/cli	1.053s
```

This matches the expected failure: connector namespace help is not routed through command-surface metadata yet.

## Green result

```bash
gofmt -w cmd internal
go test ./internal/cli -run TestFreshchatCommandSurfaceHelp
go run ./cmd/connectorgen validate internal/connectors/defs
go build ./cmd/pm
./pm help connectors
./pm freshchat
./pm freshchat --help
./pm docs validate --connectors-dir docs/connectors
```

Results:

- `go test ./internal/cli -run TestFreshchatCommandSurfaceHelp`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.
- `go build ./cmd/pm`: pass.
- `./pm help connectors`: pass.
- `./pm freshchat`: pass; output includes `COMMAND SURFACE`, `pm freshchat <command> [flags]`, `user list`, and `conversation update`.
- `./pm freshchat --help`: pass with the same credential-free command-surface help.
- `./pm docs validate --connectors-dir docs/connectors`: pass.

Green changes:

- Connector namespace help routes to connector manual/command surface without credential resolution.
- Freshchat generated manual includes the `COMMAND SURFACE` section.
- CLI docs/website docs mention Freshchat command-surface usage and safety constraints.

## Planned verification

```bash
gofmt -w cmd internal
go test ./internal/cli -run 'TestFreshchatCommandSurfaceHelp'
go test ./internal/connectors -run 'TestRenderConnectorManual|TestFreshchat'
go run ./cmd/connectorgen validate internal/connectors/defs
./pm docs validate --connectors-dir docs/connectors
```

Full handoff gates when practical:

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```
