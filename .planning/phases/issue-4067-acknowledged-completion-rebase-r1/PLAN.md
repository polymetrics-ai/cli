---
phase: issue-4067-acknowledged-completion-rebase-r1
plan: 01
type: tdd
---

# #4067 — acknowledged transport completion rebase plan

**Issue:** [#4067](https://github.com/polymetrics-ai/cli/issues/4067)  
**Parent chain:** #3864 → #3862 → #4015  
**Stacked PR:** #4059 → `feat/3862-any-to-any-transport`  
**Starting rejected candidate (preserved):** `883a86cf0040d559edcd4777413d1c2de20cd94a`

## Manual GSD lifecycle

The issue phase is absent from the archived numeric roadmap and role spawning is prohibited by the repository contract. The following inline/manual fallback preserves the required lifecycle:

1. `scripts/gsd prompt discuss-phase issue-4067-acknowledged-completion-rebase-r1 --auto` — decisions recorded in `CONTEXT.md` and `DISCUSSION-LOG.md`.
2. `scripts/gsd prompt plan-phase issue-4067-acknowledged-completion-rebase-r1 --tdd` — this TDD plan.
3. `scripts/gsd prompt execute-phase issue-4067-acknowledged-completion-rebase-r1` — execution record and RED/GREEN commits in `EXECUTION.md`.
4. `scripts/gsd prompt verify-work issue-4067-acknowledged-completion-rebase-r1` — goal-backward verifier evidence in `VERIFICATION.md` and `UAT.md`.
5. `scripts/gsd prompt code-review issue-4067-acknowledged-completion-rebase-r1` — finding dispositions in `REVIEW.md`.

## Required skills

`golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-context`, `golang-concurrency`, `golang-safety`, `golang-security`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-cli`, `github-issue-first-delivery`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, and `no-mistakes`.

## Objective

Make an acknowledged transport run complete truthfully despite an unrelated post-checkpoint writer, without permitting checkpoint overwrite, destination replay, or generic state refresh.

## RED — deterministic durable completion leak

1. Locate the closest real persisted two-App fixture in `internal/app/transport_dispatch_test.go`.
2. Add a test that pauses the first App after its checkpoint acknowledgement, lets a second App persist only unrelated state, then releases ordinary successful completion.
3. Before production mutation, assert the exact red symptom: returned `Run{}`/non-terminal identity, ordinary revision-conflict error chain, and after reopen the generated run is still `running`; establish first that the acknowledged stream and unrelated writer state are unchanged.
4. Parameterize the exact same interleaving across all seven canonical modes.
5. Run the focused test with a 20-minute timeout. It must fail for the durable symptom, not compilation or an assertion that only checks an error string.
6. Commit the behavioral RED and ledger evidence before a production edit.

## GREEN — strict completion rebase

1. Trace `RunETL → runTransportETL → completeRun` and bind final completion to the precise checkpoint already acknowledged by this run.
2. In the normal completion path only, read latest locked state and terminalize only if the current target run is `running` and current target stream is exactly the acknowledged checkpoint.
3. Change only the target run's terminal fields and the run's own final stream/run metadata. Retain all other current state directly from the locked latest state.
4. Preserve ordinary revision guard behavior for any completion not meeting both eligibility predicates; do not alter #4046's typed failure handling.
5. Preserve state-store commit outcome truth: definite-not-committed cannot return a speculative terminal run; committed or indeterminate outcomes return a durable-consistent terminal identity; error identity remains detectable.
6. Do not modify checkpoint CAS/retry or destination-apply behavior. If a minimal metadata propagation is demonstrably required, keep it transport-finalization-local and prove it does not change R7/R8 semantics.
7. Re-run the RED test to green and commit the smallest coherent implementation and evidence update.

## Focused proof expansion

- Run/reopen proof in all seven modes.
- Cancellation delivered after checkpoint acknowledgement but before completion; it must retain ordering and not create a false terminal result.
- Target-stream changed/missing/non-running cases must fail closed and preserve current state.
- Distinct write after checkpoint acknowledgement must retain winner checkpoint and all unrelated state byte/semantically unchanged.
- Returned run must match reopened durable terminal run and error chain must remain detectable.
- Existing #4046 typed-conflict tests and R7/R8 identity/CAS suite must remain green.
- Run focused interleaving tests under `-race` before heavier validation.

## Generated-artifact remediation

Only after focused behavior is green, discover and invoke the canonical generators for:

1. `website/lib/docs.generated.ts`;
2. `internal/connectors/certifications/flow-matrix.json`.

Use their documented check commands after generation. Inspect the resulting diff and retain only candidate-owned output. Never hand-edit either generated file.

## Ordered validation and delivery

1. Focused RED → matching GREEN → repeat/interleaving/reopen/cancellation/all-seven-mode/race/R7-R8 tests.
2. Before resource-heavy full repository validation, report exactly: `working: transport focused gates green; requesting heavy validation window`.
3. Run GSD verifier, lint, generator checks, affected package tests, vet/build, and the required individual repository gates.
4. Run manual GSD code review and disposition every finding.
5. Confirm no active competing no-mistakes run, then start a fresh #4067 0/5 run without `--yes`; follow every synchronous return. Do not touch the immutable old run.
6. Push only the existing branch normally and update only draft #4059. Do not merge, retarget, force-push, or open another PR.
7. Wait for exact-head CI. On all green, report the immutable candidate SHA and request a fresh independent Sol audit.

## Success criteria

- RED reproduces zero returned run and durable reopened `running` leak after a real acknowledged checkpoint plus unrelated writer.
- GREEN returns a terminal run matching the durable reopened record only where latest target run remains `running` and latest target stream equals the acknowledged checkpoint.
- Winner/acknowledged stream and unrelated state survive; no destination replay, checkpoint overwrite, or generic last-writer-wins path appears.
- All seven modes, cancellation, restart/reopen, race, #4046, and R7/R8 focused evidence pass.
- Generated outputs are canonical-generator-only; no separate issue is created for them.
- Fresh no-mistakes ledger remains at most 5 loops; #4059 stays draft and unmerged.

