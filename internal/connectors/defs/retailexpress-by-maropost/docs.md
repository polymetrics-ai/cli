# Overview

Reads Retail Express products, customers, orders, stock levels, and stores through the Maropost API.
Read-only.

Readable streams: `products`, `customers`, `orders`, `stock_levels`, `stores`.

This connector is read-only; no write actions are declared.

Service API documentation: https://retailexpress.atlassian.net/wiki/spaces/APIDOC/overview.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); OAuth-style access token, sent as a Bearer token. Never
  logged.
- `api_key` (optional, secret, string); API key, sent as the X-API-Key header. Used only when
  access_token is not configured. Never logged.
- `base_url` (required, string); format `uri`; Retail Express account base URL, e.g.
  https://<account>.retailexpress.com.au/api/v2.
- `created_after` (optional, string); Optional created-after filter applied to every stream's
  request.
- `status` (optional, string); Optional status filter applied to every stream's request.
- `store_id` (optional, string); Optional store ID filter applied to every stream's request.
- `updated_after` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- API key authentication in `X-API-Key` using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `products`: GET `/products` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `store_id` from template `{{ config.store_id }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental
  cursor `updated_at`; formatted as `rfc3339`; computed output fields `stream`; emits passthrough
  records.
- `customers`: GET `/customers` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `store_id` from template `{{ config.store_id }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental
  cursor `updated_at`; formatted as `rfc3339`; computed output fields `stream`; emits passthrough
  records.
- `orders`: GET `/orders` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `store_id` from template `{{ config.store_id }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental
  cursor `updated_at`; formatted as `rfc3339`; computed output fields `stream`; emits passthrough
  records.
- `stock_levels`: GET `/stock-levels` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `store_id` from template `{{ config.store_id }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental
  cursor `updated_at`; formatted as `rfc3339`; computed output fields `stream`; emits passthrough
  records.
- `stores`: GET `/stores` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `store_id` from template `{{ config.store_id }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `stream`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Retail Express by Maropost API read of product,
customer, order, and stock data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=1.
