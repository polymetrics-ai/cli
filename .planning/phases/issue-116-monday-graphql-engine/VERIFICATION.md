# Verification — issue #116 Monday GraphQL/direct-read engine

```bash
go test ./internal/connectors/engine -run 'TestDirectReadGraphQL' -count=1
go test ./internal/connectors/commandrunner -run 'TestRunDirectReadGraphQLOperation' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `go test ./internal/connectors/engine -run 'TestDirectReadGraphQL' -count=1` — pass.
- `go test ./internal/connectors/commandrunner -run 'TestRunDirectReadGraphQLOperation' -count=1` — pass.
- `go test ./internal/connectors/engine -run 'TestDirectRead' -count=1` — pass.
- `go test ./internal/connectors/commandrunner -run 'TestRun.*DirectRead' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — pass: 547 connectors, 0 findings, 0 warnings.
