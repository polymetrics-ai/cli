# Overview

Reads Kissmetrics products, reports, events, and properties through the Kissmetrics query API using
HTTP Basic authentication.

Readable streams: `products`, `reports`, `events`, `properties`.

This connector is read-only; no write actions are declared.

Service API documentation: https://support.kissmetrics.io/reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://query.kissmetrics.io/v3`; format `uri`;
  Kissmetrics query API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-200).
- `password` (required, secret, string); Kissmetrics API secret, sent via HTTP Basic auth. Never
  logged.
- `product_id` (optional, string); Kissmetrics product (account) id. Required for the reports,
  events, and properties streams, which are scoped to a single product partition
  (products/{product_id}/<resource>).
- `username` (required, string); Kissmetrics API key/username, sent via HTTP Basic auth.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://query.kissmetrics.io/v3`, `max_pages=0`,
`page_size=50`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 50.

- `products`: GET `/products` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50.
- `reports`: GET `/products/{{ config.product_id }}/reports` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50.
- `events`: GET `/products/{{ config.product_id }}/events` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50.
- `properties`: GET `/products/{{ config.product_id }}/properties` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 50.

## Write actions & risks

This connector is read-only. Read behavior: external Kissmetrics query API read of product analytics
metadata.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
