# Verification checklist — provider-cited rate-limit declarations R1

## Planned commands

- `jq empty internal/connectors/defs/*/rate_limits.json`
- `go test ./internal/connectors/engine -run TestProductionDefinitionsEmbedEveryRateLimitDeclaration`
- `go test ./internal/connectors/engine`
- `go build ./cmd/pm`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`

## Recorded evidence

- 2026-08-06: `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` failed as expected when
  `harvest/rate_limits.json` existed before the `defs.FS` wildcard. It passed after the wildcard
  was added.
- `make tidy-check`
- `make lint`
- `make docs-check`
- `make smoke-no-build`
- `make agent-contract-check`
- `make connector-boundary`
- `make release-workflow-check`

## Results

- `jq empty internal/connectors/defs/*/rate_limits.json` — pass (25 files).
- `go test ./internal/connectors/engine -run TestProductionDefinitionsEmbedEveryRateLimitDeclaration -count=1` — pass after the planned red result.
- `go test ./internal/connectors/engine -count=1` — pass.
- `go test ./internal/connectors/commandrunner -count=1` — pass.
- `go vet ./...` and `go build ./cmd/pm` — pass.
- `go run ./cmd/connectorgen validate` — 550 connectors checked, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` — 550 connectors scanned, 0 changes.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check` — pass.
- `git diff --check` — pass.
- Population audit — pass: all 25 `rate_limits.json` directories join to the current authoritative
  sweep ledger as `status: done`, with a nonblank provider artifact URL and
  `scope_in_current_defs: true`. No selected connector matches a dead/retired sweep reason.

The full `go test ./...` and aggregate `make verify` are intentionally left to CI because this
repository's full suite exceeds the worker command timeout; all relevant package tests and each
non-test verification gate were run individually.
