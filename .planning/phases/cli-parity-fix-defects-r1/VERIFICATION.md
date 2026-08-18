# Verification checklist — GitHub command certification defect fixes

## Local gates

- [x] Exact-integer red/green regression proves the persisted `9007199254740993` path.
- [x] Required-body red/green tests cover blob, commit, tree, check-run, branch protection, and commit status.
- [x] Wrong-path red/green tests cover at least three Actions paths and the project-draft declaration.
- [x] False-success red/green tests prove 404 is not counted as a write and 2xx remains successful.
- [x] `gofmt -w` for changed Go files.
- [x] Targeted changed-package tests with `-timeout 20m`.
- [x] `go test -timeout 20m ./internal/cli` as its own command.
- [x] `go vet ./...` and `go build ./cmd/pm`.
- [x] `go run ./cmd/connectorgen certification-matrix --check`.
- [x] `go run ./cmd/connectorgen validate`.
- [x] `go run ./cmd/connectorgen surface-sync --check`.
- [x] Repository non-suite verification gates and generated/docs/help checks.
- [x] `scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1`.
- [x] `git diff --check` and credential scan.

## Live conversion and containment

- [x] Three exact-integer commands change the asserted provider state.
- [x] Three required-body commands replay existing provider objects with the asserted content-addressed identity.
- [x] Three corrected-path commands change the asserted provider state.
- [x] Org webhook create/delete and a third false-success-family command have honest counters/status and independent read-back.
- [x] Every task-created fixture is deleted and independently returns 404/absence.
- [x] Live evidence contains no credential, approval token, or secret value.

## Delivery

- [x] Code review completed and every actionable finding dispositioned.
- [x] Commits pushed only to `fm/cli-parity-fix-defects`.
- [x] Direct PR [#4234](https://github.com/polymetrics-ai/cli/pull/4234) opened with `Refs #4015`.
- [x] An API-filtered `gh-axi pr list --base integration/4015-mvp-flat-r1 --head fm/cli-parity-fix-defects` returned exactly PR #4234.

## CI gap closure

- [x] The certification tests identify `projects_create_draft_item_for_authenticated_user` as the sole retired REST write action and assert its GraphQL replacement/blocked duplicate accounting.
- [x] `go test -timeout 20m ./internal/connectors/certify -count=1` passes (`9.458s`).
- [x] The connectorgen candidate, API ledger, and parity-completion tests reconcile the same REST-to-GraphQL transition by command and operation identity.
- [x] `go test -timeout 20m ./cmd/connectorgen -count=1` passes (`98.702s`).
- [ ] PR #4234's `verify` check passes on the pushed gap-fix commit.

## Verification evidence

- `go test -timeout 20m ./internal/state ./internal/connectors ./internal/connectors/engine ./internal/connectors/conformance ./internal/connectors/commandrunner -count=1`: PASS, including all 552 bundles.
- `go test -timeout 20m ./internal/app -count=1`: PASS (`392.483s`).
- `go test -timeout 20m ./internal/cli -count=1`: PASS on the final rebased/generated tree (`748.968s`).
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, and `make lint`: PASS (`0 issues`).
- `make docs-check-no-build`, `make smoke-no-build`, `make agent-contract-check`, `make connector-runtime-preflight`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, `make connectorgen-certification-sweep`, `make github-parity-artifacts-check`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check`: PASS.
- `website: pnpm typecheck`, `pnpm test:unit`, and `pnpm test:scripts`: PASS (80 unit tests and 28 script tests).
- `./pm help github`, `./pm github`, and `./pm github git blobs create --help`: PASS; the bare namespace exits successfully and generated help exposes the required body flags.
- `scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1`: PASS.
- `go test -timeout 20m ./internal/connectors/certify -count=1`: PASS (`9.458s`) after the CI inventory reconciliation.
- `go test -timeout 20m ./internal/connectors/engine -run '^TestGitHub(CorrectedCommandDeclarations|UserDraftCommandBuildsFixedGraphQLMutation)$' -count=1`: PASS.
- `go run ./cmd/connectorgen certification-matrix --check`, `go run ./cmd/connectorgen validate`, and `go run ./cmd/connectorgen surface-sync --check`: PASS after the CI inventory reconciliation.
- `go test -timeout 20m ./cmd/connectorgen -count=1`: PASS (`98.702s`) after reconciling the candidate, fixture, API-ledger, and completion projections.
- The aggregate `go test -timeout 20m ./...` and `make verify` commands were intentionally not run as single commands: repository instructions require changed packages plus `internal/cli` separately and the non-suite Make gates individually because the 550+ connector suite exceeds per-command budgets. The equivalent scoped suites and individual gates above were run.

## Manual code-review disposition

- Finding: the false-success accounting change made the closed issue-label cleanup's already-absent inverse fail its intentional idempotence contract. Fixed by accepting exactly one written-or-unchanged result in that cleanup only, emitting the unchanged count, and covering both app and binary CLI behavior.
- Finding: the initial user-project correction tightened `user_id` but left the live-proven 404 REST transport in place. Fixed by binding the command to the existing closed `addProjectV2DraftIssue` GraphQL operation, removing the invalid write action, classifying the source-lock REST row as the duplicate, and adding a command-builder regression.
- No remaining actionable findings after the generated-artifact review, scoped tests, boundary scan, static analysis, and help/docs checks.
