# Overview

Reads eBay seller financial data - transactions, payouts, transfers, and the seller funds summary -
through the eBay Sell Finances REST API.

Readable streams: `transactions`, `payouts`, `transfers`, `seller_funds_summary`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.ebay.com/api-docs/sell/finances/overview.html.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://apiz.ebay.com/sell/finances/v1`; format `uri`;
  eBay Sell Finances API base URL. Defaults to the production endpoint; override for eBay's sandbox
  environment or test proxies.
- `client_access_token` (optional, secret, string); eBay OAuth user access token with the
  sell.finances scope, sent as a Bearer token (Authorization: Bearer <client_access_token>). Never
  logged.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for
  transactions/payouts/transfers, sent as a date-range filter (e.g.
  transactionDate:[<start_date>..]). Only used on a fresh sync with no persisted cursor.

Secret fields are redacted in logs and write previews: `client_access_token`.

Default configuration values: `base_url=https://apiz.ebay.com/sell/finances/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.client_access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/seller_funds_summary`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 200.

Pagination by stream: none: `seller_funds_summary`; offset_limit: `transactions`, `payouts`,
`transfers`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `transactions`: GET `/transaction` - records path `transactions`; query `filter` from template
  `transactionDate:[{{ incremental.lower_bound }}..]`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 2; incremental cursor
  `transactionDate`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `amount_currency`, `amount_value`.
- `payouts`: GET `/payout` - records path `payouts`; query `filter` from template `payoutDate:[{{
  incremental.lower_bound }}..]`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 200; incremental cursor `payoutDate`; formatted as
  `rfc3339`; initial lower bound from `start_date`; computed output fields `amount_currency`,
  `amount_value`, `payoutInstrument_accountLastFourDigits`, `payoutInstrument_nickname`.
- `transfers`: GET `/transfer` - records path `transfers`; query `filter` from template
  `transactionDate:[{{ incremental.lower_bound }}..]`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 200; incremental cursor
  `transferDate`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `amount_currency`, `amount_value`.
- `seller_funds_summary`: GET `/seller_funds_summary` - records at response root; computed output
  fields `availableFunds_currency`, `availableFunds_value`, `fundsOnHold_currency`,
  `fundsOnHold_value`, `processingFunds_currency`, `processingFunds_value`, `totalFunds_currency`,
  `totalFunds_value`.

## Write actions & risks

This connector is read-only. Read behavior: external eBay Sell Finances API read of a seller's
monetary records.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=2.
