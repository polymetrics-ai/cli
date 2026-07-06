---
name: pm-plaid
description: Plaid connector knowledge and safe action guide.
---

# pm-plaid

## Purpose

Reads Plaid institutions and category metadata through read-only POST endpoints. All credentials and pagination/filter state travel in the JSON request body (Plaid's own convention), driven by a StreamHook.

## Icon

- asset: icons/plaid.svg
- source: official
- review_status: official_verified
- review_url: https://plaid.com/docs/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- country_codes
- max_pages
- mode
- page_size
- client_id (secret)
- secret (secret)

## ETL Streams

- institutions:
  - primary key: institution_id
  - fields: country_codes(), institution_id(), name()
- categories:
  - primary key: category_id
  - fields: category_id(), group(), hierarchy()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Plaid API read of institution/category metadata
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect plaid
```

### Inspect as structured JSON

```bash
pm connectors inspect plaid --json
```

## Agent Rules

- Run pm connectors inspect plaid before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
