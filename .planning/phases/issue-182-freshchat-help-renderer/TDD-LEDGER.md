# TDD Ledger — Issue #182 Freshchat help renderer

## Red target

Add CLI tests proving connector namespace help works without credentials:

- `pm freshchat` exits 0 and prints Freshchat command-surface help.
- `pm freshchat --help` exits 0 and prints Freshchat command-surface help.
- Output includes `pm freshchat <command> [flags]`, `user list`, `conversation update`, and reverse-ETL approval guidance.

Expected current failure:

- `pm freshchat --help` returns `help topic "freshchat" not found`.
- `pm freshchat` returns a missing connector command path usage error.

## Green target

- Connector namespace help routes to connector manual/command surface without credential resolution.
- Freshchat generated manual includes the `Command Surface` section.
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
