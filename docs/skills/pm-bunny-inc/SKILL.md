---
name: pm-bunny-inc
description: Bunny, Inc. connector knowledge and safe action guide.
---

# pm-bunny-inc

## Purpose

Reads Bunny subscription-billing data through declared per-tenant GraphQL connection routes.

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

- subdomain (required)
- apikey (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - cursor: updatedAt
  - fields: id(string), updatedAt(string)
- contacts:
  - primary key: id
  - cursor: updatedAt
  - fields: id(string), updatedAt(string)
- invoices:
  - primary key: id
  - cursor: updatedAt
  - fields: id(string), updatedAt(string)
- payments:
  - primary key: id
  - cursor: updatedAt
  - fields: id(string), updatedAt(string)
- subscriptions:
  - primary key: id
  - cursor: updatedAt
  - fields: id(string), updatedAt(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded declared Bunny GraphQL reads use a source-validated tenant subdomain and bearer API key.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bunny-inc
```

### Inspect as structured JSON

```bash
pm connectors inspect bunny-inc --json
```

## Agent Rules

- Run pm connectors inspect bunny-inc before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
