# Verification — issue #4342 binary upload CLI and certification foundation

## Planned gates

- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine` — pass (11.533s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner` — pass (39.064s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/certify` — pass (14.284s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` — pass after the final lint refactor (227.914s).
- [x] `GOFLAGS='-p=3' go test -timeout 20m ./internal/cli -run '^TestSkillsGenerateMatchesTrackedSkills$' -count=1` — pass (7.631s) after `GOFLAGS='-p=3' go run ./cmd/pm skills generate --dir docs/skills` regenerated the affected GitHub skill.
- [x] `GOFLAGS='-p=3' go run ./cmd/connectorgen validate internal/connectors/defs --json` — 552 connectors, 0 findings.
- [x] `GOFLAGS='-p=3' go run ./cmd/connectorgen surface-sync --check`, `operation-evidence --check`, `certification-candidates --connector github --check`, and `certification-sweep --connector github --check` — all current.
- [x] Generator/docs parity checks from `make verify`: `make docs-check`, `make connectorgen-certification-subject`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, `make connectorgen-certification-sweep`, `make github-parity-artifacts-check`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check` — pass.
- [x] `GOFLAGS='-p=3' go vet ./...`, `GOFLAGS='-p=3' go build ./cmd/pm`, `make tidy-check`, `make lint`, `make smoke-no-build`, and `make agent-contract-check` — pass.
- [x] GSD `verify-work` and `code-review` prompts generated and executed inline; no gap phase was needed. Manual review checked declaration-only source admission, non-executable `file_upload`, evidence classification by intent (not operation name), no direct upload dispatch, and the no-false-pass certification behavior.

## Non-green whole-suite result

- `GOFLAGS='-p=3' go test -timeout 20m ./internal/cli` was attempted and ran for 484.491s. It initially identified generated-skill drift, fixed by the documented skills generation above, but the package still exits non-zero because its broader runtime tests repeatedly dial unavailable local Redis endpoints `127.0.0.1:1` and `127.0.0.1:2`. No task code is changed to suppress those integration failures.
- `GOFLAGS='-p=3' go test -timeout 20m ./...` and aggregate `make verify` were deliberately not invoked as single commands: repository guidance says their 550+ connector suite routinely exceeds the agent command window on this memory-bound machine. All individual non-test gates and every changed-package suite ran; CI retains the aggregate suite.

## Main merge recheck

- Merged current `origin/main` cleanly before publication, then reran `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine`, `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner ./internal/connectors/certify`, and `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` — all passed. Regenerated `operation-evidence --write-fixed-100` for the merged declaration fingerprint; its check and the certification subject/matrix/candidates/sweep checks then passed.

## Post-PR website correction

- GitHub's `Website checks` and `Website generated data` initially failed because this intent changed three generated website artifacts. Ran `cd website && pnpm run gen:website-data`, then verified `pnpm run lint` (warnings only, exit 0), `pnpm run typecheck`, `pnpm run test:unit` (80 tests), `pnpm run test:scripts` (34 tests), and `pnpm run build` — all passed. The build emits existing Better Auth default-secret warnings while statically collecting pages but exits 0.

## Constraint

No credentialed provider call is authorized for this task. Tests must use the existing declaration-bound fixture/provider doubles and assert actual byte transfer through that real application path. A missing live candidate is `not_live`, not a passing transfer assertion.
