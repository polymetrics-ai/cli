# Overview

Reads inFlow Inventory products, customers, vendors, sales orders, and categories through the inFlow
cloud REST API.

Readable streams: `products`, `customers`, `vendors`, `sales_orders`, `categories`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.inflowinventory.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); inFlow Inventory API key, sent as the raw value of the
  Authorization header (no Bearer prefix). Never logged.
- `base_url` (optional, string); default `https://cloudapi.inflowinventory.com`; format `uri`;
  inFlow Inventory API base URL override for tests or proxies.
- `companyid` (required, string); inFlow company id, embedded as the first path segment of every
  resource request.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://cloudapi.inflowinventory.com`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/{{ config.companyid }}/categories` with query `count`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`.

- `products`: GET `/{{ config.companyid }}/products` - records path `.`; query `count`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next cursor from last record
  field `productId`.
- `customers`: GET `/{{ config.companyid }}/customers` - records path `.`; query `count`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next cursor from last record
  field `customerId`.
- `vendors`: GET `/{{ config.companyid }}/vendors` - records path `.`; query `count`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next cursor from last record
  field `vendorId`.
- `sales_orders`: GET `/{{ config.companyid }}/sales-orders` - records path `.`; query `count`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next cursor from last record
  field `salesOrderId`.
- `categories`: GET `/{{ config.companyid }}/categories` - records path `.`; query `count`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next cursor from last record
  field `categoryId`.

## Write actions & risks

This connector is read-only. Read behavior: external inFlow Inventory API read of products,
customers, vendors, sales orders, and categories.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
