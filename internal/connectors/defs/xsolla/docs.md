# Overview

Reads Xsolla merchant transaction search/registry, payouts, payout currency breakdown, and financial
report data, and writes full/partial transaction refunds through the Xsolla Pay Station API.

Readable streams: `projects`, `orders`, `transactions`, `transactions_search`,
`transactions_registry`, `payouts`, `payout_currency_breakdown`, `financial_reports`.

Write actions: `request_refund`, `request_partial_refund`.

Service API documentation: https://developers.xsolla.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Xsolla merchant (or project) API key, used as the Basic auth
  password. Never logged.
- `base_url` (optional, string); default `https://api.xsolla.com/merchant/v2`; format `uri`; Xsolla
  Pay Station / Merchant API v2 base URL override for tests or proxies.
- `datetime_from` (optional, string); Period start (YYYY-MM-DD), passed as the 'datetime_from' query
  param on date-bounded reads (transactions_registry, financial_reports, payouts,
  payout_currency_breakdown).
- `datetime_to` (optional, string); Period end (YYYY-MM-DD), passed as the 'datetime_to' query param
  on date-bounded reads. financial_reports requires the datetime_from/datetime_to window to be 92
  days or less (Xsolla API constraint, not enforced client-side).
- `merchant_id` (required, string); Xsolla merchant ID (integer, shown in Publisher Account >
  Company settings > Company). Used as the Basic auth username and substituted into every
  merchant-scoped path segment. Not a credential by itself; never sufficient to authenticate without
  api_key.
- `mode` (optional, string).
- `project_id` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.xsolla.com/merchant/v2`.

Authentication behavior:

- HTTP Basic authentication using `config.merchant_id`, `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `transactions_registry`, `payouts`, `payout_currency_breakdown`,
`financial_reports`; offset_limit: `transactions_search`; page_number: `projects`, `orders`,
`transactions`.

- `projects`: GET `/projects` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `orders`: GET `/projects/{{ config.project_id }}/orders` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); emits passthrough records.
- `transactions`: GET `/projects/{{ config.project_id }}/transactions` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  maximum 1 page(s); emits passthrough records.
- `transactions_search`: GET `/merchants/{{ config.merchant_id }}/reports/transactions/search.json`
  - records at response root; query `datetime_from` from template `{{ config.datetime_from }}`,
  omitted when absent; `datetime_to` from template `{{ config.datetime_to }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  computed output fields `transaction_create_date`, `transaction_id`; emits passthrough records.
- `transactions_registry`: GET `/merchants/{{ config.merchant_id
  }}/reports/transactions/registry.json` - records at response root; query `datetime_from` from
  template `{{ config.datetime_from }}`, omitted when absent; `datetime_to` from template `{{
  config.datetime_to }}`, omitted when absent; `in_transfer_currency`=`0`; computed output fields
  `transaction_id`, `transaction_transfer_date`; emits passthrough records.
- `payouts`: GET `/merchants/{{ config.merchant_id }}/reports/transfers` - records at response root;
  query `datetime_from` from template `{{ config.datetime_from }}`, omitted when absent;
  `datetime_to` from template `{{ config.datetime_to }}`, omitted when absent; computed output
  fields `payout_date`, `payout_id`; emits passthrough records.
- `payout_currency_breakdown`: GET `/merchants/{{ config.merchant_id
  }}/reports/transactions/summary/transfer` - records at response root; query `datetime_from` from
  template `{{ config.datetime_from }}`, omitted when absent; `datetime_to` from template `{{
  config.datetime_to }}`, omitted when absent; emits passthrough records.
- `financial_reports`: GET `/merchants/{{ config.merchant_id }}/reports` - records at response root;
  query `datetime_from`=`{{ config.datetime_from }}`; `datetime_to`=`{{ config.datetime_to }}`;
  emits passthrough records.

## Write actions & risks

Overall write risk: external mutation: issues full or partial refunds to end users for completed
transactions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `request_refund`: PUT `/merchants/{{ config.merchant_id }}/reports/transactions/{{
  record.transaction_id }}/refund` - kind `update`; body type `json`; path fields `transaction_id`;
  required record fields `transaction_id`, `description`; accepted fields `description`, `email`,
  `transaction_id`; risk: irreversible external mutation; issues a full refund to the user for the
  given transaction; approval required.
- `request_partial_refund`: PUT `/merchants/{{ config.merchant_id }}/reports/transactions/{{
  record.transaction_id }}/partial_refund` - kind `update`; body type `json`; path fields
  `transaction_id`; required record fields `transaction_id`, `description`, `refund_amount`;
  accepted fields `description`, `refund_amount`, `transaction_id`; risk: irreversible external
  mutation; issues a partial refund to the user for the given transaction; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, out_of_scope=3, requires_elevated_scope=2.
