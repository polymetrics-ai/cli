# Overview

Reads Klarna settlement payouts and transactions through the Klarna Settlements API.

Readable streams: `payouts`, `transactions`, `payout_details`, `payout_summaries`, `payout_summary`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.klarna.com/api/.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Klarna Settlements API base URL, e.g.
  https://api.klarna.com (EU production), https://api-na.klarna.com (NA), https://api-oc.klarna.com
  (OC), or the matching *.playground.klarna.com host for test merchants.
- `mode` (optional, string).
- `password` (required, secret, string); Klarna API shared secret (password), sent via HTTP Basic
  auth. Never logged.
- `payment_references` (optional, string); Comma-separated Klarna payment_reference values to read
  through the payout_details stream.
- `summary_currency_code` (optional, string); Optional currency_code filter for Klarna's
  /payouts/summary endpoint.
- `summary_end_date` (optional, string); Required when reading payout_summaries. ISO 8601 end
  date/time for Klarna's /payouts/summary endpoint.
- `summary_start_date` (optional, string); Required when reading payout_summaries. ISO 8601 start
  date/time for Klarna's /payouts/summary endpoint.
- `username` (required, secret, string); Klarna merchant UID (username), sent via HTTP Basic auth.

Secret fields are redacted in logs and write previews: `password`, `username`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/settlements/v1/payouts` with query `size`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `size`; page
size 100.

Pagination by stream: none: `payout_details`, `payout_summaries`; offset_limit: `payouts`,
`transactions`, `payout_summary`.

- `payouts`: GET `/settlements/v1/payouts` - records path `payouts`; offset/limit pagination; offset
  parameter `offset`; limit parameter `size`; page size 100; computed output fields `currency_code`,
  `merchant_settlement_type`, `payment_reference`, `payout_reference`, `settlement_amount`,
  `totals`.
- `transactions`: GET `/settlements/v1/transactions` - records path `transactions`; offset/limit
  pagination; offset parameter `offset`; limit parameter `size`; page size 100.
- `payout_details`: GET `/settlements/v1/payouts/{{ fanout.id }}` - single-object response; records
  at response root; fan-out; ids from config field `payment_references`; id inserted into the
  request path.
- `payout_summaries`: GET `/settlements/v1/payouts/summary` - records path `.`; query
  `currency_code` from template `{{ config.summary_currency_code }}`, omitted when absent;
  `end_date`=`{{ config.summary_end_date }}`; `start_date`=`{{ config.summary_start_date }}`.
- `payout_summary`: GET `/settlements/v1/payouts` - records path `payouts`; offset/limit pagination;
  offset parameter `offset`; limit parameter `size`; page size 100; computed output fields
  `currency_code`, `fee_amount`, `payout_reference`, `return_amount`, `sale_amount`,
  `settlement_amount`.

## Write actions & risks

This connector is read-only. Read behavior: external Klarna Settlements API read of payout and
transaction data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4.
