---
name: pm-my-hours
description: My Hours connector knowledge and safe action guide.
---

# pm-my-hours

## Purpose

Reads My Hours clients, projects, team members, tags, and time log activity through the My Hours REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: my-hours
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
- email (required)
- logs_batch_size
- mode
- start_date (required)
- password (secret) (required)

## ETL Streams

- clients:
  - primary key: id
  - fields: archived(boolean), custom_id(string), date_archived(string), id(integer), name(string)
- projects:
  - primary key: id
  - fields: archived(boolean), billable(boolean), client_id(integer), client_name(string), date_archived(string), date_created(string), id(integer), name(string)
- users:
  - primary key: id
  - fields: account_owner(boolean), active(boolean), admin(boolean), archived(boolean), billable_rate(number), email(string), id(integer), name(string), rate(number)
- tags:
  - primary key: id
  - fields: archived(boolean), date_archived(string), id(integer), name(string)
- time_logs:
  - primary key: logId
  - cursor: date
  - fields: amount(number), billable(boolean), billable_amount(number), billable_hours(number), client_id(integer), client_name(string), date(string), invoiced(boolean), labor_hours(number), logId(integer), log_duration(number), note(string), project_id(integer), project_name(string), rate(number), tags(string), task_id(integer), task_name(string), user_id(integer), user_name(string)

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
