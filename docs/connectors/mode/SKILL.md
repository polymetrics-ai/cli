---
name: pm-mode
description: Mode connector knowledge and safe action guide.
---

# pm-mode

## Purpose

Reads Mode workspace collections through fixed HAL+JSON REST routes.

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

- workspace (required)
- api_secret (secret) (required)
- api_token (secret) (required)

## ETL Streams

- spaces:
  - primary key: token
  - cursor: updated_at
  - fields: created_at(string), description(string), id(integer), name(string), restricted(boolean), space_type(string), state(string), token(string), updated_at(string)
- reports:
  - primary key: token
  - cursor: updated_at
  - fields: account_username(string), archived(boolean), created_at(string), description(string), id(integer), last_run_at(string), name(string), public(boolean), space_token(string), token(string), updated_at(string)
- data_sources:
  - primary key: token
  - cursor: updated_at
  - fields: adapter(string), asleep(boolean), created_at(string), database(string), description(string), host(string), id(integer), name(string), public(boolean), queryable(boolean), token(string), updated_at(string)
- groups:
  - primary key: token
  - cursor: updated_at
  - fields: created_at(string), description(string), id(integer), name(string), state(string), token(string), updated_at(string)
- memberships:
  - primary key: token
  - cursor: created_at
  - fields: admin(boolean), created_at(string), email(string), id(integer), token(string), username(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded HAL+JSON reads use the fixed Mode origin and declared Basic authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mode
```

### Inspect as structured JSON

```bash
pm connectors inspect mode --json
```

## Agent Rules

- Run pm connectors inspect mode before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
