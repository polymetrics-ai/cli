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

## Pre-#390 implementation slice — registry, decisions, proof, recovery, workspace primitives

- RED: added official unit-registry parsing/routing tests for GSD Pi 1.11 metadata, malformed/partial registry rejection, and symlinked registry rejection.
- GREEN: added `internal/gsd` unit registry loader, metadata-driven command/model routing, prompt-contract registry validation, and supervise/run rebinding to the canonical unit's expected model before execution.
- RED: added durable decision request tests for restart survival, exact-once consumption, stale generation/head/expiry rejection, and GitHub reply filtering for unauthorized/bot/edited/malformed replies.
- GREEN: added `RunAwaitingDecision`, durable `decision_requests`, marker-owned GitHub question comments, exact reply parsing, and retry-exhaustion transition to `awaiting_decision` with a durable GitHub question publication path.
- RED: added recovery-budget tests proving per-class persistence across restarts and durable exhaustion.
- GREEN: added `recovery_budgets` keyed by failure class, generation, unit, and head with bounded failure/recovery-plan evidence.
- RED: added artifact proof and attestation persistence tests for exact-head validation and validator downgrade rejection.
- GREEN: added durable artifact proof and Sol/high attestation store APIs.
- RED: added disposable attempt worktree tests for promotion/discard, stale-head rejection, and out-of-scope rejection.
- GREEN: added `internal/workspace` attempt worktree manager with owned attempt roots, branch isolation, promote/discard semantics, and scope/head checks.
- Refactor/security: canonicalized path boundary checks, verified SQLite WAL mode, added delivery-run column migration, prevented child model events from overwriting first observed model proof, and bounded/redacted decision evidence before GitHub publication.
- GREEN wiring follow-up: host-runtime canonical `runHeadless` now creates disposable attempt worktrees, runs the unit in the attempt worktree, promotes only scoped output, and fails closed for Podman canonical units until equivalent isolation exists.
- GREEN wiring follow-up: retryable failures now update per-class recovery budgets and use that remaining budget for retry versus durable `awaiting_decision`.
- GREEN wiring follow-up: `supervise` now polls GitHub decision replies for open requests and resumes accepted `retry`/`continue` replies or blocks accepted `stop` replies.
- GREEN wiring follow-up: successful canonical units now build an artifact manifest from scoped changed files between exact start/end heads, bind it to official GSD unit metadata (`phase_chain` and `required_workflow_tools`), hash it, and persist Sol/high artifact-proof plus attestation records.
- GREEN wiring follow-up: GitHub question publication now goes through outbox enqueue, claim, execute, and mark-sent/failed semantics before the marker-owned comment is considered published.
- GREEN wiring follow-up: added a full fake-runtime `shepherd supervise` test that drives the real supervise loop through an execution unit in a disposable worktree and reaches `final_human_gate`.
- Local review: reviewer/security subagents found missing production wiring; follow-up passes wired host worktrees, reply consumption, recovery budget use, real artifact manifests, outbox claim/mark-sent, and fake-runtime final-gate integration. Remaining work is merge-disabled Twenty/Asana canaries and post-canary cleanup/migration.
- RED canary: merge-disabled Asana canary stalled in `research-slice/M001/S03`; attempt worktree started without the canonical `.gsd` workflow state, then official GSD emitted `Cannot dispatch: no active milestone` while Shepherd heartbeats stayed alive with no model/tool/child activity. The run was stopped and does not count as acceptance evidence.
- GREEN canary fix: attempt worktrees now copy canonical `.gsd` state before dispatch, run a pre-dispatch attempt query that must match the canonical milestone/phase/unit, adopt successful attempt `.gsd` state back after scoped promotion, ignore `.gsd` workflow state during source promotion, and hash official `.gsd` state artifacts for proof when code head is unchanged.
- GREEN stall/retry fix: Runner now fails startup with a typed runtime-contract error when no model/tool/child activity appears before the startup deadline, and unit retry budgets now use the configured `max_unit_attempts` directly instead of multiplying by 20.

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
- PASS `cd agent-runtime/shepherd && go test ./...`
- PASS `cd agent-runtime/shepherd && go vet ./... && go test -race ./... && go build ./cmd/shepherd && make verify`
- PASS `cd agent-runtime/shepherd && go test ./cmd/shepherd -run TestSuperviseFakeRuntimeToFinalHumanGate -count=1 -v`
- PASS `cd agent-runtime/shepherd && go test ./... && go vet ./... && go test -race ./... && go build -o shepherd ./cmd/shepherd && make verify`
