# PLAN — #3858 page-safe polling source executor

## Task Delivery Header

- Issue: `Closes #3858 — feat(sync): implement page-safe polling source executor`
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: PR opened against `integration/4015-mvp-flat-r1` with green CI.
- Working branch: `fm/cli-3858-polling-executor-r1`
- Task: Implement the shared page-safe `polling_watermark` source executor.
- Verification: `go test -timeout 20m ./internal/connectors/engine/... ./internal/connectors/...` plus the individual repository gates and CI.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The immutable common polling corpus runs through this executor | fake | This shared foundation has no native driver or catalog-selected live binding. The fake transport records every tuple request, and the fake destination records every delivered stable identity and durable acknowledgement. |
| An interruption between pages resumes without a gap or duplicate | fake | The source records ordered rows with an equal-watermark page split; the test stops after a durable first page and asserts the combined delivered identities are exactly `a`, `b`, `c` once each. |
| A checkpoint never advances beyond a durably delivered page | fake | The destination and state store have separate counters. Refusals assert zero page delivery/state writes; persistence failure after acknowledgement asserts replay from the prior committed tuple. |
| Tuple, codec, resume, overlap, schema/source, and soft-delete rules are fail-closed | fake | The descriptor and source adapter are structured in-process contracts. Fakes make source requests, delivered records, checkpoints, and no-I/O/no-write refusals observable without an unauthorized database driver. |
| The executor accepts no generic SQL, HTTP, shell, path, or raw-protocol input | fake | Its public request contains only a preflight-resolved declaration, catalog-selected object binding, typed records, page sink, and #3810 checkpoint envelope; compile-time interface tests prevent bypass parameters. |

## Scope and ownership

This is the connector-neutral source half of #3855. It consumes #3857's
`PollingPreflight` and #3810 checkpoint contract. It owns the shared executor,
its closed source/sink/store seams, test fixtures that drive the real executor,
and delivery evidence only. The existing #3880 legacy `ChangefeedExecutor`
uses state maps and an unrelated changefeed descriptor; it is a migration
reference, not a second implementation to extend or a claim of CDC support.

Excluded: a provider/native adapter, target apply/DML (#3859), connector
definition/bundle changes, command/help/docs surfaces (#3860), generic SQL,
HTTP, shell, path, or raw query input, live databases, credentials, and all
unrelated defects (#4125, #4136, #4090).

## Decisions captured inline

- The registered source remains preflight-owned: the executor receives a
  `ResolvedPollingWatermark` only after #3857 validates the declaration,
  catalog object, canonical mode, registered source/apply evidence, and
  bounded keyset policy. It must not reproduce that rule set.
- A page is delivered as one bounded unit. The destination reports durable
  acknowledgement for the full page before the checkpoint store is called.
  If persistence fails after acknowledgement, the old envelope remains the
  resume point and destination idempotency handles the replay.
- Resume uses the complete lossless `(watermark, tie_breaker)` tuple. No
  float conversion, display formatting, cursor-only predicate, or mutable
  offset boundary is introduced.
- Every source adapter receives only typed, declaration-selected fields and a
  cloned checkpoint tuple; it cannot receive caller-authored SQL, HTTP, shell,
  URL, method, body, or filesystem path.
- #3810 typed `RecoveryOutcome` is surfaced for incompatible/expired resume
  state. Invalid state is never cleared or converted into a full scan.

## GSD lifecycle and skills

Passed: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`,
`plan-phase`, `execute-phase`, `verify-work`, and `code-review`; and
`go run ./cmd/agentcontractgen check`. Generated prompts for all five steps
were resolved with `scripts/gsd prompt ...`.

Manual GSD fallback: `gsd-sdk query init.phase-op
issue-3858-polling-source-executor-r1` reports `phase_found: false`, so the
registry cannot produce phase artifacts for this issue ID. The task explicitly
requires autonomous delivery and the canonical single-worker contract forbids
role spawning. This issue directory therefore records the inline
discuss → plan (`--tdd`) → execute → verify → review lifecycle; the same
red/green, verification, and review gates are retained.

Loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
and `golang-database`. No CLI surface changes are planned, so CLI
help/manual/website parity is not applicable.

## TDD slices

1. **Red — real preflight-to-source execution.** Write a closed source
   executor test harness that resolves #3857, fetches page tuples, delivers a
   bounded page, and persists a #3810 envelope. Assert exact source requests,
   records, acknowledgements, and committed tuples.
2. **Red — page boundary and durability.** Prove equal-watermark interruption
   and state-save failure resume at the prior committed tuple; assert a
   combined once-only logical result and no checkpoint after failed/partial
   acknowledgement.
3. **Red — refuse unsafe state before I/O.** Cover null/precision violations,
   incomplete/non-advancing tuple, stale source/schema state, and unsafe
   overlap. Each refusal asserts no fetch, no destination acknowledgement,
   and no checkpoint mutation.
4. **Green — smallest declaration-driven executor.** Add only the typed
   source, durable page sink, and checkpoint-store seams required to satisfy
   the tests, with defensive cloning and `context.Context` propagation.
5. **Green — immutable corpus lane.** Route every source-relevant #3856
   fixture through the real executor rather than the reference-only lane;
   keep delete visibility explicit and never advertise polling as CDC.
6. **Refactor/review.** Remove no behavior. Audit no raw input parameters,
   no scalar/float checkpoint fallback, no unbounded pages, and no hidden
   checkpoint advancement.

## Planned files

- `internal/connectors/engine/polling_source.go` — new shared executor and
  closed source/sink/checkpoint interfaces.
- `internal/connectors/engine/polling_source_test.go` — observable
  red/green conformance and failure-boundary tests.
- `internal/connectors/engine/polling_conformance_test.go` — only if needed to
  make the immutable suite exercise this executor rather than the old
  reference lane.
- `.planning/phases/issue-3858-polling-source-executor-r1/*` — lifecycle
  evidence.

## Commit checkpoints

1. Plan/TDD/verification evidence.
2. Red test checkpoint with command output recorded in `TDD-LEDGER.md`.
3. Green executor and focused test checkpoint.
4. Review/final-verification checkpoint and PR delivery.
