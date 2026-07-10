# Verification: Zendesk Sensitive/Admin Policy

## Planned checks

```bash
go test ./internal/connectors/engine -run 'ZendeskWrite|ZendeskStream|ZendeskDirectRead|ZendeskOperationLedger' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface|Write' -count=1
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

Targeted checks passed:

- `go test ./internal/connectors/engine -run 'ZendeskWrite|ZendeskStream|ZendeskDirectRead|ZendeskOperationLedger' -count=1`
- `go test ./cmd/connectorgen -run 'CLISurface|Surface|Write' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 548 connector(s) checked, 0 findings.
- `./pm docs validate --connectors-dir docs/connectors`

Broad handoff checks are pending.

## Safety verification

- No secrets requested, printed, summarized, or stored.
- No credentialed Zendesk check is run.
- No reverse ETL execution is run.
- No new dependencies are added.
