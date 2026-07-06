# Overview

Reads Katana MRP (Cloud Inventory) products, materials, variants, sales orders, and customers
through the Katana REST API.

Readable streams: `products`, `materials`, `variants`, `sales_orders`, `customers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.katanamrp.com/reference/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Katana API key. Used only for Bearer auth; never logged.
- `base_url` (optional, string); default `https://api.katanamrp.com/v1`; format `uri`; Katana API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.katanamrp.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 50.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `products`: GET `/products` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 50; incremental cursor `updated_at`; formatted as
  `rfc3339`.
- `materials`: GET `/materials` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `variants`: GET `/variants` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 50; incremental cursor `updated_at`; formatted as
  `rfc3339`.
- `sales_orders`: GET `/sales_orders` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `customers`: GET `/customers` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; incremental cursor `updated_at`;
  formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Katana MRP API read of inventory, sales, and
customer data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
