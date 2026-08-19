---
name: pm-lightspeed-retail
description: Lightspeed Retail connector knowledge and safe action guide.
---

# pm-lightspeed-retail

## Purpose

Reads Lightspeed Retail (X-Series) products, customers, sales, outlets, and registers through the Lightspeed REST API. Read-only.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- mode
- subdomain (required)
- api_key (secret) (required)

## ETL Streams

- products:
  - primary key: id
  - cursor: version
  - fields: brand_id(string), created_at(string), description(string), handle(string), has_variants(boolean), id(string), is_active(boolean), is_composite(boolean), name(string), price_excluding_tax(number), price_including_tax(number), product_category(string), sku(string), supplier_id(string), supply_price(number), updated_at(string), version(integer)
- customers:
  - primary key: id
  - cursor: version
  - fields: balance(number), created_at(string), customer_code(string), customer_group_id(string), do_not_email(boolean), enable_loyalty(boolean), id(string), loyalty_balance(number), updated_at(string), version(integer), year_to_date(number)
- sales:
  - primary key: id
  - cursor: version
  - fields: created_at(string), customer_id(string), id(string), invoice_number(string), register_id(string), sale_date(string), status(string), total_price(number), total_tax(number), updated_at(string), user_id(string), version(integer)
- outlets:
  - primary key: id
  - cursor: version
  - fields: currency(string), currency_symbol(string), default_tax_id(string), display_prices(string), id(string), name(string), time_zone(string), version(integer)
- registers:
  - primary key: id
  - cursor: version
  - fields: email_receipt(boolean), id(string), invoice_prefix(string), invoice_sequence(integer), is_open(boolean), name(string), outlet_id(string), print_receipt(boolean), version(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Lightspeed Retail API read of product, customer, and sales data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect lightspeed-retail
```

### Inspect as structured JSON

```bash
pm connectors inspect lightspeed-retail --json
```

## Agent Rules

- Run pm connectors inspect lightspeed-retail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
