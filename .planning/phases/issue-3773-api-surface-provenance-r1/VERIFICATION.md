# Verification — Issue #3773 api_surface provenance

## Checklist

- [x] Structural v2 loader contract and v1 compatibility are covered with red/green evidence.
- [x] Shared semantic validator is consumed by conformance and connectorgen, with no duplicated
  consumer rule.
- [x] `covered_by` resolution and capability derivation remain unchanged and proven by focused
  tests.
- [x] Certification emits evidence-only status/counts for legacy, complete, and invalid ledgers.
- [x] CLI/manual/website parity is updated after inspecting the existing generated certification
  docs.
- [x] No definition bundles, credentials, live calls, redaction code, executors, or approval paths
  changed.
- [ ] Rebased on current `main` before final delivery; no sibling-foundation base used.

## Completed pre-rebase evidence

- `go test ./internal/connectors/engine -count=1`
- `go test ./internal/connectors/conformance -count=1`
- `go test ./cmd/connectorgen -count=1`
- `go test ./internal/connectors/certify -count=1`
- Scoped CLI parity tests for bare `connectors`, `connectors certify --help`, report text, and JSON.
- `go vet ./...`; `go build ./cmd/pm`
- `make tidy-check`; `make lint`; `make docs-check`; `make smoke-no-build`
- `make agent-contract-check`; `make connectorgen-validate` (550 connectors, 0 findings)
- `make connectorgen-surface-sync` (550 connectors, 0 corrections); `make connector-boundary`
- `make release-workflow-check`
- `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- Runtime help checks: `pm help connectors`, bare `pm connectors`, and
  `pm connectors certify --help` all exit successfully and describe v2 provenance.

The same focused test matrix and relevant gates are rerun after the required rebase on `main`.

## Required local gates

```sh
go test ./internal/connectors/engine -count=1
go test ./internal/connectors/conformance -count=1
go test ./cmd/connectorgen -count=1
go test ./internal/connectors/certify -count=1
go test ./internal/cli -run 'TestBareCommandShowsManualInsteadOfUsageError|TestConnectorsManualDocumentsConnectorArchitectureAndGithubExamples|TestCertifyCLIHelpShowsProvenanceContract|TestRenderCertifyReportTextIncludesSurfaceProvenanceEvidence|TestWriteCertifyReportJSONIncludesSurfaceProvenanceEvidence' -count=1
go vet ./...
go build ./cmd/pm
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```

Aggregate `go test ./...` and `make verify` are intentionally reserved for CI because the suite
exceeds the per-command time budget documented in `AGENTS.md`.
