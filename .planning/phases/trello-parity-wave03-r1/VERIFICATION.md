# Trello parity wave 03 verification

## Required gates

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/trello
go test ./internal/connectors/conformance -run 'TestConformance/trello' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

## Additional focused gates

```bash
go test ./cmd/connectorgen -run TrelloAPISurfaceOperationLedger -count=1
pm help trello
pm trello
pm trello --help
pm connectors inspect trello --json
```

## Results

- PASS: `go run ./cmd/connectorgen validate internal/connectors/defs/trello` → 1 connector checked, 0 findings.
- PASS: `go test ./cmd/connectorgen -run TrelloAPISurfaceOperationLedger -count=1`.
- PASS: `go test ./internal/connectors/conformance -run 'TestConformance/trello' -count=1`.
- PASS: `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`.
- PASS: `go build ./cmd/pm`.
- PASS: `make connector-boundary` → clean boundary report.
- PASS: `make verify` → fmt, tidy-check, vet, full tests, build, docs-check, smoke, lint, connectorgen validate, boundary, release workflow check.
- PASS: `git diff --check`.
- PASS: `./pm help trello`, `./pm trello`, `./pm trello --help`, and `./pm connectors inspect trello --json`.

## Notes

- No credentialed Trello checks, no live Trello API calls, and no provider writes were executed.
- `make verify` took advantage of Go test caching for the certify package after a dedicated uncached certify run passed within the package timeout.
