---
name: pm-cart
description: Cart.com connector knowledge and safe action guide.
---

# pm-cart

## Purpose

Reads Cart.com orders, customers, products, and inventory through a read-only REST API.

## Icon

- asset: icons/cart.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.cart.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- access_token (secret)

## ETL Streams

- orders:
  - primary key: id
  - fields: id(), order_number(), updated_at()
- customers:
  - primary key: id
  - fields: id(), order_number(), updated_at()
- products:
  - primary key: id
  - fields: id(), order_number(), updated_at()
- inventory:
  - primary key: id
  - fields: id(), product_id(), quantity(), sku(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Cart.com API read of order, customer, product, and inventory data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect cart
```

### Inspect as structured JSON

```bash
pm connectors inspect cart --json
```

## Agent Rules

- Run pm connectors inspect cart before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
