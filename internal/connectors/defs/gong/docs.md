# Overview

Reads Gong users, calls, and scorecards through the Gong REST API (read-only).

Readable streams: `users`, `calls`, `scorecards`.

This connector is read-only; no write actions are declared.

Service API documentation: https://us-66463.app.gong.io/settings/api/documentation.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Gong generated access key. Used for HTTP Basic auth;
  never logged.
- `access_key_secret` (required, secret, string); Gong generated access key secret. Used for HTTP
  Basic auth; never logged.
- `base_url` (optional, string); default `https://api.gong.io/v2`; format `uri`; Gong API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects
  created/started at or after this time are read.

Secret fields are redacted in logs and write previews: `access_key`, `access_key_secret`.

Default configuration values: `base_url=https://api.gong.io/v2`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.access_key`, `secrets.access_key_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `records.cursor`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `users`: GET `/users` - records path `users`; query `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `cursor`; next token from `records.cursor`; incremental cursor
  `created`; sent as `fromDateTime`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `email_address`, `first_name`, `last_name`, `manager_id`, `phone_number`.
- `calls`: GET `/calls` - records path `calls`; query `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `cursor`; next token from `records.cursor`; incremental cursor
  `started`; sent as `fromDateTime`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `is_private`.
- `scorecards`: GET `/settings/scorecards` - records path `scorecards`; query `limit`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `cursor`; next token from
  `records.cursor`; incremental cursor `updated`; sent as `fromDateTime`; formatted as `rfc3339`;
  initial lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Gong API read of call, user, and scorecard
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
