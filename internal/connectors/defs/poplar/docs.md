# Overview

Reads Poplar campaigns and orders through read-only REST list endpoints.

Readable streams: `campaigns`, `orders`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.poplar.studio/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Poplar API token, sent as a Bearer token (Authorization:
  Bearer <api_token>). Never logged.
- `base_url` (optional, string); default `https://api.heypoplar.com/v1`; format `uri`; Poplar API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.heypoplar.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/campaigns` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `meta.next_page`;
maximum 3 page(s).

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `campaigns`: GET `/campaigns` - records path `data`; query `page`=`1`; `per_page`=`100`; cursor
  pagination; cursor parameter `page`; next token from `meta.next_page`; maximum 3 page(s); computed
  output fields `created_at`.
- `orders`: GET `/orders` - records path `data`; query `page`=`1`; `per_page`=`100`; cursor
  pagination; cursor parameter `page`; next token from `meta.next_page`; maximum 3 page(s);
  incremental cursor `created_at`; formatted as `rfc3339`; computed output fields `created_at`,
  `name`.

## Write actions & risks

This connector is read-only. Read behavior: external Poplar API read of campaign and order data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
