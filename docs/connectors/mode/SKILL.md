---
name: pm-mode
description: Mode connector knowledge and safe action guide.
---

# pm-mode

## Purpose

Reads Mode collections (spaces), reports, data sources, groups, and memberships through the Mode REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
- workspace
- api_secret (secret)
- api_token (secret)

## ETL Streams

- spaces:
  - primary key: token
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), restricted(), space_type(), state(), token(), updated_at()
- reports:
  - primary key: token
  - cursor: updated_at
  - fields: account_username(), archived(), created_at(), description(), id(), last_run_at(), name(), public(), space_token(), token(), updated_at()
- data_sources:
  - primary key: token
  - cursor: updated_at
  - fields: adapter(), asleep(), created_at(), database(), description(), host(), id(), name(), public(), queryable(), token(), updated_at()
- groups:
  - primary key: token
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), state(), token(), updated_at()
- memberships:
  - primary key: token
  - cursor: created_at
  - fields: admin(), created_at(), email(), id(), token(), username()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Mode API reads performed by the legacy connector via a Tier-2 hook
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
