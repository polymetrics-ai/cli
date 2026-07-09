# Verification: Chatwoot Operation Ledger

## Planned gates

```bash
go test ./cmd/connectorgen -run ChatwootAPISurfaceOperationLedgerMetrics -count=1
go test ./cmd/connectorgen -run 'GitHubAPISurface|ChatwootAPISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1
git diff --check
```

## Full handoff gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Results

Pending.
