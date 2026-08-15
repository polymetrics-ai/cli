# Context — PostgreSQL apply/history club

## Task Delivery Header

- Issue: Closes #4094 — PostgreSQL incremental dedupe history; Closes #4095 — PostgreSQL CDC delete binding; Refs #3859 — native database apply strategies
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, required local evidence recorded, and the API-reported base observed as exactly `integration/4015-mvp-flat-r1`.
- Working branch: `fm/cli-pg-apply-history-club-r1`
- Task: Preserve the already-landed PostgreSQL history target and CDC delete binding while closing #3859's audited residual: the registered database polling adapter must seal the required PostgreSQL source/destination history route into its write plan. The adapter path must produce durable history writes and typed non-PostgreSQL refusals before I/O.
- Verification: Focused red/green engine and database tests; live Docker/Colima PostgreSQL tests that query validity-window state and durable receipts; the requested tagged databaseintegration suite; scoped Go/static/repository gates; GSD verify/review; GitHub API base read-back.

## Evidence Table

| Item | Acceptance criterion | Evidence | Observable assertion, or why this individual fake is necessary |
| --- | --- | --- | --- |
| #4094 | PostgreSQL → PostgreSQL maintains keyed versions and correct validity windows | live | Query the real target after two adapter-applied versions; assert the first interval closes exactly when the second opens and only the second is current. |
| #4094 | A soft delete closes history without physically deleting it | live | Apply a CDC-derived tombstone through the adapter and query the retained row with `_is_current=false` and non-null `_valid_to`. |
| #4094 | Late/replayed history input is deterministic | live | Replay an older/equal source position after the newer version/delete and assert the queried history row set is unchanged. |
| #4094 | History receipts and restart recovery are durable | live | Recreate the driver/ledger/executor, read the persisted receipt, then re-read the same history state and successfully continue from it. |
| #4094 | The three non-PostgreSQL route cells refuse with typed source/destination reasons before I/O | fake | Deterministic route-matrix fakes are necessary because this acceptance requires unsupported engines not to be contacted; assert `DatabaseWriteHistoryRouteError` reason and zero begin/batch/commit/rollback/ledger counters. |
| #4095 | R3/R4 insert, update, and delete reach keyed apply/history close | live | Existing tagged PostgreSQL scenarios read back inserted/updated rows and the tombstone-closed history interval; this change reruns them against the real driver. |
| #4095 | Receipt precedes source acknowledgement/LSN advancement and remains replay-safe across restart | live | The transport/apply result must contain a persisted delivery receipt before returning acknowledgement; a fresh executor replays and queries unchanged target state. |
| #4095 | R1/R2 and destination CDC declarations refuse before I/O | fake | Closed preflight fakes are necessary to prove forbidden routes without opening those sources/destinations; assert the typed refusal and zero source/target/send/session counters. |
| #3859 residual | The database polling adapter can plan the required history strategy | live | Apply an `incremental_dedupe_history` page through `DatabasePollingApplyExecutor`; without the fix it returns `ErrDatabaseWritePlanInvalid`, while green evidence queries the inserted history row and durable acknowledgement. |
| #3859 residual | Adapter route identity cannot be invented or mismatched | fake | Use loaded PostgreSQL/non-PostgreSQL definitions with a recording write boundary; assert typed history-route refusal and zero preview/session/ledger mutations before a provider can be contacted. |

## Locked decisions

- The integration base already contains the staging commits for #4094, #3859, and #4095. Do not duplicate or redesign their history, CDC tombstone, mapping, receipt, or PostgreSQL driver machinery.
- The precise remaining defect is in `internal/connectors/engine/database_polling_apply.go`: it constructs history plans without `DatabaseWritePlanRequest.HistoryRoute`.
- The adapter will receive a loaded, immutable source `database.Definition`, retain the existing destination definition, and derive the history route from those definitions only. It will not infer driver authority from free-form executor IDs.
- Every refusal stays before preview/session/ledger I/O. No public CLI, generic SQL write, connector surface, dependency, or schema change is authorized.
- #4125 and #4158 are explicitly excluded. The known `TestPostgresManagedTargetDriverLiveControlAssertions` base failure will be recorded, not fixed.

## Lifecycle and skills

The adapter was validated with `scripts/gsd doctor`, every required `scripts/gsd sources` lookup, and `go run ./cmd/agentcontractgen check`. `discuss-phase 4094 --auto` and `plan-phase 4094 --tdd` were resolved through `scripts/gsd prompt` and executed inline. The combined issue phase is not a standalone numeric roadmap phase, and the repository's single-worker contract forbids GSD role spawning, so inline/manual execution is the documented compatible fallback.

Loaded skills: `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`.

CLI help/manual/website parity is not applicable because this task changes no command, flag, connector bundle surface, help topic, manual, or website content.
