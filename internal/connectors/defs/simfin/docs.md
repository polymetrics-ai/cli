# Overview

Reads SimFin company, financial statement, price, share, filing, and database-change data through
the SimFin REST API.

Readable streams: `companies`, `statements`, `markets`, `company_general_compact`,
`company_general_verbose`, `company_statements_compact`, `company_statements_verbose`,
`company_prices_compact`, `company_prices_verbose`, `common_shares_outstanding`,
`weighted_shares_outstanding`, `filings_by_company`, `filings`, `changed_companies`,
`data_change_log`.

This connector is read-only; no write actions are declared.

Service API documentation: https://simfin.readme.io/reference/getting-started-1.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); SimFin API key.
- `as_reported` (optional, string); Optional boolean string for endpoints that accept the asreported
  flag.
- `base_url` (optional, string); default `https://backend.simfin.com`; format `uri`; SimFin API base
  URL override for tests or proxies.
- `company_ids` (optional, string); Optional comma-separated SimFin company IDs for company-scoped
  v3 endpoints.
- `end_date` (optional, string); Optional end date filter for endpoints that accept end/end-date.
- `filing_company_id` (optional, string); Optional SimFin ID for the filings-by-company endpoint.
- `filing_company_ticker` (optional, string); Optional ticker for the filings-by-company endpoint.
- `fiscal_years` (optional, string); Optional comma-separated fiscal years.
- `include_details` (optional, string); Optional boolean string for the compact statements details
  flag.
- `include_ratios` (optional, string); Optional boolean string for price endpoints that can include
  ratios and derived metrics.
- `include_ttm` (optional, string); Optional boolean string for endpoints that accept the ttm flag.
- `periods` (optional, string); Optional comma-separated fiscal periods such as Q1,Q2,Q3,Q4,FY.
- `start_date` (optional, string); Optional start date filter for endpoints that accept
  start/start-date.
- `statements` (optional, string); Optional comma-separated statement set for statement endpoints,
  for example PL,BS,CF,DERIVED. Defaults to all documented statement families when absent.
- `tickers` (optional, string); Optional comma-separated ticker symbols for company-scoped v3
  endpoints.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://backend.simfin.com`.

Authentication behavior:

- API key authentication in query parameter `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v3/companies/list` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `company_general_compact`, `company_general_verbose`,
`company_statements_compact`, `company_statements_verbose`, `company_prices_compact`,
`company_prices_verbose`, `common_shares_outstanding`, `weighted_shares_outstanding`,
`filings_by_company`, `changed_companies`; page_number: `companies`, `statements`, `markets`,
`filings`, `data_change_log`.

- `companies`: GET `/api/v3/companies/list` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`,
  `name`, `updated_at`.
- `statements`: GET `/api/v3/statements/list` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`,
  `name`, `updated_at`.
