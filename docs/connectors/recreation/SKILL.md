---
name: pm-recreation
description: Recreation.gov connector knowledge and safe action guide.
---

# pm-recreation

## Purpose

Reads Recreation.gov RIDB facilities, campsites, activities, organizations, and recreation areas through the RIDB REST API.

## Icon

- asset: icons/recreation.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- facilities:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), type(), updated_at()
- campsites:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), type(), updated_at()
- activities:
  - primary key: id
  - fields: id(), name()
- organizations:
  - primary key: id
  - fields: id(), name()
- recareas:
  - primary key: id
  - fields: id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Recreation.gov RIDB API read of public facility, campsite, activity, organization, and recreation-area data
- approval: none; read-only public-data API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect recreation
```

### Inspect as structured JSON

```bash
pm connectors inspect recreation --json
```

## Agent Rules

- Run pm connectors inspect recreation before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
