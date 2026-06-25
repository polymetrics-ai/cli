# PRD: Airbyte-Style Sync Modes

## Goal

Add complete Airbyte-style ETL sync-mode semantics to the `pm` Go CLI for the dependency-free local JSONL warehouse path.

## User Value

Agents and humans can choose the correct sync behavior for a stream without guessing how the connector writes data. The CLI will support append, overwrite, incremental, and deduped final outputs with safe checkpoint behavior.

## Required Modes

- `full_refresh_append`
- `full_refresh_overwrite`
- `full_refresh_overwrite_deduped`
- `incremental_append`
- `incremental_append_deduped`

## Requirements

- Split source read mode from destination write mode internally.
- Validate cursor and primary-key requirements before running ETL.
- Keep connector code focused on capabilities and record emission.
- Implement raw JSONL history and final JSONL materialization for the local warehouse.
- Preserve previous final output when overwrite or deduped overwrite runs fail.
- Commit cursor state only after successful raw writes and final materialization.
- Update connector manifests, docs, generated skills, and benchmarks.

## Non-Goals

- PostgreSQL-backed final table implementation.
- Temporal orchestration changes.
- Live external API benchmark runs.
- New Go module dependencies.

