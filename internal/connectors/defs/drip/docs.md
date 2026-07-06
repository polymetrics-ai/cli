# Overview

Reads Drip subscribers, campaigns, broadcasts, accounts, workflows, forms, tags, and webhooks, and
writes subscriber/tag/broadcast/workflow/event/webhook mutations through the Drip REST API.

Readable streams: `subscribers`, `campaigns`, `broadcasts`, `accounts`, `workflows`, `forms`,
`webhooks`.

Write actions: `create_or_update_subscriber`, `delete_subscriber`, `unsubscribe_subscriber`,
`apply_tag`, `remove_tag`, `record_event`, `create_broadcast`, `update_broadcast`,
`delete_broadcast`, `activate_workflow`, `pause_workflow`, `create_webhook`, `delete_webhook`.

Service API documentation: https://developer.drip.com/.

## Auth setup

Connection fields:

- `account_id` (required, string); Drip account ID; path-scopes the subscribers/campaigns/broadcasts
  streams (the accounts stream itself is account-agnostic).
- `api_key` (required, secret, string); Drip API key. Sent as the HTTP Basic username with a blank
  password; never logged.
- `base_url` (optional, string); default `https://api.getdrip.com/v2`; format `uri`; Drip API base
  URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.getdrip.com/v2`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `accounts`; page_number: `subscribers`, `campaigns`, `broadcasts`,
`workflows`, `forms`, `webhooks`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `subscribers`: GET `/{{ config.account_id }}/subscribers` - records path `subscribers`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; incremental cursor `created_at`; formatted as `rfc3339`.
- `campaigns`: GET `/{{ config.account_id }}/campaigns` - records path `campaigns`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `created_at`; formatted as `rfc3339`.
- `broadcasts`: GET `/{{ config.account_id }}/broadcasts` - records path `broadcasts`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `created_at`; formatted as `rfc3339`.
- `accounts`: GET `/accounts` - records path `accounts`; incremental cursor `created_at`; formatted
  as `rfc3339`.
- `workflows`: GET `/{{ config.account_id }}/workflows` - records path `workflows`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `forms`: GET `/{{ config.account_id }}/forms` - records path `forms`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `webhooks`: GET `/{{ config.account_id }}/webhooks` - records path `webhooks`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external Drip API mutation of subscribers, tags, broadcasts, workflows, custom
events, and webhooks; delete_subscriber and delete_broadcast/delete_webhook are destructive and
require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_or_update_subscriber`: POST `/{{ config.account_id }}/subscribers` - kind `upsert`; body
  type `json`; body fields `subscribers`; required record fields `subscribers`; accepted fields
  `subscribers`; risk: creates a new subscriber or updates an existing one matched by email;
  low-risk external mutation, no approval required. Drip's real API strictly requires the
  one-subscriber-per-write body to still be an ARRAY under the "subscribers" key ({"subscribers":
  [{...}]}), not a bare object - record_schema's "subscribers" field is itself that array, and
  body_fields copies it verbatim so the wire body matches Drip's real shape exactly; the engine's
  schema dialect has no minItems/maxItems keyword to mechanically cap it to exactly one element, so
  callers are expected to supply exactly one (this action is not the true batch endpoint).
- `delete_subscriber`: DELETE `/{{ config.account_id }}/subscribers/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: permanently
  removes a subscriber and their event/tag history; destructive, approval required.
- `unsubscribe_subscriber`: POST `/{{ config.account_id }}/unsubscribes` - kind `custom`; body type
  `json`; body fields `subscribers`; required record fields `subscribers`; accepted fields
  `subscribers`; confirmation `destructive`; risk: unsubscribes the named email from ALL mailings in
  the account; stops all future campaign/broadcast/workflow sends to them, approval required.
- `apply_tag`: POST `/{{ config.account_id }}/subscribers/{{ record.subscriber_id }}/tags` - kind
  `custom`; body type `json`; path fields `subscriber_id`; body fields `tags`; required record
  fields `subscriber_id`, `tags`; accepted fields `subscriber_id`, `tags`; risk: applies one or more
  tags to a subscriber, potentially triggering any workflow with a matching tag-applied trigger;
  low-risk external mutation, no approval required.
- `remove_tag`: DELETE `/{{ config.account_id }}/subscribers/{{ record.subscriber_id }}/tags/{{
  record.tag }}` - kind `custom`; body type `none`; path fields `subscriber_id`, `tag`; required
  record fields `subscriber_id`, `tag`; accepted fields `subscriber_id`, `tag`; risk: removes a tag
  from a subscriber, potentially triggering any workflow with a matching tag-removed trigger;
  low-risk external mutation, no approval required.
- `record_event`: POST `/{{ config.account_id }}/events` - kind `create`; body type `json`; body
  fields `events`; required record fields `events`; accepted fields `events`; risk: records a custom
  behavioral event on a subscriber, which can trigger any workflow with a matching event trigger;
  low-risk external mutation, no approval required. Drip's real API requires the body to be an ARRAY
  under the "events" key even for one event; record_schema's "events" field is itself that array
  (callers are expected to supply exactly one element; this action is not the true batch endpoint).
- `create_broadcast`: POST `/{{ config.account_id }}/broadcasts` - kind `create`; body type `json`;
  body fields `broadcasts`; required record fields `broadcasts`; accepted fields `broadcasts`; risk:
  creates a new draft single-email campaign (broadcast); low-risk in draft form (not sent), no
  approval required. Drip's real API requires the body to be an ARRAY under the "broadcasts" key
  even for one broadcast; record_schema's "broadcasts" field is itself that array (callers are
  expected to supply exactly one element; this action is not the true batch endpoint).
- `update_broadcast`: PATCH `/{{ config.account_id }}/broadcasts/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; body fields `broadcasts`; required record fields `id`,
  `broadcasts`; accepted fields `broadcasts`, `id`; risk: mutates an existing draft broadcast's
  subject/content; Drip only allows updating broadcasts still in draft status, so this cannot alter
  an already-sent email; external mutation, approval required.
- `delete_broadcast`: DELETE `/{{ config.account_id }}/broadcasts/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: permanently removes
  a broadcast; destructive if it was already sent (removes historical record, though delivered
  emails cannot be recalled), approval required.
- `activate_workflow`: POST `/{{ config.account_id }}/workflows/{{ record.id }}/activate` - kind
  `custom`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: activates a paused workflow, resuming automated sends to everyone currently enrolled and
  allowing new triggers to enroll people; external mutation, approval required.
- `pause_workflow`: POST `/{{ config.account_id }}/workflows/{{ record.id }}/pause` - kind `custom`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  pauses an active workflow, stopping automated sends to everyone currently enrolled and disabling
  new triggers; low-risk (reversible via activate_workflow), no approval required.
- `create_webhook`: POST `/{{ config.account_id }}/webhooks` - kind `create`; body type `json`; body
  fields `webhooks`; required record fields `webhooks`; accepted fields `webhooks`; risk: registers
  a new outbound webhook that will POST live subscriber/campaign event data to an external URL of
  the caller's choosing; verify the target endpoint before enabling. Drip's real API requires the
  body to be an ARRAY under the "webhooks" key even for one webhook; record_schema's "webhooks"
  field is itself that array (callers are expected to supply exactly one element; this action is not
  the true batch endpoint).
- `delete_webhook`: DELETE `/{{ config.account_id }}/webhooks/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: permanently removes a
  webhook subscription; stops future event delivery to its post_url; destructive, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=9, non_data_endpoint=3, out_of_scope=20.
