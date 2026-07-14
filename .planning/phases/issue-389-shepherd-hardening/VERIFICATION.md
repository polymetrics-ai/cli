# Issue 389 Verification

## Focused gates

- [x] Prompt-advertised tools are a subset of the active unit registry. Covered by `internal/gsd/prompt_contract_test.go` and `make verify`.
- [x] Two issues cannot share one canonical Shepherd/GSD project identity. Covered by `internal/gsd/project_test.go`.
- [x] Same-issue restart adopts the exact existing identity. Covered by `internal/gsd/project_test.go` and `internal/store/store_test.go`.
- [x] Unit attempt budget survives process restart and exhausted retryable failures stop as typed `retry_exhausted` blocked state. Covered by `TestUnitAttemptBudgetSurvivesStoreRestart`, `TestFinalUnitRunStateBlocksWhenRetryBudgetExhausted`, and `TestFinalUnitRunStateRetriesWhileBudgetRemains`.
- [x] Signal reconciliation interrupts orphaned nested runs. Covered by `internal/gsd/subagents_test.go`.
- [x] Nested activity is visible through bounded heartbeats. Covered by existing runner/telemetry tests and `make verify`.
- [x] Success rejects missing artifacts, stale heads, unchanged canonical state, and live children. Covered by completion proof tests in `cmd/shepherd` and final module gates.
- [x] `supervise` dispatches the canonical sequence and stops at the final human gate. Covered by `internal/supervisor/policy_test.go` and CLI wiring in `cmd/shepherd/main.go`.
- [x] Planning and validation observe GPT-5.6 Sol/high; execution observes GPT-5.5/high. Covered by model selection tests, `TestRunnerCanDeriveGovernedImplementationModel`, and runtime identity validation in `cmd/shepherd`.
- [x] Runtime event and headless patch resource-shape failures preserve `runtime_contract_mismatch` typing. Covered by `internal/gsd/events_test.go` and `TestApplyPinnedHeadlessToolPatchReportsContractMismatch`.

## Pre-#390 implementation slice gates

- [x] Official GSD Pi 1.11 unit-registry metadata is parsed and used for model routing. Covered by `internal/gsd/registry_test.go`, `internal/supervisor/policy_test.go`, and `cmd/shepherd/main_test.go`.
- [x] Runtime prompt/tool compatibility rejects malformed or partial registry metadata. Covered by `internal/gsd/prompt_contract_test.go` and `internal/gsd/registry_test.go`.
- [x] `awaiting_decision` is a durable run state. Covered by `internal/domain` and `internal/store/decision_recovery_test.go`.
- [x] Durable decision requests survive restart, reject stale generation/head/expiry, and consume answers exactly once. Covered by `internal/store/decision_recovery_test.go`.
- [x] GitHub question comments are marker-owned, bounded, mention the configured human, and parse only authorized exact unedited replies. Covered by `internal/github/decision_broker_test.go`.
- [x] Recovery budgets are persisted per failure class. Covered by `internal/store/decision_recovery_test.go`.
- [x] Artifact proof and Sol/high attestation persistence reject moved heads and validator downgrades. Covered by `internal/store/proof_test.go`.
- [x] Disposable attempt worktree primitives reject stale head and out-of-scope promotion. Covered by `internal/workspace/worktree_test.go`.
- [x] Path boundary checks reject symlink final components and canonicalize missing-path ancestors before comparing work/state roots. Covered indirectly by existing config/runtime tests.

## Module gates

- [x] `gofmt -w cmd internal` — PASS
- [x] `go test ./...` — PASS
- [x] `go test -race ./...` — PASS
- [x] `go vet ./...` — PASS
- [x] `go build ./cmd/shepherd` — PASS
- [x] `make verify` — PASS
- [x] `cd ../.. && go list ./...` — PASS
- [x] `scripts/tests/shepherd-module-boundary.sh` — PASS
- [x] `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify` — PASS

## Notes

- No commits, pushes, GitHub mutations, credential reads, or merge actions were performed.
- Broader root test gates were not run because the changes are isolated to the nested Shepherd module and phase artifacts.
- Read-only reviewer subagent flagged retry exhaustion and runtime-contract typing gaps; both were fixed and covered by focused tests.
- Final local reviewer pass found no critical findings and flagged two stale `.gsd/phases/01-m001/01-CONTEXT.md` remote-review references; both were updated to local review policy.
- Repository policy now treats automated review coverage as local reviewer/verifier/security evidence; remote PR-bot review is not required by default.
- Pre-#390 local reviewer/security passes found no credential exposure or critical issue. Follow-up wiring added host disposable worktrees, GitHub reply consumption, and per-class recovery budget use in the real `supervise`/`runHeadless` path.
- Follow-up completed the prior pre-canary blockers: real artifact manifest hashing from scoped changed files plus official unit metadata, outbox claim/execute/mark-sent for GitHub questions, and a fake-runtime `shepherd supervise` integration test to `final_human_gate`.
- Remaining before parent PR #390 merge: run merge-disabled Twenty and Asana canaries to `final_human_gate`, then perform only post-canary Podman cleanup and `.gsd/.planning` migration if approved.
