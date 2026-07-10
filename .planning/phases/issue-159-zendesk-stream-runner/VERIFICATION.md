# Verification: Zendesk Stream Runner

## Planned checks

```bash
go test ./internal/connectors/engine -run 'ZendeskStream|ZendeskDirectRead|ZendeskOperationLedger' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface|Schema' -count=1
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

- `go test ./internal/connectors/engine -run 'ZendeskStream|ZendeskDirectRead|ZendeskOperationLedger' -count=1`
- `go test ./cmd/connectorgen -run 'CLISurface|Surface|Schema' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs` — 548 connector(s) checked, 0 findings.
- `./pm docs validate --connectors-dir docs/connectors`
- `go test ./internal/connectors/conformance -run 'TestConformance/zendesk$' -count=1` after adding the synthetic `list_activities` replay fixture.

Broad handoff checks pass on the cumulative #162 stack; this #159 branch has targeted stream/conformance verification.

## Safety verification

- No secrets requested, printed, summarized, or stored.
- No credentialed Zendesk check is run.
- No reverse ETL execution is run.
- No new dependencies are added.
