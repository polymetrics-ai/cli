---
name: pm-opinion-stage
description: Opinion Stage connector knowledge and safe action guide.
---

# pm-opinion-stage

## Purpose

Reads Opinion Stage items (polls, quizzes, and forms) through the Opinion Stage Public Result API. Read-only.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- items:
  - primary key: id
  - fields: created(string), embed(object), id(string), links(object), modified(string), relationships(object), status(string), title(string), type(string)
- responses:
  - primary key: id
  - fields: answers(array), created(string), duration(number), id(string), item_id(string), links(object), result(object), result_text(string), result_title(string), type(string), utm(object)
- questions:
  - primary key: id
  - fields: created(string), id(string), item_id(string), kind(string), lead(boolean), modified(string), title(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
