# Overview

Reads Linkrunner mobile attribution campaigns and attributed users from the Linkrunner Data API.

Readable streams: `campaigns`, `attributed_users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.linkrunner.io/sdk-less/api-reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.linkrunner.io/api/v1`; format `uri`;
  Linkrunner Data API base URL override for tests or proxies.
- `channel` (optional, string); Optional campaigns stream channel filter passed through as-is.
- `display_id` (optional, string); Campaign display_id to scope the attributed_users stream.
  Required when reading attributed_users.
- `end_timestamp` (optional, string); Optional attributed_users stream upper-bound timestamp filter.
- `filter` (optional, string); Optional campaigns stream filter passed through as-is.
- `linkrunner-key` (required, secret, string); Linkrunner API key, sent as the linkrunner-key
  request header. Never logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_timestamp` (optional, string); Optional attributed_users stream lower-bound timestamp
  filter.
- `timezone` (optional, string); Optional attributed_users stream timezone for timestamp filters.

Secret fields are redacted in logs and write previews: `linkrunner-key`.

Default configuration values: `base_url=https://api.linkrunner.io/api/v1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `linkrunner-key` using `secrets.linkrunner-key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/campaigns`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `campaigns`: GET `/campaigns` - records path `data.campaigns`; query `channel` from template `{{
  config.channel }}`, omitted when absent; `filter` from template `{{ config.filter }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1;
  page size 100; incremental cursor `update_at`; formatted as `rfc3339`.
- `attributed_users`: GET `/attributed-users` - records path `data.users`; query `display_id`=`{{
  config.display_id }}`; `end_timestamp` from template `{{ config.end_timestamp }}`, omitted when
  absent; `start_timestamp` from template `{{ config.start_timestamp }}`, omitted when absent;
  `timezone` from template `{{ config.timezone }}`, omitted when absent; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental cursor
  `attributed_at`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Linkrunner API read of mobile attribution
campaign and user data.

## Known limits

- Published rate limit metadata: requests_per_minute=60.
- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=2, out_of_scope=6, requires_elevated_scope=1.
