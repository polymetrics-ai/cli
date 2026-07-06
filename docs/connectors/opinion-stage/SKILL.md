---
name: pm-opinion-stage
description: Opinion Stage connector knowledge and safe action guide.
---

# pm-opinion-stage

## Purpose

Reads Opinion Stage items (polls, quizzes, and forms) through the Opinion Stage Public Result API. Read-only.

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
- api_key (secret)

## ETL Streams

- items:
  - primary key: id
  - fields: created(), embed(), id(), links(), modified(), relationships(), status(), title(), type()
- responses:
  - primary key: id
  - fields: answers(), created(), duration(), id(), item_id(), links(), result(), result_text(), result_title(), type(), utm()
- questions:
  - primary key: id
  - fields: created(), id(), item_id(), kind(), lead(), modified(), title(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Opinion Stage API read of item directory
- approval: none; read-only API-key access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect opinion-stage
```

### Inspect as structured JSON

```bash
pm connectors inspect opinion-stage --json
```

## Agent Rules

- Run pm connectors inspect opinion-stage before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
