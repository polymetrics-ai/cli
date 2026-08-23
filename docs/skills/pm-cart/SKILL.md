---
name: pm-cart
description: Cart.com connector knowledge and safe action guide.
---

# pm-cart

## Purpose

Reads Cart.com orders, customers, products, and inventory through a read-only REST API.

## Icon

- id: cart
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

- base_url (required)
- mode
- page_size
- access_token (secret) (required)

## ETL Streams

- orders:
  - primary key: id
  - fields: id(string), order_number(string), updated_at(string)
- customers:
  - primary key: id
  - fields: id(string), order_number(string), updated_at(string)
- products:
  - primary key: id
  - fields: id(string), order_number(string), updated_at(string)
- inventory:
  - primary key: id
  - fields: id(string), product_id(string), quantity(integer), sku(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
