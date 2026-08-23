---
name: pm-marketo
description: Marketo connector knowledge and safe action guide.
---

# pm-marketo

## Purpose

Reads Marketo leads, programs, and activities through Marketo REST endpoints. Read-only; does not refresh OAuth tokens internally.

## Icon

- id: marketo
- asset: icons/marketo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.marketo.com/rest-api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- activity_type_ids
- base_url (required)
- max_pages
- mode
- page_size
- access_token (secret)

## ETL Streams

- leads:
  - primary key: id
  - fields: createdAt(string), email(string), id(integer), updatedAt(string)
- programs:
  - primary key: id
  - fields: createdAt(string), id(integer), name(string), updatedAt(string)
- activities:
  - primary key: id
  - fields: activityDate(string), activityTypeId(integer), id(integer), leadId(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Marketo REST API read of lead, program, and activity data
- approval: none; read-only Marketo REST API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect marketo
```

### Inspect as structured JSON

```bash
pm connectors inspect marketo --json
```

## Agent Rules

- Run pm connectors inspect marketo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
