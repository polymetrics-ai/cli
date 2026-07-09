# Verification — Issue #183

## Planned focused gates

```bash
go test ./internal/connectors/conformance -run TestFreshchatImplementedETLCommandsHaveReplayFixtures
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance ./internal/connectors/engine ./cmd/connectorgen
```

## Planned full gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

No credentialed Freshchat checks, no secret inspection, and no reverse ETL execution are in scope.
