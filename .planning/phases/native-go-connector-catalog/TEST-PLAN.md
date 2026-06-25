# Native Go Connector Catalog TEST PLAN

- `go test ./internal/connectors -run TestConnectorCatalog`
- `go test ./internal/cli -run TestConnectorCatalog`
- `go test ./...`
- `go build ./cmd/pm`
- `./pm docs generate --dir docs/cli --connectors-dir docs/connectors`
- `./pm docs validate --connectors-dir docs/connectors`
- `make verify`
