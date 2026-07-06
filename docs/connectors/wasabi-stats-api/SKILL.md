---
name: pm-wasabi-stats-api
description: Wasabi Stats API connector knowledge and safe action guide.
---

# pm-wasabi-stats-api

## Purpose

Reads Wasabi account and bucket storage statistics from the Wasabi Stats API.

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
- start_date
- api_key (secret)

## ETL Streams

- bucket_stats:
  - primary key: id
  - cursor: date
  - fields: bucket(), date(), id(), storage_bytes()
- account_stats:
  - primary key: id
  - cursor: date
  - fields: date(), id(), object_count(), storage_bytes()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Wasabi Stats API read of account/bucket storage usage metrics
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect wasabi-stats-api
```

### Inspect as structured JSON

```bash
pm connectors inspect wasabi-stats-api --json
```

## Agent Rules

- Run pm connectors inspect wasabi-stats-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
