# VERIFICATION — engine configuration-time spec-constraint validation

## Required evidence

| Area | Command / check | Result |
| --- | --- | --- |
| Red reproduction | focused app test before production code | passed — recorded in `TDD-LEDGER.md` |
| Engine constraint unit tests | `go test ./internal/connectors/engine/... -count=1` | pending |
| App boundary tests | `go test ./internal/app/... -count=1` | pending |
| Connector contracts | `go test ./internal/connectors/... -count=1` | pending |
| CLI regression package | `go test ./internal/cli/... -count=1` | pending |
| Format | `gofmt -w` changed Go files then `gofmt -l cmd internal` | pending |
| Vet | `go vet ./...` | pending |
| PM build | `go build ./cmd/pm` | pending |
| Tidy | `make tidy-check` | pending |
| Lint | `make lint` | pending |
| Docs | `make docs-check` | pending / not expected to change |
| Smoke | `make smoke-no-build` | pending |
| Connector validation | `make connectorgen-validate` | pending |
| Surface sync | `make connectorgen-surface-sync` | pending |
| Boundary | `make connector-boundary` | pending |
| Release workflow | `make release-workflow-check` | pending |

`go test ./...` and `make verify` are intentionally not run as single commands
in this per-command environment; CI carries the complete suite after the
scoped package matrix and individual `make verify` gates pass.

## Explicitly not applicable

- CLI help/manual/website parity: no command, flag, output, or connector
  surface changes are planned.
- Runtime services and credentialed checks: this is deterministic local
  validation and uses no secret values.
- Storage/vault verification: excluded by task ownership; the only assertion
  is that invalid configuration is rejected before storage is touched.
