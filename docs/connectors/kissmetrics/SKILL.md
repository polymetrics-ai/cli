---
name: pm-kissmetrics
description: Kissmetrics connector knowledge and safe action guide.
---

# pm-kissmetrics

## Purpose

Reads Kissmetrics products, reports, events, and properties through the Kissmetrics query API using HTTP Basic authentication.

## Icon

- asset: icons/kissmetrics.svg
- source: official
- review_status: official_verified
- review_url: https://support.kissmetrics.io/reference

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- product_id
- username
- password (secret)

## ETL Streams

- products:
  - primary key: id
  - fields: created_at(), id(), name(), updated_at()
- reports:
  - primary key: id
  - fields: created_at(), id(), name(), product_id(), type(), updated_at()
- events:
  - primary key: id
  - fields: created_at(), display_name(), id(), name(), product_id()
- properties:
  - primary key: id
  - fields: created_at(), display_name(), id(), name(), product_id(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Kissmetrics query API read of product analytics metadata
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect kissmetrics
```

### Inspect as structured JSON

```bash
pm connectors inspect kissmetrics --json
```

## Agent Rules

- Run pm connectors inspect kissmetrics before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
