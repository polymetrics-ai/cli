# Overview

Reads Paddle customers, subscriptions, transactions, and products through the Paddle REST API.

Readable streams: `transactions`, `customers`, `subscriptions`, `products`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.paddle.com/api-reference/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Paddle API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.paddle.com`; format `uri`; Paddle API base URL
  override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.paddle.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/transactions` with query `per_page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next token from
`meta.pagination.next`.

- `transactions`: GET `/transactions` - records path `data`; query `per_page`=`100`; cursor
  pagination; cursor parameter `after`; next token from `meta.pagination.next`.
- `customers`: GET `/customers` - records path `data`; query `per_page`=`100`; cursor pagination;
  cursor parameter `after`; next token from `meta.pagination.next`.
- `subscriptions`: GET `/subscriptions` - records path `data`; query `per_page`=`100`; cursor
  pagination; cursor parameter `after`; next token from `meta.pagination.next`.
- `products`: GET `/products` - records path `data`; query `per_page`=`100`; cursor pagination;
  cursor parameter `after`; next token from `meta.pagination.next`.

## Write actions & risks

This connector is read-only. Read behavior: external Paddle API read of customer, subscription,
transaction, and product data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=3.
