# Issue 389 TDD Ledger

## Baseline

- Command: `scripts/gsd doctor && scripts/gsd prompt programming-loop init --phase issue-389-shepherd-hardening --dry-run`
- Result: PASS; adapter resources resolved from current repo-local `.gsd` prompt surface.
- Command: `cd agent-runtime/shepherd && go test ./...`
- Result: PASS in the reconciled worktree before the final fix slice.

## Runtime contract admission

- RED: Existing focused tests in `internal/gsd/prompt_contract_test.go` cover forbidden advertised tools and pinned prompt patch failure modes. Added `TestApplyPinnedHeadlessToolPatchReportsContractMismatch` so headless patch qualification errors must preserve `runtime_contract_mismatch` typing.
- GREEN: Wrapped pinned headless resource shape/version failures and projected event contract failures with `gsd.ErrRuntimeContractMismatch`; focused `go test ./internal/gsd -run 'TestRunnerCanDeriveGovernedImplementationModel|TestApplyPinnedHeadlessToolPatchReportsContractMismatch|TestApplyPinnedHeadlessToolPatchIsExactAndIdempotent' -count=1` passed.
- Refactor: Kept compatibility checks version-qualified and side-effect free; final module gates passed.

## Issue identity and bootstrap

- RED: Existing focused tests in `internal/gsd/project_test.go` cover atomic `.gsd/ISSUE.json`, same-issue restart adoption, and cross-issue identity rejection.
- GREEN: Reused `ensureIssueDelivery` in `supervise` so the validated context is materialized once and immutable Shepherd/GSD identity is adopted or bootstrapped before any dispatch.
- Refactor: Kept Shepherd SQLite as controller truth and native `.gsd/ISSUE.json` as issue-local GSD identity proof.

## Durable attempts and lifecycle

- RED: Existing interrupted prompt recorded missing `classifyUnitFailure`/`isAutomaticallyRetryable` and a missing delivery fixture for `TestUnitAttemptBudgetSurvivesStoreRestart`; reviewer then found exhausted retryable attempts could leave `RunFailed` instead of typed `retry_exhausted` blocked state.
- GREEN: Added typed failure classification/retry policy, initialized the canonical delivery before durable unit-attempt rows, and added `finalUnitRunState` tests proving retryable failures return `RunReady` only while budget remains and become `RunBlocked` with `store.ErrRetryBudgetExhausted` at exhaustion.
- Refactor: Added table-driven classification coverage for retryable runtime/artifact/interruption failures and fail-closed contract, stale-head, scope, model, orphan-child, and exhausted-budget failures.

## Completion proof

- RED: Existing completion proof tests cover missing artifacts, unchanged canonical state, stale heads, model drift, write-scope breach, and live children.
- GREEN: Preserved the exact completion proof path and mapped proof failures into bounded typed classes.
- Refactor: `targetRunState` continues to convert canonical completion into `RunHumanGate`; final verification passed.

## One-command supervision

- RED: `supervise` was absent from CLI usage/dispatch and lacked canonical policy coverage; added `TestRunnerCanDeriveGovernedImplementationModel`, which failed to compile until the runner could derive a GPT-5.5 execution runner from the coordinator runner.
- GREEN: Added `internal/supervisor` policy tests, CLI wiring for `shepherd supervise --config <absolute-config> --issue <N> --context <validated-context.json>`, per-unit supervised model derivation, and bounded blocked status emission for non-retryable failures.
- Refactor: Kept command selection in a small policy package: dispatch only canonical units, map `discuss-milestone` to targeted `discuss`, stop unsafe `skip`, avoid subagent model events overwriting top-level model proof, and emit final human-gate status at `phase=complete,next.action=stop`.

## Policy review routing cleanup

- RED: `rg 'claude-review-loop|@claude|claude-review.yml|GitHub Copilot review|Copilot backup|Claude automatic review|Claude review' AGENTS.md CLAUDE.md .agents .github .codex .opencode .pi website/content docs/plans` found active GitHub-hosted review requirements.
- GREEN: Replaced active policy with local automated review routing, local review loop, local review disposition agents/prompts, and removed the GitHub Claude review workflow/rubric.
- Refactor: Historical planning artifacts remain historical; active contracts, workflows, templates, runtime adapters, and website source now avoid GitHub-hosted review as a default gate.

## Final verification evidence

- PASS `gofmt -w cmd internal`
- PASS `go test ./...`
- PASS `go test -race ./...`
- PASS `go vet ./...`
- PASS `go build ./cmd/shepherd`
- PASS `make verify`
- PASS `cd ../.. && go list ./...`
- PASS `scripts/tests/shepherd-module-boundary.sh`
- PASS `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify`
