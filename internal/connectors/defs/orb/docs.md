# Overview

Reads Orb customers, subscriptions, plans, and invoices.

Readable streams: `customers`, `subscriptions`, `plans`, `invoices`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.withorb.com/reference/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Orb API key. Used only for Bearer auth (Authorization:
  Bearer <api_key>); never logged.
- `base_url` (optional, string); default `https://api.withorb.com/v1`; format `uri`; Orb API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects created at
  or after this time are read on a fresh sync.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.withorb.com/v1`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`pagination_metadata.next_cursor`; stop flag `pagination_metadata.has_more`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination_metadata.next_cursor`; stop flag `pagination_metadata.has_more`; incremental cursor
  `created_at`; sent as `created_at[gte]`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `subscriptions`: GET `/subscriptions` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination_metadata.next_cursor`; stop flag `pagination_metadata.has_more`; incremental cursor
  `created_at`; sent as `created_at[gte]`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `plans`: GET `/plans` - records path `data`; query `limit` from template `{{ config.page_size }}`,
  default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination_metadata.next_cursor`; stop flag `pagination_metadata.has_more`; incremental cursor
  `created_at`; sent as `created_at[gte]`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `invoices`: GET `/invoices` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination_metadata.next_cursor`; stop flag `pagination_metadata.has_more`; incremental cursor
  `created_at`; sent as `created_at[gte]`; formatted as `rfc3339`; initial lower bound from
  `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Orb API read of customer and billing data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
