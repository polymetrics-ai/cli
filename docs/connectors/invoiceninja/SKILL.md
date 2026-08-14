---
name: pm-invoiceninja
description: Invoice Ninja connector knowledge and safe action guide.
---

# pm-invoiceninja

## Purpose

Reads Invoice Ninja clients, invoices, products, payments, and quotes through the Invoice Ninja v5 REST API.

## Icon

- id: simple-icons-invoiceninja
- asset: icons/simple-icons/invoiceninja.svg
- title: Invoice Ninja
- simple_icon_slug: invoiceninja
- simple_icon_hex: 000000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Invoice%20Ninja
- match: exact-name-or-slug
- matched_by: invoiceninja

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
- api_key (secret) (required)

## ETL Streams

- clients:
  - primary key: id
  - fields: archived_at(integer), balance(number), created_at(integer), currency_id(string), display_name(string), id(string), is_deleted(boolean), name(string), number(string), paid_to_date(number), phone(string), updated_at(integer), vat_number(string), website(string)
- invoices:
  - primary key: id
  - fields: amount(number), balance(number), client_id(string), created_at(integer), currency_id(string), date(string), due_date(string), id(string), is_deleted(boolean), number(string), paid_to_date(number), status_id(string), updated_at(integer)
- products:
  - primary key: id
  - fields: cost(number), created_at(integer), id(string), is_deleted(boolean), notes(string), price(number), product_key(string), quantity(number), tax_name1(string), tax_rate1(number), updated_at(integer)
- payments:
  - primary key: id
  - fields: amount(number), applied(number), client_id(string), created_at(integer), currency_id(string), date(string), id(string), is_deleted(boolean), number(string), refunded(number), status_id(string), transaction_reference(string), updated_at(integer)
- quotes:
  - primary key: id
  - fields: amount(number), balance(number), client_id(string), created_at(integer), currency_id(string), date(string), due_date(string), id(string), is_deleted(boolean), number(string), status_id(string), updated_at(integer), valid_until(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Invoice Ninja API read of client and billing data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect invoiceninja
```

### Inspect as structured JSON

```bash
pm connectors inspect invoiceninja --json
```

## Agent Rules

- Run pm connectors inspect invoiceninja before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
