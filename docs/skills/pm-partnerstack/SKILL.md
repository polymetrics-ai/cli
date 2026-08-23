---
name: pm-partnerstack
description: PartnerStack connector knowledge and safe action guide.
---

# pm-partnerstack

## Purpose

Reads PartnerStack partnerships, customers, transactions, and groups through the REST API.

## Icon

- id: partnerstack
- asset: icons/partnerstack.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.partnerstack.com/docs/api-overview

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- limit
- max_pages
- mode
- api_key (secret) (required)

## ETL Streams

- partnerships:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), email(string), id(string), status(string)
- customers:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), email(string), id(string), name(string)
- transactions:
  - primary key: id
  - cursor: created_at
  - fields: amount(number), created_at(string), currency(string), customer_id(string), id(string)
- groups:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external PartnerStack API read of partnership and referral-customer data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect partnerstack
```

### Inspect as structured JSON

```bash
pm connectors inspect partnerstack --json
```

## Agent Rules

- Run pm connectors inspect partnerstack before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
