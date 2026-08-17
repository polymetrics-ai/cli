---
name: pm-genesys
description: Genesys connector knowledge and safe action guide.
---

# pm-genesys

## Purpose

Reads Genesys Cloud users, queues, groups, and divisions through the Genesys Cloud Platform API.

## Icon

- asset: icons/genesys.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.genesys.cloud/api/

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
- scope
- token_url
- client_id (secret)
- client_secret (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: display_name(), email(), id(), name(), state()
- queues:
  - primary key: id
  - fields: description(), id(), name()
- groups:
  - primary key: id
  - fields: description(), id(), name()
- divisions:
  - primary key: id
  - fields: description(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Genesys Cloud Platform API read of user, queue, group, and division data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect genesys
```

### Inspect as structured JSON

```bash
pm connectors inspect genesys --json
```

## Agent Rules

- Run pm connectors inspect genesys before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
