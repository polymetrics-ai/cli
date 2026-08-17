---
name: pm-recurly
description: Recurly connector knowledge and safe action guide.
---

# pm-recurly

## Purpose

Reads Recurly accounts, subscriptions, invoices, transactions, and plans through the Recurly v3 REST API.

## Icon

- asset: icons/recurly.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.recurly.com/api/v2021-02-25/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- accounts:
  - primary key: id
  - cursor: updated_at
  - fields: code(), created_at(), email(), id(), state(), updated_at()
- subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(), created_at(), id(), plan_id(), state(), updated_at()
- invoices:
  - primary key: id
  - cursor: created_at
  - fields: account_id(), created_at(), id(), state(), total()
- transactions:
  - primary key: id
  - cursor: created_at
  - fields: account_id(), amount(), created_at(), id(), status()
- plans:
  - primary key: id
  - cursor: updated_at
  - fields: code(), id(), name(), state(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Recurly API read of subscription billing data
- approval: none; read-only billing API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect recurly
```

### Inspect as structured JSON

```bash
pm connectors inspect recurly --json
```

## Agent Rules

- Run pm connectors inspect recurly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
