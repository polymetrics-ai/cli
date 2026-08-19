# Overview

Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse
ETL customer create, update, and typed destructive delete actions through the Stripe REST API.

Readable streams: `customers`, `charges`, `invoices`, `subscriptions`, `products`.

Write actions: `create_customer`, `update_customer`, `delete_customer`.

The operation ledger was refreshed against Stripe OpenAPI spec3 version `2026-07-29.dahlia`: 416
paths and 589 documented HTTP operations are tracked exactly once in `api_surface.json`. Operations
that are not declared as streams, writes, or provider commands remain blocked/planned metadata until
their typed schemas, bounds, fixtures, and safety evidence are authored.

The exact public source URL, retrieval time, byte count, SHA-256, and operation inventory are pinned
in `sources/stripe-operation-source-lock.json` (retrieved 2026-08-19).

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

Overall write risk: external Stripe API mutation, including typed destructive customer deletion when
explicitly confirmed.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `form`; accepted non-empty
  string fields `description`, `email`, `name`, `phone`; additional record fields are rejected;
  risk: external mutation; approval required.
- `update_customer`: POST `/customers/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id` (matching `cus_...`) plus at least one non-empty mutable
  field; accepted fields `description`, `email`, `id`, `name`, `phone`; additional record fields are
  rejected; risk: external mutation; approval required.
- `delete_customer`: DELETE `/customers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id` (matching `cus_...`); `id` is redacted from
  operator-visible previews and write errors; idempotent missing-status handling for `404`; typed
  confirmation `destructive`; risk: destructive external mutation; deletes a Stripe customer;
  reverse ETL approval and typed destructive confirmation required.

## Known limits

- Published rate limit metadata: requests_per_minute=100.
- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 3 write-backed endpoint group(s), and
  581 blocked/planned operation-ledger row(s) from the official Stripe OpenAPI source.
- Official lane counts represented in the ledger: `etl_read=242`, `reverse_etl_write=316`,
  `direct_read_query_search=9`, `binary_file=7`, `cdc_changefeed=7`,
  `excluded_not_applicable=8`, `total=589`.
- Fixture-backed executable coverage remains intentionally narrow and uncertified: 5 ETL streams and
  3 customer write actions. No live Stripe credential check, provider call, write, or certification
  evidence is claimed by this bundle.
- DELETE and destructive operations are not blanket-excluded as unsafe. They are in scope only when
  exposed as typed write actions with `confirm: "destructive"` and the existing reverse ETL plan →
  preview → explicit approval → execute path.
- Most official Stripe POST operations remain blocked/planned because their OpenAPI form schemas need
  connector-reviewed field flattening, redaction, idempotency, and fixtures before they can be
  truthfully exposed as executable writes. Exact dependency: a Stripe-compatible nested
  `application/x-www-form-urlencoded` encoder for object/array parameters, or explicit per-action
  flattened `body_fields` schemas, before complex writes are declared executable.
- Direct/search rows remain blocked/planned on the provider search/query boundary foundation tracked
  by issue #2985 plus connector-owned fixed-target command metadata and redaction policies.
- CDC/changefeed rows remain blocked/planned on CDC truthfulness/state foundations tracked by issues
  #2986 and #2988 plus connector-owned event/webhook fixtures.
- Binary/file rows remain blocked/planned until bounded provider command metadata, response
  redaction, size limits, and conformance fixtures are authored.
