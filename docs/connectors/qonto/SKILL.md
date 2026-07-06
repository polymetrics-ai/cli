---
name: pm-qonto
description: Qonto connector knowledge and safe action guide.
---

# pm-qonto

## Purpose

Reads Qonto bank transactions, memberships, and accounts through the Qonto REST API (read-only).

## Icon

- asset: icons/qonto.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api-doc.qonto.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- iban
- start_date
- api_key (secret)

## ETL Streams

- transactions:
  - primary key: id
  - cursor: settled_at
  - fields: amount(), id(), settled_at(), side(), updated_at()
- memberships:
  - primary key: id
  - fields: amount(), id(), settled_at(), side(), updated_at()
- accounts:
  - primary key: id
  - fields: amount(), id(), settled_at(), side(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Qonto API read of bank transaction and account data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect qonto
```

### Inspect as structured JSON

```bash
pm connectors inspect qonto --json
```

## Agent Rules

- Run pm connectors inspect qonto before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
