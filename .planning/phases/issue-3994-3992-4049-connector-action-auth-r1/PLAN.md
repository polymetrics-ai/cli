# Plan — connector action identity, per-fire grants, and typed rate refusal

## Task Delivery Header

- Issue: Refs #3994 — prepared connector-action execution identity; Refs #3992 — per-fire grants,
  approval, and cleanup ordering; Refs #4049 — typed shared-coordinator rate-budget refusal.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Pull request open from `fm/cli-flow-identity-grant-club-r1` against the stated base,
  committed with focused checks green and its API-reported base verified.
- Working branch: `fm/cli-flow-identity-grant-club-r1`
- Task: Add a payload-bound prepared identity and one-consume firing grant to the production
  connector action path, make scheduled action firing use and retain safe evidence from that grant,
  and expose the exact typed rate-budget refusal required when shared coordination is unavailable.
- Verification: focused app/flow/schedule/CLI/engine/hook/connsdk tests with `-timeout 20m`, selected
  race tests, fresh-binary production composition, available credentialed GitHub and PostgreSQL
  evidence, CLI/docs/website parity, generated drift gates, vet/build/lint, and PR base API read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Prepared action identity reaches a durable receipt | live | A fresh `pm` binary executes an approved scheduled action and the persisted safe receipt has the expected non-empty payload-bound identity; without the implementation the field and grant do not exist. |
| Every firing gets exactly one grant | fake | An isolated project and recording connector are necessary to deterministically race the same in-memory prepared firing; exactly one write/read-back/receipt/checkpoint succeeds and the loser gets a typed consumed-grant error. |
| Pre-I/O refusals do not consume the grant | fake | Cancellation, invalid approval binding, expiry, revocation, and scope drift need injected state/clock control unavailable against a real provider; tests assert typed errors, zero provider events, no checkpoint, and an unconsumed marker or later legitimate consume. |
| Process death/partial failure cannot replay | fake | A connector failure after durable grant consumption models the ambiguous crash boundary; a reopened App refuses the same grant and records no second write/checkpoint. |
| Schedule state/lock cleanup follows write outcome | live | The production schedule path writes running state before dispatch, parks or completes terminal state before lock removal, and tests inspect exact persisted state plus lock/target ordering. |
| `require_shared` coordinator loss returns the specified SDK contract | fake | A local counting transport and absent coordinator deterministically assert `*connsdk.RateBudgetRefusalError`, code `shared_coordinator_unavailable`, wrapped safe coordinator reason, and zero sends. |
| Real GitHub behavior remains valid | live | When `PM_CERT_GITHUB_TOKEN`/`GITHUB_TOKEN` is present, the existing immutable lab boundary is used for a fresh-binary approved write/read-back/cleanup; credentials and tokens never enter output. |
| Real PostgreSQL behavior remains valid | fake | The base has no published PostgreSQL destination transport (`write=false`, source-only descriptor), and #4158 is explicitly excluded. A live source/control check can prove only the read side; R2 destination proof is recorded as unavailable rather than fabricated. |

## TDD slices

1. **RED — prepared identity and grant contract.** Add app tests that require a non-empty stable
   prepared identity, payload drift changing it, a signed per-firing grant, receipt propagation,
   typed replay/expiry/refusal errors, zero events/checkpoint on every pre-I/O refusal, one winner
   under concurrent consumption, and no replay after write failure/process reopen.
2. **GREEN — approval authority.** Add the smallest authenticated execution-grant type to the
   project write-approval authority, reuse its durable exclusive consumption marker, and bind the
   resulting connector evidence to the exact preview target/digest.
3. **GREEN — production action path.** Split action execution into prepare and execute boundaries;
   revalidate scope/runtime immediately before grant consumption, consume immediately before write,
   and persist safe prepared/firing identities only after acknowledgement and read-back.
4. **RED/GREEN — schedule production composition.** Extend schedule/CLI tests and fresh-binary
   proof so a firing observes one per-action grant, retains safe identities, rejects cancellation,
   overlap, replay, revoked/expired authorization, and partial write without checkpoint advance,
   and persists terminal state before lock cleanup.
5. **RED/GREEN — rate refusal.** Add the `connsdk` error/code contract and require every existing
   `require_shared` coordinator-unavailable path, including GitHub WriteHook, to wrap with it while
   preserving context cancellation and the existing coordinator cause.
6. **REFACTOR/PARITY.** Update flow/schedule help, CLI manual, website reference, generated
   transcript/data, and lifecycle artifacts only where public safe status changes. No new command,
   raw destination input, generic writer, or dependency is introduced.
7. **VERIFY/LIVE/REVIEW.** Run focused suites and race checks, a fresh binary, available live
   credentials without rendering them, the one-pass generators and drift checks, then execute
   `verify-work` and `code-review` inline. Any gap uses `plan-phase --gaps` and
   `execute-phase --gaps-only`.

## Edge-case matrix

| Edge | Typed outcome | Required negative evidence |
| --- | --- | --- |
| cancellation before consume | `context.Canceled` | zero send/write, no receipt/checkpoint, grant remains consumable |
| process dies / write outcome ambiguous | consumed/replay error on restart | no automatic second write or checkpoint |
| already-granted/already-fired replay | `ExecutionGrantConsumedError` / parked schedule | one total write and one total receipt |
| expired grant / revoked authorization | `ExecutionGrantExpiredError` / `AuthorizationRevokedError` | zero provider events; no grant consumption/checkpoint |
| refused approval or binding | `ExecutionGrantRefusedError` | zero provider events; no grant consumption/checkpoint |
| coordinator unavailable | `RateBudgetRefusalError(shared_coordinator_unavailable)` | zero HTTP sends and zero checkpoint advance |
| concurrent same grant | exactly one success, one consumed error | exactly one write/read-back/receipt/checkpoint |
| write fails partway | original typed/connector failure; schedule parked | terminal parked state precedes lock cleanup; no checkpoint; grant cannot replay |

## Required skills and workflow

Loaded: `golang-how-to`, `golang-cli`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-testing`, `golang-context`, and `golang-concurrency`. `golang-lint` is loaded before final
review as required by #4049. CLI parity and runtime/PostgreSQL references are active.

Resolved GSD path: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` →
`code-review`, executed inline because this issue-club is not a roadmap-numbered phase and the
canonical delivery contract forbids spawning roles.

## CLI parity checklist

- [ ] `pm flow`, `pm help flow`, and `pm flow --help` remain contextual and accurate.
- [ ] `pm schedule`, `pm help schedule`, and `pm schedule --help` remain contextual and accurate.
- [ ] JSON/human output exposes only opaque safe identities, never grant/token/MAC material.
- [ ] `docs/cli/flow.md`, `docs/cli/schedule.md`, website references, embedded manual, and generated
      transcripts/data are updated or explicitly unchanged after inspection.
- [ ] Invalid actions remain usage errors and bare namespaces exit successfully.

## Commit checkpoints

1. plan/TDD evidence;
2. RED tests and captured failures;
3. approval/grant plus action/schedule GREEN implementation;
4. typed rate-refusal GREEN implementation;
5. regenerated artifacts, verification/review evidence, and any review fixes.

## Scope guards

- Do not edit #4125 duration-overflow behavior or #4158 managed-target route matching.
- Do not publish PostgreSQL write capability or invent the missing destination descriptor.
- Do not change GitHub rate-limit declarations or GraphQL exclusion.
- Do not add a generic HTTP/SQL/shell write surface, dependency, credential field, or raw token
  carrier.

