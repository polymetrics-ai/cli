# Overview

Reads Eventzilla events, categories, users, attendees, ticket types, and transactions, and writes
attendee check-in and event sales-page toggle mutations, through the Eventzilla v2 REST API.

Readable streams: `events`, `categories`, `users`, `attendees`, `tickets`, `transactions`.

Write actions: `checkin_attendee`, `toggle_event_sales`.

Service API documentation: https://www.eventzilla.net/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Eventzilla API key, sent as the x-api-key request header.
  Never logged.
- `base_url` (optional, string); default `https://www.eventzillaapi.net/api/v2`; format `uri`;
  Eventzilla API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://www.eventzillaapi.net/api/v2`.

Authentication behavior:

- API key authentication in `x-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/events` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `events`: GET `/events` - records path `events`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `categories`: GET `/categories` - records path `categories`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `users`: GET `/users` - records path `users`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `attendees`: GET `/events/{{ fanout.id }}/attendees` - records path `attendees`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; fan-out; ids from
  request `/events`; id-list records path `events`; id field `id`; id inserted into the request
  path.
- `tickets`: GET `/events/{{ fanout.id }}/tickets` - records path `tickets`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; fan-out; ids from
  request `/events`; id-list records path `events`; id field `id`; id inserted into the request
  path.
- `transactions`: GET `/events/{{ fanout.id }}/transactions` - records path `transactions`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  fan-out; ids from request `/events`; id-list records path `events`; id field `id`; id inserted
  into the request path.

## Write actions & risks

Overall write risk: external mutation of attendee check-in state and event sales-page publish
status; every write ships with an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `checkin_attendee`: POST `/attendees/checkin` - kind `update`; body type `json`; required record
  fields `barcode`, `eventcheckin`; accepted fields `barcode`, `eventcheckin`; risk: marks an
  attendee checked in or reverts check-in at the door; low-risk operational mutation, no approval
  required.
- `toggle_event_sales`: POST `/events/togglesales` - kind `update`; body type `json`; required
  record fields `eventid`, `status`; accepted fields `eventid`, `status`; risk: publishes or
  unpublishes an event's public sales page; setting status false immediately stops new ticket sales
  for that event, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=5, out_of_scope=4.
