---
name: pm-rocketlane
description: Rocketlane connector knowledge and safe action guide.
---

# pm-rocketlane

## Purpose

Reads Rocketlane projects, tasks, customers, users, and time entries through the REST API.

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
- created_after
- mode
- project_id
- status
- updated_after
- api_key (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(string), id(string), name(string), status(string), stream(string), updated_at(string)
- tasks:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), name(string), project_id(string), status(string), stream(string), updated_at(string)
- customers:
  - primary key: id
  - cursor: updated_at
  - fields: domain(string), id(string), name(string), stream(string), updated_at(string)
- users:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), id(string), name(string), status(string), stream(string), updated_at(string)
- time_entries:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), minutes(integer), project_id(string), stream(string), task_id(string), updated_at(string), user_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Rocketlane API read of project, task, customer, and time-entry data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect rocketlane
```

### Inspect as structured JSON

```bash
pm connectors inspect rocketlane --json
```

## Agent Rules

- Run pm connectors inspect rocketlane before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
