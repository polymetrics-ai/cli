---
name: pm-rocketlane
description: Rocketlane connector knowledge and safe action guide.
---

# pm-rocketlane

## Purpose

Reads Rocketlane projects, tasks, customers, users, and time entries through the REST API.

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
- created_after
- mode
- project_id
- status
- updated_after
- api_key (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(), id(), name(), status(), stream(), updated_at()
- tasks:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), project_id(), status(), stream(), updated_at()
- customers:
  - primary key: id
  - cursor: updated_at
  - fields: domain(), id(), name(), stream(), updated_at()
- users:
  - primary key: id
  - cursor: updated_at
  - fields: email(), id(), name(), status(), stream(), updated_at()
- time_entries:
  - primary key: id
  - cursor: updated_at
  - fields: id(), minutes(), project_id(), stream(), task_id(), updated_at(), user_id()

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
