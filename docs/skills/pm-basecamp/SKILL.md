---
name: pm-basecamp
description: Basecamp connector knowledge and safe action guide.
---

# pm-basecamp

## Purpose

Reads Basecamp 3 projects, people, and account activity events through the Basecamp REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: simple-icons-basecamp
- asset: icons/simple-icons/basecamp.svg
- title: Basecamp
- simple_icon_slug: basecamp
- simple_icon_hex: 1D2D35
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Basecamp
- match: exact-name-or-slug
- matched_by: basecamp

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url
- mode
- start_date (required)
- client_id (secret) (required)
- client_refresh_token_2 (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - cursor: updated_at
  - fields: app_url(string), bookmark_url(string), created_at(string), description(string), id(integer), name(string), purpose(string), status(string), updated_at(string), url(string)
- people:
  - primary key: id
  - cursor: updated_at
  - fields: admin(boolean), client(boolean), created_at(string), email_address(string), id(integer), name(string), owner(boolean), personable_type(string), time_zone(string), title(string), updated_at(string)
- events:
  - primary key: id
  - cursor: created_at
  - fields: action(string), created_at(string), id(integer), kind(string), recording_id(integer), summary(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Basecamp API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect basecamp
```

### Inspect as structured JSON

```bash
pm connectors inspect basecamp --json
```

## Agent Rules

- Run pm connectors inspect basecamp before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
