---
name: pm-ruddr
description: Ruddr connector knowledge and safe action guide.
---

# pm-ruddr

## Purpose

Reads Ruddr clients, projects, and time entries through the Ruddr API. Read-only.

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
- mode
- workspace_id
- api_key (secret)

## ETL Streams

- clients:
  - primary key: id
  - fields: id(), name(), stream()
- projects:
  - primary key: id
  - fields: id(), name(), project_id(), stream()
- time_entries:
  - primary key: id
  - fields: hours(), id(), name(), project_id(), stream()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Ruddr API read of client, project, and time-entry data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ruddr
```

### Inspect as structured JSON

```bash
pm connectors inspect ruddr --json
```

## Agent Rules

- Run pm connectors inspect ruddr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
