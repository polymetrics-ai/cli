---
name: pm-primetric
description: Primetric connector knowledge and safe action guide.
---

# pm-primetric

## Purpose

Reads Primetric employees, projects, clients, and roles through OAuth-authenticated REST list endpoints.

## Icon

- asset: icons/primetric.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.primetric.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- token_url
- client_id (secret)
- client_secret (secret)

## ETL Streams

- employees:
  - primary key: id
  - fields: created_at(), email(), first_name(), id(), last_name(), name(), updated_at()
- projects:
  - primary key: id
  - fields: created_at(), id(), name(), updated_at()
- clients:
  - primary key: id
  - fields: created_at(), id(), name(), updated_at()
- roles:
  - primary key: id
  - fields: created_at(), id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Primetric API read of employee, project, client, and role data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect primetric
```

### Inspect as structured JSON

```bash
pm connectors inspect primetric --json
```

## Agent Rules

- Run pm connectors inspect primetric before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
