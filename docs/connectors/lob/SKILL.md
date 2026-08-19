---
name: pm-lob
description: Lob connector knowledge and safe action guide.
---

# pm-lob

## Purpose

Reads Lob addresses, postcards, letters, checks, and bank accounts through the Lob print & mail REST API.

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

- base_url
- max_pages
- mode
- page_size
- api_key (secret) (required)

## ETL Streams

- addresses:
  - primary key: id
  - cursor: date_created
  - fields: address_city(string), address_country(string), address_line1(string), address_line2(string), address_state(string), address_zip(string), company(string), date_created(string), date_modified(string), deleted(boolean), description(string), email(string), id(string), name(string), object(string), phone(string)
- postcards:
  - primary key: id
  - cursor: date_created
  - fields: carrier(string), date_created(string), date_modified(string), deleted(boolean), description(string), expected_delivery_date(string), id(string), object(string), send_date(string), status(string), url(string)
- letters:
  - primary key: id
  - cursor: date_created
  - fields: carrier(string), date_created(string), date_modified(string), deleted(boolean), description(string), expected_delivery_date(string), id(string), object(string), send_date(string), status(string), url(string)
- checks:
  - primary key: id
  - cursor: date_created
  - fields: carrier(string), date_created(string), date_modified(string), deleted(boolean), description(string), expected_delivery_date(string), id(string), object(string), send_date(string), status(string), url(string)
- bank_accounts:
  - primary key: id
  - cursor: date_created
  - fields: account_number(string), account_type(string), bank_name(string), date_created(string), date_modified(string), deleted(boolean), description(string), id(string), object(string), routing_number(string), signatory(string), verified(boolean)

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
