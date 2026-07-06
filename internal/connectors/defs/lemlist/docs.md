# Overview

Reads lemlist campaigns, activities, team metadata, CRM contacts/companies, schedules, tasks,
webhooks, unsubscribes, field definitions, and signal-agent data through the lemlist REST API.

Readable streams: `campaigns`, `team`, `team_senders`, `team_credits`, `team_crm_users`,
`activities`, `unsubscribes`, `schedules`, `database_filters`, `tasks`, `inbox_labels`, `contacts`,
`contact_lists`, `companies`, `webhooks`, `unsubscribed_variables`, `watchlist_signals`,
`user_channels`, `fields_contact`, `fields_company`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.lemlist.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Lemlist API key. Sent as the access_token query parameter on
  every request; never logged.
- `base_url` (optional, string); default `https://api.lemlist.com/api`; format `uri`; Lemlist API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.lemlist.com/api`.

Authentication behavior:

- API key authentication in query parameter `access_token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/team`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `team`, `team_senders`, `team_credits`, `team_crm_users`,
`database_filters`, `inbox_labels`, `contact_lists`, `webhooks`, `user_channels`, `fields_contact`,
`fields_company`; offset_limit: `campaigns`, `activities`, `unsubscribes`, `schedules`, `contacts`,
`companies`, `unsubscribed_variables`, `watchlist_signals`; page_number: `tasks`.

- `campaigns`: GET `/campaigns` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `team`: GET `/team` - single-object response; records at response root.
- `team_senders`: GET `/team/senders` - records at response root.
- `team_credits`: GET `/team/credits` - single-object response; records at response root.
- `team_crm_users`: GET `/team/crmUsers` - records at response root.
- `activities`: GET `/activities` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `unsubscribes`: GET `/unsubscribes` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `schedules`: GET `/schedules` - records path `schedules`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `database_filters`: GET `/database/filters` - records at response root.
- `tasks`: GET `/tasks` - records path `results`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 0; page size 100.
- `inbox_labels`: GET `/inbox/labels` - records at response root.
- `contacts`: GET `/contacts` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `contact_lists`: GET `/contacts/lists` - records at response root.
- `companies`: GET `/companies` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `webhooks`: GET `/hooks` - records at response root.
- `unsubscribed_variables`: GET `/v2/unsubscribes/variables` - records at response root;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `watchlist_signals`: GET `/watchlist/signals` - records path `signals`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `user_channels`: GET `/user/channels` - single-object response; records at response root.
- `fields_contact`: GET `/fields` - records path `data.contact`.
- `fields_company`: GET `/fields` - records path `data.company`.

## Write actions & risks

This connector is read-only. Read behavior: external lemlist API read of campaign, outreach, CRM,
inbox metadata, unsubscribe, webhook, and signal-agent data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 20 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=4, non_data_endpoint=1, out_of_scope=4,
  requires_elevated_scope=10.
