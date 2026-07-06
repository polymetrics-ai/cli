# Overview

Reads tweets and their authors matching a search query from the Twitter (X) API v2 recent search
endpoint using an App-only Bearer token.

Readable streams: `tweets`, `authors`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.twitter.com/en/docs/twitter-api.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Twitter (X) API v2 App-only Bearer token. Used only for
  Bearer auth; never logged.
- `base_url` (optional, string); default `https://api.twitter.com/2`; format `uri`; Twitter API base
  URL override for tests or proxies.
- `end_date` (optional, string); format `date-time`; Optional RFC3339 upper bound sent as end_time;
  only tweets created at or before this time are returned.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (10-100, Twitter's max_results
  bounds).
- `query` (required, string); Recent-search query string (Twitter v2 search syntax), e.g.
  "from:example".
- `start_date` (optional, string); format `date-time`; Optional RFC3339 lower bound sent as
  start_time; only tweets created at or after this time are returned.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.twitter.com/2`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/tweets/search/recent` with query `max_results`=`10`; `query`=`{{
config.query }}`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `next_token`; next token from
`meta.next_token`.

- `tweets`: GET `/tweets/search/recent` - records path `data`; query `end_time` from template `{{
  config.end_date }}`, omitted when absent; `expansions`=`author_id`; `max_results`=`{{
  config.page_size }}`; `query`=`{{ config.query }}`; `start_time` from template `{{
  config.start_date }}`, omitted when absent;
  `tweet.fields`=`id,text,author_id,created_at,conversation_id,lang,source,in_reply_to_user_id,possibly_sensitive,public_metrics`;
  `user.fields`=`id,name,username,created_at,description,location,verified,protected,url,public_metrics`;
  cursor pagination; cursor parameter `next_token`; next token from `meta.next_token`.
- `authors`: GET `/tweets/search/recent` - records path `includes.users`; query `end_time` from
  template `{{ config.end_date }}`, omitted when absent; `expansions`=`author_id`; `max_results`=`{{
  config.page_size }}`; `query`=`{{ config.query }}`; `start_time` from template `{{
  config.start_date }}`, omitted when absent;
  `tweet.fields`=`id,text,author_id,created_at,conversation_id,lang,source,in_reply_to_user_id,possibly_sensitive,public_metrics`;
  `user.fields`=`id,name,username,created_at,description,location,verified,protected,url,public_metrics`;
  cursor pagination; cursor parameter `next_token`; next token from `meta.next_token`.

## Write actions & risks

This connector is read-only. Read behavior: external Twitter (X) API read of tweets and author
profiles matching a search query.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4, requires_elevated_scope=1.
