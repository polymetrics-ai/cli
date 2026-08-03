# Overview

Reads subreddit posts and comments through the Reddit OAuth API listing endpoints.

Readable streams: `posts`, `comments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.reddit.com/dev/api/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Reddit OAuth access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). OAuth token acquisition/refresh is out of scope; the
  caller supplies a valid token. Reddit bearer tokens expire 1 hour after issuance and this
  connector does not refresh them, so any sync expected to run longer than an hour, or any
  recurring/scheduled sync, needs a freshly issued token supplied before each run. Never logged.
- `base_url` (optional, string); default `https://oauth.reddit.com`; format `uri`; Reddit OAuth API
  base URL override for tests or proxies.
- `reddit_username` (required, string); Reddit username of the account that authorized
  access_token, e.g. "my_bot_account" without the /u/ prefix. Sent as the required identity
  component of the User-Agent header Reddit's API rules mandate (`<platform>:<app ID>:<version>
  (by /u/<reddit username>)`); non-conforming User-Agents are drastically rate-limited. Not a
  secret; used for no other purpose.
- `subreddit` (required, string); Subreddit name to read posts/comments from (path-scoped as
  /r/{subreddit}/...).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://oauth.reddit.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.
- Every request sends `User-Agent: go:ai.polymetrics.cli:v1 (by /u/<reddit_username>)`, matching
  Reddit's required `<platform>:<app ID>:<version string> (by /u/<reddit username>)` format.

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
  non_data_endpoint=1, out_of_scope=6.
- The `comments` stream calls `GET /r/{subreddit}/comments`, a route that is live and working but
  is not among Reddit's documented endpoints at https://www.reddit.com/dev/api/; the only
  documented comments listing is the per-article tree `GET [/r/subreddit]/comments/{article}` (see
  `api_surface.json` for the full citation and why this connector does not migrate to it).
- `access_token` is caller-supplied and never refreshed by this connector; Reddit bearer tokens
  expire 1 hour after issuance, so any sync longer than an hour, or any scheduled sync, needs a
  fresh token supplied before each run.
- Reddit enforces 100 queries per minute per OAuth client id; on HTTP 429 check the
  `X-Ratelimit-Used`, `X-Ratelimit-Remaining`, and `X-Ratelimit-Reset` response headers for the
  current budget and reset time.
