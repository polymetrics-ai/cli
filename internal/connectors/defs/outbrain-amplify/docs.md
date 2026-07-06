# Overview

Reads Outbrain Amplify marketers, campaigns, and performance reports via the Outbrain Amplify REST
API. Read-only.

Readable streams: `marketers`, `campaigns`, `performance_reports`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.outbrain.com/home-page/amplify-api/documentation/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Outbrain Amplify API access token. Takes precedence
  over username/password when set; sent as a Bearer token.
- `base_url` (optional, string); default `https://api.outbrain.com/amplify/v0.1`; format `uri`;
  Outbrain Amplify API base URL override for tests or proxies.
- `conversion_count` (optional, string); Definition of conversion count in performance reports (e.g.
  view, click, viewAndClick).
- `end_date` (optional, string); Date in the format YYYY-MM-DD; sent as the end_date report filter.
- `geo_location_breakdown` (optional, string); Granularity used for geo location data in performance
  reports (e.g. country, region, dma).
- `marketer_id` (optional, string); Optional marketer ID.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-500).
- `password` (optional, secret, string); Outbrain Amplify password; used with the username config
  value for HTTP Basic auth when access_token is not set.
- `report_granularity` (optional, string); Granularity used for periodic data in performance reports
  (e.g. daily, weekly, monthly, all).
- `start_date` (optional, string); Date in the format YYYY-MM-DD; sent as the start_date report
  filter. Any data before this date will not be replicated.
- `username` (optional, string); Outbrain Amplify username; used with the password secret for HTTP
  Basic auth when access_token is not set.

Secret fields are redacted in logs and write previews: `access_token`, `password`.

Default configuration values: `base_url=https://api.outbrain.com/amplify/v0.1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- HTTP Basic authentication using `config.username`, `secrets.password` when `{{ secrets.password
  }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/marketers` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `marketers`: GET `/marketers` - records path `marketers`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `campaigns`: GET `/campaigns` - records path `campaigns`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `performance_reports`: GET `/reports` - records path `results`; query `conversion_count` from
  template `{{ config.conversion_count }}`, omitted when absent; `end_date` from template `{{
  config.end_date }}`, omitted when absent; `geo_location_breakdown` from template `{{
  config.geo_location_breakdown }}`, omitted when absent; `report_granularity` from template `{{
  config.report_granularity }}`, omitted when absent; `start_date` from template `{{
  config.start_date }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Outbrain Amplify API read of marketer,
campaign, and performance report data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
