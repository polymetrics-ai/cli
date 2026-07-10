# Verification: issue-212 Help Scout all operations

## Targeted checks

```bash
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json

go test ./internal/connectors/engine -run 'DirectRead|Operations|CLISurface'
go test ./cmd/connectorgen -run 'HelpScout|CLISurface|Operation'
go test ./internal/connectors/commandrunner -run 'DirectRead|Operation|WriteCommand'
go test ./internal/cli -run 'HelpScoutConnectorNamespaceRendersCommandSurfaceHelp|HelpScoutCommandSurface|Manual'
go test ./internal/connectors/conformance -run 'TestConformance/help-scout'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: passed.

## CLI/docs/website parity checks

```bash
go build ./cmd/pm
./pm help docs
./pm help connectors
./pm connectors
./pm help help-scout
./pm help-scout
./pm help-scout --help
./pm connectors inspect help-scout --help
./pm connectors inspect help-scout --json
./pm docs generate --dir docs/cli
./pm docs validate --connectors-dir docs/connectors
cd website && pnpm run gen:website-data
cd website && pnpm run typecheck
rg -n "help-scout|Help Scout" docs/cli docs/connectors website
```

Results so far:

- `go build ./cmd/pm`: passed.
- `./pm help docs`: passed.
- `./pm help connectors`: passed.
- `./pm connectors`: passed.
- `./pm help help-scout`: passed; renders connector manual/command surface without credentials.
- `./pm help-scout`: passed; renders connector manual/command surface without credentials.
- `./pm help-scout --help`: passed; renders connector manual/command surface without credentials.
- `./pm connectors inspect help-scout --help`: passed.
- `./pm connectors inspect help-scout --json`: passed; connector name `help-scout`, write capability `true`, no credentials read.
- `./pm docs generate --dir docs/cli`: ran; broad unrelated connector manual rewrites were reverted, retaining Help Scout/catalog outputs plus the targeted `docs/cli/connectors.md` help note.
- `./pm docs validate --connectors-dir docs/connectors`: passed.
- `cd website && pnpm run gen:website-data`: passed.
- `cd website && pnpm run typecheck`: blocked because `tsc` is unavailable and `website/node_modules` is missing. No dependency install was run.
- `rg -n "help-scout|Help Scout" docs/cli docs/connectors website`: passed; found updated generated Help Scout/catalog/website references.

## Full gate before PR handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: passed.
