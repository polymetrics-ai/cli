---
name: pm-pocket
description: Pocket connector knowledge and safe action guide.
---

# pm-pocket

## Purpose

Reads saved Pocket items through the fixed v3 retrieve API.

## Icon

- id: pocket
- asset: icons/pocket.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- contentType
- detail_type
- favorite
- since
- sort
- state
- tag
- access_token (secret) (required)
- consumer_key (secret) (required)

## ETL Streams

- items:
  - primary key: item_id
  - cursor: updated_at
  - fields: excerpt(string), item_id(string), title(string), updated_at(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: Bounded POST reads use the fixed Pocket origin and source-declared request credentials.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pocket
```

### Inspect as structured JSON

```bash
pm connectors inspect pocket --json
```

## Agent Rules

- Run pm connectors inspect pocket before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
