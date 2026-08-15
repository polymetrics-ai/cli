# Plan — connector action identity, standing job approval, and typed rate refusal

## Task Delivery Header

- Issues: closes #3994, #3992, and #4049.
- Base: `integration/4015-mvp-flat-r1`; working branch:
  `fm/cli-flow-identity-grant-club-r1`; merges base → `main` only through the human-gated parent.
- Delivery: direct PR with focused tests, regenerated artifacts, inline GSD verification/review,
  call-chain evidence, edge-case table, and API-confirmed base.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`, and
  `golang-naming`. CLI parity and runtime/PostgreSQL routing references are active.

## Corrected design

- Keep one standing `AuthorizationRecord`, minted by the existing one-time reverse-plan
  plan/preview/proceed lifecycle. Do not add a schedule authorization, per-fire grant, or token.
- Add explicit flow job references. Sync references resolve to existing connections; action
  references resolve to existing reverse plans with a non-empty, currently valid standing
  authorization. Runtime action fields are derived from the stored plan on every execution.
- Add an atomic flow-create path. It resolves every job before writing. The same resolution runs on
  every flow run so credential, manifest/scope, mapping, action, confirmation, expiry, and
  revocation drift fail before provider I/O.
- A schedule manifest stores only name/cron/flow/root/timestamps. Create and install resolve the
  named stored flow before schedule/backend writes. A flow may have one schedule so the rendered
  direct `flow run` command can deterministically recover its fire lease.
- Keep the exact crontab payload `pm --root <root> flow run <name> --json`. `flowRun` associates a
  named scheduled flow before execution, begins the existing persistent fire lease, forces a fresh
  firing checkpoint namespace, and completes or parks that lease only after the terminal flow
  result.
- Derive a payload-bound prepared-execution identity and persist it only after connector
  acknowledgement plus independent read-back. Use a safe exclusive execution lease to reject
  concurrent/replayed identical prepared execution without treating that lease as approval.
- Return `*connsdk.RateBudgetRefusalError` / `shared_coordinator_unavailable` for every
  `require_shared` unavailable path, including GitHub WriteHook routes, before any HTTP send.

## TDD slices

1. **RED correction:** replace the obsolete grant RED tests with missing/unapproved job creation,
   missing-flow scheduling, direct rendered firing, standing-authorization revalidation, prepared
   identity, replay/ambiguity, and named rate-refusal assertions. Retain the original failed command
   as superseded historical evidence.
2. **GREEN flow jobs:** add typed job-reference errors and resolver; hydrate action scope from the
   approved reverse plan; atomically create the flow only after all positive checks; resolve again
   at plan/run time.
3. **GREEN schedule inheritance:** remove authorization from manifests/argv/output; validate stored
   flow on create/install; render direct `flow run`; associate one schedule per flow and preserve
   running/parked/succeeded ordering around the terminal result.
4. **GREEN prepared identity:** keep prepare/execute separation, remove grant issuance/consumption,
   revalidate live authorization and destructive preview immediately before write, enforce the
   non-authoritative prepared execution lease, and persist safe identity evidence after read-back.
5. **GREEN rate refusal:** expose the named SDK error/code while retaining context cancellation and
   the existing coordinator cause; prove zero transport sends on engine and GitHub hook routes.
6. **PARITY:** update flow/schedule help, generated manual, `docs/cli`, website docs/data, golden
   transcripts, and the #3992 issue comment. Run generators once after source/docs settle, then all
   drift checks until clean.
7. **VERIFY/REVIEW:** focused tests/race checks, fresh binary call-chain proof, available live checks,
   scoped vet/build/lint and non-suite gates, inline `verify-work` and `code-review`; close gaps with
   the required GSD gap loop.

## Acceptance and negative side-effect evidence

| Edge | Typed outcome | Negative/terminal proof |
| --- | --- | --- |
| missing/malformed/unrecognised flow job | `*flow.JobReferenceError` naming reference/reason | no flow file, provider event, receipt, or checkpoint |
| unapproved/expired/revoked action job | `*flow.JobReferenceError` wrapping the precise App authorization error | no flow file/send/write/checkpoint |
| missing/malformed scheduled flow | `*schedule.FlowReferenceError` | no schedule file, crontab entry, or sentinel |
| cancellation before dispatch | `context.Canceled` | zero send/write/receipt/checkpoint; prepared lease released |
| process death or ambiguous write | persisted running/parked lease | no automatic replay or checkpoint advance |
| same prepared execution replay | `*app.PreparedExecutionReplayError` | one total write/read-back/receipt/checkpoint |
| coordinator unavailable | `*connsdk.RateBudgetRefusalError` with named code | zero HTTP sends and no checkpoint |
| concurrent schedule/prepared firing | `schedule.ErrFireInProgress` / prepared replay error | exactly one provider dispatch |
| cleanup failure after terminal write | schedule `FireStopCleanup` park | terminal state is stored before lock cleanup; no replay |

## CLI/docs parity checklist

- [ ] `pm flow`, `pm help flow`, `pm flow --help`, and new creation help are accurate.
- [ ] `pm schedule`, `pm help schedule`, and `pm schedule --help` describe flow inheritance and no
      crontab approval token/reference.
- [ ] Bare namespaces remain successful contextual help; invalid actions remain usage errors.
- [ ] Human/JSON output contains only safe job/flow/prepared/receipt identities.
- [ ] Generated manual, CLI docs, website docs/data, and golden transcripts are regenerated once and
      pass drift checks.

## Scope guards

No generic HTTP/SQL/shell writer, raw destination control, new credential carrier, dependency,
PostgreSQL destination publication, #4125 change, or #4158 change.

## Gap closure — PR #4168 verify failure

The CI `certify-timing` gate exposed one branch-owned proof gap after the corrected schedule model
landed: the production certification harness still supplied and required the superseded
`--authorization auth_...` schedule reference. The production CLI correctly removed that carrier,
so certification failed while looking for a field that must not exist.

1. **RED:** retain the exact failing fresh-binary test,
   `TestCertifyCLISingleConnectorPassExitsZero`, and make the scripted certification driver reject
   any schedule authorization flag or response field.
2. **GREEN:** remove the obsolete carrier from `stageScheduleRoundtrip`; assert create/list/install/
   remove envelopes and the rendered crontab contain no approval token/reference/credential
   material while preserving sentinel cleanup and byte-identical restoration.
3. **VERIFY:** rerun the exact CLI test, focused certification tests, and `certify-timing`; then
   regenerate all derived artifacts in one pass and run every applicable drift check to a clean
   worktree before the rebased branch is force-pushed with lease.

Inline/manual GSD gap fallback is used because this worker has no compatible isolated Pi-agent
runtime. Required skills remain those listed in the delivery header.
