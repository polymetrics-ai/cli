---
name: pm-rentcast
description: RentCast connector knowledge and safe action guide.
---

# pm-rentcast

## Purpose

Reads RentCast properties, sale listings, rental listings, market data, and value/rental estimates through the RentCast REST API. Read-only.

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

- address
- base_url
- city
- property_type
- state
- zip_code
- api_key (secret)

## ETL Streams

- properties:
  - primary key: id
  - cursor: last_seen_date
  - fields: address(), city(), id(), last_seen_date(), property_type(), state(), zip_code()
- sale_listings:
  - primary key: id
  - cursor: last_seen_date
  - fields: address(), id(), last_seen_date(), price(), property_type()
- rental_listings:
  - primary key: id
  - cursor: last_seen_date
  - fields: address(), id(), last_seen_date(), property_type(), rent()
- markets:
  - primary key: id
  - fields: city(), id(), state(), zip_code()
- value_estimates:
  - primary key: id
  - fields: address(), id(), price()
- rental_estimates:
  - primary key: id
  - fields: address(), id(), rent()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external RentCast API read of property, listing, market, and valuation data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect rentcast
```

### Inspect as structured JSON

```bash
pm connectors inspect rentcast --json
```

## Agent Rules

- Run pm connectors inspect rentcast before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
