---
name: pm-stripe
description: Stripe connector knowledge and safe action guide.
---

# pm-stripe

## Purpose

Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse ETL customer actions through the Stripe REST API.

## Icon

- asset: icons/stripe.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://stripe.com/docs/api

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- max_pages
- mode
- page_size
- start_date
- client_secret (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: created
  - fields: balance(), created(), currency(), delinquent(), description(), email(), id(), livemode(), name(), object(), phone()
- charges:
  - primary key: id
  - cursor: created
  - fields: amount(), amount_captured(), amount_refunded(), created(), currency(), customer(), id(), livemode(), object(), paid(), refunded(), status()
- invoices:
  - primary key: id
  - cursor: created
  - fields: amount_due(), amount_paid(), amount_remaining(), created(), currency(), customer(), id(), livemode(), object(), paid(), status(), subscription(), total()
- subscriptions:
  - primary key: id
  - cursor: created
  - fields: cancel_at_period_end(), canceled_at(), created(), currency(), current_period_end(), current_period_start(), customer(), id(), livemode(), object(), status()
- products:
  - primary key: id
  - cursor: created
  - fields: active(), created(), description(), id(), livemode(), name(), object(), type(), updated()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation; approval required
- update_customer:
  - endpoint: POST /customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required

## Security

- read risk: external Stripe API read of customer and billing data
- write risk: external Stripe API mutation
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect stripe
```

### Inspect as structured JSON

```bash
pm connectors inspect stripe --json
```

## Agent Rules

- Run pm connectors inspect stripe before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
