---
name: pm-freshbooks
description: FreshBooks connector knowledge and safe action guide.
---

# pm-freshbooks

## Purpose

Reads FreshBooks clients, invoices, expenses, payments, and items through the FreshBooks accounting REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url
- max_pages
- mode
- page_size
- oauth_access_token (secret) (required)

## ETL Streams

- clients:
  - primary key: id
  - cursor: updated
  - fields: currency_code(string), email(string), fname(string), id(integer), lname(string), organization(string), updated(string), userid(integer), vis_state(integer)
- invoices:
  - primary key: id
  - cursor: updated
  - fields: amount(object), create_date(string), currency_code(string), customerid(integer), id(integer), invoice_number(string), invoiceid(integer), outstanding(object), status(integer), updated(string)
- expenses:
  - primary key: id
  - cursor: updated
  - fields: amount(object), categoryid(integer), date(string), expenseid(integer), id(integer), notes(string), staffid(integer), updated(string), vendor(string)
- payments:
  - primary key: id
  - cursor: updated
  - fields: amount(object), date(string), id(integer), invoiceid(integer), note(string), type(string), updated(string)
- items:
  - primary key: id
  - cursor: updated
  - fields: description(string), id(integer), inventory(string), itemid(integer), name(string), qty(string), unit_cost(object), updated(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external FreshBooks API read of accounting data (clients, invoices, expenses, payments, items)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freshbooks
```

### Inspect as structured JSON

```bash
pm connectors inspect freshbooks --json
```

## Agent Rules

- Run pm connectors inspect freshbooks before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
