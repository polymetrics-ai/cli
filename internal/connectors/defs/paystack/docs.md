# Overview

Reads Paystack customers, transactions, subscriptions, invoices, and disputes through the Paystack
REST API.

Readable streams: `customers`, `transactions`, `subscriptions`, `invoices`, `disputes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://paystack.com/docs/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.paystack.co`; format `uri`; Paystack API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `secret_key` (required, secret, string); Paystack secret key (sk_...). Used only for Bearer auth;
  never logged.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; sent as the `from`
  filter on a fresh sync with no persisted cursor.

Secret fields are redacted in logs and write previews: `secret_key`.

Default configuration values: `base_url=https://api.paystack.co`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.secret_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customer` with query `perPage`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `meta.next`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customer` - records path `data`; query `from` from template `{{
  incremental.lower_bound }}`, omitted when absent; `perPage`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `meta.next`; incremental cursor `createdAt`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `transactions`: GET `/transaction` - records path `data`; query `from` from template `{{
  incremental.lower_bound }}`, omitted when absent; `perPage`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `meta.next`; incremental cursor `createdAt`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `subscriptions`: GET `/subscription` - records path `data`; query `from` from template `{{
  incremental.lower_bound }}`, omitted when absent; `perPage`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `meta.next`; incremental cursor `createdAt`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `invoices`: GET `/paymentrequest` - records path `data`; query `from` from template `{{
  incremental.lower_bound }}`, omitted when absent; `perPage`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `meta.next`; incremental cursor `createdAt`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `disputes`: GET `/dispute` - records path `data`; query `from` from template `{{
  incremental.lower_bound }}`, omitted when absent; `perPage`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page`; next token from `meta.next`; incremental cursor `createdAt`;
  formatted as `rfc3339`; initial lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Paystack API read of customer and payment data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=4.
