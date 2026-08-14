---
name: pm-timely
description: Timely connector knowledge and safe action guide.
---

# pm-timely

## Purpose

Reads users, projects, clients, calendar/time events, time entries (hours), tags (labels), and teams from the Timely API. Read-only: every Timely mutation endpoint requires a nested single-key JSON body envelope (e.g. {"client": {...}}) the engine's declarative write dialect cannot express.

## Icon

- id: timely
- asset: icons/timely.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://dev.timelyapp.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url
- start_date
- bearer_token (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - fields: created_at(string), email(string), id(string), name(string), updated_at(string)
- projects:
  - primary key: id
  - fields: client_id(string), created_at(string), id(string), name(string), updated_at(string)
- clients:
  - primary key: id
  - fields: created_at(string), id(string), name(string), updated_at(string)
- events:
  - primary key: id
  - fields: created_at(string), duration(string), id(string), project_id(string), updated_at(string), user_id(string)
- hours:
  - primary key: id
  - fields: billable(boolean), billed(boolean), created_at(integer), day(string), deleted(boolean), external_id(string), from(string), id(integer), note(string), project_id(integer), to(string), uid(string), updated_at(integer), user_id(integer)
- labels:
  - primary key: id
  - fields: active(boolean), created_at(string), emoji(string), external_id(string), id(integer), name(string), parent_id(integer), sequence(integer), updated_at(string)
- teams:
  - primary key: id
  - fields: color(string), emoji(string), external_id(string), id(integer), name(string), project_ids(array), user_ids(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Timely API read of user, project, client, time event/entry, tag, and team data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect timely
```

### Inspect as structured JSON

```bash
pm connectors inspect timely --json
```

## Agent Rules

- Run pm connectors inspect timely before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
