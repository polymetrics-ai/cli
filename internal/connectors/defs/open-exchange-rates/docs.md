# Overview

Reads Open Exchange Rates account usage/plan status through the Open Exchange Rates JSON API
(read-only).

Readable streams: `usage`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.openexchangerates.org/.

## Auth setup

Connection fields:

- `app_id` (required, secret, string); Open Exchange Rates app_id, sent as the app_id query
  parameter. Never logged.
- `base_url` (optional, string); default `https://openexchangerates.org/api`; format `uri`; Open
  Exchange Rates API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `app_id`.

Default configuration values: `base_url=https://openexchangerates.org/api`.

Authentication behavior:

- API key authentication in query parameter `app_id` using `secrets.app_id`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/usage.json`.

## Streams notes

Default pagination: single request; no pagination.

- `usage`: GET `/usage.json` - single-object response; records path `data`; computed output fields
  `app_id`, `daily_average`, `days_elapsed`, `days_remaining`, `plan`, `requests`, `requests_quota`,
  `requests_remaining`, `status`.

## Write actions & risks

This connector is read-only. Read behavior: external Open Exchange Rates API read of account
usage/plan status.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
