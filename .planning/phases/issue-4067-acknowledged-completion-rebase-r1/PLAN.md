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
**R2 rejected candidate (preserved):** `3f84693bfbc128523a66e22653db7227fb9c0869`

## Manual GSD lifecycle

The issue phase is absent from the archived numeric roadmap and role spawning is prohibited by the repository contract. The following inline/manual fallback preserves the required lifecycle:

1. `scripts/gsd prompt discuss-phase issue-4067-acknowledged-completion-rebase-r1 --auto` — decisions recorded in `CONTEXT.md` and `DISCUSSION-LOG.md`.
2. `scripts/gsd prompt plan-phase issue-4067-acknowledged-completion-rebase-r1 --tdd` — this TDD plan.
3. `scripts/gsd prompt execute-phase issue-4067-acknowledged-completion-rebase-r1` — execution record and RED/GREEN commits in `EXECUTION.md`.
4. `scripts/gsd prompt verify-work issue-4067-acknowledged-completion-rebase-r1` — goal-backward verifier evidence in `VERIFICATION.md` and `UAT.md`.
5. `scripts/gsd prompt code-review issue-4067-acknowledged-completion-rebase-r1` — finding dispositions in `REVIEW.md`.

On the r2 continuation, the `discuss-phase --auto` and `plan-phase --tdd` prompts were resolved again after #4067 was amended and the complete Sol r2 handoff was read. Both official phase lookups report `phase_found: false` for this named issue phase; the documented inline/manual fallback therefore updates this existing phase directory rather than spawning a role or creating a numeric-roadmap phase.

On the r3 continuation, the same five generated prompts were resolved after the complete Sol r3 report, current loop-4 directive, and live #4067 amendment/readback. This workspace is not a Pi runtime and the named phase is not a numeric roadmap phase; role spawning is forbidden in this custody lane. The lifecycle therefore remains an inline/manual fallback with the same required discussion, plan/TDD, execution, verify-work, and review evidence.

## Required skills

`golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-context`, `golang-concurrency`, `golang-safety`, `golang-security`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-cli`, `github-issue-first-delivery`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, and `no-mistakes`.

The r3 continuation additionally loaded `golang-code-style` and `golang-lint`; no CLI command, help, public documentation, connector definition, or generated surface is planned to change, so CLI/help parity work is explicitly not applicable to the production slice.

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

## R2 RED — post-acknowledgement error finalization and missing target

1. Before a production edit, add a persisted two-App all-seven-mode test that pauses after the real checkpoint callback, lets a second App persist only an unrelated stream/checkpoint/run, cancels the source context, and releases the source.
2. The RED must fail by observing the actual durable symptom: `RunETL` returns `Run{}` while `errors.Is(err, context.Canceled)` remains true and reopen finds the original run still `running`. Before asserting that symptom, prove the acknowledged target stream and unrelated write are preserved and the destination applied exactly once.
3. Add one representative post-ack source-error witness with the same unrelated revision, preserving a sentinel error identity, one apply, exact state preservation, returned/reopened failed-run identity, and no replay.
4. Add an independent all-seven-mode completion witness that removes the exact target run after acknowledgement. It must initially fail because the current error does not satisfy `errors.Is(err, errStateRevisionConflict)`; it also proves `Run{}`, one apply, and unchanged reopened state.
5. Commit the behavioral RED and its ledger evidence before touching `internal/app/app.go` or `internal/app/transport_dispatch.go`.

## R2 GREEN — smallest acknowledged-error repair

1. In `runTransportETL`, preserve a non-zero `etlExecutionResult` only if a real durable checkpoint acknowledgement exists before the orchestrator returns its error; pre-ack failures still return zero result.
2. In the declared-transport `RunETL` error branch, use a dedicated acknowledged-failure finalizer only when that result carries the post-ack witness. All other errors continue through unmodified ordinary `failRun`.
3. The dedicated finalizer reads locked current state through the existing state update path and mutates only the matching target run to `failed` when the current run is still `running` and the current stream exactly equals the captured acknowledged witness. It retains the acknowledged stream/checkpoint and every unrelated entry, returns the original error chain, and honors definite/committed/indeterminate outcomes without speculative runs.
4. In `completeRunWithAcknowledgedTransportState`, wrap a missing target run as `errStateRevisionConflict` only when the rebase is an acknowledged rebase; preserve ordinary missing-run behavior elsewhere.
5. Do not alter `failRun`, checkpoint CAS, destination apply, #4046 typed-conflict terminalization, or source identity. Re-run the exact RED commands to GREEN before broader coverage.

## R3 RED — stale second-page writer after an earlier acknowledgement

1. Add `TestRunETLTransportAcknowledgedPageThenStaleSecondPageFinalizesLosingRunForAllModes` to `internal/app/transport_dispatch_test.go` using the existing real JSON-state fixture and all seven `synccontract.AllModes()` values.
2. Configure a two-page loser. Pause only after its first real checkpoint callback returns; use a separately opened App to persist unrelated state and run a real one-page winner that advances the same target stream.
3. Release the loser. It must apply its second page exactly once and then receive `errTransportStreamStateConflict` from the existing per-stream CAS. Before production mutation, the test must expose the durable symptom: typed error but `Run{}` returned and a reopened loser still `running`.
4. In every subtest, assert page-one acknowledgement, winner checkpoint/run identity, unrelated stream/checkpoint/run preservation, exactly two loser applies (and one winner apply), non-overwrite, no retry/replay, and a truthful returned/reopened failed loser as the intended green contract.
5. Run the selector with a 20-minute timeout. Its RED must fail on the observable zero-run/durable-running condition, never compilation or an error-string-only assertion. Commit test and ledger evidence before `internal/app/app.go` changes.

## R3 GREEN — typed-conflict-only finalization ordering

1. At the beginning of `failAcknowledgedTransportRun`, route `errors.Is(runErr, errTransportStreamStateConflict)` directly to `a.failRun(runID, runErr)` before examining `result.PendingStreamState` or `a.state`.
2. Change no other production behavior: ordinary acknowledged cancellation/source-error handling retains the r2 exact stream and running-run guard; #4046 remains the sole typed-conflict latest-state terminalization owner.
3. Re-run the identical all-seven witness green, proving the returned failure, reopened durable failure, typed error preservation, winner/unrelated preservation, and two loser applies.
4. Do not retry/overwrite a checkpoint, replay destination work, introduce a state refresh, touch R7/R8, or expand production registration.

## Focused proof expansion

- Run/reopen proof in all seven modes.
- Cancellation delivered after checkpoint acknowledgement but before completion; it must retain ordering and not create a false terminal result.
- Target-stream changed/missing/non-running cases must fail closed and preserve current state.
- Distinct write after checkpoint acknowledgement must retain winner checkpoint and all unrelated state byte/semantically unchanged.
- Returned run must match reopened durable terminal run and error chain must remain detectable.
- Existing #4046 typed-conflict tests and R7/R8 identity/CAS suite must remain green.
- Run focused interleaving tests under `-race` before heavier validation.
- Repeat the r2 all-mode cancellation/unrelated, source-error, and missing-run selectors under `-race`; assert `applyCalls == 1` in every witness.
- Repeat the new r3 two-page all-mode selector at least three times normally and at least three times under `-race`; its loser-specific `applyCalls == 2` is the no-replay assertion.
- Treat deterministic core proof as primary. Separately build the real binary, run current GitHub definition/hook tests and command-runner runtime-preflight tests, inspect current production registration, and use only any repository-owned bounded GitHub harness valid at this branch.
- A credentialed GitHub smoke is conditional on an already-approved secret channel. It is bounded and read-only only; no token is copied to a command line, file, output, or evidence. Absence of that channel is a recorded limitation, not an excuse to fabricate an integration result.

## Generated-artifact remediation

Only after focused behavior is green, discover and invoke the canonical generators for:

1. `website/lib/docs.generated.ts`;
2. `internal/connectors/certifications/flow-matrix.json`.

Use their documented check commands after generation. Inspect the resulting diff and retain only candidate-owned output. Never hand-edit either generated file.

## Ordered validation and delivery

1. Focused r3 planning → behavioral RED commit → matching GREEN → repeat/interleaving/reopen/cancellation/source-error/missing-run/all-seven-mode/race/R7-R8 tests.
2. Before resource-heavy full repository validation, report exactly: `working: transport focused gates green; requesting heavy validation window`.
3. Run GSD verifier, lint, generator checks, affected package tests, vet/build, and the required individual repository gates.
4. Run manual GSD code review and disposition every finding.
5. Confirm no active run on this branch, then use only loop 4/5 for #4067 without `--yes`; follow every synchronous return. Do not touch the immutable old run or an unrelated branch run.
6. Current no-mistakes help lacks a safe existing-stacked-PR route. Use local validation with push/PR/CI skipped; also skip document rather than retaining another unrelated architecture rewrite. Stop for Firstmate after the clean submitted local head. Do not merge, retarget, force-push, push, or open another PR.

## Success criteria

- RED reproduces zero returned run and durable reopened `running` leak after a real acknowledged checkpoint plus unrelated writer.
- GREEN returns a terminal run matching the durable reopened record only where latest target run remains `running` and latest target stream equals the acknowledged checkpoint.
- Winner/acknowledged stream and unrelated state survive; no destination replay, checkpoint overwrite, or generic last-writer-wins path appears.
- All seven modes, cancellation, restart/reopen, race, #4046, and R7/R8 focused evidence pass.
- Generated outputs are canonical-generator-only; no separate issue is created for them.
- Fresh no-mistakes ledger remains at most 5 loops; #4059 stays draft and unmerged.
- R2 post-ack cancellation after an unrelated revision and representative source error both return matching failed runs while preserving original error identity and exact current state; missing exact target run returns `Run{}` with `errStateRevisionConflict` detectable.
- R3 two-page stale-page conflict returns the matching durable failed loser in all modes while preserving the typed error, winner checkpoint, unrelated state, and exactly two loser applies; current GitHub/registration evidence is reported honestly without treating a provider smoke as core proof or expanding #4059 production scope.
