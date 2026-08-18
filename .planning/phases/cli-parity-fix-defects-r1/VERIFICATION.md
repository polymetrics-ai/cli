# Verification checklist — GitHub command certification defect fixes

## Local gates

- [ ] Exact-integer red/green regression proves the persisted `9007199254740993` path.
- [ ] Required-body red/green tests cover blob, commit, tree, check-run, branch protection, and commit status.
- [ ] Wrong-path red/green tests cover at least three Actions paths and the project-draft declaration.
- [ ] False-success red/green tests prove 404 is not counted as a write and 2xx remains successful.
- [ ] `gofmt -w` for changed Go files.
- [ ] Targeted changed-package tests with `-timeout 20m`.
- [ ] `go test -timeout 20m ./internal/cli` as its own command.
- [ ] `go vet ./...` and `go build ./cmd/pm`.
- [ ] `go run ./cmd/connectorgen certification-matrix --check`.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [ ] `go run ./cmd/connectorgen surface-sync --check`.
- [ ] Repository non-suite verification gates and generated/docs/help checks.
- [ ] `scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1`.
- [ ] `git diff --check` and credential scan.

## Live conversion and containment

- [ ] Three exact-integer commands change the asserted provider state.
- [ ] Three required-body commands create/update the asserted provider object.
- [ ] Three corrected-path commands change the asserted provider state.
- [ ] Org webhook create/delete and a third false-success-family command have honest counters/status and independent read-back.
- [ ] Every task-created fixture is deleted and independently returns 404/absence.
- [ ] Live evidence contains no credential, approval token, or secret value.

## Delivery

- [ ] Code review completed and every actionable finding dispositioned.
- [ ] Commits pushed only to `fm/cli-parity-fix-defects`.
- [ ] Direct PR opened with `Refs #4015`.
- [ ] GitHub API reports base `integration/4015-mvp-flat-r1`.

