---
name: pm-taboola
description: Taboola connector knowledge and safe action guide.
---

# pm-taboola

## Purpose

Reads Taboola campaigns through the Backstage API. Read-only.

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

- account_id
- base_url
- max_pages
- mode
- page_size
- client_id (secret)
- client_secret (secret)

## ETL Streams

- campaigns:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Taboola Backstage API read of campaign data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect taboola
```

### Inspect as structured JSON

```bash
pm connectors inspect taboola --json
```

## Agent Rules

- Run pm connectors inspect taboola before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
