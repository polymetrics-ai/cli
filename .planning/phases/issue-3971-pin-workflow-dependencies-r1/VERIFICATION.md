# Verification checklist: pin build dependencies

## Pre-edit research

- [x] Alert #135 remediation guidance read from GitHub REST API.
- [x] Alert location confirmed: `.github/workflows/verify.yml:86`.
- [x] All workflow `uses:` references, Dockerfile `FROM` lines, and literal workflow images inventoried locally.
- [x] Current action commits and image manifest-list digests captured in `TDD-LEDGER.md` before edits.

## Targeted gates

- [x] Red: `./scripts/tests/pinned-build-dependencies.sh` failed against the baseline and named mutable action and image references.
- [x] Green: `make pinned-build-dependencies-check` passes after pinning.
- [x] `make release-workflow-check` passes.
- [x] Workflow YAML parser check passes for every `.github/workflows/*.{yml,yaml}` file as part of the focused gate.
- [x] `git diff --check` passes.
- [x] `go run ./cmd/agentcontractgen check` passes.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` pass individually where applicable.
- [x] `go build ./cmd/pm`, `go vet ./...`, `bash -n scripts/tests/pinned-build-dependencies.sh`, and `shellcheck scripts/tests/pinned-build-dependencies.sh` pass.
- [x] `docker build --check --file website/Dockerfile website` accepts the digest-pinned Dockerfile (it reports only one pre-existing non-blocking advisory outside this scope).
- [x] `scripts/verify-gsd-workflow` reports no `cmd/` or `internal/` implementation changes, while the phase evidence remains present for this manual-GSD fallback.

## Remote follow-up (last, after local work)

- [x] Search GitHub for an existing equivalent issue before creating a child of #3971; queries for `PinnedDependencies` and `pin workflow dependencies` returned no matching issue.
- [x] Create focused child [#3986](https://github.com/polymetrics-ai/cli/issues/3986) and attach it to #3971.
- [x] Re-read alert #135 and inventory the other four open alerts after GitHub REST quota permitted: #134 at `verify.yml:81`, #133 at `release.yml:480`, #132 at `release.yml:422`, and #131 at `release.yml:415`; all are the same Scorecard `PinnedDependenciesID` rule and are covered by this same change.
- [x] State alert closure honestly: #135 remains open on `main` at the pre-change commit `f96a47e801b89f25386c33951a53a93d1a4c7c8d`; it cannot close until this branch is pushed, merged, and Scorecard analyzes the default branch.

## Post-rebase follow-up

- [x] Rebases cleanly onto `origin/main` at `4df0b0416` after resolving the `Makefile` add/add target list by retaining both `github-parity-artifacts-check` and `pinned-build-dependencies-check`.
- [x] Re-audit catches and fixes the two mutable refs newly introduced by #3970 in `.github/workflows/github-source-drift.yml`; no new Dockerfile or build-image manifest was introduced.

## Not applicable

- CLI help, manual, and website documentation parity: no CLI surface or user-facing documentation changes.
- Full `go test ./...`: not required for a workflow/image-only change; the task-specific and existing `make verify` component gates are the local evidence.
