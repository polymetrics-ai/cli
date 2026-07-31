---
name: pm-quickbooks
description: QuickBooks connector knowledge and safe action guide.
---

# pm-quickbooks

## Purpose

Reads QuickBooks Online customers, invoices, payments, accounts, and vendors through the v3 Query API via the OAuth 2.0 refresh-token grant; tracks the complete r2 official qbo operation ledger with non-stream operations blocked by default.

## Icon

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
- realm_id
- sandbox
- start_date
- token_url
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- customers:
  - primary key: id
  - fields: active(), balance(), display_name(), id()
- invoices:
  - primary key: id
  - fields: balance(), customer_ref(), doc_number(), id(), total_amt()
- payments:
  - primary key: id
  - fields: customer_ref(), id(), total_amt(), txn_date()
- accounts:
  - primary key: id
  - fields: account_type(), classification(), id(), name()
- vendors:
  - primary key: id
  - fields: active(), balance(), display_name(), id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external QuickBooks Online v3 Query API read of customer/invoice/payment/account/vendor entities
- approval: read streams require no write approval; official write, binary, provider-query, and CDC ledger rows remain blocked until typed schemas, bounds, redaction, fixtures, and plan -> preview -> approval -> execute evidence exist
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect QuickBooks Online accounting streams and the blocked official operation ledger.
- Usage: pm quickbooks <group> <command> [flags]
- Source CLI: Intuit Developer QuickBooks Online Accounting API (EntityJsonObject_v1 qbo operation inventory)
- Fixture-backed ETL streams
- Blocked official operation lanes
- Other Commands
  - accounts list - Read QuickBooks chart-of-account records through the bounded Query API stream. [intent=etl availability=implemented stream=accounts]; approval: no write approval required; risk: read-only Query API access to account records
  - customers list - Read QuickBooks customer records through the bounded Query API stream. [intent=etl availability=implemented stream=customers]; approval: no write approval required; risk: read-only Query API access to customer records
  - invoices list - Read QuickBooks invoice records through the bounded Query API stream. [intent=etl availability=implemented stream=invoices]; approval: no write approval required; risk: read-only Query API access to invoice records
  - payments list - Read QuickBooks payment records through the bounded Query API stream. [intent=etl availability=implemented stream=payments]; approval: no write approval required; risk: read-only Query API access to payment records
  - vendors list - Read QuickBooks vendor records through the bounded Query API stream. [intent=etl availability=implemented stream=vendors]; approval: no write approval required; risk: read-only Query API access to vendor records
  - ledger status - Inspect the complete r2 official operation ledger and blocked/planned lane counts. [intent=docs_only availability=planned]; approval: no provider call; risk: local metadata inspection only; notes: Use pm connectors inspect quickbooks --json; docs-only dynamic connector execution is intentionally not exposed.
  - writes planned - QuickBooks create/update/delete/send/batch operations are tracked but blocked until typed write schemas and fixtures are authored. [intent=reverse_etl availability=planned]; approval: future reverse ETL must use plan -> preview -> approval -> execute with destructive confirmation where applicable; risk: high; mutations and destructive deletes are blocked by default; notes: No raw write passthrough exists.
  - query planned - Additional QuickBooks provider-query/search reads are tracked but blocked until fixed-target bounded direct-read commands exist. [intent=direct_read availability=planned]; approval: no write approval; future direct reads must remain fixed-target and redacted; risk: medium; bounded read only once implemented; notes: No arbitrary SQL/query escape hatch is exposed.
  - binary planned - QuickBooks PDF, attachment, and report/file payload surfaces are tracked but blocked until bounded binary/report transfer support exists. [intent=direct_read availability=planned]; approval: blocked by default; risk: high; file payloads require size limits, path-safe storage, and redaction before execution; notes: No generic file read/write surface is exposed.
  - cdc planned - QuickBooks change data capture is tracked but blocked until shared CDC truthfulness and state semantics are available. [intent=direct_read availability=planned]; approval: blocked by default; risk: medium; read-only changefeed once a typed cursor contract exists; notes: No CDC capability is advertised by metadata.
- Help topics:
  - quickbooks-auth - Use OAuth 2.0 refresh-token credentials from environment variables or stdin; never paste secrets into prompts.
  - quickbooks-ledger - The operation ledger is complete for the r2 qbo audit, but only five read streams execute locally.

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
