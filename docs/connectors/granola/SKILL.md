---
name: pm-granola
description: Granola connector knowledge and safe action guide.
---

# pm-granola

## Purpose

Reads Granola meeting notes metadata and full note detail (summary, owner, attendees, calendar event) through the Granola public API (read-only).

## Icon

- asset: icons/source-granola.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.granola.ai/introduction

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- start_date
- api_key (secret)

## ETL Streams

- notes:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), object(), owner_email(), owner_name(), title(), updated_at()
- detailed_notes:
  - primary key: id
  - cursor: created_at
  - fields: attendees(), calendar_event(), created_at(), folders(), id(), object(), owner_email(), owner_name(), summary(), title(), transcript(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Granola API read of meeting notes metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect granola
```

### Inspect as structured JSON

```bash
pm connectors inspect granola --json
```

## Agent Rules

- Run pm connectors inspect granola before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
