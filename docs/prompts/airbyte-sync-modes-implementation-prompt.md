# Prompt: Airbyte-Style Sync Modes for PM

Use this prompt when asking Codex, Claude, or another implementation agent to add complete ETL sync-mode semantics to the Polymetrics `pm` Go CLI using strict TDD and the GSD universal programming loop.

## Research Summary

Airbyte models sync behavior as a combination of source read mode and destination write mode:

- Source read modes: `full_refresh`, `incremental`
- Destination write modes: `append`, `overwrite`, `append_dedup`
- Product-facing combinations:
  - `full_refresh_append`
  - `full_refresh_overwrite`
  - `full_refresh_overwrite_deduped`
  - `incremental_append`
  - `incremental_append_deduped`

The important architecture lesson is that connectors should describe stream capabilities and emit records; they should not own destination semantics. The sync engine should validate the chosen mode, write raw/staged data with run metadata, then materialize the final destination table or file according to the mode.

Airbyte code reviewed locally from `airbytehq/airbyte` commit `124f26099572d954ea6f257780145b8e1fb5bd68`:

- `SyncMode.yaml`: source protocol modes are `full_refresh` and `incremental`.
- `DestinationSyncMode.yaml`: destination protocol modes are `append`, `overwrite`, and `append_dedup`.
- `StreamConfig.kt`: stream execution config includes import action, primary key, cursor, generation ID, minimum generation ID, and sync ID.
- `CatalogParser.kt`: maps `append`/`overwrite` to append behavior and `append_dedup` to dedupe; overwrite behavior is driven by generation/min-generation refresh semantics.
- `AbstractStreamOperation.kt`: uses temporary raw/final tables for truncate/full-refresh work and swaps them into place only after successful stream completion.
- `StorageOperation.kt`: abstracts stage preparation, stage overwrite, temp-stage transfer, final-table creation, soft reset, final overwrite, and type/dedupe.
- `DefaultSyncOperation.kt`: prepares per-stream operations, writes batches, adds sync metadata, and finalizes streams concurrently.

## Difficulty Assessment

Easy:

- `full_refresh_append`: read all records and append them. It intentionally allows duplicates.
- Basic `full_refresh_overwrite`: write all records to a temporary destination and atomically replace the final destination on success.

Medium:

- `incremental_append`: requires cursor configuration, per-stream durable state, source capability validation, and checkpoint commit only after successful destination writes.
- Mode validation and CLI docs: straightforward, but must be exhaustive and deterministic for agents.

Medium-hard:

- `full_refresh_overwrite_deduped`: requires primary key and cursor validation, full replacement semantics, deterministic dedupe ordering, temp final output, and failure-safe swap behavior.

Hard:

- `incremental_append_deduped`: requires raw history, final materialization/upsert by primary key, cursor ordering, retry-safe checkpoints, tie breakers, and delete/tombstone handling.
- Schema evolution and typed final tables: needs a stable destination abstraction and tests around added/removed/renamed fields.

Very hard:

- CDC parity with production warehouses: requires delete semantics, source-emitted tombstones, ordering guarantees, reprocessing unprocessed raw records, raw/final migrations, and resumable failed full-refresh generations.

The practical PM plan should implement the dependency-free file-backed store first, with the same interfaces required for a later PostgreSQL-backed store. This gives us correct semantics and test coverage without blocking on runtime dependencies.

## Implementation Prompt

