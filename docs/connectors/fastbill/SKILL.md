---
name: pm-fastbill
description: FastBill connector knowledge and safe action guide.
---

# pm-fastbill

## Purpose

Reads FastBill customers, invoices, products, recurring invoices, and revenues through the FastBill JSON API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/fastbill.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://apidocs.fastbill.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- username
- api_key (secret)

## ETL Streams

- customers:
  - primary key: customer_id
  - fields: country_code(), created(), currency_code(), customer_id(), customer_number(), customer_type(), email(), first_name(), last_name(), organization(), phone()
- invoices:
  - primary key: invoice_id
  - fields: currency_code(), customer_id(), due_date(), invoice_date(), invoice_id(), invoice_number(), is_canceled(), sub_total(), total(), type(), vat_total()
- products:
  - primary key: article_number
  - fields: article_number(), currency_code(), description(), is_greedy(), title(), unit_price(), vat_percent()
- recurring_invoices:
  - primary key: invoice_id
  - fields: currency_code(), customer_id(), due_date(), invoice_date(), invoice_id(), invoice_number(), is_canceled(), sub_total(), total(), type(), vat_total()
- revenues:
  - primary key: invoice_id
  - fields: currency_code(), customer_id(), invoice_date(), invoice_id(), invoice_number(), total(), vat_total()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external FastBill API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect fastbill
```

### Inspect as structured JSON

```bash
pm connectors inspect fastbill --json
```

## Agent Rules

- Run pm connectors inspect fastbill before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
