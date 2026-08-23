# Verification checklist — issue 4338

## Before push

- [x] Focused red test command is recorded before production implementation.
- [x] Focused green source-projection/coverage behavior tests pass.
- [x] Asana absent-action and Jira incomplete-contract cases are separately
  proven through the real projection and validation paths.
- [x] Complete mutation, read-only rejection, no-partial, and byte-stability
  controls pass.
- [x] `gofmt`, `go vet`, applicable `cmd/connectorgen` tests, and required
  generated/snapshot checks pass without increasing the 20-minute budget.
- [x] `git diff --check` passes; `internal/connectors/defs/**` and
  `internal/connectors/defs/github/rate_limits.json` are unchanged.
- [ ] `origin/main` merge is recorded immediately before push.
- [ ] Inline GSD code review has no unresolved actionable findings.

## PR read-back

- [ ] GitHub API reports PR base exactly `main`.
- [ ] PR body includes `Refs #4338`, GSD/TDD evidence, skills, verification,
  safety, scope, and review-route status.
- [ ] No merge is performed.

## Automated verification record (2026-08-24)

All commands below ran serially (one process at a time) without changing the
repository's shared 20-minute test budget.

- `GOCACHE=<isolated-cache> go test -count=1 -timeout 20m ./cmd/connectorgen`
  — pass.
- `GOCACHE=<isolated-cache> go test -count=1 -timeout 20m ./internal/cli`
  — pass.
- `GOCACHE=<isolated-cache> go vet ./...` — pass.
- `GOCACHE=<isolated-cache> go build ./cmd/pm` — pass.
- `make tidy-check`, `make lint`, `make agent-contract-check` — pass.
- `make connectorgen-validate`, `make connectorgen-surface-sync`, and
  `make connectorgen-operation-evidence` — pass.
- `make connectorgen-certification-subject`,
  `make connectorgen-certification-matrix`,
  `make connectorgen-certification-candidates`, and
  `make connectorgen-certification-sweep` — pass.
- `make connector-boundary`, `make connector-canon-check`, and
  `make connector-runtime-preflight` — pass.
- `make release-workflow-check`, `make docs-check`, and `make smoke-no-build`
  — pass.

`make verify` and `go test ./...` were deliberately not run as single commands:
the repository's AGENTS.md directs per-command-timeout agents to run the
affected packages plus `internal/cli` and `make verify`'s other gates
individually, because the full suite routinely exceeds a command wrapper.

## Inline UAT fallback

`scripts/gsd prompt verify-work 4338` was resolved on 2026-08-24. The
adapter's isolated Pi worker runtime is unavailable in this direct-PR session,
so UAT is recorded inline. All deliverables are automated and need no visual or
credentialed human judgment:

- source-cited mutation gaps: pass — focused behavioural suite exercises
  descriptor marshal, projection, and executable-coverage validation;
- closed safety boundary: pass — complete action, implemented incomplete
  command, read-only substitution, and GraphQL POST-query controls fail closed;
- non-regression: pass — existing GitHub projection stays byte-identical and
  generated/snapshot gates pass.
