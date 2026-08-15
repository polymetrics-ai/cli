# Plan — #3977 pgoutput v2 CDC

## Delivery and lifecycle

This is an inline/manual-GSD fallback because the active roadmap has no phase `3977`. The following command path was resolved through `scripts/gsd`: `discuss-phase`, `plan-phase 3977 --tdd`, `execute-phase 3977 --interactive`, `verify-work 3977`, and `code-review 3977`. The canonical single-worker contract prohibits role spawning. `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` pass.

Required skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and `golang-database`. The PostgreSQL/dbtest runtime reference was read. No CLI command changes are planned; the connector manual, catalog, and website-generated connector data are regenerated for the promoted CDC capability.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A stream abort publishes, checkpoints, and acknowledges nothing | fake | A deterministic v2-frame test injects `StreamStart`/DML/`StreamAbort` and asserts all three counters remain zero. PostgreSQL does not expose a practical SQL-only way to force an abort frame deterministically. |
| No staged transaction is visible before stream commit | fake | The frame test injects a segment and `StreamStop`, then asserts the downstream event and receipt counters are zero before `StreamCommit`. |
| Receipt durability precedes checkpoint and acknowledgement | fake | The checkpoint test observes the stage receipt during checkpoint persistence and records the strict `emit → receipt → checkpoint → acknowledge` order. |
| PostgreSQL 14+ v2 replication executes and cleanup is scoped | live | A `dbtest` PostgreSQL container uses `wal_level=logical` and small logical-decoding memory, produces a multi-segment transaction, checks ordered event delivery and durable LSN resume, then observes that `TeardownCDC` removes the derived slot. |
| Capability promotion matches executable runtime | live | Focused capability tests load the real PostgreSQL bundle and native descriptor after the live proof; `connectorgen validate` and `surface-sync --check` observe the same declaration. |

## TDD slices

1. **Red:** add deterministic protocol-v2 stream tests for no pre-commit visibility, abort discard, finite stage quota, and receipt/checkpoint/acknowledgement ordering. Run the focused PostgreSQL package; it must fail while `ReadCDC` returns the planned-stage fence.
2. **Green:** add a PostgreSQL-private v2 state machine and durable transaction receiver. Open a bounded stage below `RuntimeConfig.ProjectDir`, use source-derived transaction identities, negotiate `proto_version '2'` plus `streaming 'on'`, and process only the valid v2 lifecycle transitions.
3. **Refactor:** keep framing/receipt code in a dedicated file, preserve typed resume/rebootstrap and lifecycle helpers, and ensure error paths close the connection and do not acknowledge a candidate LSN.
4. **Red:** replace the historical skipped test with a `databaseintegration` PostgreSQL/dbtest proof that asserts an actual multi-segment transaction, ordering, restart, and slot cleanup.
5. **Green:** implement the harness config and live fixture. Run the explicit Docker/Colima command supplied by the task.
6. **Promotion:** only after the live test passes, change the native descriptor and PostgreSQL bundle `changefeed.json`/`metadata.json` to the matching implemented contract. Regenerate derived connector surface data if required.

## Commit checkpoints

- Planning checkpoint: these GSD records.
- Red regression checkpoint: focused test fails for the planned execution fence.
- Green executor checkpoint: unit tests and targeted package checks pass.
- Live/projection checkpoint: databaseintegration proof and descriptor parity pass.
- Review-fix checkpoint: code review findings, if any, are fixed with repeated verification.
