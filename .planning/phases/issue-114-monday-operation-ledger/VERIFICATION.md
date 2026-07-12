# Verification — issue #114 Monday operation ledger

```bash
go test ./cmd/connectorgen -run 'TestMondayOperationLedger' -count=1
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayOperationLedger' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `go test ./cmd/connectorgen -run 'TestMondayOperationLedger' -count=1` — pass.
- `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayOperationLedger' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings.
- Secrets check: no real secret values stored; repeated `access_token`/`api_token` strings appear only as redaction field names in policy metadata.
