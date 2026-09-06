---
name: pm-copper
description: Copper connector knowledge and safe action guide.
---

# pm-copper

## Purpose

Reads Copper CRM records through fixed typed search routes.

## Icon

- id: copper
- asset: icons/copper.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.copper.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- user_email (required)
- api_key (secret) (required)

## ETL Streams

- people:
  - primary key: id
  - cursor: date_modified
  - fields: date_modified(integer), id(integer)
- companies:
  - primary key: id
  - cursor: date_modified
  - fields: date_modified(integer), id(integer)
- opportunities:
  - primary key: id
  - cursor: date_modified
  - fields: date_modified(integer), id(integer)
- leads:
  - primary key: id
  - cursor: date_modified
  - fields: date_modified(integer), id(integer)
- tasks:
  - primary key: id
  - cursor: date_modified
  - fields: date_modified(integer), id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded Copper search requests use fixed API routes and declared three-header authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect copper
```

### Inspect as structured JSON

```bash
pm connectors inspect copper --json
```

## Agent Rules

- Run pm connectors inspect copper before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
