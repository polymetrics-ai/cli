# Verification — issue #117 Monday sensitive/admin policy

```bash
go test ./cmd/connectorgen -run 'TestMondaySensitiveAdminPolicy' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `go test ./cmd/connectorgen -run 'TestMondaySensitiveAdminPolicy' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings.
