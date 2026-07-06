---
name: pm-workday
description: Workday connector knowledge and safe action guide.
---

# pm-workday

## Purpose

Reads Workday tenant data (workers, organizations, positions) through conservative Workday API endpoints. Read-only.

## Icon

- asset: icons/workday.svg
- source: official
- review_status: official_verified
- review_url: https://community.workday.com/sites/default/files/file-hosting/productionapi/index.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- tenant
- password (secret)
- username (secret)

## ETL Streams

- workers:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), updated_at()
- organizations:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), type(), updated_at()
- positions:
  - primary key: id
  - cursor: updated_at
  - fields: id(), title(), updated_at(), worker_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Workday tenant API read of worker, organization, and position data (HR/PII-adjacent)
- approval: none; read-only, HTTP Basic auth
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect workday
```

### Inspect as structured JSON

```bash
pm connectors inspect workday --json
```

## Agent Rules

- Run pm connectors inspect workday before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
