---
name: pm-retailexpress-by-maropost
description: Retail Express by Maropost connector knowledge and safe action guide.
---

# pm-retailexpress-by-maropost

## Purpose

Reads Retail Express products, customers, orders, stock levels, and stores through the Maropost API. Read-only.

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

- base_url (required)
- created_after
- status
- store_id
- updated_after
- access_token (secret)
- api_key (secret)

## ETL Streams

- products:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), name(string), sku(string), status(string), stream(string), updated_at(string)
- customers:
  - primary key: id
  - cursor: updated_at
  - fields: email(string), first_name(string), id(string), last_name(string), stream(string), updated_at(string)
- orders:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(string), id(string), order_number(string), status(string), stream(string), total(number), updated_at(string)
- stock_levels:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), product_id(string), quantity(number), store_id(string), stream(string), updated_at(string)
- stores:
  - primary key: id
  - fields: code(string), id(string), name(string), status(string), stream(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Retail Express by Maropost API read of product, customer, order, and stock data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect retailexpress-by-maropost
```

### Inspect as structured JSON

```bash
pm connectors inspect retailexpress-by-maropost --json
```

## Agent Rules

- Run pm connectors inspect retailexpress-by-maropost before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
