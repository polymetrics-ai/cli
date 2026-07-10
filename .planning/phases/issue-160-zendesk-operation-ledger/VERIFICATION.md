# Verification: Zendesk Operation Ledger

## Planned checks

```bash
go test ./internal/connectors/engine -run 'Zendesk|Operation' -count=1
go test ./cmd/connectorgen -run 'Operations|Surface|Zendesk' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
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

- No secrets requested, printed, summarized, or stored.
- No credentialed Zendesk check is run.
- No reverse ETL execution is run.
- No new dependencies are added.
