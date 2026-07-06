# Overview

Reads configured datasets from the US Census Bureau's API via a caller-supplied query path and
query-string qualifier, and reads the Bureau's own published dataset catalog.

Readable streams: `query`, `datasets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.census.gov/data/developers/guidance/api-user-guide.html.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); US Census Bureau API key, sent as the 'key' query-string
  parameter on every request. Never logged.
- `base_url` (optional, string); default `https://api.census.gov`; format `uri`; US Census API base
  URL override for tests or proxies.
- `query_params` (required, string); Raw URL-encoded query string appended to the request (e.g.
  'get=NAME,ESTAB&for=us:*').
- `query_path` (required, string); Dataset path segment appended to base_url, e.g. 'data/2019/cbp'.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.census.gov`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.api_key` when `{{ secrets.api_key
  }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `{{ config.query_path }}`.

## Streams notes

Default pagination: single request; no pagination.

- `query`: GET `{{ config.query_path }}` - records path `.`.
- `datasets`: GET `data.json` - records path `dataset`; computed output fields `dataset_path`.

## Write actions & risks

This connector is read-only. Read behavior: external US Census Bureau API read of a
caller-configured dataset endpoint, plus the Bureau's own public dataset catalog (no auth required
for the catalog).

## Known limits

- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
