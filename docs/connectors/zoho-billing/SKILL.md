---
name: pm-zoho-billing
description: Zoho Billing connector knowledge and safe action guide.
---

# pm-zoho-billing

## Purpose

Reads Zoho Billing customers, subscriptions, and invoices through the Zoho Billing REST API.

## Icon

- id: simple-icons-zoho-billing
- asset: icons/simple-icons/zoho-billing.svg
- title: Zoho
- simple_icon_slug: zoho
- simple_icon_hex: E42527
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Zoho
- match: curated-alias
- matched_by: zoho

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- organization_id
- access_token (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(string), display_name(string), id(string), name(string), status(string), updated_at(string), updated_time(string)
- subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), name(string), status(string), subscription_id(string), updated_at(string), updated_time(string)
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), invoice_id(string), invoice_number(string), name(string), status(string), updated_at(string), updated_time(string)

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
