---
name: pm-iterable
description: Iterable connector knowledge and safe action guide.
---

# pm-iterable

## Purpose

Reads Iterable lists, campaigns, and templates through the Iterable REST API. Read-only.

## Icon

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
- api_key (secret)

## ETL Streams

- lists:
  - primary key: id
  - fields: createdAt(), id(), listType(), name(), updatedAt()
- campaigns:
  - primary key: id
  - fields: createdAt(), id(), name(), updatedAt()
- templates:
  - primary key: id
  - fields: createdAt(), id(), name(), updatedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
