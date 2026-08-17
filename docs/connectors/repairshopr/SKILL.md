---
name: pm-repairshopr
description: RepairShopr connector knowledge and safe action guide.
---

# pm-repairshopr

## Purpose

Reads RepairShopr customers, tickets, invoices, estimates, and assets through the REST API.

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
- created_after
- query
- updated_after
- api_token (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), email(), id(), name(), phone(), stream(), updated_at()
- tickets:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(), id(), number(), status(), stream(), subject(), updated_at()
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(), id(), number(), status(), stream(), total(), updated_at()
- estimates:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(), id(), number(), status(), stream(), total(), updated_at()
- assets:
  - primary key: id
  - cursor: updated_at
  - fields: customer_id(), id(), name(), serial_number(), stream(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external RepairShopr API read of customer and shop-management data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect repairshopr
```

### Inspect as structured JSON

```bash
pm connectors inspect repairshopr --json
```

## Agent Rules

- Run pm connectors inspect repairshopr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
