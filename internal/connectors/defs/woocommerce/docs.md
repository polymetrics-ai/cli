# Overview

Reads WooCommerce orders, products, customers, and coupons through the WooCommerce REST API (wc/v3).

Readable streams: `orders`, `products`, `customers`, `coupons`.

This connector is read-only; no write actions are declared.

Service API documentation: https://woocommerce.github.io/woocommerce-rest-api-docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); WooCommerce REST API consumer key. Sent as the HTTP Basic
  auth username; never logged.
- `api_secret` (required, secret, string); WooCommerce REST API consumer secret. Sent as the HTTP
  Basic auth password; never logged.
- `base_url` (required, string); format `uri`; WooCommerce store REST API base URL, e.g.
  https://example.com/wp-json/wc/v3.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `page_size` (optional, string); default `10`; Records per page (1-100), sent as per_page.
- `start_date` (optional, string); ISO8601 or YYYY-MM-DD lower bound; only objects modified at or
  after this time are read (modified_after).

Secret fields are redacted in logs and write previews: `api_key`, `api_secret`.

Default configuration values: `max_pages=0`, `page_size=10`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.api_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `orders` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 10.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `orders`: GET `orders` - records at response root; query `after` from template `{{
  incremental.lower_bound }}`, omitted when absent; `order`=`asc`; `orderby`=`id`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 10;
  incremental cursor `date_modified_gmt`; sent as `modified_after`; formatted as `rfc3339`; initial
  lower bound from `start_date`.
- `products`: GET `products` - records at response root; query `after` from template `{{
  incremental.lower_bound }}`, omitted when absent; `order`=`asc`; `orderby`=`id`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 10;
  incremental cursor `date_modified_gmt`; sent as `modified_after`; formatted as `rfc3339`; initial
  lower bound from `start_date`.
- `customers`: GET `customers` - records at response root; query `after` from template `{{
  incremental.lower_bound }}`, omitted when absent; `order`=`asc`; `orderby`=`id`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 10;
  incremental cursor `date_modified_gmt`; sent as `modified_after`; formatted as `rfc3339`; initial
  lower bound from `start_date`.
- `coupons`: GET `coupons` - records at response root; query `after` from template `{{
  incremental.lower_bound }}`, omitted when absent; `order`=`asc`; `orderby`=`id`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 10;
  incremental cursor `date_modified_gmt`; sent as `modified_after`; formatted as `rfc3339`; initial
  lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external WooCommerce store read of orders, products,
customers, and coupons.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=5.
