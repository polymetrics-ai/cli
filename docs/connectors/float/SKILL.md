---
name: pm-float
description: Float connector knowledge and safe action guide.
---

# pm-float

## Purpose

Reads Float people, projects, clients, tasks, and departments through the Float v3 REST API.

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
- page_size
- access_token (secret)

## ETL Streams

- people:
  - primary key: people_id
  - fields: active(), created(), department_id(), email(), employee_type(), job_title(), modified(), name(), people_id(), people_type_id(), role_id(), start_date()
- projects:
  - primary key: project_id
  - fields: active(), budget_total(), budget_type(), client_id(), color(), created(), default_hourly_rate(), modified(), name(), non_billable(), notes(), project_id(), project_manager(), tags()
- clients:
  - primary key: client_id
  - fields: client_id(), created(), modified(), name()
- tasks:
  - primary key: task_id
  - fields: billable(), created(), modified(), name(), project_id(), task_id(), task_meta_id()
- departments:
  - primary key: department_id
  - fields: department_id(), name(), parent_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Float API read of resource-planning and staffing data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect float
```

### Inspect as structured JSON

```bash
pm connectors inspect float --json
```

## Agent Rules

- Run pm connectors inspect float before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
