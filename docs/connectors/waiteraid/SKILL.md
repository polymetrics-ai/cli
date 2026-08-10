---
name: pm-waiteraid
description: WaiterAid connector knowledge and safe action guide.
---

# pm-waiteraid

## Purpose

Reads and writes WaiterAid restaurant reservations, meals, guests, and queue entries.

## Icon

- id: waiteraid
- asset: icons/waiteraid.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://waiteraid.com/api

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- restid (required)
- start_date
- auth_hash (secret) (required)

## ETL Streams

- reservations:
  - primary key: id
  - cursor: date
  - fields: date(string), guest_name(string), id(string), status(string)
- meals:
  - primary key: id
  - fields: id(string), max_end(string), min_start(string), name(string)
- queue:
  - primary key: queue_id
  - fields: added_date(string), amount(string), comment(string), cust_id(string), firstname(string), lastname(string), mobile(string), queue_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_booking:
  - endpoint: POST /wa-api/addBooking?restid={{ config.restid }}&start_time={{ record.start_time }}&amount={{ record.amount }}&date={{ record.date }}&mealid={{ record.mealid }}
  - required fields: start_time, amount, date, mealid
  - risk: creates a new restaurant reservation, visible to restaurant staff and the guest; external mutation, approval required
- set_booking_status:
  - endpoint: POST /wa-api/setBookingStatus?restid={{ config.restid }}&bookingId={{ record.id }}&status={{ record.status }}
  - required fields: id, status
  - risk: changes a reservation's status (including marking it deleted); external mutation, approval required
- edit_booking:
  - endpoint: POST /wa-api/editBooking?restid={{ config.restid }}&bookingId={{ record.id }}&start_time={{ record.start_time }}
  - required fields: id, start_time
  - risk: edits an existing reservation's start time; external mutation, approval required
- add_guest:
  - endpoint: POST /wa-api/addGuest?restid={{ config.restid }}&firstname={{ record.firstname }}&lastname={{ record.lastname }}
  - required fields: firstname, lastname
  - risk: creates a new guest record; external mutation, approval required
- add_to_queue:
  - endpoint: POST /wa-api/queue/add?restid={{ config.restid }}&name={{ record.name }}&amount={{ record.amount }}
  - required fields: name, amount
  - risk: adds a guest to the restaurant's walk-in queue; external mutation, approval required
- delete_from_queue:
  - endpoint: POST /wa-api/queue/delete?restid={{ config.restid }}&queue_id={{ record.queue_id }}
  - required fields: queue_id
  - risk: removes a guest from the restaurant's walk-in queue; external mutation, approval required

## Security

- read risk: external WaiterAid API read of restaurant reservation/meal data
- write risk: external mutation of WaiterAid reservations, guests, and walk-in queue; approval required
- approval: read: none, read-only sync surface. write: required for all mutating actions (create/update reservations, guests, and queue entries).
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect waiteraid
```

### Inspect as structured JSON

```bash
pm connectors inspect waiteraid --json
```

## Agent Rules

- Run pm connectors inspect waiteraid before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
