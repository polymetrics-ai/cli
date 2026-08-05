---
name: pm-zoho-inventory
description: Zoho Inventory connector knowledge and safe action guide.
---

# pm-zoho-inventory

## Purpose

Reads Zoho Inventory contacts, items, and sales orders through the Zoho Inventory REST API.

## Icon

- id: simple-icons-zoho-inventory
- asset: icons/simple-icons/zoho-inventory.svg
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
- max_pages
- mode
- organization_id
- page_size
- access_token (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: company_name(), contact_id(), contact_name(), contact_type(), created_time(), currency_code(), email(), id(), last_modified_time(), outstanding_receivable_amount(), phone(), status(), updated_at()
- items:
  - primary key: id
  - cursor: updated_at
  - fields: created_time(), description(), id(), item_id(), item_name(), last_modified_time(), name(), rate(), sku(), status(), unit(), updated_at()
- salesorders:
  - primary key: id
  - cursor: updated_at
  - fields: balance(), created_time(), currency_code(), customer_id(), customer_name(), date(), id(), last_modified_time(), salesorder_id(), salesorder_number(), shipment_date(), status(), total(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zoho Inventory API read of contact/item/sales-order data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Mechanism

- kind: official_api
- sanctioned_by_provider: true (official)

## Commands

### Inspect as a manual

```bash
pm connectors inspect zoho-inventory
```

### Inspect as structured JSON

```bash
pm connectors inspect zoho-inventory --json
```

## Agent Rules

- Run pm connectors inspect zoho-inventory before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
