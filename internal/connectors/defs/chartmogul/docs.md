# Overview

Reads and writes ChartMogul customers, contacts, subscription activities, plans, invoices, tasks,
customer-count metrics, and account details through the ChartMogul REST API.

Readable streams: `customers`, `activities`, `customer_count`, `account`, `plans`, `contacts`,
`tasks`, `invoices`.

Write actions: `create_customer`, `update_customer`.

Service API documentation: https://dev.chartmogul.com/reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); ChartMogul API key. Sent as the HTTP Basic username with an
  empty password; never logged.
- `base_url` (optional, string); default `https://api.chartmogul.com/v1`; format `uri`; ChartMogul
  API base URL.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for the activities
  stream's start-date filter and the metrics window's start date.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.chartmogul.com/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `customers`, `activities`, `plans`, `contacts`, `tasks`, `invoices`;
none: `customer_count`, `account`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers` - records path `entries`; query `per_page`=`2`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `activities`: GET `/activities` - records path `entries`; query `per_page`=`2`; `start-date` from
  template `{{ incremental.lower_bound }}`, omitted when absent; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; stop flag `has_more`; incremental cursor `date`; formatted as
  `rfc3339`; initial lower bound from `start_date`.
- `customer_count`: GET `/metrics/customer-count` - records path `entries`; query
  `end-date`=`2099-12-31`; `interval`=`month`; `start-date` from template `{{
  incremental.lower_bound }}`, default `2026-01-01`; incremental cursor `date`; formatted as
  YYYY-MM-DD date; initial lower bound from `start_date`.
- `account`: GET `/account` - single-object response; records path `.`.
- `plans`: GET `/plans` - records path `plans`; query `per_page`=`2`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `contacts`: GET `/contacts` - records path `entries`; query `per_page`=`2`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `tasks`: GET `/tasks` - records path `entries`; query `per_page`=`2`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `invoices`: GET `/invoices` - records path `invoices`; query `per_page`=`2`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`.

## Write actions & risks

Overall write risk: external mutation of ChartMogul customer records; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `data_source_uuid`, `external_id`; accepted fields `city`, `company`, `country`,
  `data_source_uuid`, `email`, `external_id`, `free_trial_started_at`, `lead_created_at`, `name`,
  `owner`, `website_url`; risk: external mutation; approval required.
- `update_customer`: PUT `/customers/{{ record.uuid }}` - kind `update`; body type `json`; path
  fields `uuid`; required record fields `uuid`; accepted fields `city`, `company`, `country`,
  `email`, `name`, `owner`, `uuid`, `website_url`; risk: external mutation; approval required.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 8 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=12, duplicate_of=11, non_data_endpoint=3, out_of_scope=59,
  requires_elevated_scope=5.
