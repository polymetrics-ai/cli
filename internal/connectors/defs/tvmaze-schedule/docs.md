# Overview

Reads public TVmaze broadcast and web schedules without credentials.

Readable streams: `schedule`, `web_schedule`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.tvmaze.com/api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.tvmaze.com`; format `uri`; TVmaze API base URL
  override for tests or proxies.
- `country` (optional, string); ISO 3166-1 country code to scope the schedule to (e.g. US).
  Optional; omitted entirely when unset, matching TVmaze's own default-country behavior.
- `date` (optional, string); format `date`; ISO 8601 date (YYYY-MM-DD) to scope the schedule to.
  Optional; omitted entirely when unset, defaulting to today per TVmaze's own API behavior.

Default configuration values: `base_url=https://api.tvmaze.com`.

Authentication is handled by the connector-specific implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/schedule` with query `country`=`US`.

## Streams notes

Default pagination: single request; no pagination.

- `schedule`: GET `/schedule` - records path `.`; query `country` from template `{{ config.country
  }}`, omitted when absent; `date` from template `{{ config.date }}`, omitted when absent; computed
  output fields `show_id`, `show_name`.
- `web_schedule`: GET `/web/schedule` - records path `.`; query `country` from template `{{
  config.country }}`, omitted when absent; `date` from template `{{ config.date }}`, omitted when
  absent; computed output fields `show_id`, `show_name`.

## Write actions & risks

This connector is read-only. Read behavior: external TVmaze public API read of broadcast/web
schedule data.

## Known limits

- Batch defaults: read_page_size=0.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
