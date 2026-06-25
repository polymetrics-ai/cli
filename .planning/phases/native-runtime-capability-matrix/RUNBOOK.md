# Runbook

## Regenerate Catalog

```bash
go run ./cmd/pm-cataloggen
./pm docs generate --dir docs/cli --connectors-dir docs/connectors
./pm docs validate --connectors-dir docs/connectors
```

## Verify

```bash
go test ./internal/connectors -run TestConnectorCatalog
go test ./internal/cli -run TestConnectorCatalog
go test ./...
make verify
```

## Rollback

Revert the code and generated data/docs from this phase, then rerun `go test ./...`.
