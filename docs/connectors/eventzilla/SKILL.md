---
name: pm-eventzilla
description: Eventzilla connector knowledge and safe action guide.
---

# pm-eventzilla

## Purpose

Reads Eventzilla events, categories, users, attendees, ticket types, and transactions, and writes attendee check-in and event sales-page toggle mutations, through the Eventzilla v2 REST API.

## Icon

- asset: icons/eventzilla.svg
- source: official
- review_status: official_verified
- review_url: https://www.eventzilla.net/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- events:
  - primary key: id
  - fields: categories(), currency(), end_date(), end_time(), id(), start_date(), start_time(), status(), tickets_sold(), tickets_total(), time_zone(), title(), url(), venue()
- categories:
  - primary key: category
  - fields: category()
- users:
  - primary key: id
  - fields: company(), email(), first_name(), id(), last_name(), last_seen(), phone_primary(), timezone(), user_type(), username()
- attendees:
  - primary key: id
  - fields: email(), event_id(), first_name(), id(), is_attended(), last_name(), refno(), ticket_type(), transaction_amount(), transaction_date(), transaction_status()
- tickets:
  - primary key: id
  - fields: event_id(), id(), is_visible(), price(), quantity_total(), sales_end_date(), sales_start_date(), ticket_type(), title()
- transactions:
  - primary key: checkout_id
  - fields: buyer_first_name(), buyer_last_name(), checkout_id(), comments(), email(), event_date(), event_id(), payment_type(), promo_code(), tickets_in_transaction(), title(), transaction_amount(), transaction_date(), transaction_ref(), transaction_status(), user_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- checkin_attendee:
  - endpoint: POST /attendees/checkin
  - risk: marks an attendee checked in or reverts check-in at the door; low-risk operational mutation, no approval required
- toggle_event_sales:
  - endpoint: POST /events/togglesales
  - risk: publishes or unpublishes an event's public sales page; setting status false immediately stops new ticket sales for that event, approval required

## Security

- read risk: external Eventzilla API read of event, category, user, attendee, ticket, and transaction data
- write risk: external mutation of attendee check-in state and event sales-page publish status; every write ships with an explicit per-action risk string
- approval: required for toggle_event_sales (unpublishing stops new ticket sales immediately); checkin_attendee is a low-risk operational door-scan mutation, no approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect eventzilla
```

### Inspect as structured JSON

```bash
pm connectors inspect eventzilla --json
```

## Agent Rules

- Run pm connectors inspect eventzilla before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
