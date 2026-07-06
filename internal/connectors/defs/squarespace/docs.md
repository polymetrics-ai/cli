# Overview

Reads Squarespace orders, products, inventory, profiles, transactions, store pages, webhook
subscriptions, and contacts, and writes webhook subscription mutations through the Squarespace
Commerce API.

Readable streams: `orders`, `products`, `inventory`, `profiles`, `transactions`, `store_pages`,
`webhook_subscriptions`, `contacts`.

Write actions: `create_webhook_subscription`, `delete_webhook_subscription`.

Service API documentation: https://developers.squarespace.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Squarespace API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.squarespace.com/1.0`; format `uri`;
  Squarespace API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.squarespace.com/1.0`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/commerce/orders` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`pagination.nextPageCursor`.

Pagination by stream: cursor: `orders`, `products`, `inventory`, `profiles`, `transactions`,
`store_pages`, `contacts`; none: `webhook_subscriptions`.

- `orders`: GET `/commerce/orders` - records path `result`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `pagination.nextPageCursor`.
- `products`: GET `/commerce/products` - records path `result`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `pagination.nextPageCursor`.
- `inventory`: GET `/commerce/inventory` - records path `result`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `pagination.nextPageCursor`.
- `profiles`: GET `/profiles` - records path `profiles`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `pagination.nextPageCursor`.
- `transactions`: GET `/commerce/transactions` - records path `documents`; cursor pagination; cursor
  parameter `cursor`; next token from `pagination.nextPageCursor`.
- `store_pages`: GET `/commerce/store_pages` - records path `storePages`; cursor pagination; cursor
  parameter `cursor`; next token from `pagination.nextPageCursor`.
- `webhook_subscriptions`: GET `/webhook_subscriptions` - records path `webhookSubscriptions`.
- `contacts`: GET `https://api.squarespace.com/v1/contacts` - records path `contacts`; query
  `pageSize`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination.nextPageCursor`.

## Write actions & risks

Overall write risk: external Squarespace API mutation (webhook subscription create/delete).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_webhook_subscription`: POST `/webhook_subscriptions` - kind `create`; body type `json`;
  required record fields `endpointUrl`; accepted fields `endpointUrl`, `topics`; risk: registers a
  new HTTPS endpoint to receive live order/contact/address event notifications; low-risk external
  mutation, no approval required.
- `delete_webhook_subscription`: DELETE `/webhook_subscriptions/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: permanently removes a webhook subscription,
  stopping future event notifications to that endpoint; external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=2, duplicate_of=8, non_data_endpoint=5, out_of_scope=21.
