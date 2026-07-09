# Verification — Issue #182 Freshchat help renderer

## Planned focused gates

```bash
gofmt -w cmd internal
go test ./internal/cli -run 'TestFreshchatCommandSurfaceHelp'
go run ./cmd/connectorgen validate internal/connectors/defs
go build ./cmd/pm
./pm help connectors
./pm freshchat
./pm freshchat --help
./pm docs validate --connectors-dir docs/connectors
```

## Planned full gates before handoff

```bash
go vet ./...
go test ./...
make verify
```

No credentialed Freshchat checks, no secret inspection, and no reverse ETL execution are in scope.
