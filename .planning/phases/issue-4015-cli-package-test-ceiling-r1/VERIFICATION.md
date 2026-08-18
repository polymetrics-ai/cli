# Verification — CLI package test-ceiling foundation

## Required checks

- [x] Baseline test inventory: `go test -list '.*' ./internal/cli` (263 runnable names).
- [x] Baseline verbose timing: `/usr/bin/time -p go test -v -timeout 30m ./internal/cli` (627.73s wall; 623.128s package).
- [x] Focused red test demonstrates non-shared binary paths before fixture implementation (`red-fixture-identity.log`).
- [x] Focused fixture test passes after the implementation (`green-fixture-identity.log`).
- [x] Changed inventory equals the baseline inventory exactly (263 runnable names).
- [x] Changed verbose package duration is 532.694s / 537.29s wall, a 14.5% local reduction. Final rebased normal-topology evidence is stronger: `make verify` reports 847.535s (29.4% below 1200s), without a timeout change.
- [x] `go test -timeout 20m ./internal/cli` through final rebased `make verify` (standard aggregate run reports 847.535s).
- [x] `go test -timeout 20m ./cmd/connectorgen` (pass, 94.783s).
- [x] Repository verification gates and generated-file checks: `make tidy-check`, `go vet ./...`, `go build ./cmd/pm`, `make docs-check-no-build`, `make smoke-no-build`, `make lint`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-sweep`, `make connector-boundary`, `make connector-canon-check`, `make release-workflow-check`, and the aggregate `make verify` all passed. `pnpm --dir website run gen:docs` passed twice and `git diff --exit-code -- website` confirmed byte stability.
- [x] Static diff review and inline GSD `verify-work` evidence; `git diff --check` passed.
- [x] GSD `code-review` completed inline; `REVIEW.md` records no actionable finding and the focused race proof passed in 61.353s.
- [ ] Commit/push and PR API base read-back remain.

## Explicit non-applicable checks

- No documentation source is changed, so `gen:docs` is a deterministic generated-file guard rather than a documentation update.
- No credentialed connector run, reverse-ETL execution, or external write is authorized or required.
- `security/snyk` is not run locally because the task records its identical base-branch failure as pre-existing and out of scope.
