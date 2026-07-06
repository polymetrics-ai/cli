# Overview

Reads Instagram Business/Creator account profile, media, and stories through the Facebook Graph API.

Readable streams: `users`, `media`, `stories`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.facebook.com/docs/instagram-platform/changelog.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Instagram long-lived access token. Used only for Bearer
  auth; never logged.
- `base_url` (optional, string); default `https://graph.facebook.com/v23.0`; format `uri`; Facebook
  Graph API base URL override for tests or proxies.
- `ig_user_id` (required, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://graph.facebook.com/v23.0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/{{ config.ig_user_id }}`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `media`, `stories`; none: `users`.

- `users`: GET `/{{ config.ig_user_id }}` - single-object response; records at response root; query
  `fields`=`id,username,name,biography,website,profile_picture_url,followers_count,follows_count,media_count`.
- `media`: GET `/{{ config.ig_user_id }}/media` - records path `data`; query
  `fields`=`id,caption,media_type,media_product_type,media_url,permalink,thumbnail_url,timestamp,username,like_count,comments_count`;
  `limit`=`100`; follows a next-page URL from the response body; URL path `paging.next`; next URLs
  stay on the configured API host.
- `stories`: GET `/{{ config.ig_user_id }}/stories` - records path `data`; query
  `fields`=`id,caption,media_type,media_product_type,media_url,permalink,thumbnail_url,timestamp,username`;
  `limit`=`100`; follows a next-page URL from the response body; URL path `paging.next`; next URLs
  stay on the configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external Facebook Graph API read of Instagram
Business/Creator account data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=2.
