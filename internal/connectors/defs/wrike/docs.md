# Overview

Reads Wrike tasks, folders, and contacts through the Wrike REST API. Read-only.

Readable streams: `tasks`, `folders`, `contacts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.wrike.com/api/v4/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Wrike API access token, sent as a Bearer token on every
  request. Never logged.
- `base_url` (optional, string); default `https://www.wrike.com/api/v4`; format `uri`; Wrike API
  base URL. Defaults to the production endpoint; override for test proxies.
- `max_pages` (optional, string); default `1`.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://www.wrike.com/api/v4`, `max_pages=1`,
`page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/tasks` with query `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

- `tasks`: GET `/tasks` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `folders`: GET `/folders` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `contacts`: GET `/contacts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Wrike API read of task, folder, and contact
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
