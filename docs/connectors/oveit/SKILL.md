---
name: pm-oveit
description: Oveit connector knowledge and safe action guide.
---

# pm-oveit

## Purpose

Reads Oveit events, orders, and attendees.

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
- email
- page_size
- password (secret)

## ETL Streams

- events:
  - primary key: id
  - fields: created_at(), email(), id(), name(), starts_at(), status(), total()
- orders:
  - primary key: id
  - fields: created_at(), email(), id(), name(), starts_at(), status(), total()
- attendees:
  - primary key: id
  - fields: created_at(), email(), id(), name(), starts_at(), status(), total()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Oveit API read of event, order, and attendee data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect oveit
```

### Inspect as structured JSON

```bash
pm connectors inspect oveit --json
```

## Agent Rules

- Run pm connectors inspect oveit before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
