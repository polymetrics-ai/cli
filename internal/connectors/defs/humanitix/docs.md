# Overview

Reads Humanitix events, orders, tickets, and tags through the Humanitix public REST API.

Readable streams: `events`, `tags`, `orders`, `tickets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.humanitix.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Humanitix API key, sent as the x-api-key header. Never
  logged.
- `base_url` (optional, string); default `https://api.humanitix.com/v1`; format `uri`; Humanitix API
  base URL override for tests or proxies.
- `event_id` (optional, string); Humanitix event id the 'orders' and 'tickets' streams are scoped to
  (required for those two streams; substituted into the event-scoped path).
- `page_size` (optional, integer); default `100`; Page size for the pageSize query parameter
  (1-100).
- `since` (optional, string); Optional ISO-8601 lower bound for the 'events' stream's incremental
  `since` filter; the stream's own state cursor takes precedence once a sync has run once.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.humanitix.com/v1`, `page_size=100`.

Authentication behavior:

- API key authentication in `x-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/events` with query `page`=`1`; `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `events`: GET `/events` - records path `events`; query `since` from template `{{
  incremental.lower_bound }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100; incremental cursor `updatedAt`; sent as
  `since`; formatted as `rfc3339`; initial lower bound from `since`.
- `tags`: GET `/tags` - records path `tags`; query `since` from template `{{ incremental.lower_bound
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter
  `pageSize`; starts at 1; page size 100; incremental cursor `updatedAt`; sent as `since`; formatted
  as `rfc3339`; initial lower bound from `since`.
- `orders`: GET `/events/{{ config.event_id }}/orders` - records path `orders`; query `since` from
  template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `pageSize`; starts at 1; page size 100; incremental cursor
  `updatedAt`; sent as `since`; formatted as `rfc3339`; initial lower bound from `since`.
- `tickets`: GET `/events/{{ config.event_id }}/tickets` - records path `tickets`; query `since`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `pageSize`; starts at 1; page size 100; incremental cursor
  `updatedAt`; sent as `since`; formatted as `rfc3339`; initial lower bound from `since`.

## Write actions & risks

This connector is read-only. Read behavior: external Humanitix API read of event, order, ticket, and
tag data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
