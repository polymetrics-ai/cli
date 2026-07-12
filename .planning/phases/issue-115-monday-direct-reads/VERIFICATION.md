# Verification — issue #115 Monday direct reads

```bash
go test ./internal/connectors/commandrunner -run 'TestRunMondayDirectRead' -count=1
go test ./cmd/connectorgen -run 'TestMondayDirectRead' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `go test ./internal/connectors/commandrunner -run 'TestRunMondayDirectRead' -count=1` — pass.
- `go test ./cmd/connectorgen -run 'TestMondayDirectRead' -count=1` — pass.
- `go test ./cmd/connectorgen -run 'TestMonday(DirectRead|CLISurface|OperationLedger)' -count=1` — pass.
- `go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMonday' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings.
