# Overview

Reads FullStory segments, users, events, and user-scoped sessions; writes server-side user and
custom event data through the FullStory Server API.

Readable streams: `segments`, `users`, `events`, `sessions`.

Write actions: `create_user`, `update_user`, `create_event`.

Service API documentation: https://developer.fullstory.com/reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); FullStory API key. Sent as 'Authorization: Basic <api_key>';
  never logged.
- `base_url` (optional, string); default `https://api.fullstory.com`; format `uri`; FullStory API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `200`; Records per page (1-1000).
- `session_email` (optional, string); Optional user email filter for the sessions stream.
  FullStory's sessions endpoint requires uid and/or email.
- `session_uid` (optional, string); Optional user uid filter for the sessions stream. FullStory's
  sessions endpoint requires uid and/or email.
- `uid` (optional, secret, string); Optional FullStory user id sent as the FS-Uid header. Not
  currently wired into any request (see docs.md Known limits: the engine's header dialect always
  hard-errors on an absent secret rather than omitting the header, so an optional secret-sourced
  header cannot be expressed safely).

Secret fields are redacted in logs and write previews: `api_key`, `uid`.

Default configuration values: `base_url=https://api.fullstory.com`, `max_pages=0`, `page_size=200`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Basic` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/segments/v2`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`next_page_token`.

Pagination by stream: cursor: `segments`, `users`, `events`; none: `sessions`.

- `segments`: GET `/segments/v2` - records path `results`; query `limit`=`200`; cursor pagination;
  cursor parameter `pageToken`; next token from `next_page_token`.
- `users`: GET `/v2/users` - records path `results`; query `limit`=`200`; cursor pagination; cursor
  parameter `pageToken`; next token from `next_page_token`.
- `events`: GET `/v2/events` - records path `results`; query `limit`=`200`; cursor pagination;
  cursor parameter `pageToken`; next token from `next_page_token`.
- `sessions`: GET `/v2/sessions` - records path `results`; query `email` from template `{{
  config.session_email }}`, omitted when absent; `limit`=`200`; `uid` from template `{{
  config.session_uid }}`, omitted when absent.

## Write actions & risks

Overall write risk: creates or updates FullStory server-side user attributes and custom events used
for analytics segmentation.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/v2/users` - kind `upsert`; body type `json`; required record fields `uid`;
  accepted fields `display_name`, `email`, `properties`, `uid`; risk: creates or upserts a FullStory
  user profile and associated custom user properties.
- `update_user`: POST `/v2/users/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `display_name`, `email`, `id`, `properties`;
  risk: updates a FullStory user profile's display fields or custom properties.
- `create_event`: POST `/v2/events` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`, `properties`, `session`, `timestamp`, `user`; risk: creates a
  custom FullStory event that becomes part of analytics/session context.

## Known limits

- Batch defaults: read_page_size=200, write_batch_size=1.
- API coverage includes 4 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, deprecated=2, destructive_admin=4, duplicate_of=7, non_data_endpoint=13,
  out_of_scope=11, requires_elevated_scope=20.
