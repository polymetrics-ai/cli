---
name: pm-revolut-merchant
description: Revolut Merchant connector knowledge and safe action guide.
---

# pm-revolut-merchant

## Purpose

Reads Revolut Merchant orders, customers, settlements, and payment links through the REST API.

## Icon

- asset: icons/revolut.svg
- source: official
- review_status: official_verified
- review_url: https://developer.revolut.com/docs/guides/merchant/reference/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- customer_id
- from_created_date
- state
- to_created_date
- api_key (secret)

## ETL Streams

- orders:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), id(), state(), stream()
- customers:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), full_name(), id(), stream()
- settlements:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), id(), stream()
- payment_links:
  - primary key: id
  - cursor: created_at
  - fields: amount(), created_at(), currency(), id(), state(), stream()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Revolut Merchant API read of order, customer, settlement, and payment-link data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect revolut-merchant
```

### Inspect as structured JSON

```bash
pm connectors inspect revolut-merchant --json
```

## Agent Rules

- Run pm connectors inspect revolut-merchant before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
