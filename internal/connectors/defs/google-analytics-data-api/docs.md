# Overview

Reads Google Analytics 4 reports (active users, traffic sources, devices, pages) from the Analytics
Data API runReport endpoint. Read-only.

Readable streams: `daily_active_users`, `website_overview`, `traffic_sources`, `devices`, `pages`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developers.google.com/analytics/devguides/reporting/data/v1/changelog.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `convert_conversions_event` (optional, string); Enables conversion of `conversions:*` event
  metrics from integers to floats. This is beneficial for preventing data rounding when the API
  returns float values for any `conversions:*` fields.
- `credentials` (optional, string); Credentials for the service.
- `custom_reports_array` (optional, string); You can add your Custom Analytics report by creating
  one.
- `date_ranges_end_date` (optional, string); The end date from which to replicate report data in the
  format YYYY-MM-DD. Data generated after this date will not be included in the report. Not applied
  to custom Cohort reports. When no date is provided or the date is in the future, the date from
  today is used.
- `date_ranges_start_date` (optional, string); The start date from which to replicate report data in
  the format YYYY-MM-DD. Data generated before this date will not be included in the report. Not
  applied to custom Cohort reports.
- `keep_empty_rows` (optional, string); If false, each row with all metrics equal to 0 will not be
  returned. If true, these rows will be returned if they are not separately removed by a filter.
  More information is available in <a
  href="https://developers.google.com/analytics/devguides/reporting/data/v1/rest/v1beta/properties/runReport#request-body">the
  documentation</a>.
- `lookback_window` (optional, string); Since attribution changes after the event date, and Google
  Analytics has a data processing latency, we should specify how many days in the past we should
  refresh the data in every run. So if you set it at 5 days, in every sync it will fetch the last
  bookmark date minus 5 days.
- `mode` (optional, string).
- `property_ids` (required, string); A list of your Property IDs. The Property ID is a unique number
  assigned to each property in Google Analytics, found in your GA4 property URL. This ID allows the
  connector to track the specific events associated with your property. Refer to the <a
  href='https://developers.google.com/analytics/devguides/reporting/data/v1/property-id#what_is_my_property_id'>Google
  Analytics documentation</a> to locate your property ID.
- `subscription_tier` (optional, string); Quota tier of the Google Analytics 4 properties being
  queried. Determines the per-property rate-limit policy applied locally once the tier-aware
  rate-limit budget is activated. Select "Analytics 360 Property" only if all configured property
  IDs belong to an Analytics 360 subscription. See
  https://developers.google.com/analytics/devguides/reporting/data/v1/quotas.
- `window_in_days` (optional, string).

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `daily_active_users`: GET connector-managed request path - records path `data`; incremental cursor
  `date`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `website_overview`: GET connector-managed request path - records path `data`; incremental cursor
  `date`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `traffic_sources`: GET connector-managed request path - records path `data`; incremental cursor
  `date`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `devices`: GET connector-managed request path - records path `data`; incremental cursor `date`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `pages`: GET connector-managed request path - records path `data`; incremental cursor `date`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `daily_active_users`, `website_overview`,
  `traffic_sources`, `devices`, `pages`.
