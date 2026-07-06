# Overview

Reads Mention app metadata, accounts, alerts, mentions, alert tags, alert shares, alert preferences,
and alert tasks from the Mention social listening REST API.

Readable streams: `app_data`, `account_me`, `account`, `alert`, `mention`, `alert_tag`,
`alert_share`, `alert_preferences`, `alert_task`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.mention.com/.

## Auth setup

Connection fields:

- `account_id` (required, string); Mention account id.
- `alert_id` (optional, string); Mention alert id, required only by the 'mention' and 'alert_tag'
  streams (used to build their path: accounts/{account_id}/alerts/{alert_id}/mentions|tags).
- `api_key` (required, secret, string); Mention API key, sent as the raw Authorization header value
  (no Bearer prefix). Never logged.
- `base_url` (optional, string); default `https://api.mention.net/api`; format `uri`; Mention API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.mention.net/api`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts/me`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `alert`, `mention`; none: `app_data`, `account_me`, `account`,
`alert_tag`, `alert_share`, `alert_preferences`, `alert_task`.

- `app_data`: GET `/app/data` - single-object response; records at response root.
- `account_me`: GET `/accounts/me` - records path `account`.
- `account`: GET `/accounts/{{ config.account_id }}` - records path `account`.
- `alert`: GET `/accounts/{{ config.account_id }}/alerts` - records path `alerts`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `_links.more.params.cursor`.
- `mention`: GET `/accounts/{{ config.account_id }}/alerts/{{ config.alert_id }}/mentions` - records
  path `mentions`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token
  from `_links.more.params.cursor`.
- `alert_tag`: GET `/accounts/{{ config.account_id }}/alerts/{{ config.alert_id }}/tags` - records
  path `tags`.
- `alert_share`: GET `/accounts/{{ config.account_id }}/alerts/{{ config.alert_id }}/shares` -
  records path `shares`.
- `alert_preferences`: GET `/accounts/{{ config.account_id }}/alerts/{{ config.alert_id
  }}/preferences` - single-object response; records path `preferences`.
- `alert_task`: GET `/accounts/{{ config.account_id }}/alerts/{{ config.alert_id }}/tasks` - records
  path `tasks`.

## Write actions & risks

This connector is read-only. Read behavior: external Mention API read of app metadata, account,
alert, mention, tag, share, preference, and task data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 9 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, non_data_endpoint=1, out_of_scope=3,
  requires_elevated_scope=4.
