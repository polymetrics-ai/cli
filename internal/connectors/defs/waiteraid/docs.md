# Overview

Reads and writes WaiterAid restaurant reservations, meals, guests, and queue entries.

Readable streams: `reservations`, `meals`, `queue`.

Write actions: `add_booking`, `set_booking_status`, `edit_booking`, `add_guest`, `add_to_queue`,
`delete_from_queue`.

Service API documentation: https://app.waiteraid.com/api-docs/index.html.

## Auth setup

Connection fields:

- `auth_hash` (required, secret, string).
- `base_url` (optional, string); default `https://app.waiteraid.com`; format `uri`; WaiterAid API
  base URL.
- `restid` (required, string).
- `start_date` (optional, string); Optional date filter (YYYY-MM-DD), sent as the real
  getBooking/searchBooking endpoint's own 'date' query parameter when configured.

Secret fields are redacted in logs and write previews: `auth_hash`.

Default configuration values: `base_url=https://app.waiteraid.com`.

Authentication behavior:

- API key authentication in query parameter `auth_hash` using `secrets.auth_hash`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call POST `/wa-api/searchBooking` with query `restid`=`{{ config.restid }}`.

## Streams notes

Default pagination: single request; no pagination.

- `reservations`: POST `/wa-api/searchBooking` - records path `bookings`; query `date` from template
  `{{ config.start_date }}`, omitted when absent; `restid`=`{{ config.restid }}`; computed output
  fields `guest_name`.
- `meals`: POST `/wa-api/getMeals` - records path `meals`; query `restid`=`{{ config.restid }}`.
- `queue`: POST `/wa-api/queue/list` - records path `queue`; query `date` from template `{{
  config.start_date }}`, omitted when absent; `restid`=`{{ config.restid }}`.

## Write actions & risks

Overall write risk: external mutation of WaiterAid reservations, guests, and walk-in queue; approval
required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_booking`: POST `/wa-api/addBooking?restid={{ config.restid }}&start_time={{ record.start_time
  }}&amount={{ record.amount }}&date={{ record.date }}&mealid={{ record.mealid }}` - kind `create`;
  body type `none`; required record fields `start_time`, `amount`, `date`, `mealid`; accepted fields
  `amount`, `date`, `mealid`, `start_time`; risk: creates a new restaurant reservation, visible to
  restaurant staff and the guest; external mutation, approval required.
- `set_booking_status`: POST `/wa-api/setBookingStatus?restid={{ config.restid }}&bookingId={{
  record.id }}&status={{ record.status }}` - kind `update`; body type `none`; path fields `id`;
  required record fields `id`, `status`; accepted fields `id`, `status`; risk: changes a
  reservation's status (including marking it deleted); external mutation, approval required.
- `edit_booking`: POST `/wa-api/editBooking?restid={{ config.restid }}&bookingId={{ record.id
  }}&start_time={{ record.start_time }}` - kind `update`; body type `none`; path fields `id`;
  required record fields `id`, `start_time`; accepted fields `id`, `start_time`; risk: edits an
  existing reservation's start time; external mutation, approval required.
- `add_guest`: POST `/wa-api/addGuest?restid={{ config.restid }}&firstname={{ record.firstname
  }}&lastname={{ record.lastname }}` - kind `create`; body type `none`; required record fields
  `firstname`, `lastname`; accepted fields `firstname`, `lastname`; risk: creates a new guest
  record; external mutation, approval required.
- `add_to_queue`: POST `/wa-api/queue/add?restid={{ config.restid }}&name={{ record.name
  }}&amount={{ record.amount }}` - kind `create`; body type `none`; required record fields `name`,
  `amount`; accepted fields `amount`, `name`; risk: adds a guest to the restaurant's walk-in queue;
  external mutation, approval required.
- `delete_from_queue`: POST `/wa-api/queue/delete?restid={{ config.restid }}&queue_id={{
  record.queue_id }}` - kind `delete`; body type `none`; path fields `queue_id`; required record
  fields `queue_id`; accepted fields `queue_id`; risk: removes a guest from the restaurant's walk-in
  queue; external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, non_data_endpoint=1, out_of_scope=10.
