# PLAN — #3865/#3867 durable coordination production wiring

## Task Delivery Header

- Issue: Refs #3865 — fence cohorts after verified authentication failure; Refs #3867 — persist rate-limit parking and automatic resumption.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, local gates green, automated review routed, and the API-reported PR base recorded.
- Working branch: `fm/cli-durable-store-club-r1`.
- Task: Add crash-durable, cross-process fencing and parking stores; construct them from `app.Open`; wire ordinary admission, typed auth report, credential repair, rate-scope admission, ETL parking, and restart resume; prove both through real child-process restarts and observable state/side-effect assertions.
- Verification: red-first focused tests; real child-process CLI tests; live PostgreSQL `databaseintegration` test with Docker/Colima; race tests; package tests/vet/build; individual repository gates; one-pass derived-artifact regeneration and drift checks; GSD verify/code-review; PR-base API read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Fenced cohort state survives a real process crash/restart | live | A CLI child reaches a typed verified PostgreSQL authentication refusal, persists a fenced epoch, is killed, and a second CLI child is refused before target I/O; the durable file still contains one valid fenced record. |
| Auth coordinator is constructed and used by production admission/report/repair | live | `pm credentials test` enters `app.Open` composition; wrong PostgreSQL auth fences, a dedicated live correct-auth test advances the epoch, and a later admitted operation reaches PostgreSQL. Typed refusal plus unchanged probe table proves zero side effects. |
| Parked state/reset/checkpoint survive a real process crash/restart | live | A CLI ETL child stores authoritative reset evidence and the prior committed checkpoint, is killed, and a new CLI process reloads the record and resumes from that checkpoint. Store deletion and checkpoint/row assertions would be absent for a no-op. |
| Same rate scope blocks siblings and unrelated scopes continue | live | Concurrent child invocations race the same durable store; exactly one scope remains blocked/claimed while a distinct scope completes and changes its own observable output. Refusals assert typed errors and zero writes/checkpoint movement. |
| Resume never replays acknowledged destination apply | live | Destination row identity/count and the committed checkpoint are read after restart; an already acknowledged item remains single while only the unacknowledged next item is applied. |
| Production binary path owns registration and dispatch | live | Tests invoke the real CLI dispatcher, which opens `App`, installs durable coordination in runtime config, dispatches the connector/ETL path, and observes durable files plus destination state without directly constructing either coordinator. |

## Required skills loaded

- GSD: `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
  `gsd-verify-work`, `gsd-code-review`.
- Delivery: `github-issue-first-delivery`.
- Go: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-safety`,
  `golang-security`, `golang-context`, `golang-concurrency`,
  `golang-database`, `golang-testing`.

## Production call chains to prove

1. Fencing: `cmd/pm` → CLI `credentials test`/connector command → `app.Open`
   → durable auth store + coordinator → resolved runtime registration → engine
   or native connector admission → typed verified-auth report → durable fence.
2. Repair: `cmd/pm` → `credentials test` → `App.TestCredential` dedicated probe
   → connector check → verified-healthy repair → durable epoch advance → later
   ordinary admission.
3. Parking: `cmd/pm` → CLI `etl run` → `app.Open` → durable parking store +
   coordinator → declaration-derived rate-scope admission → ETL dispatch →
   typed rate-limit evidence → park committed checkpoint.
4. Resume: next `cmd/pm` → `app.Open` → parking reload/claim → `App.RunETL`
   → existing durable stream checkpoint → destination acknowledgement → store
   release and truthful persisted event.

## TDD slices

### Slice A — RED durable store contracts

- Add tests for schema/version validation, empty/single/large collections,
  atomic duplicate and out-of-order mutations, permission refusal, cancellation,
  and store reopen after abrupt subprocess death.
- Capture compile/test failures before file-store production code.

### Slice B — RED production auth composition

- Add real-process CLI/live-PostgreSQL tests for wrong-password fencing,
  pre-I/O restart refusal, same-cohort concurrency, unrelated cohort admission,
  correct-password repair, stale epoch refusal, and cancelled operation.
- Assert typed errors and unchanged database state for every refusal.

### Slice C — GREEN auth durability/wiring

- Add a versioned file auth store with cross-process atomic mutation.
- Add provider-neutral typed verified-auth marker and runtime coordination seam.
- Construct the coordinator in `app.Open`; attach it during credential
  resolution; wrap engine/native PostgreSQL operations; implement dedicated
  credential probe repair.

### Slice D — RED parking composition and crash resume

- Add real-process CLI/application tests for durable park/reopen, no early
  resume, exact committed checkpoint, same-scope races, claim crash recovery,
  cancellation, duplicate/out-of-order input, schema drift, zero/one/large
  batches, failed resume retention, and acknowledged replay.
- Assert typed refusals and zero destination/checkpoint mutations.

### Slice E — GREEN parking durability/wiring

- Add a versioned file parking store with atomic create, scope reads, leased
  claim/retry, conditional completion, and cancellation.
- Compose parking admission with declared rate-scope admission in the engine.
- Add typed scope resolution, park ETL errors only after a committed checkpoint,
  reload/resume from `app.Open`, and append secret-free durable events.

### Slice F — refactor, parity, verification, review

- Run `gofmt`, focused normal/race tests, live PostgreSQL integration, vet,
  build, and individual verify gates.
- Regenerate connector catalog, website data, golden transcripts, skills, and
  manuals once; run all corresponding drift checks and require clean status.
- Execute generated GSD verify/code-review prompts inline; disposition every
  finding. Commit/push coherent green slices, open the stacked PR, and read its
  base from the GitHub API.

## Edge-case coverage contract

| Edge | Planned assertion |
| --- | --- |
| Cancellation mid-operation | `context.Canceled` is preserved; no fence/park/checkpoint/write occurs unless the durable transition committed first. |
| Connection/process dies partway | SIGKILL child after persisted transition/claim; reopened process observes state and either refuses or resumes after lease expiry. |
| Empty, single-row, large input | Empty store/list and ETL page, one record, and bounded large batch have exact counts/checkpoints. |
| Duplicate/out-of-order delivery | Identical duplicate is idempotent; conflicting evidence or stale epoch/checkpoint is typed refusal with unchanged durable state. |
| Schema drift | Unknown file schema and incompatible checkpoint schema are rejected before mutation/resume. |
| Permission/auth refusal | Filesystem permission and PostgreSQL 28P01 produce typed errors; zero rows/writes/checkpoint advancement. |
| Concurrent same-target runs | Cross-process atomic update/claim admits one winner; losers are typed refusal with zero side effects. |
| Resume after interruption | Expired claim is recoverable; stored committed checkpoint is passed unchanged. |
| Already-acknowledged replay | Destination identity/count remains one and checkpoint does not move backward. |

## Safety and parity

- No dependency, secret output, GitHub credential, raw scope/cohort, provider
  response, generic SQL/HTTP/shell write, or reverse-ETL bypass.
- CLI help/docs/website parity is expected not applicable because no public
  command, flag, syntax, help text, or advertised capability changes. Derived
  artifacts are still regenerated/check-validated as the captain required.
- Findings outside touched paths, especially #4125/#4136/#4158, become a
  `needs-decision` status rather than a scope expansion.
