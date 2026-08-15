# Plan — issue #4158 / Production MVP verify green

## Task Delivery Header

- Issue: Refs #4158; Production MVP fresh-binary acceptance blocker.
- Base / head at planning: `origin/integration/4015-mvp-flat-r1` / `ef3c71caf` (`BASE-OK` verified).
- Branch: `fm/cli-mvp-verify-green-r1`.
- Target connector: `postgres` managed target; GitHub test is an upstream external-binary reproduction path.
- Delivery: direct PR to `integration/4015-mvp-flat-r1`, never `main`.

## Required skills and lifecycle

- Loaded: `golang-how-to`, `golang-troubleshooting`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-lint`.
- `scripts/gsd doctor`, all five `scripts/gsd sources` calls, all five generated command prompts, and `go run ./cmd/agentcontractgen check` passed before planning.
- Inline/manual GSD fallback: the task and canonical contract forbid lifecycle-role spawning. This records the lifecycle, not a waiver.

## Causal investigation sequence

1. **Reproduce from observable behavior.** Run `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip` from `internal/cli` twice at the shared head; record its binary-level exit code and produced artifact assertion. Run `TestPostgresManagedTargetDriverLiveControlAssertions` with its declared integration controls and record acknowledgement/refusal. These are the visible symptoms, not causes.
2. **Disconfirm first.** Test each falsifier in `CONTEXT.md`: repeatability, independent PostgreSQL control result, pre-#4150 parent result, and one-condition route-admission counterfactual.
3. **Trace and bisect.** Compare the failing execution call path against a known durable-acknowledgement path. Inspect the `#4150` and `#4155` diffs/blame and bisect their commits from the immediate `#4150` parent. Locate the earliest state/route predicate divergence, not merely the nearest changed line.
4. **TDD red.** Before production edits, add separately named tests in existing package style:
   - Happy: valid PostgreSQL managed-target control path reaches a durable acknowledgement through its real constructor/driver path; assert the acknowledgement/receipt value.
   - Bad: non-PostgreSQL history route returns `*database.DatabaseWriteHistoryRouteError` with the applicable refusal reason before any driver/side effect.
   - Edge: one boundary selected from the actual cause (for example, absent history route on a non-history mode, or replayed acknowledgement); assert its exact durable/refusal result and zero unintended I/O.
   The external-binary GitHub test remains required production-entry-point evidence, not a hand-constructed substitute.
5. **Green smallest correction.** Change only the proven causal condition. Keep typed errors and fail-closed behavior. Rerun the red tests, the PostgreSQL control test, and the fresh-binary test.
6. **Verification and review.** Run targeted package tests with `-timeout 20m`, relevant standalone `make verify` gates, build the binary, re-run original acceptance, execute `verify-work` evidence inline, then deep code review. Create a PR body containing reproduction, trigger/mask/symptom, divergent path, counterfactual, falsifiers, TDD classes, skill/lifecycle evidence, and review route.

## CLI help/docs/website parity

Not applicable unless investigation changes a public command, flag, output, connector surface, or help topic. The intended admission correction is internal driver behavior only; this exemption will be re-evaluated before PR creation.

## Decision gate reached

No production or fixture change is authorized yet. The external-binary
acceptance test's inline action configuration conflicts with the job-only
contract introduced by `#4168`; a one-condition job-reference counterfactual
passes. Firstmate must decide whether this is a fixture migration or a request
to restore legacy inline-action compatibility. The latter changes the public
flow contract and needs a new, non-PostgreSQL target ownership decision.

## Commit checkpoints

1. Planning artifacts committed before production edits.
2. Red regression tests and recorded causal evidence.
3. Smallest green implementation and targeted verification.
4. Review / final verification evidence.
