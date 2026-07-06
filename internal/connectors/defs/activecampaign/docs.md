# Overview

Reads ActiveCampaign contacts, lists, deals, campaigns, tags, automations, custom fields, accounts,
users, deal stages, and deal tasks through the ActiveCampaign v3 REST API.

Readable streams: `contacts`, `lists`, `deals`, `campaigns`, `tags`, `automations`, `fields`,
`accounts`, `users`, `deal_stages`, `deal_tasks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.activecampaign.com/reference/overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); ActiveCampaign API key, sent as the Api-Token request
  header. Never logged.
- `base_url` (required, string); format `uri`; ActiveCampaign API base URL, e.g.
  https://<account>.api-us1.com/api/3.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in `Api-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 20.

Pagination by stream: none: `users`; offset_limit: `contacts`, `lists`, `deals`, `campaigns`,
`tags`, `automations`, `fields`, `accounts`, `deal_stages`, `deal_tasks`.

- `contacts`: GET `/contacts` - records path `contacts`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20.
- `lists`: GET `/lists` - records path `lists`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 20.
- `deals`: GET `/deals` - records path `deals`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 20.
- `campaigns`: GET `/campaigns` - records path `campaigns`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 20.
- `tags`: GET `/tags` - records path `tags`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 20.
- `automations`: GET `/automations` - records path `automations`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 20.
- `fields`: GET `/fields` - records path `fields`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20.
- `accounts`: GET `/accounts` - records path `accounts`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20.
- `users`: GET `/users` - records path `users`.
- `deal_stages`: GET `/dealStages` - records path `dealStages`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 20.
- `deal_tasks`: GET `/dealTasks` - records path `dealTasks`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 20.

## Write actions & risks

This connector is read-only. Read behavior: external ActiveCampaign API read of contacts, lists,
deals, campaigns, tags, automations, custom fields, accounts, users, deal stages, and deal tasks.

## Known limits

- Batch defaults: read_page_size=20.
- API coverage includes 11 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=9, duplicate_of=11, out_of_scope=30.
