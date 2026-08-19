---
name: pm-genesys
description: Genesys connector knowledge and safe action guide.
---

# pm-genesys

## Purpose

Reads Genesys Cloud users, queues, groups, and divisions through the Genesys Cloud Platform API.

## Icon

- id: genesys
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

- base_url (required)
- max_pages
- mode
- page_size
- scope
- token_url (required)
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - fields: display_name(string), email(string), id(string), name(string), state(string)
- queues:
  - primary key: id
  - fields: description(string), id(string), name(string)
- groups:
  - primary key: id
  - fields: description(string), id(string), name(string)
- divisions:
  - primary key: id
  - fields: description(string), id(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
