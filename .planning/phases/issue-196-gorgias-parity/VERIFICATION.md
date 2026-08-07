# Verification — issue 196 Gorgias parity

Commands to run and record:

```bash
go run ./cmd/connectorgen validate
go test ./internal/connectors/conformance -run 'TestConformance/gorgias' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go test ./internal/cli -run 'TestDynamicConnector|TestCobraRouterShellPreservesDynamic|TestRootHelpListsDynamic|TestConnectorInspect|TestConnectorCatalog|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1 -timeout=180s
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
git diff --check
```

Full-repo gates from `AGENTS.md` remain the broader local standard but may be too broad for this connector-only crewmate stop point. Record any skipped broad gate honestly; do not claim certification or live behavior.

## Results

- ✅ `go run ./cmd/connectorgen validate` — pass, `549 connector(s) checked, 0 findings`.
- ✅ `go test ./internal/connectors/conformance -run 'TestConformance/gorgias' -count=1` — pass.
- ⚠️ `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=60s` — timed out in existing certification batch coverage (`TestCertifyCLIBatchModeRunsCredsFileConnectors`); not claimed green.
- ✅ Scoped CLI parity gate: `go test ./internal/cli -run 'TestDynamicConnector|TestCobraRouterShellPreservesDynamic|TestRootHelpListsDynamic|TestConnectorInspect|TestConnectorCatalog|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1 -timeout=180s` — pass.
- ✅ `go vet ./internal/connectors/... ./internal/cli/...` — pass.
- ✅ `go build ./cmd/pm` — pass.
- ✅ `make connector-boundary` — pass (`outcome: clean`; pre-existing GitHub/Bahmni exceptions only).
- ✅ `git diff --check` — pass.
- ✅ Smoke checks: `./pm connectors inspect gorgias --json`, `./pm gorgias --help`, `./pm gorgias tickets list --help`, and `./pm gorgias search --help` returned successfully.

No live Gorgias credentials, provider calls, writes, certification, push, or PR were performed.
