# Native Go Connector Catalog Runbook

Regenerate catalog data:

```bash
go run ./cmd/pm-cataloggen --out internal/connectors/catalog_data.json --docs-dir docs/connectors/catalog
```

Validate:

```bash
go test ./...
go build -o pm ./cmd/pm
./pm docs generate --dir docs/cli --connectors-dir docs/connectors
./pm docs validate --connectors-dir docs/connectors
```
