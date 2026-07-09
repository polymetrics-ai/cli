# Verification — issue #113 Monday stream runner

```bash
go test ./internal/connectors/commandrunner -run 'TestRunMonday' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Results:

- `go test ./internal/connectors/commandrunner -run 'TestRunMonday' -count=1` — pass.
- Test uses a local `httptest.Server`, synthetic GraphQL response, `max_pages=1`, and no credentials.
