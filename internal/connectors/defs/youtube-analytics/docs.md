# Overview

Reads YouTube Reporting API jobs, report types, and generated reports via the Google OAuth 2.0
refresh-token grant.

Readable streams: `jobs`, `report_types`, `reports`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/youtube/reporting/v1/reports.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://youtubereporting.googleapis.com/v1`; format `uri`;
  YouTube Reporting API base URL override for tests or proxies.
- `client_id` (required, secret, string); Google OAuth 2.0 client ID for the refresh-token grant.
  Used only in the token-request form; never logged.
- `client_secret` (optional, secret, string); Google OAuth 2.0 client secret (optional for some
  client types). Used only in the token-request form; never logged.
- `content_owner_id` (optional, string); Optional content-owner ID; sent as the
  onBehalfOfContentOwner query parameter for content-owner-scoped accounts.
- `job_id` (optional, string); Reporting job ID the 'reports' stream is scoped to (required only
  when reading the reports stream).
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `refresh_token` (required, secret, string); Long-lived Google OAuth 2.0 refresh token. Exchanged
  for a short-lived access token at token_url; never logged. The 3-legged consent/acquisition dance
  is out of scope for this connector (credentials layer already owns it).
- `scopes` (optional, string); default `https://www.googleapis.com/auth/yt-analytics.readonly`;
  OAuth scope requested on the token-refresh grant.
- `token_url` (optional, string); default `https://oauth2.googleapis.com/token`; format `uri`;
  Google OAuth 2.0 token endpoint override. MUST be https in production; the hook fails closed on a
  non-https or unparseable value to prevent exfiltrating the refresh token/client secret to an
  attacker-chosen endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://youtubereporting.googleapis.com/v1`, `max_pages=0`,
`page_size=100`, `scopes=https://www.googleapis.com/auth/yt-analytics.readonly`,
`token_url=https://oauth2.googleapis.com/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`, `config.scopes`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/reportTypes`.

## Streams notes

All streams are read in full-refresh mode only; incremental sync is not available.

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

- `jobs`: GET `/jobs` - records path `jobs`; query `onBehalfOfContentOwner` from template `{{
  config.content_owner_id }}`, omitted when absent; `pageSize`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `pageToken`; next token from `nextPageToken`; computed output fields
  `create_time`, `expire_time`, `report_type_id`, `system_managed`.
- `report_types`: GET `/reportTypes` - records path `reportTypes`; query `onBehalfOfContentOwner`
  from template `{{ config.content_owner_id }}`, omitted when absent; `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; computed output fields `deprecate_time`, `system_managed`.
- `reports`: GET `/jobs/{{ config.job_id }}/reports` - records path `reports`; query
  `onBehalfOfContentOwner` from template `{{ config.content_owner_id }}`, omitted when absent;
  `pageSize`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; computed output fields `create_time`, `download_url`, `end_time`,
  `job_expire_time`, `job_id`, `start_time`.

## Write actions & risks

This connector is read-only. Read behavior: external YouTube Reporting API read of
reporting-job/report-type/report metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, out_of_scope=2.
