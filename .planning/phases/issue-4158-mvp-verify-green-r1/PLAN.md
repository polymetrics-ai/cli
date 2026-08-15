# Plan — issue #4158 / Production MVP verify green

## Task Delivery Header

- Issue: Production MVP fresh-binary acceptance blocker (#4170 fixture); #4158 was investigated as a suspected shared root and is independent.
- Base / head at planning: `origin/integration/4015-mvp-flat-r1` / `ef3c71caf` (`BASE-OK` verified).
- Branch: `fm/cli-mvp-verify-green-r1`.
- Target connector: GitHub → local warehouse → GitHub action through a fresh external binary.
- Delivery: direct PR to `integration/4015-mvp-flat-r1`, never `main`.

## Required skills and lifecycle

- Loaded: `golang-how-to`, `golang-troubleshooting`, `golang-cli`, `golang-testing`, `golang-stretchr-testify`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-lint`.
- `scripts/gsd doctor`, all five `scripts/gsd sources` calls, all five generated command prompts, and `go run ./cmd/agentcontractgen check` passed before planning.
- Inline/manual GSD fallback: the task and canonical contract forbid lifecycle-role spawning. This records the lifecycle, not a waiver.

## Causal investigation sequence

1. **Reproduce from observable behavior.** Run `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip` from `internal/cli` twice at the shared head; record its binary-level exit code and produced artifact assertion. Run `TestPostgresManagedTargetDriverLiveControlAssertions` with its declared integration controls and record acknowledgement/refusal. These are the visible symptoms, not causes.
2. **Disconfirm first.** Test each falsifier in `CONTEXT.md`: repeatability, independent PostgreSQL control result, pre-#4150 parent result, and one-condition route-admission counterfactual.
3. **Trace and bisect.** Compare the failing execution call path against a known durable-acknowledgement path. Inspect the `#4150` and `#4155` diffs/blame and bisect their commits from the immediate `#4150` parent. Locate the earliest state/route predicate divergence, not merely the nearest changed line.
4. **TDD regression coverage.** Keep the red fresh-binary test as the happy-path acceptance contract and add separately named contract guards:
   - Happy: the fresh binary builds and completes the job-backed GitHub → warehouse → GitHub flow, asserting emitted records, warehouse row, checkpoint, and receipt.
   - Bad: inline action scope without an approved job returns `*flow.JobReferenceError` with `flow.JobReferenceMalformed`, maps to `validation/flow_job_reference_refused`, and creates no stored flow or target event.
   - Edge: revoked and stale approved job references return their exact typed reasons (`unapproved` and `missing`) before target I/O.
5. **Green smallest correction.** Migrate only the #4170 fixture action to its already-created approved reverse-plan `job`; retain `read_back_stream` as the sole action-local setting. Keep typed errors and fail-closed behavior. Rerun the new guard tests and the fresh-binary test.
6. **Verification and review.** Run targeted package tests with `-timeout 20m`, relevant standalone `make verify` gates, build the binary, re-run original acceptance, execute `verify-work` evidence inline, then deep code review. Create a PR body containing reproduction, trigger/mask/symptom, divergent path, counterfactual, falsifiers, TDD classes, skill/lifecycle evidence, and review route.

## CLI help/docs/website parity

No CLI surface changed. The fixture now conforms to the already shipped #4168 job-only flow contract, documented in `docs/cli/flow.md`; no help/manual/website regeneration is required. This is rechecked during verification.

## Decision resolution

Firstmate selected **option 1: fixture migration**. The fixture migrates to
the job-only #4168 contract. Restoring inline-action compatibility was refused:
it would create a public authorization surface that permits an action without
an approved job, contrary to the binding approval model. No production
admission predicate is widened.

## Commit checkpoints

1. Planning artifacts committed before production edits.
2. Red regression tests and recorded causal evidence.
3. Smallest green implementation and targeted verification.
4. Review / final verification evidence.
