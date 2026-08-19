---
name: pm-katana
description: Katana connector knowledge and safe action guide.
---

# pm-katana

## Purpose

Reads Katana MRP (Cloud Inventory) products, materials, variants, sales orders, and customers through the Katana REST API.

## Icon

- id: simple-icons-katana
- asset: icons/simple-icons/katana.svg
- title: Katana
- simple_icon_slug: katana
- simple_icon_hex: 000000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Katana
- match: exact-name-or-slug
- matched_by: katana

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- products:
  - primary key: id
  - cursor: updated_at
  - fields: additional_info(string), category_name(string), created_at(string), default_supplier_id(integer), id(integer), is_producible(boolean), is_purchasable(boolean), is_sellable(boolean), name(string), uom(string), updated_at(string)
- materials:
  - primary key: id
  - cursor: updated_at
  - fields: additional_info(string), category_name(string), created_at(string), default_supplier_id(integer), id(integer), is_sellable(boolean), name(string), uom(string), updated_at(string)
- variants:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), material_id(integer), product_id(integer), purchase_price(number), sales_price(number), sku(string), type(string), updated_at(string)
- sales_orders:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), currency(string), customer_id(integer), delivery_date(string), id(integer), order_created_date(string), order_no(string), status(string), total(number), total_in_base_currency(number), updated_at(string)
- customers:
  - primary key: id
  - cursor: updated_at
  - fields: category(string), created_at(string), currency(string), email(string), id(integer), name(string), phone(string), reference_id(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Katana MRP API read of inventory, sales, and customer data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect katana
```

### Inspect as structured JSON

```bash
pm connectors inspect katana --json
```

## Agent Rules

- Run pm connectors inspect katana before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
