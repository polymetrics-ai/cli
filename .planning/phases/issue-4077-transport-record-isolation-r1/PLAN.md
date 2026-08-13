---
phase: issue-4077-transport-record-isolation-r1
plan: 01
type: tdd
---

# #4077 — Transport record-isolation correction plan

**Issue:** [#4077](https://github.com/polymetrics-ai/cli/issues/4077)
**Base:** `c67f40a5ff67a131950f3123e70527027dca8493` from
`feat/3862-any-to-any-transport`
**Child branch:** `fix/4077-transport-record-isolation`

## Manual GSD lifecycle

The required prompts were resolved through `scripts/gsd prompt`. The named issue phase is absent
from the numeric roadmap and the current runtime cannot provide compatible isolated roles; the
repository's single-worker contract independently forbids role spawning. Execute the lifecycle
inline and record it here:

1. `scripts/gsd prompt discuss-phase issue-4077-transport-record-isolation-r1 --auto` — recorded
   in `CONTEXT.md`, `DISCUSSION-LOG.md`, and `PROBLEM.md`.
2. `scripts/gsd prompt plan-phase issue-4077-transport-record-isolation-r1 --tdd` — this plan and
   `TDD-LEDGER.md`.
3. `scripts/gsd prompt execute-phase issue-4077-transport-record-isolation-r1` — record RED,
   GREEN, refactor, and commits in `EXECUTION.md`.
4. `scripts/gsd prompt verify-work issue-4077-transport-record-isolation-r1` — record
   goal-backward verification in `VERIFICATION.md` and `UAT.md`.
5. `scripts/gsd prompt code-review issue-4077-transport-record-isolation-r1` — record findings and
   dispositions in `REVIEW.md`.

## Required skills

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-safety`, `golang-security`, `golang-testing`,
`github-issue-first-delivery`, `gsd-discuss-phase`, `gsd-plan-phase`,
`gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, and `no-mistakes`.
`golang-lint` will be loaded before the final review. CLI/help/manual/website parity is not
applicable because no public command, output, docs, generated surface, or website artifact changes.

## TDD execution slices

### 1. Planning checkpoint — complete before production edits

- Preserve exact-head reproduction, causal analysis, scope fence, lifecycle fallback, skills, and
  verification plan in this phase directory.
- Commit only GSD evidence.

### 2. RED — focused isolation regression

- Extend `internal/synctransport/transport_test.go` only.
- Prove direct `cloneRecord` isolation for `json.RawMessage` and `map[string]string`, including
  nested values inside existing `map[string]any`, `[]any`, and `[]connectors.Record` containers.
- Preserve passing controls for `[]byte`, `map[string]any`, `[]any`, and `[]connectors.Record`.
- Prove a mutating `WarehouseStage` cannot change source raw JSON and a mutating
  `DestinationExecutor` cannot change source string-map storage.
- Prove an unknown mutable value (for example `map[string]int`) fails before it can reach the stage
  or destination; a stage-injected unknown fails before destination application.
- Run the narrow selector. It must fail for aliasing/unsafe-forwarding behavior, not compilation.
- Commit this test-only RED plus ledger evidence before touching `types.go` or `orchestrator.go`.

### 3. GREEN — explicit copy/rejection contract

- In `internal/synctransport/types.go`, add explicit cloning for `json.RawMessage` and
  `map[string]string` while preserving existing recursive cases.
- Make the closed clone helpers return contextual errors for non-enumerated values instead of
  forwarding `any` values by alias. Preserve scalar values only when they are explicitly immutable
  JSON-like types; do not use reflection or add a new provider value schema.
- Make the minimal `internal/synctransport/orchestrator.go` changes required to stop before
  `WarehouseStage.Stage` or `DestinationExecutor.ApplyDestination` when copy validation fails.
- Keep checkpoint, acknowledgement, CAS, source/destination dispatch, and result accounting
  unchanged. Do not add retry, normalization, provider, or registry behavior.
- Run the RED selector until green, then package tests and race selector.

### 4. Refactor and verification

- Use `gofmt` on changed Go files. Keep error strings lowercase and contextual; do not panic or
  drop values.
- Rerun focused normal/race tests, all `internal/synctransport`, Transport-focused `internal/app`,
  and the all-seven canonical-mode selector.
- Run `go vet ./...`, `go build ./cmd/pm`, `scripts/verify-gsd-workflow` against the parent base,
  and the repository's split non-mutating gates. Do not run credentialed/live checks or a warehouse
  mutation smoke; record why they cannot prove this boundary.
- Complete manual verify-work and standard-depth code review. Run no-mistakes without `--yes`, no
  more than five loops, and with the canonical `--skip=push,pr,ci` child topology. Stop at
  `needs-decision` before any manual delivery exception.

## Commit checkpoints

| Checkpoint | Allowed paths | Evidence |
|---|---|---|
| Plan | `.planning/phases/issue-4077-transport-record-isolation-r1/**` | GSD context, plan, diagnosis, initial ledger |
| RED | `internal/synctransport/transport_test.go` + phase artifacts | exact source/stage/destination failure; committed before code |
| GREEN | `internal/synctransport/types.go`, `internal/synctransport/orchestrator.go`, focused test, phase artifacts | normal/race focused tests pass |
| Review/fix | only files identified by in-scope gate findings | review disposition and rerun evidence |

## Verification plan

```text
go test -timeout 20m ./internal/synctransport
go test -race -timeout 20m ./internal/synctransport
go test -timeout 20m ./internal/app -run '<Transport canonical-mode selector>'
go vet ./...
go build ./cmd/pm
scripts/verify-gsd-workflow c67f40a5ff67a131950f3123e70527027dca8493
make tidy-check
make lint
make docs-check
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```

`make verify` and `go test ./...` are intentionally split under the repository per-command-timeout
rule. Live GitHub/PostgreSQL credentials cannot prove a clone that happens entirely before any
provider/database call; no provider E2E claim will be made.
