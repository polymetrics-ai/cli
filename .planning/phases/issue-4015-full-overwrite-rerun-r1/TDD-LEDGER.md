# TDD Ledger

## Source-semantic checkpoint eligibility

- Red: Complete. `TestOrchestratorSourceCheckpointFollowsRefreshSemantics` failed only for `full_append` and `full_overwrite`: both received the prior checkpoint when the source contract requires a full refresh. All four incremental modes passed with the prior checkpoint intact. The live binary route first transferred `3/3` records into target IDs `[1 2 3]`; after the source became IDs `[2 3 4]`, the second run reported completed with `records_read=0` and `records_loaded=0`. Its independently queried target had zero rows rather than the required replacement rows. See `traces/red-full-refresh-checkpoint.md`.
- Green: Complete. A single `sourceCheckpointForMode` boundary now omits prior positions for both full-refresh modes and preserves them for every incremental mode. Live rerun transferred `3/3` changed rows, independently queried target IDs `[2 3 4]`, named sample `id=2 label="replacement-two"`, and proved deleted ID 1 absent. A dedicated live incremental-upsert replay remained `0/0` with the same three target rows and named sample.
- Refactor: Complete. The shared selector replaced both Arrow-only special cases and now governs generic, run-scoped overwrite, serial Arrow, and pipelined Arrow extraction. Formatting, focused and broad package tests, live proofs, repository gates, and inline standard review are green.

## Required observable evidence

- First full-overwrite run: exact `records_read`, `records_loaded`, target count, and named sample.
- Changed source: exact source count; deleted ID 1, updated ID 2 named sample, inserted ID 4.
- Second full-overwrite run: exact `records_read`, `records_loaded`, target count, IDs, named sample, and ID 1 absence.
- Incremental replay: exact `records_read=0`, `records_loaded=0`, unchanged target count, and unchanged named sample.
