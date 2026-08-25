---
name: pm-looker
description: Looker connector knowledge and safe action guide.
---

# pm-looker

## Purpose

Reads Looker users, groups, folders, looks, and dashboards through the Looker API 4.0.

## Icon

- id: looker
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

- base_url (required)
- mode
- token_url
- access_token (secret)
- client_id (secret)
- client_secret (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: display_name(string), email(string), id(string)
- groups:
  - primary key: id
  - fields: id(string), name(string)
- folders:
  - primary key: id
  - fields: id(string), name(string)
- looks:
  - primary key: id
  - fields: folder_id(string), id(string), title(string)
- dashboards:
  - primary key: id
  - fields: folder_id(string), id(string), title(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Looker API read of users, groups, folders, looks, and dashboards
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Looker's declared typed write actions.
- Usage: pm looker <command> [flags]

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

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
