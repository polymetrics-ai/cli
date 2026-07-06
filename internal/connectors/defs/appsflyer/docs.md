# Overview

Reads AppsFlyer raw-data CSV export reports (installs, in-app events) through the AppsFlyer Pull
API. Read-only.

Readable streams: `installs_report`, `in_app_events_report`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.appsflyer.com/hc/reference.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); AppsFlyer Pull API token, sent as a Bearer token
  (Authorization: Bearer <api_token>). Never logged.
- `app_id` (required, string); AppsFlyer app identifier the raw-data export reports are scoped to
  (path-substituted).
- `base_url` (optional, string); default `https://hq1.appsflyer.com`; format `uri`; AppsFlyer API
  base URL override for tests or proxies.
- `end_date` (optional, string); Report end date (YYYY-MM-DD); defaults to start_date when unset.
- `mode` (optional, string).
- `start_date` (optional, string); Report start date (YYYY-MM-DD or a value with a time component;
  only the date portion is sent as the 'from' query param).
- `timezone` (optional, string); Optional AppsFlyer report timezone override.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://hq1.appsflyer.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/raw-data/export/app/{{ config.app_id }}/installs_report/v5`.

## Streams notes

Default pagination: single request; no pagination.

- `installs_report`: GET `/api/raw-data/export/app/{{ config.app_id }}/installs_report/v5` - records
  at response root; emits passthrough records.
- `in_app_events_report`: GET `/api/raw-data/export/app/{{ config.app_id }}/in_app_events_report/v5`
  - records at response root; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external AppsFlyer API read of raw installs/in-app-event
export reports.

## Known limits

- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
