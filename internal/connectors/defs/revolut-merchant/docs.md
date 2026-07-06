# Overview

Reads Revolut Merchant orders, customers, settlements, and payment links through the REST API.

Readable streams: `orders`, `customers`, `settlements`, `payment_links`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.revolut.com/docs/guides/merchant/reference/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Revolut Merchant API secret key, sent as an Authorization:
  Bearer <api_key> header. Never logged.
- `base_url` (optional, string); default `https://merchant.revolut.com/api/1.0`; format `uri`;
  Revolut Merchant API base URL override for tests or proxies.
- `customer_id` (optional, string); Optional passthrough filter: scope the orders stream to a single
  customer id.
- `from_created_date` (optional, string); Optional passthrough filter: only return records created
  at or after this value.
- `state` (optional, string); Optional passthrough filter: scope the orders stream to a single order
  state.
- `to_created_date` (optional, string); Optional passthrough filter: only return records created at
  or before this value.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://merchant.revolut.com/api/1.0`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/orders` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `orders`: GET `/orders` - records at response root; query `customer_id` from template `{{
  config.customer_id }}`, omitted when absent; `from_created_date` from template `{{
  config.from_created_date }}`, omitted when absent; `state` from template `{{ config.state }}`,
  omitted when absent; `to_created_date` from template `{{ config.to_created_date }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; computed output fields `stream`; emits passthrough records.
- `customers`: GET `/customers` - records at response root; query `customer_id` from template `{{
  config.customer_id }}`, omitted when absent; `from_created_date` from template `{{
  config.from_created_date }}`, omitted when absent; `state` from template `{{ config.state }}`,
  omitted when absent; `to_created_date` from template `{{ config.to_created_date }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; computed output fields `stream`; emits passthrough records.
- `settlements`: GET `/settlements` - records at response root; query `customer_id` from template
  `{{ config.customer_id }}`, omitted when absent; `from_created_date` from template `{{
  config.from_created_date }}`, omitted when absent; `state` from template `{{ config.state }}`,
  omitted when absent; `to_created_date` from template `{{ config.to_created_date }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; computed output fields `stream`; emits passthrough records.
- `payment_links`: GET `/payment-links` - records at response root; query `customer_id` from
  template `{{ config.customer_id }}`, omitted when absent; `from_created_date` from template `{{
  config.from_created_date }}`, omitted when absent; `state` from template `{{ config.state }}`,
  omitted when absent; `to_created_date` from template `{{ config.to_created_date }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; computed output fields `stream`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Revolut Merchant API read of order, customer,
settlement, and payment-link data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, out_of_scope=5.
