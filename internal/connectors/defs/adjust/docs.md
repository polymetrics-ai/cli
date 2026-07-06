# Overview

Reads Adjust report-service report rows for configured dimensions and metrics. Read-only.

Readable streams: `reports`.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.adjust.com/en/article/reports-endpoint.

## Auth setup

Connection fields:

- `additional_metrics` (optional, string); Optional comma-separated list of additional Adjust
  metrics sent as additional_metrics.
- `api_token` (required, secret, string); Adjust Report Service API token. Used only for Bearer
  auth; never logged.
- `base_url` (optional, string); default `https://automate.adjust.com`; format `uri`; Adjust Report
  Service base URL override for tests or proxies.
- `dimensions` (optional, string); default `country`; Comma-separated list of Adjust report
  dimensions (e.g. country,app). Defaults to country.
- `end_date` (optional, string); format `date`; Report period upper bound (YYYY-MM-DD), sent as the
  end of date_period. Defaults to start_date when unset.
- `metrics` (optional, string); default `installs`; Comma-separated list of Adjust report metrics
  (e.g. installs,clicks,cost). Defaults to installs.
- `mode` (optional, string).
- `start_date` (optional, string); format `date`; Report period lower bound (YYYY-MM-DD), sent as
  the start of date_period.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://automate.adjust.com`, `dimensions=country`,
`metrics=installs`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/reports-service/report` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next_page`.

- `reports`: GET `/reports-service/report` - records path `rows`; query `additional_metrics` from
  template `{{ config.additional_metrics }}`, omitted when absent; `date_period` from template `{{
  config.start_date }}:{{ config.end_date }}`, omitted when absent; `dimensions`=`{{
  config.dimensions }}`; `metrics`=`{{ config.metrics }}`; cursor pagination; cursor parameter
  `page`; next token from `next_page`; computed output fields `app`, `clicks`, `cost`, `country`,
  `date`, `installs`.

## Write actions & risks

This connector is read-only. Read behavior: external Adjust report-service read of configured
dimensions/metrics.

## Known limits

- API coverage includes 1 stream-backed endpoint group(s).
