---
name: pm-basecamp
description: Basecamp connector knowledge and safe action guide.
---

# pm-basecamp

## Purpose

Reads Basecamp 3 projects, people, and account activity events through the Basecamp REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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

- account_id
- base_url
- mode
- start_date
- client_id (secret)
- client_refresh_token_2 (secret)
- client_secret (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: updated_at
  - fields: app_url(), bookmark_url(), created_at(), description(), id(), name(), purpose(), status(), updated_at(), url()
- people:
  - primary key: id
  - cursor: updated_at
  - fields: admin(), client(), created_at(), email_address(), id(), name(), owner(), personable_type(), time_zone(), title(), updated_at()
- events:
  - primary key: id
  - cursor: created_at
  - fields: action(), created_at(), id(), kind(), recording_id(), summary()

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
