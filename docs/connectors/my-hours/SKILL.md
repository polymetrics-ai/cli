---
name: pm-my-hours
description: My Hours connector knowledge and safe action guide.
---

# pm-my-hours

## Purpose

Reads My Hours clients, projects, team members, tags, and time log activity through the My Hours REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/my-hours.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://myhours.com/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- email
- logs_batch_size
- mode
- start_date
- password (secret)

## ETL Streams

- clients:
  - primary key: id
  - fields: archived(), custom_id(), date_archived(), id(), name()
- projects:
  - primary key: id
  - fields: archived(), billable(), client_id(), client_name(), date_archived(), date_created(), id(), name()
- users:
  - primary key: id
  - fields: account_owner(), active(), admin(), archived(), billable_rate(), email(), id(), name(), rate()
- tags:
  - primary key: id
  - fields: archived(), date_archived(), id(), name()
- time_logs:
  - primary key: logId
  - cursor: date
  - fields: amount(), billable(), billable_amount(), billable_hours(), client_id(), client_name(), date(), invoiced(), labor_hours(), logId(), log_duration(), note(), project_id(), project_name(), rate(), tags(), task_id(), task_name(), user_id(), user_name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external My Hours API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect my-hours
```

### Inspect as structured JSON

```bash
pm connectors inspect my-hours --json
```

## Agent Rules

- Run pm connectors inspect my-hours before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
