---
name: pm-navan
description: Navan connector knowledge and safe action guide.
---

# pm-navan

## Purpose

Reads Navan flight, hotel, car, and rail travel bookings through the Navan REST API using OAuth2 client-credentials authentication.

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
- mode
- start_date
- client_id (secret)
- client_secret (secret)

## ETL Streams

- bookings:
  - primary key: uuid
  - cursor: last_modified
  - fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()
- hotel_bookings:
  - primary key: uuid
  - cursor: last_modified
  - fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()
- car_bookings:
  - primary key: uuid
  - cursor: last_modified
  - fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()
- rail_bookings:
  - primary key: uuid
  - cursor: last_modified
  - fields: approval_status(), base_price(), booking_fee(), booking_id(), booking_method(), booking_status(), booking_type(), cancelled_at(), confirmation_number(), created(), currency(), destination(), domestic(), end_date(), expensed(), grand_total(), last_modified(), start_date(), uuid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Navan API read of travel booking data (flight, hotel, car, rail)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect navan
```

### Inspect as structured JSON

```bash
pm connectors inspect navan --json
```

## Agent Rules

- Run pm connectors inspect navan before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
