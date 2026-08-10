---
name: pm-orb
description: Orb connector knowledge and safe action guide.
---

# pm-orb

## Purpose

Reads Orb customers, subscriptions, plans, and invoices.

## Icon

- id: orb
- asset: icons/orb.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.withorb.com/reference/api-reference

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
- start_date
- api_key (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: created_at
  - fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)
- subscriptions:
  - primary key: id
  - cursor: created_at
  - fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)
- plans:
  - primary key: id
  - cursor: created_at
  - fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)
- invoices:
  - primary key: id
  - cursor: created_at
  - fields: amount(integer), created_at(string), currency(string), customer_id(string), email(string), id(string), name(string), plan_id(string), status(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Orb API read of customer and billing data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect orb
```

### Inspect as structured JSON

```bash
pm connectors inspect orb --json
```

## Agent Rules

- Run pm connectors inspect orb before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
