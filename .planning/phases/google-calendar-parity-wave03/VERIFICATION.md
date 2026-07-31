# Google Calendar parity verification checklist

## Required local gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-calendar' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `make verify`
- [x] `git diff --check`

## Additional fixture-only checks

- [x] `go test ./internal/connectors/hooks/google-calendar -count=1`
- [x] `go test ./internal/connectors/native/nativeset -count=1`
- [x] `go test ./cmd/connectorgen -count=1`
- [x] `go test ./internal/connectors/engine -run 'TestSchemaAnyOfRequiredAlternatives|TestSchemaOneOfRejectsMultipleMatches' -count=1`
- [x] `./pm connectors inspect google-calendar --json` after build (no credentials; inspect only)
- [x] `pm help connectors` before connector CLI inspection

## Gate evidence notes

- `make verify` completed after warming long-running package test caches for `internal/cli` and `internal/connectors/certify`; the final `make verify` invocation passed its own `go test -timeout 20m ./...`, docs validation, smoke test, golangci-lint, full connectorgen validation, connector-boundary, and release notification assertions.
- Bundled reviewer subagent found no remaining findings after auth, EventDateTime, query-conformance, and CLI-help fixes.
- No live Google Calendar endpoint calls, credentialed connector checks, or provider writes were performed.
