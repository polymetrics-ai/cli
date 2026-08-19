---
name: pm-kissmetrics
description: Kissmetrics connector knowledge and safe action guide.
---

# pm-kissmetrics

## Purpose

Reads Kissmetrics products, reports, events, and properties through the Kissmetrics query API using HTTP Basic authentication.

## Icon

- id: kissmetrics
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
- username (required)
- password (secret) (required)

## ETL Streams

- products:
  - primary key: id
  - fields: created_at(string), id(string), name(string), updated_at(string)
- reports:
  - primary key: id
  - fields: created_at(string), id(string), name(string), product_id(string), type(string), updated_at(string)
- events:
  - primary key: id
  - fields: created_at(string), display_name(string), id(string), name(string), product_id(string)
- properties:
  - primary key: id
  - fields: created_at(string), display_name(string), id(string), name(string), product_id(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
