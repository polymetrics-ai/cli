# Overview

Reads Lago customers, invoices, subscriptions, plans, and billable metrics through the Lago REST
API.

Readable streams: `customers`, `invoices`, `subscriptions`, `plans`, `billable_metrics`.

This connector is read-only; no write actions are declared.

Service API documentation: https://doc.getlago.com/api-reference/intro.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Lago API key. Used only for Bearer auth (Authorization:
  Bearer <api_key>); never logged.
- `api_url` (optional, string); default `https://api.getlago.com/api/v1`; format `uri`; Lago API
  base URL override for tests, self-hosted instances, or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `api_url=https://api.getlago.com/api/v1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `api_url` value after applying defaults.

Connection checks call GET `/customers` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `customers`: GET `/customers` - records path `customers`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `invoices`: GET `/invoices` - records path `invoices`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `subscriptions`: GET `/subscriptions` - records path `subscriptions`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `plans`: GET `/plans` - records path `plans`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100.
- `billable_metrics`: GET `/billable_metrics` - records path `billable_metrics`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Lago API read of billing and subscription data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
