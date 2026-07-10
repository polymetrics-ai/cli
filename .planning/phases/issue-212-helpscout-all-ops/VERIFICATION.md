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

## PR / CI

Draft stacked PR: https://github.com/polymetrics-ai/cli/pull/258

Base: `feat/213-helpscout-cli-surface-metadata` (#236). Head: `feat/212-helpscout-all-ops`.

PR CI status after polling: green (`verify`, website checks/image, security, PR guard, conventions all success; website deploy skipped as expected).

Automated review routing: CodeRabbit posted `Review skipped` because the PR is draft. This is not review coverage. Review blocker comment: https://github.com/polymetrics-ai/cli/pull/258#issuecomment-4933338552
