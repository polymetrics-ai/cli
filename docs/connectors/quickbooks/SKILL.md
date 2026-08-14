---
name: pm-quickbooks
description: QuickBooks connector knowledge and safe action guide.
---

# pm-quickbooks

## Purpose

Reads QuickBooks Online customers, invoices, payments, accounts, and vendors through the v3 Query API via the OAuth 2.0 refresh-token grant. Read-only.

## Icon

- id: quickbooks
- asset: icons/quickbooks.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- page_size
- sandbox
- start_date
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)
- realm_id (secret) (required)
- refresh_token (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - fields: active(boolean), balance(number), display_name(string), id(string)
- invoices:
  - primary key: id
  - fields: balance(number), customer_ref(string), doc_number(string), id(string), total_amt(number)
- payments:
  - primary key: id
  - fields: customer_ref(string), id(string), total_amt(number), txn_date(string)
- accounts:
  - primary key: id
  - fields: account_type(string), classification(string), id(string), name(string)
- vendors:
  - primary key: id
  - fields: active(boolean), balance(number), display_name(string), id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external QuickBooks Online v3 Query API read of customer/invoice/payment/account/vendor entities
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect quickbooks
```

### Inspect as structured JSON

```bash
pm connectors inspect quickbooks --json
```

## Agent Rules

- Run pm connectors inspect quickbooks before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
