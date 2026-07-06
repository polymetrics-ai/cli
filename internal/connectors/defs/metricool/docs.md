# Overview

Reads Metricool brand profiles and per-brand Instagram, Facebook, LinkedIn, and TikTok post
analytics through the Metricool REST API.

Readable streams: `brands`, `instagram_posts`, `facebook_posts`, `linkedin_posts`, `tiktok_posts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.metricool.com/en/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://app.metricool.com/api`; format `uri`; Metricool
  API base URL override for tests or proxies.
- `blog_ids` (optional, string); Comma-separated list of Metricool blogId values to fan out over for
  the instagram_posts, facebook_posts, linkedin_posts, and tiktok_posts streams (Metricool's
  analytics API is not paginated; each blog is a separate request). Required for those four streams
  (no auto-discovery fallback); the brands stream is account-wide and unaffected.
- `end_date` (optional, string); Optional upper bound of the analytics date range, same per-stream
  format rules as start_date.
- `start_date` (optional, string); Optional lower bound of the analytics date range.
  instagram_posts/facebook_posts/linkedin_posts (Metricool's /stats/* endpoints) expect YYYYMMDD
  (e.g. 20260101); tiktok_posts (the /v2/analytics endpoint) expects YYYY-MM-DDTHH:MM:SS (e.g.
  2026-01-01T00:00:00). Supply the value already formatted for the stream being read; the engine
  sends it verbatim with no reformatting.
- `user_id` (required, string); Metricool account userId, sent as the userId query parameter on
  every request.
- `user_token` (optional, secret, string); Metricool user token, sent as the X-Mc-Auth header on
  every request. Never logged.

Secret fields are redacted in logs and write previews: `user_token`.

Default configuration values: `base_url=https://app.metricool.com/api`.

Authentication behavior:

- API key authentication in `X-Mc-Auth` using `secrets.user_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/admin/simpleProfiles` with query `userId`=`{{ config.user_id }}`.

## Streams notes

Default pagination: single request; no pagination.

- `brands`: GET `/admin/simpleProfiles` - records path `.`; query `userId`=`{{ config.user_id }}`.
- `instagram_posts`: GET `/stats/instagram/posts` - records path `.`; query `end` from template `{{
  config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `userId`=`{{ config.user_id }}`; fan-out; ids from config field `blog_ids`; id sent
  as query parameter `blogId`; stamps `blogId`.
- `facebook_posts`: GET `/stats/facebook/posts` - records path `.`; query `end` from template `{{
  config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `userId`=`{{ config.user_id }}`; fan-out; ids from config field `blog_ids`; id sent
  as query parameter `blogId`; stamps `blogId`.
- `linkedin_posts`: GET `/stats/linkedin/posts` - records path `.`; query `end` from template `{{
  config.end_date }}`, omitted when absent; `start` from template `{{ config.start_date }}`, omitted
  when absent; `userId`=`{{ config.user_id }}`; fan-out; ids from config field `blog_ids`; id sent
  as query parameter `blogId`; stamps `blogId`.
- `tiktok_posts`: GET `/v2/analytics/posts/tiktok` - records path `data`; query `from` from template
  `{{ config.start_date }}`, omitted when absent; `to` from template `{{ config.end_date }}`,
  omitted when absent; `userId`=`{{ config.user_id }}`; fan-out; ids from config field `blog_ids`;
  id sent as query parameter `blogId`; stamps `blogId`.

## Write actions & risks

This connector is read-only. Read behavior: external Metricool API read of brand-scoped social
analytics for the configured user_id/blog_ids.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
