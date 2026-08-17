---
name: pm-shipstation
description: ShipStation connector knowledge and safe action guide.
---

# pm-shipstation

## Purpose

Reads ShipStation orders, shipments, products, and customers through the ShipStation REST API.

## Icon

- asset: icons/shipstation.svg
- source: official
- review_status: official_verified
- review_url: https://www.shipstation.com/docs/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)
- api_secret (secret)

## ETL Streams

- orders:
  - primary key: id
  - cursor: modified_at
  - fields: id(), modified_at(), order_number(), status()
- shipments:
  - primary key: id
  - cursor: modified_at
  - fields: id(), modified_at(), order_number(), status()
- products:
  - primary key: id
  - cursor: modified_at
  - fields: id(), modified_at(), name(), sku()
- customers:
  - primary key: id
  - cursor: modified_at
  - fields: email(), id(), modified_at(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external ShipStation API read of order, shipment, product, and customer data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shipstation
```

### Inspect as structured JSON

```bash
pm connectors inspect shipstation --json
```

## Agent Rules

- Run pm connectors inspect shipstation before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
