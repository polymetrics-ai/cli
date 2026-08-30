# Issue #4390 — Asana connector-definition artifact projection plan

## Task Delivery Header

- Issue: Refs #4390 — Asana — complete connector-definition artifact projection
- Base branch: `codex/asana-track-a-matrix-r1` at `60b85d6a2cf50c1d2fca542199a94c0b651e447f`
- Merges into: `codex/asana-track-a-matrix-r1` → `main`
- Delivery: Scoped branch committed and pushed with focused checks green; ready for independent review and parent-branch integration. No PR is opened or merged by this task.
- Working branch: `codex/asana-track-b-projection-r1`
- Task: Project the Track A 249-row Asana source matrix into the actual connector-definition artifacts for direct read, direct write, binary download, binary upload, ETL through DuckDB, reverse ETL from DuckDB, and sync transport through DuckDB. Preserve existing provider/runtime behavior.
- Verification: Deliberate missing-artifact and missing-backlink red checks; JSON parsing; source-lock/descriptor/matrix reconciliation; focused Asana tests; enabled-contract validation; available source/import/projection checks; and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or bounded disposition |
| --- | --- | --- |
| Every matrix source identity remains the sole denominator | live | The focused projection test reconciles exactly 249 unique source IDs, 7 cells each, against the immutable lock and descriptor. |
| Every loader-recognized lane artifact is present and matrix-linked | live | The enabled contract declares all 7 lanes and its actual artifact paths exist; each applicable matrix cell resolves to the exact existing definition artifact(s). |
| ETL does not promote pageable source rows to streams | live | All 64 candidates are explicit: 12 implemented stream/schema/API bindings and 52 `mapped_unproven` descriptor-only cells. |
| Mutations preserve direct-write and reverse-ETL routing, including deletes | live | All 130 mutation rows retain both cells and resolve to existing named action/API/CLI evidence; the 23 DELETE rows remain present. |
| Binary facts and sync facts stay source-bounded | live | Binary download remains 0/not applicable; the single upload binds `createAttachmentForObject`; sync binds only getEvents/getTask/getTasks with no ordered-history claim. |
| Gap evidence does not create a runtime foundation | live | Seven pre-existing foundation records retain their IDs, states, reasons, and artifact-role inventory while gaining exact matrix links; one new 52-cell ETL mapping gap is source-cited. No shared package or Go runtime file changes. |

## GSD and skills

- Resolved through the repository adapter: `gsd-discuss-phase`, `gsd-plan-phase --tdd`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.
- Inline/manual execution is used because this assignment forbids spawning runtime roles. This is an artifact-only projection; no runtime or shared-foundation implementation is in scope.
- Required Go routing skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- The `connector-lane-build-order` procedure governs lock → matrix → definition artifact → proof ordering.

## Foundation Atlas disposition

- `source.retention-import.v1`: reuse — the retained Asana lock and descriptor remain the sole provider-fact inputs.
- `source.projection-admission.v1`: reuse — source-preserving matrix/backlink validation; it does not admit commands or streams.
- `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, `warehouse.reverse-etl.v1`, and `transport.sync-contract.v1`: existing runtime evidence only; no shared change is proposed.
- `asana.event-token-source.v1`: reuse — the closed event contract selects only the three source-cited task event/hydration/snapshot cells.
- Actual shared-runtime gap: none. The 52 ETL cells are `mapped_unproven` source-mapping deficits, not a missing runtime foundation.

## Plan

1. Reconcile the Track A matrix with existing Asana definitions and make implemented ETL cells point to their exact stream schema artifacts. Keep every non-applicable and mapped-unproven cell untouched in identity and disposition.
2. Tighten the enabled connector contract's artifact inventory to cover the actual ETL stream schema/API projections while retaining all seven established lanes and their bounded source counts.
3. Upgrade the connector-local gap ledger into a source-matrix companion without dropping its seven existing foundation entries. Preserve their states and reasons, attach exact source IDs/lane cells plus Atlas reuse lookups, retain the 11-stream incremental-event limitation in its existing entry, and add only the 52-cell ETL mapping deficit.
4. Add a focused Asana-only projection test that validates actual artifact availability plus lane-specific backlinks, and deliberately fails for a missing artifact and a removed backlink before the green assertions.
5. Correct connector-local docs so direct-route implementation is not misreported as 52 additional executable ETL streams; run the focused validation stack, inline verify/review, commit, push, and hand off to the parent branch.
