---
name: pm-pocket
description: Pocket connector knowledge and safe action guide.
---

# pm-pocket

## Purpose

Reads saved Pocket items through the v3 retrieve API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/pocket.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- content_type
- detail_type
- domain
- favorite
- mode
- search
- since
- sort
- state
- tag
- access_token (secret)
- consumer_key (secret)

## ETL Streams

- items:
  - primary key: item_id
  - cursor: updated_at
  - fields: excerpt(), item_id(), title(), updated_at(), url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Pocket API reads performed by the legacy connector via a Tier-2 hook
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
