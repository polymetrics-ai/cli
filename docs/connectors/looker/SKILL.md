---
name: pm-looker
description: Looker connector knowledge and safe action guide.
---

# pm-looker

## Purpose

Reads Looker users, groups, folders, looks, and dashboards through the Looker API 4.0.

## Icon

- asset: icons/looker.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://cloud.google.com/looker/docs/reference/looker-api/latest

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- token_url
- access_token (secret)
- client_id (secret)
- client_secret (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: display_name(), email(), id()
- groups:
  - primary key: id
  - fields: id(), name()
- folders:
  - primary key: id
  - fields: id(), name()
- looks:
  - primary key: id
  - fields: folder_id(), id(), title()
- dashboards:
  - primary key: id
  - fields: folder_id(), id(), title()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Looker API read of users, groups, folders, looks, and dashboards
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect looker
```

### Inspect as structured JSON

```bash
pm connectors inspect looker --json
```

## Agent Rules

- Run pm connectors inspect looker before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
