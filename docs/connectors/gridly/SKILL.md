---
name: pm-gridly
description: Gridly connector knowledge and safe action guide.
---

# pm-gridly

## Purpose

Reads Gridly views, per-view records (with flattened column cells), and per-view branches through the Gridly REST API.

## Icon

- asset: icons/gridly.svg
- source: official
- review_status: official_verified
- review_url: https://www.gridly.com/docs/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- view_ids
- api_key (secret)

## ETL Streams

- views:
  - primary key: id
  - fields: id(), name()
- records:
  - primary key: view_id, id
  - fields: cells(), id(), path(), view_id()
- branches:
  - primary key: view_id, id
  - fields: id(), name(), view_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Gridly API read of view/grid content
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gridly
```

### Inspect as structured JSON

```bash
pm connectors inspect gridly --json
```

## Agent Rules

- Run pm connectors inspect gridly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
