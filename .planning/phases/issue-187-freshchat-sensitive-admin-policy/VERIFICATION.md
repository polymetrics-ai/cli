# Verification — Issue #187

## Planned focused gates

```bash
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestFreshchatSensitiveAdminWritesRequireTypedConfirmation'
go test ./internal/connectors/commandrunner -run TestFreshchatSensitiveAdminWriteCommandsCarryConfirmationChallenges
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

No credentialed Freshchat checks, no reverse ETL execution, and no live external mutations are in scope.
