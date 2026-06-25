# Runbook

## Inspect Plans

```bash
pm connectors port-plan --all --json
pm connectors port-plan source-postgres
pm connectors port-plan source-mysql
pm connectors port-plan source-mongodb-v2
```

## Verify

```bash
go test ./internal/connectors -run TestNativePort
go test ./internal/cli -run TestConnectorPortPlan
go test ./...
make verify
```

## Enable A Native Port

1. Implement the connector as a Go `connectors.Connector`.
2. Add fixture-backed and integration conformance tests.
3. Ensure docs and skills validate.
4. Flip implementation status only after tests pass.
