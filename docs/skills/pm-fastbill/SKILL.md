---
name: pm-fastbill
description: FastBill connector knowledge and safe action guide.
---

# pm-fastbill

## Purpose

Reads FastBill billing records through fixed JSON SERVICE envelopes.

## Icon

- id: fastbill
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

- username (required)
- api_key (secret) (required)

## ETL Streams

- customers:
  - primary key: CUSTOMER_ID
  - fields: CUSTOMER_ID(string)
- invoices:
  - primary key: INVOICE_ID
  - fields: INVOICE_ID(string)
- products:
  - primary key: ARTICLE_NUMBER
  - fields: ARTICLE_NUMBER(string)
- recurring_invoices:
  - primary key: INVOICE_ID
  - fields: INVOICE_ID(string)
- revenues:
  - primary key: REVENUE_ID
  - fields: REVENUE_ID(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: Bounded FastBill JSON API requests use fixed origin and declared Basic authentication.
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
