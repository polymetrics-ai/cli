---
name: pm-monday
description: Monday connector knowledge and safe action guide.
---

# pm-monday

## Purpose

Reads monday.com boards, items, users, teams, and tags through the monday.com GraphQL API. Read-only.

## Icon

- id: monday
- asset: icons/monday.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.monday.com/api-reference/docs

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_version
- base_url
- max_pages
- page_size
- access_token (secret)
- api_token (secret)

## ETL Streams

- boards:
  - primary key: id
  - cursor: updated_at
  - fields: board_kind(string), description(string), id(string), name(string), state(string), type(string), updated_at(string), workspace_id(string)
- items:
  - primary key: id
  - cursor: updated_at
  - fields: board_id(string), board_name(string), created_at(string), group_id(string), group_title(string), id(string), name(string), state(string), updated_at(string)
- users:
  - primary key: id
  - fields: created_at(string), email(string), enabled(boolean), id(string), is_admin(boolean), is_guest(boolean), is_pending(boolean), name(string)
- teams:
  - primary key: id
  - fields: id(string), name(string), picture_url(string)
- tags:
  - primary key: id
  - fields: color(string), id(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external monday.com GraphQL API read of boards/items/users/teams/tags
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect monday
```

### Inspect as structured JSON

```bash
pm connectors inspect monday --json
```

## Agent Rules

- Run pm connectors inspect monday before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
