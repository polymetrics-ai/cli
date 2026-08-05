# Local Verification

## Phase-specific planned checks

- `go test ./internal/connectors/engine`
- `go test ./internal/connectors/connsdk`
- `go test ./internal/connectors/commandrunner`
- `go test ./internal/app`
- `go test ./cmd/connectorgen`
- `go test ./internal/cli`
- Red/green no-redaction checks scoped to `internal/connectors/engine`,
  `internal/connectors/commandrunner`, and `internal/app`
- `gofmt -w` over changed Go files and `go vet` for changed packages
- Individual `make verify` gates: `tidy-check`, `lint`, `docs-check`,
  `smoke-no-build`, `connectorgen-validate`, `connectorgen-surface-sync`,
  `connector-boundary`, and `release-workflow-check`

Full `go test ./...` and aggregate `make verify` are intentionally left to CI
because this repository's documented timeout guidance says they routinely
exceed an agent command window.

## Completed local checks

All commands below passed on 2026-08-05.

- `go test ./internal/connectors/engine -count=1`
- `go test ./internal/connectors/connsdk -count=1`
- `go test ./internal/connectors/commandrunner -count=1`
- `go test ./internal/app -count=1`
- `go test ./cmd/connectorgen -count=1`
- `go test ./internal/cli -count=1` (387.469s on the first scoped run)
- `go vet ./...`
- `go build ./cmd/pm`
- `gofmt -l cmd internal` and `git diff --check`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check`

`connectorgen validate` checked 550 connector bundles with zero findings;
`connectorgen surface-sync --check` found zero drift.

## No-redaction correction checks (2026-08-05)

Red evidence was captured before production edits:

- Focused engine tests failed because response tokens were absent and a 500
  response body became `[redacted]`.
- Focused commandrunner tests failed because the preview record and input error
  became `***`.
- Focused app tests failed because the persisted plan sample and failed-run
  error were stripped.

The following green checks passed after the correction:

- `go test ./internal/connectors/engine -count=1`
- `go test ./internal/connectors/connsdk -count=1`
- `go test ./internal/connectors/commandrunner -count=1`
- `go test ./internal/app -count=1`
- `go test ./cmd/connectorgen -count=1`
- `go test ./internal/cli -count=1` (450.810s)
- `go vet ./...`
- `go build ./cmd/pm`
- `gofmt -l` on changed Go files and `git diff --check`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`, `make
  connector-boundary`, and `make release-workflow-check`

The direct-read file has no diff, no connector definition file has a diff, and
the direct-write executor contains no call to `redactJSONValue`,
`redactNamedJSONFields`, or `safety.RedactErrorText`.

CLI help/manual/website parity is not applicable to production content in this
slice: no production CLI command, flag, or connector bundle was added or
changed. The captain's no-redaction correction preserves every declared output
policy name and command surface; it changes only runtime content handling. The
test-only fixture exercises the existing connector-command lifecycle.

## Not performed

- Aggregate `go test ./...` and aggregate `make verify` were not performed;
  repository guidance assigns their long full-suite coverage to CI.
- Runtime-backed integration checks (`scripts/runtime.sh ...`) were not
  performed; they require local services and are optional.
- CI for this no-redaction correction has not been observed yet; the existing
  PR and its earlier push predate this correction. GitHub CI is expected after
  this corrective commit is pushed.
- The no-mistakes pipeline was not performed for this correction because the
  earlier pipeline run retained stale cancelled-run state during the outage.
- The three historical `data/...` documents named by the task could not be
  read because they are absent from this checkout and supplied workspace.
- The new GSD programming-loop adapter command was not available despite a
  passing `scripts/gsd doctor` (`unknown GSD command: programming-loop`), so
  the repository-permitted manual GSD/TDD fallback is recorded in this phase.
- Project-instructions maintenance was not performed: the required
  `fm-ensure-agents-md.sh .` check stopped because this worktree has both a
  real `AGENTS.md` and `CLAUDE.md`. It made no instruction-file change.

- CI detected: no
- Local harness required: yes

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Install or lockfile validation | missing | TBD | No matching command was detected in the repo profile. |
| Format check | configured | gofmt -l cmd internal |  |
| Lint | configured | go vet ./... |  |
| Typecheck or static analysis | configured | go vet ./... |  |
| Unit tests | configured | go test ./... |  |
| Integration tests | configured | make smoke |  |
| E2E or smoke tests | configured | make smoke |  |
| Build | configured | go build ./... |  |
| Dependency vulnerability scan | missing_optional_tool | TBD | Add or configure dependency scanning before production release. |
| Secret scan | configured | git diff --check |  |
| Accessibility check | missing_optional_tool | TBD | Required for user-facing UI work. |
| Load or benchmark | missing_optional_tool | TBD | Required for performance-sensitive paths. |
