---
name: pm-orb
description: Orb connector knowledge and safe action guide.
---

# pm-orb

## Purpose

Reads Orb customers, subscriptions, plans, and invoices.

## Icon

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
- api_key (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), customer_id(), email(), id(), name(), plan_id(), status(), updated_at()
- subscriptions:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), customer_id(), email(), id(), name(), plan_id(), status(), updated_at()
- plans:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), customer_id(), email(), id(), name(), plan_id(), status(), updated_at()
- invoices:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), customer_id(), email(), id(), name(), plan_id(), status(), updated_at()

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
