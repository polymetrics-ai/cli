---
name: pm-marketo
description: Marketo connector knowledge and safe action guide.
---

# pm-marketo

## Purpose

Reads Marketo leads, programs, and activities through Marketo REST endpoints. Read-only; does not refresh OAuth tokens internally.

## Icon

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
- base_url
- max_pages
- mode
- page_size
- access_token (secret)

## ETL Streams

- leads:
  - primary key: id
  - fields: createdAt(), email(), id(), updatedAt()
- programs:
  - primary key: id
  - fields: createdAt(), id(), name(), updatedAt()
- activities:
  - primary key: id
  - fields: activityDate(), activityTypeId(), id(), leadId()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
