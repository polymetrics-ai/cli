---
name: pm-qualaroo
description: Qualaroo connector knowledge and safe action guide.
---

# pm-qualaroo

## Purpose

Reads Qualaroo nudges and reporting response records through the Qualaroo API. Read-only.

## Icon

- asset: icons/qualaroo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://help.qualaroo.com/hc/en-us/articles/201969438-The-Qualaroo-API

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- survey_id
- api_key (secret)
- api_secret (secret)

## ETL Streams

- nudges:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), status(), updated_at()
- responses:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), id(), nudge_id(), updated_at()
- survey_responses:
  - primary key: id
  - cursor: time
  - fields: answered_questions(), id(), identity(), ip_address(), page(), properties(), referrer(), time(), token(), user_agent()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Qualaroo API read of survey nudge and reporting response data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect qualaroo
```

### Inspect as structured JSON

```bash
pm connectors inspect qualaroo --json
```

## Agent Rules

- Run pm connectors inspect qualaroo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
