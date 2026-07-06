# Overview

Reads Paperform forms and form submissions through the Paperform REST API.

Readable streams: `forms`, `submissions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://paperform.co/help/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Paperform API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.paperform.co/v1`; format `uri`; Paperform API
  base URL override for tests or proxies.
- `form_id` (optional, string); Paperform form id the 'submissions' stream is scoped to (required
  for that stream; substituted into the form-scoped submissions path).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.paperform.co/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `forms`: GET `/forms` - records path `results`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`.
- `submissions`: GET `/forms/{{ config.form_id }}/submissions` - records path `results`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental
  cursor `created_at`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Paperform API read of form and submission data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=2.
