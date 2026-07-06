# Overview

Reads Plausible Analytics sites and stats reports through the Stats API.

Readable streams: `sites`, `aggregate`, `timeseries`, `breakdown`.

This connector is read-only; no write actions are declared.

Service API documentation: https://plausible.io/docs/stats-api.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Plausible Analytics API token, sent as a Bearer token.
  Never logged.
- `base_url` (optional, string); default `https://plausible.io/api/v1`; format `uri`; Plausible API
  base URL override for self-hosted instances, tests, or proxies.
- `compare` (optional, string); Optional Plausible comparison period for the stats streams.
- `date` (optional, string); Optional date (or date range, for period=custom) for the stats streams.
- `filters` (optional, string); Optional Plausible filter expression for the stats streams.
- `metrics` (optional, string); Optional comma-separated metrics override for the stats streams
  (e.g. visitors,pageviews).
- `mode` (optional, string).
- `period` (optional, string); default `30d`; Plausible reporting period (e.g. 30d, 7d, month,
  custom) for the stats streams.
- `property` (optional, string); default `event:page`; Breakdown dimension for the breakdown stream
  (e.g. event:page, visit:source, visit:country).
- `site_id` (optional, string); Plausible site domain (site id); required for the aggregate,
  timeseries, and breakdown stats streams.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://plausible.io/api/v1`, `period=30d`,
`property=event:page`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sites`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `sites`, `aggregate`, `timeseries`; page_number: `breakdown`.

- `sites`: GET `/sites` - records path `sites`; computed output fields `site_id`.
- `aggregate`: GET `/stats/aggregate` - records path `results`; query `compare` from template `{{
  config.compare }}`, omitted when absent; `date` from template `{{ config.date }}`, omitted when
  absent; `filters` from template `{{ config.filters }}`, omitted when absent; `metrics` from
  template `{{ config.metrics }}`, omitted when absent; `period`=`{{ config.period }}`;
  `site_id`=`{{ config.site_id }}`; computed output fields `bounce_rate`, `events`, `pageviews`,
  `site_id`, `visit_duration`, `visitors`, `visits`.
- `timeseries`: GET `/stats/timeseries` - records path `results`; query `compare` from template `{{
  config.compare }}`, omitted when absent; `date` from template `{{ config.date }}`, omitted when
  absent; `filters` from template `{{ config.filters }}`, omitted when absent; `metrics` from
  template `{{ config.metrics }}`, omitted when absent; `period`=`{{ config.period }}`;
  `site_id`=`{{ config.site_id }}`; computed output fields `bounce_rate`, `events`, `pageviews`,
  `site_id`, `visit_duration`, `visitors`, `visits`.
- `breakdown`: GET `/stats/breakdown` - records path `results`; query `compare` from template `{{
  config.compare }}`, omitted when absent; `date` from template `{{ config.date }}`, omitted when
  absent; `filters` from template `{{ config.filters }}`, omitted when absent; `metrics` from
  template `{{ config.metrics }}`, omitted when absent; `period`=`{{ config.period }}`;
  `property`=`{{ config.property }}`; `site_id`=`{{ config.site_id }}`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `bounce_rate`, `events`, `pageviews`, `property_value`, `site_id`, `visit_duration`, `visitors`,
  `visits`.

## Write actions & risks

This connector is read-only. Read behavior: external Plausible Analytics API read of site analytics
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=3.