- `markets`: GET `/api/v3/markets/list` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`,
  `name`, `updated_at`.
- `company_general_compact`: GET `/api/v3/companies/general/compact` - records path `.`; query `id`
  from template `{{ config.company_ids }}`, omitted when absent; `ticker` from template `{{
  config.tickers }}`, omitted when absent; emits passthrough records.
- `company_general_verbose`: GET `/api/v3/companies/general/verbose` - records path `.`; query `id`
  from template `{{ config.company_ids }}`, omitted when absent; `ticker` from template `{{
  config.tickers }}`, omitted when absent; emits passthrough records.
- `company_statements_compact`: GET `/api/v3/companies/statements/compact` - records path `.`; query
  `asreported` from template `{{ config.as_reported }}`, omitted when absent; `details` from
  template `{{ config.include_details }}`, omitted when absent; `end` from template `{{
  config.end_date }}`, omitted when absent; `fyear` from template `{{ config.fiscal_years }}`,
  omitted when absent; `id` from template `{{ config.company_ids }}`, omitted when absent; `period`
  from template `{{ config.periods }}`, omitted when absent; `start` from template `{{
  config.start_date }}`, omitted when absent; `statements` from template `{{ config.statements }}`,
  default `PL,BS,CF,DERIVED`; `ticker` from template `{{ config.tickers }}`, omitted when absent;
  `ttm` from template `{{ config.include_ttm }}`, omitted when absent; emits passthrough records.
- `company_statements_verbose`: GET `/api/v3/companies/statements/verbose` - records path `.`; query
  `asreported` from template `{{ config.as_reported }}`, omitted when absent; `end` from template
  `{{ config.end_date }}`, omitted when absent; `fyear` from template `{{ config.fiscal_years }}`,
  omitted when absent; `id` from template `{{ config.company_ids }}`, omitted when absent; `period`
  from template `{{ config.periods }}`, omitted when absent; `start` from template `{{
  config.start_date }}`, omitted when absent; `statements` from template `{{ config.statements }}`,
  default `PL,BS,CF,DERIVED`; `ticker` from template `{{ config.tickers }}`, omitted when absent;
  `ttm` from template `{{ config.include_ttm }}`, omitted when absent; emits passthrough records.
- `company_prices_compact`: GET `/api/v3/companies/prices/compact` - records path `.`; query
  `asreported` from template `{{ config.as_reported }}`, omitted when absent; `end` from template
  `{{ config.end_date }}`, omitted when absent; `id` from template `{{ config.company_ids }}`,
  omitted when absent; `ratios` from template `{{ config.include_ratios }}`, omitted when absent;
  `start` from template `{{ config.start_date }}`, omitted when absent; `ticker` from template `{{
  config.tickers }}`, omitted when absent; emits passthrough records.
- `company_prices_verbose`: GET `/api/v3/companies/prices/verbose` - records path `.`; query
  `asreported` from template `{{ config.as_reported }}`, omitted when absent; `end` from template
  `{{ config.end_date }}`, omitted when absent; `id` from template `{{ config.company_ids }}`,
  omitted when absent; `ratios` from template `{{ config.include_ratios }}`, omitted when absent;
  `start` from template `{{ config.start_date }}`, omitted when absent; `ticker` from template `{{
  config.tickers }}`, omitted when absent; emits passthrough records.
- `common_shares_outstanding`: GET `/api/v3/companies/common-shares-outstanding` - records path `.`;
  query `end` from template `{{ config.end_date }}`, omitted when absent; `id` from template `{{
  config.company_ids }}`, omitted when absent; `start` from template `{{ config.start_date }}`,
  omitted when absent; `ticker` from template `{{ config.tickers }}`, omitted when absent; emits
  passthrough records.
- `weighted_shares_outstanding`: GET `/api/v3/companies/weighted-shares-outstanding` - records path
  `.`; query `end` from template `{{ config.end_date }}`, omitted when absent; `fyear` from template
  `{{ config.fiscal_years }}`, omitted when absent; `id` from template `{{ config.company_ids }}`,
  omitted when absent; `period` from template `{{ config.periods }}`, omitted when absent; `start`
  from template `{{ config.start_date }}`, omitted when absent; `ticker` from template `{{
  config.tickers }}`, omitted when absent; `ttm` from template `{{ config.include_ttm }}`, omitted
  when absent; emits passthrough records.
- `filings_by_company`: GET `/api/v3/filings/by-company` - records path `.`; query `id` from
  template `{{ config.filing_company_id }}`, omitted when absent; `ticker` from template `{{
  config.filing_company_ticker }}`, omitted when absent; emits passthrough records.
- `filings`: GET `/api/v3/filings/list` - records path `data`; query `end-date` from template `{{
  config.end_date }}`, omitted when absent; `start-date` from template `{{ config.start_date }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per-page`;
  starts at 0; page size 100; emits passthrough records.
- `changed_companies`: GET `/api/v3/companies/changed-companies` - records path `.`; query `start`
  from template `{{ config.start_date }}`, omitted when absent; emits passthrough records.
- `data_change_log`: GET `/api/v3/companies/data-change-log` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `itemsPerPage`; starts at 0; page size
  100; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external SimFin API read of company, statement, price,
share, filing, and change-log data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 15 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1.
