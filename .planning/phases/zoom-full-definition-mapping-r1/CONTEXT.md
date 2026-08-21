# Context — Zoom full definition mapping continuation

## Task Delivery Header

- Issue: Refs #4265 — Zoom five-class parity foundation cohort.
- Base branch: `fm/cli-reverse-etl-destination-r1`.
- Merges into: `fm/cli-zoom-full-definition-mapping-r1` → `fm/cli-reverse-etl-destination-r1` → `main`.
- Delivery: PR #4285 remains draft against the exact stacked base, with committed Zoom-local definitions, generated artifacts, and a complete seven-surface readiness ledger. Final local and live gates remain pending the final #4304 dispatch head. It is never merged by this lane.
- Working branch: `fm/cli-zoom-full-definition-mapping-r1`.
- Task: Reconcile the 1,937-operation source lock and 206 typed action candidates, then declare Zoom's first production typed reverse-ETL destination on top of #4304 without changing shared engine code.
- Verification: connector-local red/green tests; `connectorgen validate`, certification artifact generation/checks, `surface-sync --check`, `connector-boundary`, focused CLI timeout reproduction, and `make verify`.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The source and destination declarations select only #4304's closed adapters | local | The Zoom bundle test loads `sync_transport.json` and asserts exact executor IDs, source binding, mode strategy, acknowledgement, delivery, conformance run ID, and named action. |
| The destination action remains guarded and user-reachable | local | The test cross-checks the named `writes.json` action and `cli_surface.json` reverse-ETL command, including destructive confirmation; no provider mutation is attempted. |
| Every source/ledger operation is auditable across all seven surfaces | local | The committed readiness ledger and its test assert one record per 1,937 provider source operation and every 1,913 ledger operation with source, parity, binding, transport, implementation, certification, and exact gap fields. |
| Certification is not overstated | live | Matrix and readiness assertions require fixture and accepted live proof before a row can claim certification; no new live evidence is written. |

## Locked decisions

- PR #4304 is the temporary stacked base. Merge it locally without rewriting published history, retarget #4285 with the GitHub API, and do not merge either PR.
- Definitions, tests, fixtures, generated artifacts, and readiness evidence remain under `internal/connectors/defs/zoom/`. No shared engine hook is required by the closed #4304 destination contract.
- Eight `user_id`-only actions are destination-eligible because that normalized input maps faithfully from the existing Zoom users stream `id`; `zoom_users_userssotokendelete` is the one selected by the closed current strategy, while the other seven are explicitly pending action multiplicity. Every remaining action has a precise direct-CLI-only semantic exclusion and remains independently CLI-reachable.
- A destructive action may be declared and command-reachable, but live execution remains limited to lane-created resources with plan → preview → approval → execute. No destructive Zoom call is part of this continuation.
- Certification remains fixture-plus-live-proof only. Existing evidence is retained as observed/live proof where applicable; no uncertified cell is promoted.
- The GSD lifecycle is executed inline because the canonical single-worker contract forbids role spawning. `no-mistakes` is not run because the direct-PR brief expressly excludes it.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
`golang-concurrency`, and `golang-documentation`.
