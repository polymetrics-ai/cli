---
name: pm-zoho-billing
description: Zoho Billing connector knowledge and safe action guide.
---

# pm-zoho-billing

## Purpose

Reads Zoho Billing customers, subscriptions, and invoices through the Zoho Billing REST API.

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
- organization_id
- access_token (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(), display_name(), id(), name(), status(), updated_at(), updated_time()
- subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), subscription_id(), updated_at(), updated_time()
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: id(), invoice_id(), invoice_number(), name(), status(), updated_at(), updated_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zoho Billing API read of customer and billing data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-billing
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-billing --json
```

## Agent Rules

- Run pm connectors inspect zoho-billing before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
