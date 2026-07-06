# Overview

Reads Omnisend contacts, campaigns, carts, orders, and products through the Omnisend REST API.

Readable streams: `contacts`, `campaigns`, `carts`, `orders`, `products`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-docs.omnisend.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Omnisend API key. Sent as the X-API-KEY header; never
  logged.
- `base_url` (optional, string); default `https://api.omnisend.com/v3`; format `uri`; Omnisend API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-250).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.omnisend.com/v3`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `paging.next`; next
URLs stay on the configured API host.

- `contacts`: GET `/contacts` - records path `contacts`; query `limit`=`{{ config.page_size }}`;
  follows a next-page URL from the response body; URL path `paging.next`; next URLs stay on the
  configured API host.
- `campaigns`: GET `/campaigns` - records path `campaign`; query `limit`=`{{ config.page_size }}`;
  follows a next-page URL from the response body; URL path `paging.next`; next URLs stay on the
  configured API host.
- `carts`: GET `/carts` - records path `carts`; query `limit`=`{{ config.page_size }}`; follows a
  next-page URL from the response body; URL path `paging.next`; next URLs stay on the configured API
  host.
- `orders`: GET `/orders` - records path `orders`; query `limit`=`{{ config.page_size }}`; follows a
  next-page URL from the response body; URL path `paging.next`; next URLs stay on the configured API
  host.
- `products`: GET `/products` - records path `products`; query `limit`=`{{ config.page_size }}`;
  follows a next-page URL from the response body; URL path `paging.next`; next URLs stay on the
  configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external Omnisend API read of contact, campaign, and
ecommerce order data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
