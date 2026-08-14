---
name: pm-kyriba
description: Kyriba connector knowledge and safe action guide.
---

# pm-kyriba

## Purpose

Reads Kyriba bank accounts, transactions, statements, and payments through tenant REST API collection endpoints. Read-only.

## Icon

- id: kyriba
- asset: icons/kyriba.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.kyriba.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- scope
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- bank_accounts:
  - primary key: id
  - fields: account_number(string), currency(string), id(string), status(string)
- transactions:
  - primary key: id
  - fields: account_number(string), amount(number), currency(string), id(string), status(string)
- statements:
  - primary key: id
  - fields: account_number(string), currency(string), id(string), status(string)
- payments:
  - primary key: id
  - fields: amount(number), currency(string), id(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Kyriba tenant REST API read of bank accounts/transactions/statements/payments
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect kyriba
```

### Inspect as structured JSON

```bash
pm connectors inspect kyriba --json
```

## Agent Rules

- Run pm connectors inspect kyriba before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
