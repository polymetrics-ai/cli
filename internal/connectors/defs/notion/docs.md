# Overview

Reads Notion databases, pages, and users through the Notion REST API. Read-only.

Readable streams: `databases`, `pages`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.notion.com/reference/intro.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.notion.com/v1`; format `uri`; Notion API base
  URL override for tests or proxies. Defaults to https://api.notion.com/v1.
- `max_pages` (optional, string); Permissive parse: empty, "all", "unlimited", or any
  non-positive-integer string means unbounded.
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `token` (optional, secret, string); Notion integration token, sent as Bearer auth. Never logged.

Secret fields are redacted in logs and write previews: `token`.

Default configuration values: `base_url=https://api.notion.com/v1`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users`.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `databases`: GET `/search` - records path `results`; incremental cursor `last_edited_time`;
  formatted as `rfc3339`.
- `pages`: GET `/search` - records path `results`; incremental cursor `last_edited_time`; formatted
  as `rfc3339`.
- `users`: GET `/users` - records path `results`.

## Write actions & risks

This connector is read-only. Read behavior: external Notion API read of workspace
databases/pages/users.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
