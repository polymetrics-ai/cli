# Overview

Reads Square payments, refunds, customers, and locations through the Square Connect v2 REST API.

Readable streams: `payments`, `refunds`, `customers`, `locations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.squareup.com/reference/square.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Square access token (API key or OAuth access token). Used
  only for Bearer auth; never logged.
- `base_url` (optional, string); default `https://connect.squareup.com/v2`; format `uri`; Square API
  base URL. Set explicitly to https://connect.squareupsandbox.com/v2 for sandbox testing.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); Lower bound (YYYY-MM-DD or RFC3339) for
  payments/refunds/customers time filtering; only objects at or after this time are read.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://connect.squareup.com/v2`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/locations`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `cursor`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `payments`: GET `/payments` - records path `payments`; query `begin_time` from template `{{
  incremental.lower_bound }}`, omitted when absent; `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; incremental cursor `updated_at`; formatted as
  `rfc3339`; initial lower bound from `start_date`.
- `refunds`: GET `/refunds` - records path `refunds`; query `begin_time` from template `{{
  incremental.lower_bound }}`, omitted when absent; `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; incremental cursor `updated_at`; formatted as
  `rfc3339`; initial lower bound from `start_date`.
- `customers`: GET `/customers` - records path `customers`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`.
- `locations`: GET `/locations` - records path `locations`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`.

## Write actions & risks

This connector is read-only. Read behavior: external Square API read of payments, refunds, customer,
and location data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
