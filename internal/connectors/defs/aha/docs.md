# Overview

Reads Aha! features, products, ideas, releases, initiatives, goals, epics, and users through the
Aha! REST API (read-only).

Readable streams: `features`, `products`, `ideas`, `releases`, `initiatives`, `goals`, `epics`,
`users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.aha.io/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Aha! API key, sent as a Bearer token exactly like an OAuth
  bearer token. Never logged.
- `base_url` (optional, string); default `https://secure.aha.io`; format `uri`; Aha! account base
  URL (e.g. https://<company>.aha.io). Defaults to https://secure.aha.io; because Aha! is
  account-scoped, most deployments override this.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://secure.aha.io`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/products` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 30.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `features`: GET `/api/v1/features` - records path `features`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `resource`.
- `products`: GET `/api/v1/products` - records path `products`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `resource`.
- `ideas`: GET `/api/v1/ideas` - records path `ideas`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `resource`.
- `releases`: GET `/api/v1/releases` - records path `releases`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `resource`.
- `initiatives`: GET `/api/v1/initiatives` - records path `initiatives`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `resource`.
- `goals`: GET `/api/v1/goals` - records path `goals`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `resource`.
- `epics`: GET `/api/v1/epics` - records path `epics`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `resource`.
- `users`: GET `/api/v1/users` - records path `users`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30; computed output fields `resource`.

## Write actions & risks

This connector is read-only. Read behavior: external Aha! API read of planning and roadmap data.

## Known limits

- Batch defaults: read_page_size=30.
- API coverage includes 8 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=6, duplicate_of=8, non_data_endpoint=5, out_of_scope=25,
  requires_elevated_scope=1.
