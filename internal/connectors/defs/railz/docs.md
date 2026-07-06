# Overview

Reads Railz businesses, connections, customers, invoices, and bills through the Railz REST API.
Read-only.

Readable streams: `businesses`, `connections`, `customers`, `invoices`, `bills`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.railz.ai/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Pre-issued Railz bearer access token (preferred
  credential; checked first). Never logged.
- `api_key` (optional, secret, string); Railz API key, used as a Bearer token fallback when
  access_token is not set. Never logged.
- `base_url` (optional, string); default `https://api.railz.ai/v1`; format `uri`; Railz API base URL
  override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Default configuration values: `base_url=https://api.railz.ai/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- Bearer token authentication using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/businesses` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `businesses`: GET `/businesses` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`; computed output fields `created_at`, `id`, `name`.
- `connections`: GET `/connections` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`; computed output fields `business_id`, `created_at`, `id`.
- `customers`: GET `/customers` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `business_id`, `id`,
  `name`.
- `invoices`: GET `/invoices` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `business_id`,
  `customer_id`, `id`, `total_amount`.
- `bills`: GET `/bills` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; computed output fields `business_id`, `id`,
  `total_amount`, `vendor_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Railz API read of connected-business accounting
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
