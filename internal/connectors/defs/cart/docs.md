# Overview

Reads Cart.com orders, customers, products, and inventory through a read-only REST API.

Readable streams: `orders`, `customers`, `products`, `inventory`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.cart.com.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Cart.com API access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (required, string); format `uri`; Cart.com store API base URL, e.g.
  https://{store_name}/api/v1.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/orders` with query `page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `meta.next_page`.

- `orders`: GET `/orders` - records path `data`; query `page_size`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `meta.next_page`; emits passthrough records.
- `customers`: GET `/customers` - records path `data`; query `page_size`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `page`; next token from `meta.next_page`; emits passthrough
  records.
- `products`: GET `/products` - records path `data`; query `page_size`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `page`; next token from `meta.next_page`; emits passthrough
  records.
- `inventory`: GET `/inventory` - records path `data`; query `page_size`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `page`; next token from `meta.next_page`; emits passthrough
  records.

## Write actions & risks

This connector is read-only. Read behavior: external Cart.com API read of order, customer, product,
and inventory data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
