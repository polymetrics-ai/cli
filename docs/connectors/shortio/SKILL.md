---
name: pm-shortio
description: Short.io connector knowledge and safe action guide.
---

# pm-shortio

## Purpose

Reads Short.io links and domains through the Short.io REST API.

## Icon

- asset: icons/shortio.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.short.io/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- api_key (secret)

## ETL Streams

- links:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), path(), title(), updated_at()
- domains:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), path(), title(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Short.io API read of link and domain data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shortio
```

### Inspect as structured JSON

```bash
pm connectors inspect shortio --json
```

## Agent Rules

- Run pm connectors inspect shortio before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
