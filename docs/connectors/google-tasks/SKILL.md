---
name: pm-google-tasks
description: Google Tasks connector knowledge and safe action guide.
---

# pm-google-tasks

## Purpose

Reads Google task lists and tasks through the Google Tasks REST API.

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
- records_limit
- api_key (secret)

## ETL Streams

- tasklists:
  - primary key: id
  - cursor: updated
  - fields: etag(), id(), kind(), self_link(), title(), updated()
- tasks:
  - primary key: id
  - cursor: updated
  - fields: completed(), deleted(), due(), etag(), hidden(), id(), kind(), notes(), parent(), position(), self_link(), status(), tasklist_id(), title(), updated()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Google Tasks API read of the authenticated user's task lists and tasks
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-tasks
```

### Inspect as structured JSON

```bash
pm connectors inspect google-tasks --json
```

## Agent Rules

- Run pm connectors inspect google-tasks before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
