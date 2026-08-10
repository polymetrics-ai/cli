---
name: pm-eventbrite
description: Eventbrite connector knowledge and safe action guide.
---

# pm-eventbrite

## Purpose

Reads Eventbrite organizations, events, attendees, orders, and ticket classes through the Eventbrite v3 REST API. Read-only source.

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
- event_id
- organization_id
- start_date
- private_token (secret) (required)

## ETL Streams

- organizations:
  - primary key: id
  - fields: id(string), image_id(string), locale(string), name(string), vertical(string)
- events:
  - primary key: id
  - cursor: changed
  - fields: capacity(integer), changed(string), created(string), currency(string), description(string), end(string), id(string), listed(boolean), name(string), online_event(boolean), organization_id(string), published(string), start(string), status(string), url(string), venue_id(string)
- attendees:
  - primary key: id
  - cursor: changed
  - fields: cancelled(boolean), changed(string), checked_in(boolean), created(string), email(string), event_id(string), id(string), name(string), order_id(string), quantity(integer), refunded(boolean), status(string), ticket_class_id(string), ticket_class_name(string)
- orders:
  - primary key: id
  - cursor: changed
  - fields: changed(string), created(string), email(string), event_id(string), id(string), name(string), status(string), time_remaining(integer)
- ticket_classes:
  - primary key: id
  - fields: cost(string), description(string), event_id(string), fee(string), free(boolean), hidden(boolean), id(string), name(string), on_sale_status(string), quantity_sold(integer), quantity_total(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Eventbrite API read of organization, event, attendee, and order data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect eventbrite
```

### Inspect as structured JSON

```bash
pm connectors inspect eventbrite --json
```

## Agent Rules

- Run pm connectors inspect eventbrite before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
