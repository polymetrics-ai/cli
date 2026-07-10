# Verification — Issue #186

## Planned focused gates

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatFileUploadOperations
go test ./internal/connectors/commandrunner -run TestRunFreshchatUploadCommandsBlockTypedOperationsBeforeCredentials
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestFreshchatAPISurfaceLedger|TestFreshchatBinaryUploadCommandsUseTypedOperations'
go run ./cmd/connectorgen validate internal/connectors/defs
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

No credentialed Freshchat checks, no local file reads, no binary uploads, no reverse ETL execution.
