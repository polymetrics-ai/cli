---
name: pm-freshbooks
description: FreshBooks connector knowledge and safe action guide.
---

# pm-freshbooks

## Purpose

Reads FreshBooks clients, invoices, expenses, payments, and items through the FreshBooks accounting REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- max_pages
- mode
- page_size
- oauth_access_token (secret)

## ETL Streams

- clients:
  - primary key: id
  - cursor: updated
  - fields: currency_code(), email(), fname(), id(), lname(), organization(), updated(), userid(), vis_state()
- invoices:
  - primary key: id
  - cursor: updated
  - fields: amount(), create_date(), currency_code(), customerid(), id(), invoice_number(), invoiceid(), outstanding(), status(), updated()
- expenses:
  - primary key: id
  - cursor: updated
  - fields: amount(), categoryid(), date(), expenseid(), id(), notes(), staffid(), updated(), vendor()
- payments:
  - primary key: id
  - cursor: updated
  - fields: amount(), date(), id(), invoiceid(), note(), type(), updated()
- items:
  - primary key: id
  - cursor: updated
  - fields: description(), id(), inventory(), itemid(), name(), qty(), unit_cost(), updated()

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
