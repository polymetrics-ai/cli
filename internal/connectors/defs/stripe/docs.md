# Overview

Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse
ETL customer actions through the Stripe REST API.

Readable streams: `customers`, `charges`, `invoices`, `subscriptions`, `products`.

Write actions: `create_customer`, `update_customer`.

Service API documentation: https://stripe.com/docs/api.

## Auth setup

Connection fields:

- `account_id` (optional, string); Optional Stripe account ID; sent as the Stripe-Account header for
  Connect.
- `base_url` (optional, string); default `https://api.stripe.com/v1`; format `uri`; Stripe API base
  URL override for tests or proxies.
- `client_secret` (required, secret, string); Stripe secret API key (sk_...). Used only for Bearer
  auth; never logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects created at
  or after this time are read.

Secret fields are redacted in logs and write previews: `client_secret`.

Default configuration values: `base_url=https://api.stripe.com/v1`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `starting_after`; next cursor from last
record field `id`; stop flag `has_more`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `starting_after`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created`; sent as `created[gte]`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`.
- `charges`: GET `/charges` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `starting_after`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created`; sent as `created[gte]`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`.
- `invoices`: GET `/invoices` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `starting_after`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created`; sent as `created[gte]`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`.
- `subscriptions`: GET `/subscriptions` - records path `data`; query `limit`=`100`; cursor
  pagination; cursor parameter `starting_after`; next cursor from last record field `id`; stop flag
  `has_more`; incremental cursor `created`; sent as `created[gte]`; formatted as Unix-seconds
  timestamp; initial lower bound from `start_date`.
- `products`: GET `/products` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `starting_after`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created`; sent as `created[gte]`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`.

## Write actions & risks

Overall write risk: external Stripe API mutation.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `form`; accepted fields
  `description`, `email`, `name`, `phone`; risk: external mutation; approval required.
- `update_customer`: POST `/customers/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `description`, `email`, `id`, `name`,
  `phone`; risk: external mutation; approval required.

## Known limits

- Published rate limit metadata: requests_per_minute=100.
- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=7.
