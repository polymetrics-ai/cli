# Overview

Reads PersistIQ leads, users, campaigns, mailboxes, activities, accounts, DNC domains, events, lead
fields/statuses, tags, webhook plugin settings, and per-campaign leads/replies, and creates/updates
leads and campaigns, adds/removes campaign leads, replies to campaign messages, and adds DNC
domains, through v1 REST endpoints.

Readable streams: `leads`, `users`, `campaigns`, `mailboxes`, `activities`, `accounts`,
`dnc_domains`, `events`, `lead_fields`, `lead_statuses`, `tags`, `webhook_plugin`, `campaign_leads`,
`campaign_replies`.

Write actions: `update_lead`, `create_campaign`, `duplicate_campaign`, `add_lead_to_campaign`,
`remove_lead_from_campaign`, `reply_to_campaign_message`, `add_dnc_domain`.

Service API documentation: https://persistiq.com/api-docs/index.html.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); PersistIQ API key, sent as the X-API-KEY header. Never
  logged.
- `base_url` (optional, string); default `https://api.persistiq.com`; format `uri`; PersistIQ API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.persistiq.com`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/users`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `webhook_plugin`; page_number: `leads`, `users`, `campaigns`,
`mailboxes`, `activities`, `accounts`, `dnc_domains`, `events`, `lead_fields`, `lead_statuses`,
`tags`, `campaign_leads`, `campaign_replies`.

- `leads`: GET `/v1/leads` - records path `leads`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `users`: GET `/v1/users` - records path `users`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `campaigns`: GET `/v1/campaigns` - records path `campaigns`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `mailboxes`: GET `/v1/mailboxes` - records path `mailboxes`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `activities`: GET `/v1/activities` - records path `activities`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `accounts`: GET `/v1/accounts` - records path `accounts`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `dnc_domains`: GET `/v1/dnc_domains` - records path `dnc_domains`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `events`: GET `/v1/events` - records path `events`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `lead_fields`: GET `/v1/lead_fields` - records path `lead_fields`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `lead_statuses`: GET `/v1/lead_statuses` - records path `lead_statuses`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `tags`: GET `/v1/tags` - records path `tags`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `webhook_plugin`: GET `/v1/webhook_plugin` - single-object response; records path `.`; emits
  passthrough records.
- `campaign_leads`: GET `/v1/campaigns/{{ fanout.id }}/leads` - records path `campaign_leads`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; fan-out; ids from request `/v1/campaigns`; id-list records path `campaigns`; id field `id`;
  id inserted into the request path; stamps `campaign_id`; emits passthrough records.
- `campaign_replies`: GET `/v1/campaigns/{{ fanout.id }}/replies` - records path `replies`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; fan-out; ids from request `/v1/campaigns`; id-list records path `campaigns`; id field `id`;
  id inserted into the request path; stamps `campaign_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of PersistIQ leads and campaigns: update_lead can move a lead
into or out of active outbound-sequence automation; create_campaign/duplicate_campaign create new
live campaigns; add_lead_to_campaign enrolls a lead into automated outreach immediately depending on
campaign state; remove_lead_from_campaign stops scheduled outreach to a lead;
reply_to_campaign_message sends a real outbound email on behalf of the mailbox owner; add_dnc_domain
blocks future outreach to a domain account-wide.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_lead`: PATCH `/v1/leads/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `bounced`, `data`, `id`, `optedout`,
  `owner_id`, `status`, `status_id`; risk: external mutation of an existing PersistIQ lead's fields;
  changing status/status_id/owner_id can move a lead into or out of active outbound-sequence
  automation depending on the target account's own campaign rules; approval required.
- `create_campaign`: POST `/v1/campaigns` - kind `create`; body type `json`; required record fields
  `campaign_name`, `owner_id`; accepted fields `campaign_name`, `owner_id`; risk: creates a new
  outbound-email campaign in the target PersistIQ account; approval required.
- `duplicate_campaign`: POST `/v1/campaigns/duplicate` - kind `create`; body type `json`; required
  record fields `campaign_id`, `owner_id`; accepted fields `campaign_id`, `name`, `owner_id`; risk:
  duplicates an existing campaign (including its steps/sequence) into a new campaign in the target
  account; approval required.
- `add_lead_to_campaign`: POST `/v1/campaigns/{{ record.campaign_id }}/leads` - kind `create`; body
  type `json`; path fields `campaign_id`; required record fields `campaign_id`; accepted fields
  `campaign_id`, `lead_id`, `leads`, `mailbox_id`, `override_lead_limit`, `skip_if_exists`; risk:
  enrolls a lead into a live outbound-email campaign; the lead may start receiving automated
  outreach immediately depending on campaign schedule/state; approval required.
- `remove_lead_from_campaign`: DELETE `/v1/campaigns/{{ record.campaign_id }}/leads/{{ record.id }}`
  - kind `delete`; body type `none`; path fields `campaign_id`, `id`; required record fields
  `campaign_id`, `id`; accepted fields `campaign_id`, `id`; missing records treated as success for
  status `404`; risk: removes a lead from a live outbound-email campaign, stopping any further
  scheduled automated outreach to it in that sequence; approval required.
- `reply_to_campaign_message`: POST `/v1/campaigns/{{ record.campaign_id }}/replies` - kind
  `create`; body type `json`; path fields `campaign_id`; required record fields `campaign_id`,
  `inbox_message_id`, `body`; accepted fields `body`, `campaign_id`, `inbox_message_id`; risk: sends
  a real outbound email reply on behalf of the campaign's mailbox owner; irreversible once
  delivered; approval required.
- `add_dnc_domain`: POST `/v1/dnc_domains` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: adds a domain to the account's Do-Not-Contact list; blocks
  future outreach to that domain account-wide; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 14 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=2.
