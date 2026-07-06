---
name: pm-stigg
description: Stigg connector knowledge and safe action guide.
---

# pm-stigg

## Purpose

Reads Stigg products, plans, customers, and subscriptions through the Stigg GraphQL-over-HTTP API. Read-only.

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
- api_key (secret)

## ETL Streams

- products:
  - primary key: id
  - fields: displayName(), id(), refId(), status()
- plans:
  - primary key: id
  - fields: displayName(), id(), refId(), status()
- customers:
  - primary key: id
  - fields: displayName(), id(), refId(), status()
- subscriptions:
  - primary key: id
  - fields: customerId(), id(), refId(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Stigg GraphQL API read of product/plan/customer/subscription entitlement metadata
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect stigg
```

### Inspect as structured JSON

```bash
pm connectors inspect stigg --json
```

## Agent Rules

- Run pm connectors inspect stigg before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
