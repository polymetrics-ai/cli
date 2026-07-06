# Overview

Reads PayPal transactions, balances, catalog products, and customer disputes through the PayPal REST
API using OAuth 2.0 client-credentials auth.

Readable streams: `transactions`, `balances`, `products`, `disputes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.paypal.com/api/rest/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api-m.paypal.com`; format `uri`; PayPal API base
  URL. Defaults to the production host; set to https://api-m.sandbox.paypal.com directly for
  PayPal's sandbox environment (see docs.md Known limits for why is_sandbox is not modeled as a
  separate toggle).
- `client_id` (required, secret, string); PayPal REST app client id, used as the HTTP Basic username
  on the OAuth 2.0 client-credentials token exchange. Never logged.
- `client_secret` (required, secret, string); PayPal REST app client secret, used as the HTTP Basic
  password on the OAuth 2.0 client-credentials token exchange. Never logged.
- `end_date` (optional, string); format `date-time`; RFC3339 upper bound for the transactions
  reporting stream's end_date query param.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `start_date` (required, string); format `date-time`; RFC3339 lower bound for the transactions
  reporting stream's start_date query param. Required for every read (the incremental cursor, when
  present from a prior sync, raises this bound automatically).

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api-m.paypal.com`, `max_pages=0`.

Authentication behavior:

- Connector-specific authentication.
- OAuth tokens are cached and refreshed 60 seconds before their declared expiration.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/reporting/balances`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `balances`; link-based: `disputes`; page_number: `transactions`,
`products`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `transactions`: GET `/v1/reporting/transactions` - records path `transaction_details`; query
  `end_date` from template `{{ config.end_date }}`, omitted when absent; `fields`=`all`;
  `start_date`=`{{ incremental.lower_bound }}`; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 100; incremental cursor
  `transaction_initiation_date`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `amount`, `currency_code`, `fee_amount`, `paypal_account_id`,
  `transaction_event_code`, `transaction_id`, `transaction_initiation_date`, `transaction_status`,
  `transaction_updated_date`.
- `balances`: GET `/v1/reporting/balances` - records path `balances`; computed output fields
  `available_value`, `currency`, `primary`, `total_currency_code`, `total_value`, `withheld_value`.
- `products`: GET `/v1/catalogs/products` - records path `products`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 20 (the products endpoint
  enforces a maximum page size of 20 records, while other streams use 100).
- `disputes`: GET `/v1/customer/disputes` - records path `items`; paginates by following the
  response's `rel="next"` link, starting with `page_size=50`.

## Write actions & risks

This connector is read-only. Read behavior: external PayPal REST API read of transaction, balance,
catalog, and dispute data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
