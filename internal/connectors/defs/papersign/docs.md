# Overview

Reads PaperSign documents, templates, and recipients through the REST API.

Readable streams: `documents`, `templates`, `recipients`.

This connector is read-only; no write actions are declared.

Service API documentation: https://papersign.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); PaperSign API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.papersign.com/v1`; format `uri`; PaperSign API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.papersign.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/documents` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`pagination.next_cursor`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `documents`: GET `/documents` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `pagination.next_cursor`; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `templates`: GET `/templates` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `pagination.next_cursor`; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `recipients`: GET `/recipients` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `pagination.next_cursor`; incremental cursor
  `created_at`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external PaperSign API read of document, template, and
recipient data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=2.
