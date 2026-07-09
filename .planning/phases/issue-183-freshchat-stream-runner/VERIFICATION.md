# Verification — Issue #183

## Focused gates run

```bash
go test ./internal/connectors/conformance -run TestFreshchatImplementedETLCommandsHaveReplayFixtures
go test ./internal/connectors/conformance
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- `go test ./internal/connectors/conformance -run TestFreshchatImplementedETLCommandsHaveReplayFixtures`: pass.
- `go test ./internal/connectors/conformance`: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.
- `go test ./internal/connectors/conformance ./internal/connectors/engine ./cmd/connectorgen`: pass.

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
