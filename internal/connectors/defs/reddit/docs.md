# Overview

Reads subreddit posts and comments through the Reddit OAuth API listing endpoints.

Readable streams: `posts`, `comments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.reddit.com/dev/api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Reddit OAuth access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). OAuth token acquisition/refresh is out of scope; the
  caller supplies a valid token. Never logged.
- `base_url` (optional, string); default `https://oauth.reddit.com`; format `uri`; Reddit OAuth API
  base URL override for tests or proxies.
- `subreddit` (required, string); Subreddit name to read posts/comments from (path-scoped as
  /r/{subreddit}/...).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://oauth.reddit.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/r/{{ config.subreddit }}/new` with query `limit`=`1`; `raw_json`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next token from `data.after`.

- `posts`: GET `/r/{{ config.subreddit }}/new` - records path `data.children`; query `limit`=`100`;
  `raw_json`=`1`; cursor pagination; cursor parameter `after`; next token from `data.after`;
  computed output fields `author`, `created_utc`, `id`, `name`, `permalink`, `subreddit`, `title`.
- `comments`: GET `/r/{{ config.subreddit }}/comments` - records path `data.children`; query
  `limit`=`100`; `raw_json`=`1`; cursor pagination; cursor parameter `after`; next token from
  `data.after`; computed output fields `author`, `body`, `created_utc`, `id`, `name`, `permalink`,
  `subreddit`.

## Write actions & risks

This connector is read-only. Read behavior: external Reddit OAuth API read of public subreddit posts
and comments.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
