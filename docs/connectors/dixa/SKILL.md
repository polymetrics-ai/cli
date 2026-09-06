---
name: pm-dixa
description: Dixa connector knowledge and safe action guide.
---

# pm-dixa

## Purpose

Reads Dixa conversation export records through fixed bearer-authenticated export routes.

## Icon

- id: dixa
- asset: icons/dixa.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.dixa.io/openapi/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- updated_after
- updated_before
- api_token (secret)

## ETL Streams

- conversations:
  - primary key: id
  - cursor: updated_at
  - fields: id(integer), updated_at(integer)
- conversation_queue:
  - primary key: id
  - cursor: updated_at
  - fields: id(integer), updated_at(integer)
- conversation_rating:
  - primary key: id
  - cursor: updated_at
  - fields: id(integer), updated_at(integer)
- conversation_assignment:
  - primary key: id
  - cursor: updated_at
  - fields: id(integer), updated_at(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded Dixa conversation export reads use a fixed provider origin and declared bearer authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dixa
```

### Inspect as structured JSON

```bash
pm connectors inspect dixa --json
```

## Agent Rules

- Run pm connectors inspect dixa before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
