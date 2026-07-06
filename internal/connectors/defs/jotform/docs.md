# Overview

Reads Jotform forms, submissions, reports, folders, and the account profile through the Jotform REST
API.

Readable streams: `forms`, `submissions`, `reports`, `folders`, `user`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.jotform.com/docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Jotform API key, sent as the APIKEY request header. Never
  logged.
- `base_url` (optional, string); default `https://api.jotform.com`; format `uri`; Jotform API base
  URL override for tests, proxies, or a regional site (e.g. https://eu-api.jotform.com,
  https://hipaa-api.jotform.com).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.jotform.com`.

Authentication behavior:

- API key authentication in `APIKEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `reports`, `folders`, `user`; offset_limit: `forms`, `submissions`.

- `forms`: GET `/user/forms` - records path `content`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `submissions`: GET `/user/submissions` - records path `content`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `reports`: GET `/user/reports` - records path `content`.
- `folders`: GET `/user/folders` - records path `content`.
- `user`: GET `/user` - records path `content`.

## Write actions & risks

This connector is read-only. Read behavior: external Jotform API read of form, submission, report,
and folder data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=2.
