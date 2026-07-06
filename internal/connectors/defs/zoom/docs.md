# Overview

Reads Zoom users, meetings, and webinars through the Zoom REST API.

Readable streams: `users`, `meetings`, `webinars`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.zoom.us/docs/api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoom OAuth access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.zoom.us/v2`; format `uri`; Zoom API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-300); sent as the 'page_size'
  query param.
- `user_id` (optional, string); Zoom user id or email the 'meetings' and 'webinars' streams are
  scoped to (required for those streams; substituted into the user-scoped path).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.zoom.us/v2`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `next_page_token`; next token from
`next_page_token`.

- `users`: GET `/users` - records path `users`; query `page_size`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `next_page_token`; next token from `next_page_token`; computed output
  fields `name`, `updated_at`; emits passthrough records.
- `meetings`: GET `/users/{{ config.user_id }}/meetings` - records path `meetings`; query
  `page_size`=`{{ config.page_size }}`; cursor pagination; cursor parameter `next_page_token`; next
  token from `next_page_token`; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `webinars`: GET `/users/{{ config.user_id }}/webinars` - records path `webinars`; query
  `page_size`=`{{ config.page_size }}`; cursor pagination; cursor parameter `next_page_token`; next
  token from `next_page_token`; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.

## Write actions & risks

This connector is read-only. Read behavior: external Zoom API read of user, meeting, and webinar
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
