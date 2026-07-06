---
name: pm-paddle
description: Paddle connector knowledge and safe action guide.
---

# pm-paddle

## Purpose

Reads Paddle customers, subscriptions, transactions, and products through the Paddle REST API.

## Icon

- asset: icons/paddle.svg
- source: official
- review_status: official_verified
- review_url: https://developer.paddle.com/api-reference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- transactions:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), currency_code(), customer_id(), id(), status(), subscription_id()
- customers:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), id(), name()
- subscriptions:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), customer_id(), id(), status()
- products:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Paddle API read of customer, subscription, transaction, and product data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect paddle
```

### Inspect as structured JSON

```bash
pm connectors inspect paddle --json
```

## Agent Rules

- Run pm connectors inspect paddle before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
