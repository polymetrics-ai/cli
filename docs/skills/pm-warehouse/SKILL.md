---
name: pm-warehouse
description: Local Warehouse connector knowledge and safe action guide.
---

# pm-warehouse

## Purpose

Local JSONL warehouse destination used by the dependency-free MVP.

## Capabilities

- check=true catalog=true read=true write=true query=true
- Integration type: database

## Authentication

- No secret authentication is required for this connector.

## Configuration

- path: Local warehouse directory.

## ETL Streams

- tables: Local JSONL warehouse tables.

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
- Source modes: full_refresh, incremental
- Destination modes: append, overwrite, append_dedup, overwrite_dedup

## Security

- read risk: local warehouse read
- write risk: local file write
- mutation risk: local dependency-free warehouse writes
- approval: not required for ETL destination writes; reverse ETL still requires approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect warehouse
```

### Inspect as structured JSON

```bash
pm connectors inspect warehouse --json
```

### Warehouse credential

```bash
pm credentials add warehouse-local --connector warehouse --config path=$ROOT/.polymetrics/warehouse
pm query run --table sample_customers --limit 5 --json
```

## Agent Rules

- Run pm connectors inspect warehouse before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.

