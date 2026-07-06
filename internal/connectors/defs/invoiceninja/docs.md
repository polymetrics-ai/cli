# Overview

Reads Invoice Ninja clients, invoices, products, payments, and quotes through the Invoice Ninja v5
REST API.

Readable streams: `clients`, `invoices`, `products`, `payments`, `quotes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-docs.invoicing.co/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Invoice Ninja API token. Sent as the X-API-TOKEN request
  header; never logged.
- `base_url` (optional, string); default `https://invoicing.co/api/v1`; format `uri`; Invoice Ninja
  API base URL override for self-hosted instances, tests, or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://invoicing.co/api/v1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `X-API-TOKEN` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/clients` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `clients`: GET `/clients` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `invoices`: GET `/invoices` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `products`: GET `/products` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `payments`: GET `/payments` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `quotes`: GET `/quotes` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Invoice Ninja API read of client and billing
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=3.
