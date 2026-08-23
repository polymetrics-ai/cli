---
name: pm-oncehub
description: OnceHub connector knowledge and safe action guide.
---

# pm-oncehub

## Purpose

Reads OnceHub bookings, contacts, booking pages, users, and event types through the OnceHub REST API.

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

- base_url
- max_pages
- mode
- page_size
- start_date
- api_key (secret) (required)

## ETL Streams

- bookings:
  - primary key: id
  - cursor: last_updated_time
  - fields: booking_page(string), contact(string), creation_time(string), customer_timezone(string), duration_minutes(number), event_type(string), id(string), in_trash(boolean), last_updated_time(string), location_description(string), object(string), owner(string), starting_time(string), status(string), subject(string), tracking_id(string)
- contacts:
  - primary key: id
  - cursor: last_updated_time
  - fields: creation_time(string), email(string), first_name(string), id(string), last_updated_time(string), mobile_phone(string), object(string), owner(string), timezone(string)
- booking_pages:
  - primary key: id
  - fields: active(boolean), id(string), label(string), name(string), object(string), timezone(string), url(string)
- users:
  - primary key: id
  - fields: email(string), first_name(string), id(string), last_name(string), object(string), role_name(string), status(string)
- event_types:
  - primary key: id
  - fields: id(string), label(string), name(string), object(string)

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
