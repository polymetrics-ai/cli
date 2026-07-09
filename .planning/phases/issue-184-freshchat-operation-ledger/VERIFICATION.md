# Verification — Issue #184 Freshchat operation ledger

## Planned focused gates

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run TestFreshchatAPISurfaceOperationLedger
go test ./cmd/connectorgen -run 'TestValidate_APISurfaceOperationLedger|TestValidate_CLISurface'
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Planned full gates before handoff

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

No credentialed Freshchat checks, no secret inspection, and no reverse ETL execution are in scope.
