# Overview

Reads JustCall users, call logs, SMS, contacts, and phone numbers through the JustCall REST API.

Readable streams: `users`, `calls`, `sms`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.justcall.io/.

## Auth setup

Connection fields:

- `api_key_2` (required, secret, string); JustCall API key pair (api_key:api_secret). Sent verbatim
  as the Authorization header (no Bearer prefix); never logged.
- `base_url` (optional, string); default `https://api.justcall.io`; format `uri`; JustCall API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only calls/sms at or
  after this time are read on a fresh sync.

Secret fields are redacted in logs and write previews: `api_key_2`.

Default configuration values: `base_url=https://api.justcall.io`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key_2`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2.1/users`.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `users`: GET `/v2.1/users` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 0; page size 100.
- `calls`: GET `/v2.1/calls` - records path `data`; query `from_datetime` from template `{{
  incremental.lower_bound }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 0; page size 100; incremental cursor `call_date`; formatted
  as `rfc3339`; initial lower bound from `start_date`.
- `sms`: GET `/v2.1/texts` - records path `data`; query `from_datetime` from template `{{
  incremental.lower_bound }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 0; page size 100; incremental cursor `sms_date`; formatted as
  `rfc3339`; initial lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external JustCall API read of users, call logs, SMS,
contacts, and phone numbers.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
