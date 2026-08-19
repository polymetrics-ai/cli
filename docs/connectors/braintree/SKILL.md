---
name: pm-braintree
description: Braintree connector knowledge and safe action guide.
---

# pm-braintree

## Purpose

Reads Braintree transactions, customers, subscriptions, reference data, payment methods, disputes, merchant accounts, and Apple Pay domains through the gateway HTTP API.

## Icon

- id: braintree
- asset: icons/braintree.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- merchant_id (required)
- mode
- page_size
- private_key (secret) (required)
- public_key (secret) (required)

## ETL Streams

- transactions:
  - primary key: id
  - fields: amount(string), id(string), status(string)
- customers:
  - primary key: id
  - fields: amount(string), id(string), status(string)
- subscriptions:
  - primary key: id
  - fields: amount(string), id(string), status(string)
- add_ons:
  - primary key: id
  - fields: amount(string), id(string), kind(string), name(string)
- discounts:
  - primary key: id
  - fields: amount(string), id(string), kind(string), name(string)
- plans:
  - primary key: id
  - fields: billing_frequency(integer), currency_iso_code(string), id(string), name(string), price(string)
- merchant_accounts:
  - primary key: id
  - fields: currency_iso_code(string), default(boolean), id(string), status(string)
- payment_methods:
  - primary key: token
  - fields: customer_id(string), default(boolean), payment_instrument_type(string), token(string)
- disputes:
  - primary key: id
  - fields: amount(string), id(string), reason(string), status(string)
- apple_pay_domains:
  - primary key: domain
  - fields: domain(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Braintree API read of transaction, customer, subscription, reference, dispute, payment method, and merchant account data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect braintree
```

### Inspect as structured JSON

```bash
pm connectors inspect braintree --json
```

## Agent Rules

- Run pm connectors inspect braintree before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
