# TDD ledger — issue 196 Gorgias parity

## Red / baseline evidence

Captured before production bundle edits:

- `go run ./cmd/connectorgen validate` exits 0 for the full defs root on the current baseline.
- `go run ./cmd/connectorgen validate internal/connectors/defs/gorgias` exits 1 because the current validator interprets `fixtures/` and `schemas/` under the connector directory as connector directories (`missing required file metadata.json`). This is a command-shape mismatch in the issue checklist, not a Gorgias bundle finding.
- The behavioral gap is known from #196: current surface has 11 local rows, 4 streams, no writes, no `operations.json`, no `cli_surface.json`, and no `certification.json`; official source inventory has 114 operations.

## Green evidence

Captured after connector-local implementation:

- `go run ./cmd/connectorgen validate` exits 0 (`549 connector(s) checked, 0 findings`). The per-connector path invocation remains an invalid validator shape because it treats `fixtures/` and `schemas/` as connector roots.
- `go test ./internal/connectors/conformance -run 'TestConformance/gorgias' -count=1` exits 0.
- `go test ./internal/cli -run 'TestDynamicConnector|TestCobraRouterShellPreservesDynamic|TestRootHelpListsDynamic|TestConnectorInspect|TestConnectorCatalog|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1 -timeout=180s` exits 0.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=60s` exits 1 due `panic: test timed out after 1m0s` in existing certification batch coverage (`TestCertifyCLIBatchModeRunsCredsFileConnectors`), so the narrower dynamic connector/docs/help gate was used and recorded rather than claiming the broad regex gate.
- `go vet ./internal/connectors/... ./internal/cli/...` exits 0.
- `go build ./cmd/pm` exits 0.
- `make connector-boundary` exits 0 with `outcome: clean` and only pre-existing documented GitHub/Bahmni exceptions.
- `git diff --check` exits 0.

## Refactor / parity checks

- Confirm `api_surface.json` has exactly 114 endpoint rows and no duplicate method/path/operationId-derived entries.
- Confirm lane disposition matches parent counts: 42 ETL/CDC streams, 63 reverse writes, 2 implemented direct reads, 1 direct/provider query blocked by safe-contract gap, 6 binary/file blocked/planned rows.
- Confirm no provider-specific shared code changes.
- Confirm fixtures contain only synthetic data and no secret-shaped literals.
- Confirm delete/destructive write actions carry `confirm: destructive` and delete idempotency notes.
