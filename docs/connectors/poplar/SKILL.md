---
name: pm-poplar
description: Poplar connector knowledge and safe action guide.
---

# pm-poplar

## Purpose

Reads Poplar campaigns and orders through read-only REST list endpoints.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_token (secret)

## ETL Streams

- campaigns:
  - primary key: id
  - fields: created_at(), id(), name(), status()
- orders:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Poplar API read of campaign and order data
- approval: none; read-only, no writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect poplar
```

### Inspect as structured JSON

```bash
pm connectors inspect poplar --json
```

## Agent Rules

- Run pm connectors inspect poplar before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
