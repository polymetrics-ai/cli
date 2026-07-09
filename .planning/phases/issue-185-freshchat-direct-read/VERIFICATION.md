# Verification — Issue #185

## Planned focused gates

```bash
go test ./internal/connectors/commandrunner -run 'TestRunFreshchatUsersFetchDirectReadCommand|TestRunDirectReadRejectsUnsafeEndpointMetadata|TestRunDirectReadRequiresOutputPolicy'
go test ./internal/connectors/engine -run 'TestDirectReadFreshchatUsersFetchPOST|TestDirectReadRejectsMutationMethod'
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface|TestFreshchatAPISurfaceLedger'
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

No credentialed Freshchat checks, no secret inspection, no reverse ETL execution, and no file upload/binary executor are in scope.
