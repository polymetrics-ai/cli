---
name: pm-babelforce
description: Babelforce connector knowledge and safe action guide.
---

# pm-babelforce

## Purpose

Reads Babelforce call reporting, recordings, numbers, and users through fixed Babelforce v2 REST routes.

## Icon

- id: babelforce
- asset: icons/babelforce.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api.babelforce.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- region
- access_key_id (secret)
- access_token (secret)

## ETL Streams

- calls:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string)
- calls_extended:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string)
- recordings:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string)
- numbers:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string)
- users:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded Babelforce v2 reads use a source-declared regional provider origin and dual-header authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect babelforce
```

### Inspect as structured JSON

```bash
pm connectors inspect babelforce --json
```

## Agent Rules

- Run pm connectors inspect babelforce before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
