# Overview

Reads Awin advertiser transactions, publisher-aggregated performance reports, publisher
relationships, and publisher performance reports, and creates advertiser promotion/voucher offers,
through the Awin Advertiser REST API.

Readable streams: `transactions`, `campaign_performance`, `publishers`, `publisher_performance`,
`creative_performance`.

Write actions: `create_offer`.

Service API documentation: https://wiki.awin.com/index.php/Advertiser_API.

## Auth setup

Connection fields:

- `advertiserId` (required, string); Numeric Awin advertiser account ID; substituted into every
  stream's path.
- `api_key` (required, secret, string); Awin API OAuth2 bearer token. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://api.awin.com`; format `uri`; Awin API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `publisher_id` (optional, string); Optional transactions stream filter: a single publisher ID or
  Awin's documented comma-separated list of publisher IDs, sent verbatim as the publisherId query
  parameter.
- `report_region` (optional, string); Awin region code (e.g. GB, US, DE) required by the
  creative_performance report.
- `report_start_date` (optional, string); Lower bound for the publisher_performance report's date
  window, in Awin's 'YYYY-MM-DD' (date-only) wire format for this endpoint specifically - a
  different format than start_date's timestamp shape, since Awin's /reports/publisher endpoint
  documents date-only startDate/endDate values. Optional for publisher_performance (omitted entirely
  when unset, matching that endpoint's own optional startDate).
- `start_date` (optional, string); Lower bound for the transactions stream's date window, in Awin's
  exact 'YYYY-MM-DDTHH:MM:SS' wire format (no timezone suffix).
- `transaction_status` (optional, string); Optional transactions stream filter: one of Awin's
  documented status values (pending, approved, declined, deleted), sent verbatim as the status query
  parameter.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.awin.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/advertisers/{{ config.advertiserId }}/publishers/`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

Pagination by stream: none: `publisher_performance`, `creative_performance`; page_number:
`transactions`, `campaign_performance`, `publishers`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `transactions`: GET `/advertisers/{{ config.advertiserId }}/transactions/` - records path `.`;
  query `dateType`=`transaction`; `endDate`=`2099-12-31T23:59:59`; `publisherId` from template `{{
  config.publisher_id }}`, omitted when absent; `startDate` from template `{{
  incremental.lower_bound }}`, omitted when absent; `status` from template `{{
  config.transaction_status }}`, omitted when absent; `timezone`=`UTC`; page-number pagination; page
  parameter `page`; size parameter `pageSize`; starts at 1; page size 100; incremental cursor
  `transactionDate`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `campaign_performance`: GET `/advertisers/{{ config.advertiserId }}/reports/aggregated/publisher`
  - records path `.`; page-number pagination; page parameter `page`; size parameter `pageSize`;
  starts at 1; page size 100.
- `publishers`: GET `/advertisers/{{ config.advertiserId }}/publishers/` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1; page size
  100.
- `publisher_performance`: GET `/advertisers/{{ config.advertiserId }}/reports/publisher` - records
  path `.`; query `dateType`=`transaction`; `endDate`=`2099-12-31`; `startDate` from template `{{
  config.report_start_date }}`, omitted when absent; `timezone`=`UTC`.
- `creative_performance`: GET `/advertisers/{{ config.advertiserId }}/reports/creative` - records
  path `.`; query `dateType`=`transaction`; `endDate`=`2099-12-31`; `region` from template `{{
  config.report_region }}`, default `GB`; `startDate` from template `{{ config.report_start_date
  }}`, default `2020-01-01`; `timezone`=`UTC`.

## Write actions & risks

Overall write risk: creates a new promotion or voucher offer in the advertiser's MyOffers system,
immediately visible to publishers; external mutation, approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_offer`: POST `/promotion/advertiser/{{ config.advertiserId }}` - kind `create`; body type
  `json`; required record fields `title`, `description`, `terms`, `type`, `url`, `startDate`,
  `endDate`, `appliesToAllRegions`, `promotionCategories`; accepted fields `appliesToAllRegions`,
  `campaign`, `description`, `endDate`, `endTime`, `promotionCategories`, `regions`, `startDate`,
  `startTime`, `terms`, `timeZone`, `title`, `type`, `url`, `voucherCode`; risk: creates a new
  promotion or voucher offer in the advertiser's MyOffers system, visible to publishers immediately;
  external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