```text
Act as a senior Go engineer, ETL architect, and security-focused CLI implementer.

Repository:
Polymetrics `pm` Go CLI monolith.

Primary goal:
Implement Airbyte-style sync-mode semantics for PM ETL with dependency-free local storage first. Preserve PM's explicit Go connector design. Do not add a plugin runtime or a web UI.

Use the GSD universal programming loop:
- workflow.use_worktrees=false
- workflow.tdd_mode=true
- Red-green-refactor for every behavior slice.
- Update phase artifacts before marking the work complete.

Phase name:
airbyte-style-sync-modes

Required local context to read before coding:
- README.md
- POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md
- POLYMETRICS_AGENTIC_ETL_GO_PLAN.md
- docs/architecture/repo-profile.json
- docs/prompts/gsd-agentic-etl-full-rewrite-tdd-prompt.md
- docs/prompts/universal-programming-loop-prompts.md
- docs/cli/etl.md
- docs/cli/connectors.md
- docs/cli/credentials.md
- internal/app/*
- internal/cli/*
- internal/connectors/*
- internal/warehouse/*
- internal/perf/*
- test fixtures and existing ETL tests

Required Airbyte research anchors:
- Airbyte Sync Modes docs
- Airbyte Incremental Append + Deduped docs
- Airbyte Full Refresh Overwrite + Deduped docs
- Airbyte source files:
  - SyncMode.yaml
  - DestinationSyncMode.yaml
  - StreamConfig.kt
  - CatalogParser.kt
  - AbstractStreamOperation.kt
  - StorageOperation.kt
  - DefaultSyncOperation.kt

Architectural rules:
- Separate source read mode from destination write mode in Go types.
- Keep connector implementations focused on discovery, capability declaration, and record emission.
- Keep ETL mode behavior in the sync engine, planner, state store, and destination store.
- Preserve dependency-free operation as the default.
- Design storage interfaces so PostgreSQL-backed execution can be added later without changing connector contracts.
- Never commit a cursor checkpoint until raw writes and final materialization for that batch or stream are successful.
- Never destroy a previous final output for overwrite modes until the new output is complete and validated.
- Do not log, print, snapshot, or serialize secret values.
- Machine JSON must be deterministic and versioned. Human logs go to stderr.

Mode model to implement:

1. Source sync modes:
   - `full_refresh`
   - `incremental`

2. Destination sync modes:
   - `append`
   - `overwrite`
   - `append_dedup`
   - Internal alias for PM: `overwrite_dedup`

3. User-facing PM modes:
   - `full_refresh_append`
   - `full_refresh_overwrite`
   - `full_refresh_overwrite_deduped`
   - `incremental_append`
   - `incremental_append_deduped`

4. Backward compatibility:
   - Continue accepting the existing mode string where present.
   - Normalize old and new strings into the typed model.
   - Emit a deprecation warning only on stderr when needed.

Core data model:

- `SyncMode`
  - source mode
  - destination mode
  - canonical user-facing name
  - requires cursor
  - requires primary key
  - produces raw history
  - materializes final output
  - is overwrite/truncate generation

- `StreamConfig`
  - connector
  - stream name
  - source sync mode
  - destination sync mode
  - primary key fields
  - cursor field
  - selected fields if already supported
  - generation ID
  - minimum generation ID
  - sync ID

- `RawRecord`
  - raw ID
  - connector
  - stream
  - sync ID
  - generation ID
  - extracted timestamp
  - loaded timestamp
  - source cursor value
  - primary key values
  - deleted/tombstone marker if present
  - record JSON payload

- `StreamState`
  - connector
  - stream
  - last successful sync ID
  - cursor value
  - generation ID
  - minimum generation ID
  - raw processed watermark
  - final schema hash if available
  - counts and last completed timestamp

Storage interfaces:

- `StateStore`
  - load stream state
  - begin sync
  - commit stream state
  - abort sync
  - allocate sync ID and generation ID

- `RawStore`
  - prepare stage
  - append raw batch
  - transfer temp stage to real stage
  - overwrite real stage with temp stage
  - list unprocessed raw records
  - mark raw records processed
  - cleanup failed or stale stages

- `FinalStore`
  - append final batch
  - prepare temp final output
  - replace final output atomically
  - materialize deduped final output
  - soft reset final output
  - inspect final output counts and schema

Dependency-free implementation:
- Use JSONL files for raw history and final output.
- Use temp files/directories plus atomic rename for overwrite modes.
- Use a deterministic dedupe implementation in Go:
  - Group records by canonical primary-key tuple.
  - Keep the newest record by cursor value.
  - Break cursor ties by extracted timestamp.
  - Break remaining ties by raw ID.
  - If a latest record is a delete/tombstone, omit it from the final table unless the destination explicitly requests retained tombstones.
- Keep memory bounded with batch processing. If a full in-memory map is needed for MVP dedupe, document it and add a benchmark/test that establishes the current bound.

Future PostgreSQL-backed implementation:
- Use raw tables with PM metadata columns.
- Use final tables/views materialized by SQL.
- Use transaction-protected temp table swaps for overwrite modes.
- Use `INSERT ... ON CONFLICT` or merge-equivalent behavior for incremental deduped final tables.
- Keep the same Go interfaces used by the dependency-free store.

TDD slices:

1. Mode parsing and validation
RED:
- Add tests for all canonical modes, legacy mode aliases, invalid modes, missing cursor, missing primary key, and unsupported connector capabilities.
GREEN:
- Implement typed source/destination/user-facing mode parsing.
REFACTOR:
- Move mode validation to a small package with no connector dependencies.

2. Connector manifest capability validation
RED:
- Add tests that GitHub pull requests support full refresh and incremental by updated cursor only when configured.
- Add tests that deduped modes require a declared primary key.
GREEN:
- Extend connector manifests with supported source modes, default cursor fields, default primary keys, and destination mode compatibility.
REFACTOR:
- Ensure docs/skills generation reads from the same manifest data.

3. Full refresh append
RED:
- Add an ETL test where two runs intentionally append duplicate records.
GREEN:
- Implement append behavior through the final store.
REFACTOR:
- Keep source read and final write concerns separate.

4. Full refresh overwrite
RED:
- Add a test where an existing final output remains untouched if a new full refresh fails midway.
- Add a test where successful full refresh atomically replaces the final output.
GREEN:
- Implement temp final output and atomic replacement.
REFACTOR:
- Extract overwrite finalization into a reusable operation.

5. Incremental append
RED:
- Add a test where the first incremental run behaves like full refresh.
- Add a test where the second run sends cursor state to the connector and appends only newer/resent records.
- Add a test where a failed destination write does not advance the checkpoint.
GREEN:
- Implement stream state load/commit and cursor propagation.
REFACTOR:
- Make checkpoint commit explicit and auditable.

6. Incremental append deduped
RED:
- Add tests for updated records, duplicate cursor values, composite primary keys, out-of-order records, retried records, and delete/tombstone records.
GREEN:
- Implement raw history plus deduped final materialization.
REFACTOR:
- Extract deterministic record ordering and primary-key tuple canonicalization.

7. Full refresh overwrite deduped
RED:
- Add tests proving duplicate records inside a full-refresh run collapse to the newest record.
- Add tests proving removed source records disappear from the final output after a successful run.
- Add tests proving failed full-refresh dedupe keeps the previous final output.
GREEN:
- Combine generation/temp final behavior with deterministic dedupe.
REFACTOR:
- Share dedupe finalization with incremental dedupe while preserving overwrite semantics.

8. CLI and docs
RED:
- Add golden tests for `pm etl --help`, generated docs, and generated skills listing every sync mode with cursor/primary-key requirements.
GREEN:
- Update CLI help, docs generation, and examples.
REFACTOR:
- Ensure human output is concise and machine JSON is stable.

9. Benchmarks
RED:
- Add benchmarks for append, overwrite, incremental append, and dedupe over synthetic paginated records.
GREEN:
- Implement benchmark harness and performance docs.
REFACTOR:
- Capture records/sec, pages/sec, allocations, elapsed time, and peak memory where feasible.

Acceptance criteria:
- `pm` supports all five user-facing modes.
- Modes are typed and validated before a connector run starts.
- Dependency-free JSONL store implements correct append, overwrite, incremental, and dedupe behavior.
- Failed overwrite/deduped runs do not destroy previous final output.
- Cursor state advances only after successful writes/finalization.
- Deduped final output is deterministic for primary key, cursor, extracted timestamp, and raw ID ordering.
- GitHub pull request ETL can run with `incremental_append_deduped` using PR number or node ID as primary key and `updated_at` as cursor.
- CLI docs and generated skills explain each mode in man-page-level detail.
- Benchmarks compare at least append versus deduped materialization over multiple synthetic pages.
- `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, and existing smoke checks pass.

Required verification commands:
- gofmt on touched Go files
- go vet ./...
- go test ./...
- go build ./cmd/pm
- ./pm etl --help
- ./pm docs generate --dir docs/cli
- ./pm skills generate --dir docs/skills
- benchmark command implemented for this phase

Final response required:
- Summarize the implemented sync modes.
- List behavior covered by tests.
- List verification commands and results.
- List any skipped checks with concrete reasons.
- Mention remaining gaps, especially CDC parity, schema evolution, PostgreSQL-backed store, and memory-bound dedupe if not fully solved.
```

## Implementation Notes for PM

Start with the dependency-free store because it is the smallest proof of correctness:

1. Normalize modes into typed source and destination modes.
2. Add per-stream state with cursor and generation metadata.
3. Add raw JSONL history with PM metadata.
4. Add final JSONL output with append, temp-overwrite, and deduped materialization.
5. Add the GitHub PR stream defaults:
   - primary key: `number` or `node_id`
   - cursor: `updated_at`
6. Add generated docs and skill updates.
7. Add benchmarks before optimizing.

Do not start by adding PostgreSQL, DragonflyDB, or Temporal for this feature. Those dependencies help orchestration and durability later, but they should not be required to prove sync-mode semantics.

