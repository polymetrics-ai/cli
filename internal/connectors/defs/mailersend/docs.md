# Overview

Reads MailerSend email activity, analytics, domains, messages, recipients, templates, scheduled
messages, sender identities, inbound routes, users, invites, tokens, and webhooks through the
MailerSend REST API.

Readable streams: `activity`, `domains`, `messages`, `recipients`, `templates`,
`scheduled_messages`, `sender_identities`, `inbound_routes`, `account_users`, `invites`, `tokens`,
`webhooks`, `analytics_by_date`, `analytics_country`, `analytics_user_agents`,
`analytics_reading_environment`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.mailersend.com/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); MailerSend API token. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://api.mailersend.com/v1`; format `uri`; MailerSend
  API base URL override for tests or proxies.
- `date_from` (optional, string); Unix-seconds lower bound for the activity stream's required date
  window. Falls back to start_date when unset.
- `date_to` (optional, string); Unix-seconds upper bound for the activity stream's required date
  window.
- `domain_id` (optional, string); MailerSend domain id. Required by the activity stream
  (path-scoped); optionally filters the messages stream when set.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; Fallback for date_from when date_from itself
  is unset (activity stream only).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.mailersend.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/domains` with query `limit`=`10`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 25.

Pagination by stream: none: `analytics_by_date`, `analytics_country`, `analytics_user_agents`,
`analytics_reading_environment`; page_number: `activity`, `domains`, `messages`, `recipients`,
`templates`, `scheduled_messages`, `sender_identities`, `inbound_routes`, `account_users`,
`invites`, `tokens`, `webhooks`.

- `activity`: GET `/activity/{{ config.domain_id }}` - records path `data`; query `date_from`=`{{
  config.date_from }}`; `date_to`=`{{ config.date_to }}`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 25.
- `domains`: GET `/domains` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 25.
- `messages`: GET `/messages` - records path `data`; query `domain_id` from template `{{
  config.domain_id }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 25.
- `recipients`: GET `/recipients` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 25.
- `templates`: GET `/templates` - records path `data`; query `domain_id` from template `{{
  config.domain_id }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 25.
- `scheduled_messages`: GET `/message-schedules` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 25.
- `sender_identities`: GET `/identities` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 25.
- `inbound_routes`: GET `/inbound` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 25.
- `account_users`: GET `/users` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 25.
- `invites`: GET `/invites` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 25.
- `tokens`: GET `/token` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 25.
- `webhooks`: GET `/webhooks` - records path `data`; query `domain_id`=`{{ config.domain_id }}`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 25.
- `analytics_by_date`: GET `/analytics/date` - records path `data.stats`; query `date_from`=`{{
  config.date_from }}`; `date_to`=`{{ config.date_to }}`; `domain_id` from template `{{
  config.domain_id }}`, omitted when absent.
- `analytics_country`: GET `/analytics/country` - records path `data.stats`; query `date_from`=`{{
  config.date_from }}`; `date_to`=`{{ config.date_to }}`; `domain_id` from template `{{
  config.domain_id }}`, omitted when absent.
- `analytics_user_agents`: GET `/analytics/ua-name` - records path `data.stats`; query
  `date_from`=`{{ config.date_from }}`; `date_to`=`{{ config.date_to }}`; `domain_id` from template
  `{{ config.domain_id }}`, omitted when absent.
- `analytics_reading_environment`: GET `/analytics/ua-type` - records path `data.stats`; query
  `date_from`=`{{ config.date_from }}`; `date_to`=`{{ config.date_to }}`; `domain_id` from template
  `{{ config.domain_id }}`, omitted when absent.

## Write actions & risks

This connector is read-only. Read behavior: external MailerSend API read of email activity,
analytics, domain, message, recipient, template, schedule, identity, inbound-route, account-user,
token, invite, and webhook data.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 16 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=1, non_data_endpoint=1, out_of_scope=2,
  requires_elevated_scope=4.
