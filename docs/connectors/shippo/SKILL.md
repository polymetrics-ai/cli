---
name: pm-shippo
description: Shippo connector knowledge and safe action guide.
---

# pm-shippo

## Purpose

Reads Shippo addresses, parcels, shipments, and transactions through the Shippo REST API.

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
- api_token (secret)

## ETL Streams

- addresses:
  - primary key: id
  - cursor: updated_at
  - fields: email(), id(), name(), updated_at()
- parcels:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), updated_at()
- shipments:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), updated_at()
- transactions:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Shippo API read of address, parcel, shipment, and transaction data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shippo
```

### Inspect as structured JSON

```bash
pm connectors inspect shippo --json
```

## Agent Rules

- Run pm connectors inspect shippo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
