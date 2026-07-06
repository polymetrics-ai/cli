---
name: pm-woocommerce
description: WooCommerce connector knowledge and safe action guide.
---

# pm-woocommerce

## Purpose

Reads WooCommerce orders, products, customers, and coupons through the WooCommerce REST API (wc/v3).

## Icon

- asset: icons/woocommerce.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://woocommerce.github.io/woocommerce-rest-api-docs/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- page_size
- start_date
- api_key (secret)
- api_secret (secret)

## ETL Streams

- orders:
  - primary key: id
  - cursor: date_modified_gmt
  - fields: currency(), customer_id(), date_created(), date_created_gmt(), date_modified(), date_modified_gmt(), date_paid(), id(), number(), payment_method(), status(), total(), total_tax()
- products:
  - primary key: id
  - cursor: date_modified_gmt
  - fields: date_created_gmt(), date_modified_gmt(), id(), name(), price(), regular_price(), sale_price(), sku(), slug(), status(), stock_quantity(), stock_status(), total_sales(), type()
- customers:
  - primary key: id
  - cursor: date_modified_gmt
  - fields: date_created(), date_created_gmt(), date_modified(), date_modified_gmt(), email(), first_name(), id(), is_paying_customer(), last_name(), role(), username()
- coupons:
  - primary key: id
  - cursor: date_modified_gmt
  - fields: amount(), code(), date_created(), date_created_gmt(), date_expires(), date_modified(), date_modified_gmt(), discount_type(), id(), usage_count(), usage_limit()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external WooCommerce store read of orders, products, customers, and coupons
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect woocommerce
```

### Inspect as structured JSON

```bash
pm connectors inspect woocommerce --json
```

## Agent Rules

- Run pm connectors inspect woocommerce before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
