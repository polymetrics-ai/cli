---
name: pm-netsuite
description: NetSuite connector knowledge and safe action guide.
---

# pm-netsuite

## Purpose

Reads selected NetSuite REST Record API resources (customers, vendors, items, sales orders), authenticating with OAuth 1.0a Token-Based Authentication (HMAC-SHA256 request signing). Read-only.

## Icon

- id: netsuite
- asset: icons/netsuite.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- max_pages
- mode
- page_size
- realm (required)
- consumer_key (secret) (required)
- consumer_secret (secret) (required)
- token_key (secret) (required)
- token_secret (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: last_modified_date
  - fields: email(string), entity_id(string), id(string), last_modified_date(string), name(string), status(string)
- vendors:
  - primary key: id
  - cursor: last_modified_date
  - fields: email(string), entity_id(string), id(string), last_modified_date(string), name(string), status(string)
- items:
  - primary key: id
  - cursor: last_modified_date
  - fields: email(string), entity_id(string), id(string), last_modified_date(string), name(string), status(string)
- sales_orders:
  - primary key: id
  - cursor: last_modified_date
  - fields: email(string), entity_id(string), id(string), last_modified_date(string), name(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NetSuite REST Record API read of customer, vendor, item, and sales order data
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect netsuite
```

### Inspect as structured JSON

```bash
pm connectors inspect netsuite --json
```

## Agent Rules

- Run pm connectors inspect netsuite before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
