---
name: pm-kisi
description: Kisi connector knowledge and safe action guide.
---

# pm-kisi

## Purpose

Reads Kisi physical access-control data: members, locks, groups, users, and logins via the Kisi REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
- api_key (secret)

## ETL Streams

- members:
  - primary key: id
  - fields: access_enabled(), confirmed(), created_at(), email(), id(), name(), role_id(), updated_at()
- locks:
  - primary key: id
  - fields: created_at(), description(), geofence_restriction_enabled(), id(), name(), online(), place_id(), updated_at()
- groups:
  - primary key: id
  - fields: created_at(), description(), id(), login_count(), name(), place_id(), updated_at()
- users:
  - primary key: id
  - fields: confirmed(), created_at(), email(), id(), name(), updated_at()
- logins:
  - primary key: id
  - fields: created_at(), id(), last_used_at(), name(), type(), updated_at(), user_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Kisi API read of physical access-control data
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect kisi
```

### Inspect as structured JSON

```bash
pm connectors inspect kisi --json
```

## Agent Rules

- Run pm connectors inspect kisi before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
