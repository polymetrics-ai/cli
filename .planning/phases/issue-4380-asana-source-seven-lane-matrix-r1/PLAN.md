# Issue #4380 — Asana source-to-seven-lane matrix plan

## Task Delivery Header

- Issue: Refs #4380 — Asana — source-to-seven-lane matrix
- Base branch: `fm/cli-top100-declaration-batch-r1`
- Merges into: `fm/cli-top100-declaration-batch-r1` → `main`
- Delivery: Scoped branch committed and pushed with focused checks green; it is ready for review and parent-branch integration, not merged.
- Working branch: `codex/asana-track-a-matrix-r1`
- Task: Add a source-lock-bound Asana matrix that retains all 249 source operations and every seven-lane applicability/disposition, then link the enabled connector contract to that matrix without changing execution.
- Verification: Deliberate missing-cell red check; JSON parse; source-lock/descriptor reconciliation; source-import check; connector validation; declaration/projection checks that are available; contract/matrix link and no-hidden-row checks.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every locked source ID has seven explicit lane cells | live | A local reconciliation fails after one cell is removed and passes only when the matrix has 249 unique source rows with 7 cells each. |
| Paginated ETL candidates are not inferred from GET alone | live | The descriptor reports 64 pagination objects; matrix ETL rows are exactly 12 implemented + 52 mapped-unproven, independent of the 119 direct-read GET rows. |
| Mutations have independently visible direct-write and reverse-ETL cells | live | Exact source IDs map to API-surface/write-action evidence for 130 direct-write and 130 reverse-ETL cells; attachment retains two actions for one source ID. |
| Contract preserves source-aware coverage | live | The enabled contract cites the matrix and records 64 ETL candidates with 52 visible mapped-unproven cells rather than omitting them. |

## GSD and skills

- Resolved through the repository adapter: `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.
- Inline/manual execution is used because this assignment forbids spawning runtime roles. The mapping is artifact-only; no runtime or shared-foundation code is in scope.
- Required Go skills were loaded as routing context: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
- The connector-lane-build-order procedure governs the source lock → matrix/projection → artifact → proof ordering.

## Foundation Atlas disposition

- `source.retention-import.v1`: reuse — the retained Asana lock and descriptor are the sole provider-fact inputs.
- `source.projection-admission.v1`: reuse — the matrix is a source-preserving mapping artifact, not runtime admission.
- `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, `warehouse.reverse-etl.v1`, and `transport.sync-contract.v1`: existing runtime evidence only; no shared change proposed.
- `asana.event-token-source.v1`: reuse for the three source-cited sync cells (`getEvents`, `getTask`, `getTasks`).

## Plan

1. Define the literal connector-local matrix schema and materialize all lock-backed rows from an explicit reviewed lane-decision table: direct read from API-surface evidence; direct/reverse write from source-backed API-surface/write-action evidence; ETL from descriptor pagination; binary upload only from the multipart attachment descriptor; and sync only from the event-source contract.
2. Keep every non-applicable cell explicit. Retain each source ID, method/path/citation, descriptor pagination, request/response media, path scope, and event/cursor facts. A source path variable is recorded as scope; it never becomes an invented fanout.
3. Link the matrix from `enabled_connector_contract.json`, correct the ETL denominator to the descriptor-backed 64 candidates, and preserve all existing runtime state as evidence rather than promoting it.
4. Run a deliberate red missing-cell validation, then green source reconciliation, source import, connector validation, JSON parsing, and contract checks. Refactor only serialized evidence; no runtime code or generated global artifacts.
5. Run inline verify-work and code-review evidence, commit the scoped slice, push the scoped branch, and report it for parent-branch review without opening a PR or integrating it.
