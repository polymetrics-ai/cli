---
name: pm-iterable
description: Iterable connector knowledge and safe action guide.
---

# pm-iterable

## Purpose

Reads Iterable lists, campaigns, and templates through the Iterable REST API. Read-only.

## Icon

- id: iterable
- asset: icons/iterable.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api.iterable.com/api/docs

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- api_key (secret) (required)

## ETL Streams

- lists:
  - primary key: id
  - fields: createdAt(string), id(integer), listType(string), name(string), updatedAt(string)
- campaigns:
  - primary key: id
  - fields: createdAt(string), id(integer), name(string), updatedAt(string)
- templates:
  - primary key: id
  - fields: createdAt(string), id(integer), name(string), updatedAt(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Iterable API read of lists, campaigns, and templates
- approval: none; read-only marketing-data API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect iterable
```

### Inspect as structured JSON

```bash
pm connectors inspect iterable --json
```

## Agent Rules

- Run pm connectors inspect iterable before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
