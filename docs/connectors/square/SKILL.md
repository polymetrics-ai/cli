---
name: pm-square
description: Square connector knowledge and safe action guide.
---

# pm-square

## Purpose

Reads Square payments, refunds, customers, and locations through the Square Connect v2 REST API.

## Icon

- asset: icons/square.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.squareup.com/reference/square

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

- payments:
  - primary key: id
  - cursor: updated_at
  - fields: amount_money(), created_at(), id(), location_id(), order_id(), processing_fee(), receipt_number(), source_type(), status(), total_money(), updated_at()
- refunds:
  - primary key: id
  - cursor: updated_at
  - fields: amount_money(), created_at(), id(), location_id(), order_id(), payment_id(), processing_fee(), reason(), status(), updated_at()
- customers:
  - primary key: id
  - cursor: updated_at
  - fields: company_name(), created_at(), creation_source(), email_address(), family_name(), given_name(), id(), phone_number(), reference_id(), updated_at()
- locations:
  - primary key: id
  - fields: country(), created_at(), currency(), id(), merchant_id(), name(), status(), timezone(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Square API read of payments, refunds, customer, and location data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect square
```

### Inspect as structured JSON

```bash
pm connectors inspect square --json
```

## Agent Rules

- Run pm connectors inspect square before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
