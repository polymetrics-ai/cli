---
name: pm-oncehub
description: OnceHub connector knowledge and safe action guide.
---

# pm-oncehub

## Purpose

Reads OnceHub bookings, contacts, booking pages, users, and event types through the OnceHub REST API.

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
- max_pages
- mode
- page_size
- start_date
- api_key (secret)

## ETL Streams

- bookings:
  - primary key: id
  - cursor: last_updated_time
  - fields: booking_page(), contact(), creation_time(), customer_timezone(), duration_minutes(), event_type(), id(), in_trash(), last_updated_time(), location_description(), object(), owner(), starting_time(), status(), subject(), tracking_id()
- contacts:
  - primary key: id
  - cursor: last_updated_time
  - fields: creation_time(), email(), first_name(), id(), last_updated_time(), mobile_phone(), object(), owner(), timezone()
- booking_pages:
  - primary key: id
  - fields: active(), id(), label(), name(), object(), timezone(), url()
- users:
  - primary key: id
  - fields: email(), first_name(), id(), last_name(), object(), role_name(), status()
- event_types:
  - primary key: id
  - fields: id(), label(), name(), object()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external OnceHub API read of scheduling, contact, and user data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect oncehub
```

### Inspect as structured JSON

```bash
pm connectors inspect oncehub --json
```

## Agent Rules

- Run pm connectors inspect oncehub before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
