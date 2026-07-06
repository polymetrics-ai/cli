---
name: pm-lob
description: Lob connector knowledge and safe action guide.
---

# pm-lob

## Purpose

Reads Lob addresses, postcards, letters, checks, and bank accounts through the Lob print & mail REST API.

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

- base_url
- max_pages
- mode
- page_size
- api_key (secret)

## ETL Streams

- addresses:
  - primary key: id
  - cursor: date_created
  - fields: address_city(), address_country(), address_line1(), address_line2(), address_state(), address_zip(), company(), date_created(), date_modified(), deleted(), description(), email(), id(), name(), object(), phone()
- postcards:
  - primary key: id
  - cursor: date_created
  - fields: carrier(), date_created(), date_modified(), deleted(), description(), expected_delivery_date(), id(), object(), send_date(), status(), url()
- letters:
  - primary key: id
  - cursor: date_created
  - fields: carrier(), date_created(), date_modified(), deleted(), description(), expected_delivery_date(), id(), object(), send_date(), status(), url()
- checks:
  - primary key: id
  - cursor: date_created
  - fields: carrier(), date_created(), date_modified(), deleted(), description(), expected_delivery_date(), id(), object(), send_date(), status(), url()
- bank_accounts:
  - primary key: id
  - cursor: date_created
  - fields: account_number(), account_type(), bank_name(), date_created(), date_modified(), deleted(), description(), id(), object(), routing_number(), signatory(), verified()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Lob API read of address book, mailpiece, and bank account data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect lob
```

### Inspect as structured JSON

```bash
pm connectors inspect lob --json
```

## Agent Rules

- Run pm connectors inspect lob before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
