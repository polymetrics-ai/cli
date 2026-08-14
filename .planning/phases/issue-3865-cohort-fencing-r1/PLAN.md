# PLAN — issue #3865 verified-auth cohort fencing

## Task Delivery Header

- Issue: Refs #3865 — feat(coordination): fence cohorts after verified authentication failure.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; read its base back through the GitHub API after opening.
- Working branch: `fm/cli-3865-cohort-fencing-r1`.
- Task: Add a connector-neutral, secret-free cohort-health coordinator. Only a typed, verified invalid-authentication outcome can fence a cohort; fencing rejects future admissions and cancels active siblings. A separately verified repair advances a healthy epoch, so stale members cannot resume and new work can be admitted.
- Verification: capture RED test failure before production changes; run focused tests and `go test -race -timeout 20m ./internal/coordination/...`; run targeted vet/build, formatting, individual repository gates, CI, and a base-ref API read-back for the PR.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Only a typed verified invalid-authentication result fences a cohort | fake | A deterministic in-process coordinator is needed because this connector-neutral foundation must not call a provider. The test asserts all non-verified outcomes leave a sibling admission grantable, while the verified outcome changes health to fenced. |
| Fencing cancels same-cohort siblings and rejects new work without a send | fake | A fake sender increments only after admission. The test waits for the sibling context cancellation, then asserts its post-fence send count is exactly zero and the next cohort admission is refused. |
| Unrelated cohorts remain healthy | fake | A second opaque cohort exercises the same live coordinator instance; its sender count is exactly one after the first cohort fences. |
| Repair/test opens a new healthy epoch while stale members remain refused | fake | A deterministic typed repair result is necessary because no provider adapter is in scope. The test observes a strictly larger epoch, asserts an old member cannot admit or send, then records exactly one new-epoch send. |
| Restart and races preserve the fence/epoch contract | fake | An in-memory durable health store models the persistence seam without a project-state or provider write. Under `-race`, a restarted coordinator reloads the fenced health, concurrent admissions after fencing produce zero send increments, and a stale epoch is rejected. |

## GSD path and fallback

- `scripts/gsd doctor`, all five `scripts/gsd sources` resolutions, and `go run ./cmd/agentcontractgen check` passed.
- Generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were read and are executed inline.
- The issue is not a numbered roadmap phase; compatible isolated GSD roles are unavailable and the repository's canonical single-worker contract forbids role spawning. This is the documented inline/manual fallback; it does not waive TDD, verification, or review.

## Required skills loaded

- `golang-how-to`; `golang-design-patterns`; `golang-structs-interfaces`.
- `golang-error-handling`; `golang-safety`; `golang-security`.
- `golang-context`; `golang-concurrency`; `golang-testing`.

## Decisions from discussion

1. The coordinator accepts the already-derived `connectors.AuthCohortKey`, never credentials, credential IDs, revisions, provider responses, or raw headers.
2. Authentication outcomes are a closed typed vocabulary. A plain 401-like status, transport error, timeout, or provider error has no path to mutate health; only the verified-invalid outcome may fence.
3. Each admitted member receives a derived cancellable context and an epoch. Fencing atomically changes health before cancelling same-epoch members, so a caller checking its admission boundary cannot send after the fence wins.
4. A verified repair/test is the only transition that creates a new healthy epoch. It cancels old members; their stale epoch cannot be used to admit or report a later outcome.
5. Health persistence is an injected opaque-key store seam. This task supplies a deterministic in-memory implementation and restart proof only; it does not change `internal/app` state, provider behavior, UDS rate-budget protocol, checkpointing, scheduler behavior, or #3867 rate parking.

## TDD slices

### Slice A — RED coordinator contracts

1. Add package tests for closed outcome classification, same-cohort cancellation, zero post-fence sends, unrelated cohort health, repair epoch advancement, stale-epoch refusal, and restart/race behavior.
2. Run the focused coordinator test before any production coordinator code exists. Record its compile/failure output in `TDD-LEDGER.md`.

### Slice B — minimal healthy/fenced state machine

1. Add typed outcomes, opaque cohort epoch/member types, fail-closed errors, a thread-safe in-memory health store, and a coordinator with admission/report/repair operations.
2. Keep state changes under one mutex; never hold it while invoking cancellation. Persist the health transition before releasing a post-fence admission path.
3. Make member contexts derived from caller contexts and cancel them for fence or epoch supersession. Return safe typed errors without rendering an opaque key.

### Slice C — race/restart proof and review

1. Prove a restarted coordinator reloads a persisted fence, observes no further admitted sends, and rejects an epoch held by a prior member.
2. Run the required race package test, focused normal test, targeted vet/build, formatter, and individual safe repository gates.
3. Execute inline verify-work and standard code-review. Record finding dispositions in `REVIEW.md`; create a gap plan only for a real uncovered acceptance criterion.

## Commit and delivery checkpoints

1. Commit planning/header/TDD evidence.
2. Commit the RED test checkpoint after retaining its failure output.
3. Commit the GREEN coordinator implementation with focused/race proof.
4. Commit final verification/review artifacts, push the issue branch, create the PR with explicit base, and read back `.base.ref` through the API.

## Safety and exclusions

- No credential, secret, secret-derived equality, binding preimage, credential revision, provider body/header, endpoint, or opaque cohort key appears in output, errors, tests, state evidence, argv, or a plan.
- No provider-specific branch, live provider call, generic HTTP/SQL/shell capability, UDS rate-budget modification, app durable-state wiring, checkpoint change, scheduler feature, or #3867 parking/resumption behavior is in scope.
- #4125, #4136, and #4090 are explicitly out of scope. A finding outside this diff is recorded as `needs-decision` rather than absorbed.
