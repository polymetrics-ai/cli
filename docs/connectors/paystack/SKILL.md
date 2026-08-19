---
name: pm-paystack
description: Paystack connector knowledge and safe action guide.
---

# pm-paystack

## Purpose

Reads Paystack customers, transactions, subscriptions, invoices, and disputes through the Paystack REST API.

## Icon

- id: paystack
- asset: icons/paystack.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://paystack.com/docs/api/

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
- start_date
- secret_key (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(string), customer_code(string), domain(string), email(string), first_name(string), id(integer), last_name(string), phone(string), risk_action(string), updatedAt(string)
- transactions:
  - primary key: id
  - cursor: createdAt
  - fields: amount(integer), channel(string), createdAt(string), currency(string), domain(string), gateway_response(string), id(integer), paid_at(string), reference(string), status(string)
- subscriptions:
  - primary key: id
  - cursor: createdAt
  - fields: amount(integer), createdAt(string), domain(string), email_token(string), id(integer), next_payment_date(string), status(string), subscription_code(string), updatedAt(string)
- invoices:
  - primary key: id
  - cursor: createdAt
  - fields: amount(integer), createdAt(string), currency(string), domain(string), due_date(string), id(integer), paid(boolean), request_code(string), status(string), updatedAt(string)
- disputes:
  - primary key: id
  - cursor: createdAt
  - fields: category(string), createdAt(string), currency(string), domain(string), due_at(string), id(integer), refund_amount(integer), resolution(string), status(string), updatedAt(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Paystack API read of customer and payment data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect paystack
```

### Inspect as structured JSON

```bash
pm connectors inspect paystack --json
```

## Agent Rules

- Run pm connectors inspect paystack before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
