# VERIFICATION — engine configuration-time spec-constraint validation

## Required evidence

| Area | Command / check | Result |
| --- | --- | --- |
| Red reproduction | focused app test before production code | passed — recorded in `TDD-LEDGER.md` |
| Engine constraint unit tests | `go test ./internal/connectors/engine -count=1` | passed (10.926s) |
| App boundary tests | `go test ./internal/app -count=1` | passed (49.092s) |
| Connector contracts | `go test ./internal/connectors`, `bundleregistry`, and `native/nativeset` | passed |
| CLI regression package | `go test ./internal/cli -count=1` | passed (383.991s) |
| Format | `gofmt -w` changed Go files then `gofmt -l cmd internal` | passed |
| Vet | `go vet ./...` | passed |
| PM build | `go build ./cmd/pm` | passed |
| Tidy | `make tidy-check` | passed |
| Lint | `make lint` | passed (0 issues) |
| Docs | `make docs-check` | passed |
| Smoke | `make smoke-no-build` | passed |
| Connector validation | `make connectorgen-validate` | passed (550 connectors, 0 findings) |
| Surface sync | `make connectorgen-surface-sync` | passed (550 connectors, 0 corrections) |
| Boundary | `make connector-boundary` | passed (`clean`, 0 findings) |
| Release workflow | `make release-workflow-check` | passed |
| GSD evidence | `scripts/verify-gsd-workflow origin/main` | passed |
| Scope guard | `git diff --name-only origin/main...HEAD -- internal/connectors/defs` | passed (empty; no bundle change) |
| Rebase regression | focused engine/app/native-set matrix after rebase to `d30dd4905` | passed |

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
