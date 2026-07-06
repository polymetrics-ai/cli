# Overview

Reads Facebook Page metadata and posts from the Graph API. Read-only.

Readable streams: `page`, `posts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.facebook.com/docs/pages/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Facebook Graph API page access token, sent as a Bearer
  token. Never logged.
- `base_url` (optional, string); default `https://graph.facebook.com/v19.0`; format `uri`; Facebook
  Graph API base URL override for tests or proxies.
- `page_id` (required, string); Facebook Page id whose metadata and posts are read.
- `page_size` (optional, string); default `100`; Number of posts requested per page (Graph API
  'limit' query param). Defaults to 100.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://graph.facebook.com/v19.0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/{{ config.page_id }}` with query `fields`=`id,name`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `posts`; none: `page`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `page`: GET `/{{ config.page_id }}` - single-object response; records at response root; query
  `fields`=`id,name,category,fan_count,link`.
- `posts`: GET `/{{ config.page_id }}/posts` - records path `data`; query
  `fields`=`id,message,created_time,updated_time,permalink_url`; `limit` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `paging.next`; next URLs stay on the configured API host; incremental cursor `updated_time`;
  formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Facebook Graph API read of page metadata and
posts.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
