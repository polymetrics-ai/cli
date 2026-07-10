# Verification: issue-212 Help Scout all operations

## Planned targeted checks

```bash
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json

go test ./internal/connectors/engine -run 'DirectRead|Operations|CLISurface'
go test ./cmd/connectorgen -run 'HelpScout|CLISurface|Operation'
go test ./internal/connectors/commandrunner -run 'DirectRead|Operation|WriteCommand'
go test ./internal/connectors/conformance -run 'TestConformance/help-scout'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Planned CLI/docs/website parity checks

```bash
go build ./cmd/pm
./pm help connectors
./pm connectors
./pm connectors inspect help-scout --help
./pm connectors inspect help-scout --json
./pm docs generate --dir docs/cli
./pm docs validate --connectors-dir docs/connectors
cd website && pnpm run gen:website-data
rg -n "help-scout|Help Scout" docs/cli docs/connectors website
```

## Planned full gate before PR handoff

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Results

Pending.
