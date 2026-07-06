---
name: pm-recharge
description: Recharge connector knowledge and safe action guide.
---

# pm-recharge

## Purpose

Reads Recharge customers, subscriptions, and orders through the Recharge REST API.

## Icon

- asset: icons/recharge.svg
- source: official
- review_status: official_verified
- review_url: https://docs.getrecharge.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_version
- base_url
- mode
- access_token (secret)

## ETL Streams

- customers:
  - primary key: id
  - fields: created_at(), email(), id(), updated_at()
- subscriptions:
  - primary key: id
  - fields: created_at(), customer_id(), id(), status(), updated_at()
- orders:
  - primary key: id
  - fields: created_at(), customer_id(), id(), status(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Recharge API read of customer, subscription, and order data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect recharge
```

### Inspect as structured JSON

```bash
pm connectors inspect recharge --json
```

## Agent Rules

- Run pm connectors inspect recharge before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
