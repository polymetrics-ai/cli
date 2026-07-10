# Verification: Zendesk Help Renderer

## Planned checks

```bash
go test ./internal/cli -run 'Zendesk.*Help|Connector.*Help' -count=1
./pm help zendesk
./pm zendesk
./pm zendesk read list-tickets --help
./pm connectors inspect zendesk --json
go run ./cmd/connectorgen validate internal/connectors/defs
./pm docs validate --connectors-dir docs/connectors
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Results

Pending.

## Safety verification

- Help checks are non-credentialed.
- No live Zendesk requests are run.
- No reverse ETL execution is run.
