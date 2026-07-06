# Overview

Reads Phyllo users, accounts, profiles, social content/comments, audience, and income data, and
writes user/webhook/account-config mutations using Basic-auth REST endpoints.

Readable streams: `users`, `accounts`, `profiles`, `social_contents`, `work_platforms`, `audience`,
`social_content_groups`, `social_comments`, `social_income_transactions`, `social_income_payouts`,
`commerce_income_transactions`, `commerce_income_payouts`, `commerce_income_balances`, `webhooks`.

Write actions: `create_user`, `update_account`, `disconnect_account`, `create_webhook`,
`update_webhook`, `delete_webhook`.

Service API documentation: https://docs.getphyllo.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.getphyllo.com`; format `uri`; Phyllo API base
  URL. Defaults to the production host; set explicitly to https://api.sandbox.getphyllo.com or
  https://api.staging.getphyllo.com to target a non-production environment.
- `client_id` (required, secret, string); Phyllo client ID, sent as the HTTP Basic auth username.
  Never logged.
- `client_secret` (required, secret, string); Phyllo client secret, sent as the HTTP Basic auth
  password. Never logged.
- `mode` (optional, string).
- `phyllo_account_id` (optional, string); Optional Phyllo account id filter. When set, scopes the
  profiles/social_contents/audience/social_content_groups/social_comments/social_income_transactions/social_income_payouts/commerce_income_transactions/commerce_income_payouts/commerce_income_balances
  streams to that account; omitted from the request entirely when unset.
- `phyllo_user_id` (optional, string); Optional Phyllo user id filter, scoping the accounts and
  profiles streams to that user; omitted from the request entirely when unset.
- `phyllo_work_platform_id` (optional, string); Optional Phyllo work-platform id filter, scoping the
  accounts and profiles streams to that platform (e.g. YouTube, Instagram); omitted from the request
  entirely when unset.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.getphyllo.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/users`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 50.

- `users`: GET `/v1/users` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `accounts`: GET `/v1/accounts` - records path `data`; query `user_id` from template `{{
  config.phyllo_user_id }}`, omitted when absent; `work_platform_id` from template `{{
  config.phyllo_work_platform_id }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `profiles`: GET `/v1/profiles` - records path `data`; query `account_id` from template `{{
  config.phyllo_account_id }}`, omitted when absent; `user_id` from template `{{
  config.phyllo_user_id }}`, omitted when absent; `work_platform_id` from template `{{
  config.phyllo_work_platform_id }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `social_contents`: GET `/v1/social/contents` - records path `data`; query `account_id` from
  template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `work_platforms`: GET `/v1/work-platforms` - records path `data`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `audience`: GET `/v1/audience` - records path `data`; query `account_id` from template `{{
  config.phyllo_account_id }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `social_content_groups`: GET `/v1/social/content-groups` - records path `data`; query `account_id`
  from template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `social_comments`: GET `/v1/social/comments` - records path `data`; query `account_id` from
  template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `social_income_transactions`: GET `/v1/social/income/transactions` - records path `data`; query
  `account_id` from template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50; emits passthrough
  records.
- `social_income_payouts`: GET `/v1/social/income/payouts` - records path `data`; query `account_id`
  from template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `commerce_income_transactions`: GET `/v1/commerce/income/transactions` - records path `data`;
  query `account_id` from template `{{ config.phyllo_account_id }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 50; emits
  passthrough records.
- `commerce_income_payouts`: GET `/v1/commerce/income/payouts` - records path `data`; query
  `account_id` from template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50; emits passthrough
  records.
- `commerce_income_balances`: GET `/v1/commerce/income/balances` - records path `data`; query
  `account_id` from template `{{ config.phyllo_account_id }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50; emits passthrough
  records.
- `webhooks`: GET `/v1/webhooks` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; emits passthrough records.

## Write actions & risks

Overall write risk: creates Phyllo users and webhooks, updates account monitoring configuration and
webhook subscriptions, and disconnects linked creator accounts.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/v1/users` - kind `create`; body type `json`; required record fields `name`,
  `external_id`; accepted fields `external_id`, `name`; risk: creates a new Phyllo end-user record
  that every subsequent Connect/account/profile flow is anchored to; low-risk external mutation, no
  destructive side effect, no approval required.
- `update_account`: PATCH `/v1/accounts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `data`; required record fields `id`, `data`; accepted fields `data`,
  `id`; risk: changes an account's identity/engagement/income monitoring configuration (e.g.
  STANDARD vs EXTENSIVE data collection level), affecting what data Phyllo collects going forward;
  external mutation, approval required.
- `disconnect_account`: POST `/v1/accounts/{{ record.id }}/disconnect` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; confirmation
  `destructive`; risk: revokes Phyllo's connection to the creator's linked social/creator platform
  account, permanently stopping all future data collection for it; destructive external mutation,
  approval required.
- `create_webhook`: POST `/v1/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `events`; accepted fields `events`, `url`; risk: registers a new webhook endpoint that will
  receive Phyllo event notifications; low-risk external mutation, no approval required.
- `update_webhook`: PUT `/v1/webhooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `url`, `events`; accepted fields `events`, `id`, `url`;
  risk: changes an existing webhook's target URL and/or subscribed event set, redirecting future
  event delivery; external mutation, approval required.
- `delete_webhook`: DELETE `/v1/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: permanently removes a webhook subscription,
  stopping all future event delivery to it; destructive external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 14 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=13, non_data_endpoint=2, out_of_scope=22.
