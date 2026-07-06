# Overview

Reads Wasabi account and bucket storage statistics from the Wasabi Stats API.

Readable streams: `bucket_stats`, `account_stats`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.wasabi.com/apidocs/wasabi-stats-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Wasabi Stats API credential. If this value contains exactly
  one ':' separator (a 'username:password'-shaped value), it is sent as HTTP Basic auth with the two
  split parts; otherwise it is sent as 'Authorization: Bearer <api_key>'. Never logged.
- `base_url` (optional, string); default `https://stats.wasabisys.com`; format `uri`; Wasabi Stats
  API base URL override for tests or proxies.
- `start_date` (optional, string); format `date`; Not a true incremental cursor: this connector
  sends it as a fixed filter on every request, never advancing it from persisted state.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://stats.wasabisys.com`.

Authentication behavior:

- Connector-specific authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `v1/stats` with query `start_date` from template `{{ config.start_date
}}`, default .

## Streams notes

Default pagination: single request; no pagination.

- `bucket_stats`: GET `v1/stats` - records path `data`; query `start_date` from template `{{
  config.start_date }}`, omitted when absent; computed output fields `id`; emits passthrough
  records.
- `account_stats`: GET `v1/accounts` - records path `data`; query `start_date` from template `{{
  config.start_date }}`, omitted when absent; computed output fields `id`; emits passthrough
  records.

## Write actions & risks

This connector is read-only. Read behavior: external Wasabi Stats API read of account/bucket storage
usage metrics.

## Known limits

- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=2, requires_elevated_scope=1.
