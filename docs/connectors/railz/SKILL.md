---
name: pm-railz
description: Railz connector knowledge and safe action guide.
---

# pm-railz

## Purpose

Reads Railz businesses, connections, customers, invoices, and bills through the Railz REST API. Read-only.

## Icon

- asset: icons/railz.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- access_token (secret)
- api_key (secret)

## ETL Streams

- businesses:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name(), status()
- connections:
  - primary key: id
  - cursor: created_at
  - fields: business_id(), created_at(), id(), status()
- customers:
  - primary key: id
  - fields: business_id(), email(), id(), name()
- invoices:
  - primary key: id
  - fields: business_id(), customer_id(), id(), status(), total_amount(), vendor_id()
- bills:
  - primary key: id
  - fields: business_id(), id(), status(), total_amount(), vendor_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Railz API read of connected-business accounting data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect railz
```

### Inspect as structured JSON

```bash
pm connectors inspect railz --json
```

## Agent Rules

- Run pm connectors inspect railz before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
