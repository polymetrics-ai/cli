# Verification — Zoom ETL certification parity

- [x] Targeted Zoom package test proves generator output structure.
- [x] Existing Zoom fixture/conformance behavior remains green.
- [x] `connectorgen validate internal/connectors/defs/zoom` passes.
- [x] `connectorgen surface-sync --check` passes.
- [x] `connectorgen certification-sweep --connector zoom --check` passes.
- [ ] Real runtime preflight passes.
- [ ] Connector boundary passes with target-only production paths.
- [x] `agentcontractgen certification-gate --connector zoom --transition integrate_sub_pr` returns expected `HALT capability/zoom/missing`; per the 2026-08-19 captain decision this is a firstmate-owned central scope dependency, so the wave is explicitly pending certification and not eligible for integration.
- [x] `make verify` passes before every push.

## Recorded evidence

- `go test -timeout 20m ./internal/connectors/defs/zoom -count=1` — pass.
- `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1` — pass.
- `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification(Parity|Sweep)' -count=1` — pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs/zoom` — pass; `1 connector(s) checked, 0 findings`.
- `go run ./cmd/connectorgen surface-sync --check` — pass; 552 bundles scanned with no drift.
- `go run ./cmd/connectorgen certification-sweep --connector zoom --check` — pass; 4 rows and 3 CLI commands current.
- `make connector-runtime-preflight` — pass.
- `make connector-boundary` — pass; 292 files and 552 connectors scanned, outcome `clean`.
- `./pm connectors` and `./pm zoom --help` — pass; unchanged bare namespace and Zoom help render successfully.
- `make verify` — pass (full test suite, binary build, docs validation, smoke, lint, generated artifacts, agent contract, 552-bundle validation, surface sync, certification foundation checks, connector boundary, connector canon, and release-target parity).
- No live credential was resolved or used. The fixture evidence preserves the existing ETL commands and does not certify them.
