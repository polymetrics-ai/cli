---
name: pm-workday-rest
description: Workday REST connector knowledge and safe action guide.
---

# pm-workday-rest

## Purpose

Reads Workday REST API resources (workers, organizations, job profiles) with bearer-token authentication. Read-only.

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
- tenant
- access_token (secret)

## ETL Streams

- workers:
  - primary key: id
  - cursor: updated
  - fields: descriptor(), id(), updated()
- organizations:
  - primary key: id
  - cursor: updated
  - fields: descriptor(), id(), type()
- jobs:
  - primary key: id
  - cursor: updated
  - fields: descriptor(), id(), updated()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Workday REST API read of worker, organization, and job profile data (HR/PII-adjacent)
- approval: none; read-only, bearer-token auth
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect workday-rest
```

### Inspect as structured JSON

```bash
pm connectors inspect workday-rest --json
```

## Agent Rules

- Run pm connectors inspect workday-rest before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
