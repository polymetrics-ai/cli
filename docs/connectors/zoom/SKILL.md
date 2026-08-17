---
name: pm-zoom
description: Zoom connector knowledge and safe action guide.
---

# pm-zoom

## Purpose

Reads Zoom users, meetings, and webinars through the Zoom REST API.

## Icon

- asset: icons/zoom.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.zoom.us/docs/api/

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
- user_id
- access_token (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: email(), id(), name(), updated_at()
- meetings:
  - primary key: id
  - fields: email(), id(), name(), updated_at()
- webinars:
  - primary key: id
  - fields: email(), id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zoom API read of user, meeting, and webinar data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoom
```

### Inspect as structured JSON

```bash
pm connectors inspect zoom --json
```

## Agent Rules

- Run pm connectors inspect zoom before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
