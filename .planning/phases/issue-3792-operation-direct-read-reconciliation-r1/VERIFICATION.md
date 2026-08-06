# Verification — #3792 operation direct-read preflight and surface reconciliation

## Checklist

- [x] Engine static metadata reuses the operation direct-read admission path.
- [x] Preflight rejects non-executable metadata before dispatching a command.
- [x] The repository-wide implemented-command sweep is unchanged and green.
- [x] Reconciliation reaches runtime preflight rather than duplicating it.
- [x] Fixture proves coverage requires a runnable command; failure/refusal cases do not write coverage.
- [x] #2985 six-connector report is check-only; no targeted connector definition changes.
- [x] `connectorgen` usage and migration authoring documentation are current; no `pm` user-surface generation is applicable.
- [x] Focused tests, vet, build, and individual repository gates pass.

Aggregate `go test ./...` and `make verify` remain CI work under the per-command
timeout guidance in `AGENTS.md`.

## Evidence

- `go test ./internal/connectors/commandrunner -count=1` — passes, including
  unchanged `TestEveryImplementedCommandPassesRuntimePreflight`.
- `go test ./internal/connectors/engine -count=1`, native Amazon SQS and Ashby
  packages, and `go test ./cmd/connectorgen -count=1` — pass.
- `go test ./internal/cli -count=1` — passes (410 seconds, run separately as
  required by the repository timeout guidance).
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`,
  `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check` — pass.
- `go run ./cmd/connectorgen surface-reconcile internal/connectors/defs
  --check --json --reason-contains '#2985'` exits 1 as intended for proposed
  changes and reports 574 scanned, 0 covered, 574 blocked, and 0 refused;
  `RECLASSIFICATION-REPORT.md` records the per-connector counts.
- `go run ./cmd/connectorgen --help` and `go run ./cmd/connectorgen
  surface-reconcile --help` show the maintainer command. No `pm` CLI source,
  generated manual, or website page is affected.
