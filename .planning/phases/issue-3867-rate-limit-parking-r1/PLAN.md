# PLAN — issue #3867 rate-limit parking and automatic resumption

## Task Delivery Header

- Issue: `Closes #3867 — feat(sync): persist rate-limit parking and automatic resumption`.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: PR open against `integration/4015-mvp-flat-r1` with green checks
  and observed GitHub API base read-back.
- Working branch: `fm/cli-3867-rate-parking-r1`.
- Task: persist rate-limit parking state and resume automatically without
  replaying an acknowledged destination apply.
- Verification: `go test -timeout 20m ./internal/coordination/... ./internal/connectors/engine/... -race`, targeted static/build gates, individual `make verify` gates, and CI.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Parked `parked_rate_limit` state and authoritative reset evidence survive restart | fake | A deterministic opaque store is necessary for this connector-neutral foundation. Reconstruct the coordinator/scheduler from it and assert it resumes the stored run from its committed checkpoint, not from the beginning. |
| Same `RateLimitScopeKey` blocks while unrelated scopes continue | fake | A fake sender increments only after admission; assert the parked scope has exactly zero sends before reset while an unrelated scope sends exactly once. |
| Resume begins from the last committed checkpoint without replay/overwrite | fake | The resume fake records the supplied checkpoint and a separate fake apply count; assert checkpoint equality and no repeat apply before/after scheduler reconstruction. |
| Apply outcome, idempotency, cancellation, and scheduler restart boundaries are covered | fake | Assert duplicate parking creates one dispatch, cancelled parking creates zero dispatches, failed callback retains the record, and a restored scheduler dispatches only after reset. |
| Events are truthful | fake | Capture the park event and assert the actual typed reason/source plus exact reset timestamp; reject generic/missing reset data before any parking mutation. |

## GSD Path and Fallback

- `scripts/gsd doctor`, all five `scripts/gsd sources` resolutions, and
  `go run ./cmd/agentcontractgen check` passed.
- Generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` are executed inline.
- `gsd-sdk query init.phase-op issue-3867-rate-limit-parking-r1` reports
  `phase_found: false`: the canonical roadmap intentionally has no numbered
  issue phase. The canonical single-worker contract also disallows compatible
  isolated role spawning. This issue-scoped artifact set is the documented
  inline/manual fallback and does not waive TDD, verification, or review.

## Required Skills Loaded

- `golang-how-to`; `golang-design-patterns`; `golang-structs-interfaces`.
- `golang-error-handling`; `golang-safety`; `golang-security`.
- `golang-context`; `golang-concurrency`; `golang-testing`.

## Locked Implementation Decisions

1. Treat only an error that unwraps to `*connsdk.RateLimitError` with a
   non-zero parsed reset time as parking authority. Store the reset instant and
   typed source/reason exactly; no inferred deadline, generic failure text, raw
   header, URL, credential, or provider body is stored.
2. Key shared admission by the existing opaque `connectors.RateLimitScopeKey`.
   The new state never accepts an auth cohort key or raw subject as a substitute.
3. Store an immutable clone of the last *committed* checkpoint. Reject a nil or
   invalid checkpoint; never make an uncommitted candidate resumable.
4. Use an injected store and deterministic time/scheduling seam. Persist a park
   transition before it is observable to sibling admission or a restart; do not
   hold a mutex across a resume callback.
5. Reconstructing a scheduler reloads pending records, dispatches once at or
   after reset, and leaves a failed callback parked for retry. Cancellation wins
   before dispatch and duplicate observation remains one resume operation.
6. Emit closed, secret-free park/resume transitions through the existing engine
   rate-limit event boundary or its narrow equivalent. Park events carry the
   exact typed reason and reset time; scopes and request/provider payloads stay
   private.

## TDD Slices

### Slice A — RED durable parking API and observable contracts

1. Add focused tests in `internal/coordination` for persisted parking,
   scope-isolated no-send admission, restart/re-arm, cancellation, idempotency,
   resume checkpoint identity, and no early send.
2. Add engine-focused tests that prove typed rate-limit errors become truthful
   parking requests/events and non-rate/missing-reset errors cause zero store
   mutations and no resume schedule.
3. Run those tests before production code and retain their failing compiler/test
   output in `TDD-LEDGER.md`.

### Slice B — minimum durable coordinator and event bridge

1. Add closed state/reason/event types, an opaque state-store interface with
   race-safe in-memory implementation, checkpoint clone/validation, and
   fail-closed admission errors.
2. Add the deterministic scheduling lifecycle: persist, restore/re-arm, wait
   until reset, dispatch once, retain failure, cancel safely, and release only
   after success.
3. Integrate the existing engine rate-limit classification/event seam so typed
   errors carry their real reset evidence into the coordinator without raw
   response data escaping.

### Slice C — race/restart proof and delivery review

1. Run the focused normal test and required race package command. Confirm all
   refusal paths assert zero schedule/send/mutation side effects.
2. Run formatting, targeted vet/lint/build, and the individual non-suite
   repository gates specified in `AGENTS.md`.
3. Execute generated `verify-work` and `code-review` prompts inline, record
   dispositions in `REVIEW.md`, rebase to the integration base, push, open the
   explicit-base PR, and read its base ref with the GitHub API.

## Commit and Delivery Checkpoints

1. Commit planning/header/context/TDD artifacts.
2. Commit and push the RED test checkpoint with captured failure evidence.
3. Commit and push the GREEN implementation with focused/race proof.
4. Commit review/verification evidence, rebase, push, open the explicit-base
   PR, and verify its GitHub API base.

## Safety and Exclusions

- No credentials, raw rate scope, auth cohort, provider URL/header/body,
  request payload, or secret-derived equality appears in code output, errors,
  tests, state evidence, command arguments, or planning artifacts.
- No provider-specific branch, live provider/database invocation, generic
  HTTP/SQL/shell capability, reverse-ETL bypass, CLI command/flag, or
  documentation surface is added.
- Do not fix #4125, #4136, or #4090. A defect outside the changed paths receives
  a `needs-decision` status rather than being absorbed.
