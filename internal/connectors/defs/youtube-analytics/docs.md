# Overview

Reads YouTube Reporting API jobs, report types, generated report metadata, YouTube
Analytics groups, and group items via the Google OAuth 2.0 refresh-token grant. The
connector also declares typed reverse-ETL write actions for documented YouTube Reporting job
creation/deletion and YouTube Analytics group/group-item creation, update, and deletion.

Readable streams: `jobs`, `job`, `report_types`, `reports`, `report`, `groups`, `group_items`.

Write actions: `create_job`, `delete_job`, `create_group`, `update_group`, `delete_group`,
`create_group_item`, `delete_group_item`.

Service API documentation:

- YouTube Analytics API v2: https://developers.google.com/youtube/analytics
- YouTube Reporting API v1: https://developers.google.com/youtube/reporting/v1/reports

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://youtubereporting.googleapis.com/v1`; format
  `uri`; YouTube Reporting API base URL override for tests or proxies.
- `client_id` (required, secret, string); Google OAuth 2.0 client ID for the refresh-token grant.
  Used only in the token-request form; never logged.
- `client_secret` (optional, secret, string); Google OAuth 2.0 client secret (optional for some
  client types). Used only in the token-request form; never logged.
- `content_owner_id` (optional, string); sent as the `onBehalfOfContentOwner` query parameter for
  content-owner-scoped read streams. Declarative write actions currently omit this optional query
  parameter until write-action query templates are supported.
- `created_after` (optional, string); Reporting `reports.list` `createdAfter` filter.
- `include_system_managed` (optional, string); `true`/`false` for Reporting jobs and report types.
- `job_id` (optional, string); required for `job`, `reports`, and `report` streams.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  paginated streams.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page for Reporting list operations.
- `refresh_token` (required, secret, string); Long-lived Google OAuth 2.0 refresh token. Exchanged
  for a short-lived access token at `token_url`; never logged.
- `report_id` (optional, string); required for the single `report` stream.
- `scopes` (optional, string); default `https://www.googleapis.com/auth/yt-analytics.readonly`;
  OAuth scope requested on the token-refresh grant.
- `start_time_at_or_after` (optional, string); Reporting `reports.list` start-time lower bound.
- `start_time_before` (optional, string); Reporting `reports.list` start-time upper bound.
- `token_url` (optional, string); default `https://oauth2.googleapis.com/token`; format `uri`;
  Google OAuth 2.0 token endpoint override. MUST be https in production.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`, and `config.scopes`.

Connection checks call GET `/reportTypes` on the Reporting API base URL.

## Streams notes

All streams are read in full-refresh mode only; incremental sync is not available.

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`. Single-resource streams and `group_items` use no pagination.

- `jobs`: GET `/jobs`; records path `jobs`; query `pageSize`, optional `includeSystemManaged`,
  optional `onBehalfOfContentOwner`; computed fields `report_type_id`, `create_time`,
  `expire_time`, `system_managed`.
- `job`: GET `/jobs/{{ config.job_id }}`; single object; optional `onBehalfOfContentOwner`;
  computed fields match `jobs`.
- `report_types`: GET `/reportTypes`; records path `reportTypes`; query `pageSize`, optional
  `includeSystemManaged`, optional `onBehalfOfContentOwner`; computed fields `deprecate_time`,
  `system_managed`.
- `reports`: GET `/jobs/{{ config.job_id }}/reports`; records path `reports`; query `pageSize`,
  optional `createdAfter`, `startTimeAtOrAfter`, `startTimeBefore`, and
  `onBehalfOfContentOwner`; computed fields `job_id`, `start_time`, `end_time`, `create_time`,
  `job_expire_time`, `download_url`.
- `report`: GET `/jobs/{{ config.job_id }}/reports/{{ config.report_id }}`; single object;
  optional `onBehalfOfContentOwner`; computed fields match `reports`.
- `groups`: GET `https://youtubeanalytics.googleapis.com/v2/groups`; records path `items`; requires
  the closed command selector `mine=true` before request execution and optionally sends
  `onBehalfOfContentOwner`; computed fields `title`, `published_at`, `item_count`, `item_type`.
- `group_items`: GET `https://youtubeanalytics.googleapis.com/v2/groupItems`; records path `items`;
  requires command query `groupId` before request execution and optionally sends
  `onBehalfOfContentOwner`; computed fields `group_id`, `resource_kind`, `resource_id`.

## Write actions & risks

Write actions are available only through reverse ETL plan → preview → explicit approval → execute.
Destructive delete actions also require typed `--confirm destructive` at execution time.

- `create_job`: POST `/jobs`; required `reportTypeId` and `name`.
- `delete_job`: DELETE `/jobs/{{ record.job_id }}`; required `job_id`; redacted preview/error
  field `job_id`; destructive confirmation required.
- `create_group`: POST `https://youtubeanalytics.googleapis.com/v2/groups`; required
  `snippet.title` and `contentDetails.itemType`; item type accepts `youtube#channel`,
  `youtube#playlist`, `youtube#video`, or `youtubePartner#asset`.
- `update_group`: PUT `https://youtubeanalytics.googleapis.com/v2/groups`; required `id` and
  `snippet.title`; group ID redacted in write errors.
- `delete_group`: DELETE
  `https://youtubeanalytics.googleapis.com/v2/groups?id={{ record.id | urlencode }}`; required
  `id`; redacted preview/error field `id`; destructive confirmation required.
- `create_group_item`: POST `https://youtubeanalytics.googleapis.com/v2/groupItems`; required
  `groupId` and `resource.id`; optional `resource.kind` accepts `youtube#channel`,
  `youtube#playlist`, `youtube#video`, or `youtubePartner#asset`; group/resource IDs are redacted in
  write errors.
- `delete_group_item`: DELETE
  `https://youtubeanalytics.googleapis.com/v2/groupItems?id={{ record.id | urlencode }}`; required
  `id`; redacted preview/error field `id`; destructive confirmation required.

The official APIs document optional `onBehalfOfContentOwner` for several writes. The declarative
write dialect does not yet support optional action-level query parameters, so these write actions
omit that optional query rather than hard-coding an empty or required content-owner value.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage accounts for 16 official operations: 15 executable connector surfaces and 1 planned
  operation metadata row.
- YouTube Analytics `reports.query` remains planned solely for the bounded provider-query/direct-read
  foundation tracked by issue #2985.
- Executable YouTube Analytics routes use fixed provider-owned absolute URLs. They do not interpolate
  a caller-controlled host into an authenticated request.
- YouTube Reporting `media.download` is reachable as `pm youtube-analytics reports download`.
  It requires the provider `resourceName` plus an explicit `--dest-root`, may use `--file-name`,
  streams at most 100 MiB, refuses path traversal/overwrite/archive extraction, and records file
  metadata and SHA-256 rather than placing report bytes in JSON output. The source is the Reporting
  guide's [download step](https://developers.google.com/youtube/reporting/v1/reports#step-6-download-the-report).
- Report metadata streams expose `download_url`; it is provider metadata only. The download command
  uses the documented media resource name rather than accepting arbitrary absolute URLs.
