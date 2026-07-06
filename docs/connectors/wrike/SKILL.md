---
name: pm-wrike
description: Wrike connector knowledge and safe action guide.
---

# pm-wrike

## Purpose

Reads Wrike tasks, folders, and contacts through the Wrike REST API. Read-only.

## Icon

- asset: icons/wrike.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.wrike.com/api/v4/

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
- access_token (secret)

## ETL Streams

- tasks:
  - primary key: id
  - cursor: updatedDate
  - fields: id(), title(), updatedDate()
- folders:
  - primary key: id
  - cursor: updatedDate
  - fields: id(), title(), updatedDate()
- contacts:
  - primary key: id
  - fields: firstName(), id(), lastName()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Wrike API read of task, folder, and contact data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect wrike
```

### Inspect as structured JSON

```bash
pm connectors inspect wrike --json
```

## Agent Rules

- Run pm connectors inspect wrike before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
