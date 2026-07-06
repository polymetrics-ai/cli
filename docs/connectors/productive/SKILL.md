---
name: pm-productive
description: Productive connector knowledge and safe action guide.
---

# pm-productive

## Purpose

Reads Productive projects, people, companies, and tasks through the Productive JSON:API-style REST API (read-only).

## Icon

- asset: icons/productive.svg
- source: official
- review_status: official_verified
- review_url: https://developer.productive.io/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- organization_id
- api_key (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), type(), updated_at()
- people:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), type(), updated_at()
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), type(), updated_at()
- tasks:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), type(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Productive API read of projects, people, companies, and tasks
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect productive
```

### Inspect as structured JSON

```bash
pm connectors inspect productive --json
```

## Agent Rules

- Run pm connectors inspect productive before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
