# Verification — issue #112 Monday help renderer/docs

```bash
go test ./internal/connectors/bundleregistry -run 'TestMondayGuideIncludesCLISurfaceHelp' -count=1
go run ./cmd/pm connectors inspect monday
rg -n "CLI command surface|pm monday|board list" internal/connectors/defs/monday/docs.md
```

Results:

- `go test ./internal/connectors/bundleregistry -run 'TestMondayGuideIncludesCLISurfaceHelp' -count=1` — pass.
- `go run ./cmd/pm connectors inspect monday` — pass; output includes `COMMAND SURFACE`, `Usage: pm monday <command> <subcommand> [flags]`, implemented stream commands, and planned automation commands.
- `rg -n "CLI command surface|pm monday|board list" internal/connectors/defs/monday/docs.md` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings.

Not applicable: `pm help monday` / `pm monday --help` do not exist as dynamic connector help routes in this CLI; the runtime-rendered connector manual path is `pm connectors inspect monday`.
