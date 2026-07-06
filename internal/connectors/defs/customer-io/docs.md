# Overview

Reads Customer.io campaigns, newsletters, segments, broadcasts, activities, messages, exports,
transactional templates, object types, reporting webhooks, sender identities, snippets, subscription
channels/topics, workspaces, and collections; writes snippet/webhook/segment mutations and can send
transactional email or trigger broadcasts, through the Customer.io App API.

Readable streams: `campaigns`, `newsletters`, `segments`, `broadcasts`, `activities`, `messages`,
`exports`, `transactional`, `object_types`, `reporting_webhooks`, `sender_identities`, `snippets`,
`subscription_channels`, `subscription_topics`, `workspaces`, `collections`.

Write actions: `create_snippet`, `update_snippet`, `delete_snippet`, `create_reporting_webhook`,
`update_reporting_webhook`, `delete_reporting_webhook`, `create_manual_segment`,
`delete_manual_segment`, `send_email`, `trigger_broadcast`.

Service API documentation: https://customer.io/docs/api/app/.

## Auth setup

Connection fields:

- `app_api_key` (required, secret, string); Customer.io App API key. Used only for Bearer auth;
  never logged.
- `base_url` (required, string); format `uri`; Customer.io App API base URL for your region:
  https://api.customer.io/v1 (US) or https://api-eu.customer.io/v1 (EU).
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `app_api_key`.

Default configuration values: `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.app_api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/campaigns`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `start`; next token from `next`.

Pagination by stream: cursor: `campaigns`, `newsletters`, `segments`, `broadcasts`, `activities`;
none: `messages`, `exports`, `transactional`, `object_types`, `reporting_webhooks`,
`sender_identities`, `snippets`, `subscription_channels`, `subscription_topics`, `workspaces`,
`collections`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `campaigns`: GET `/campaigns` - records path `campaigns`; query `limit`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `start`; next token from `next`; incremental cursor `updated`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `newsletters`: GET `/newsletters` - records path `newsletters`; query `limit`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `start`; next token from `next`; incremental cursor
  `updated`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `segments`: GET `/segments` - records path `segments`; query `limit`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `start`; next token from `next`; incremental cursor `updated`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `broadcasts`: GET `/broadcasts` - records path `broadcasts`; query `limit`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `start`; next token from `next`; incremental cursor
  `updated`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `activities`: GET `/activities` - records path `activities`; query `limit`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `start`; next token from `next`; incremental cursor
  `timestamp`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `messages`: GET `/messages` - records path `messages`.
- `exports`: GET `/exports` - records path `exports`.
- `transactional`: GET `/transactional` - records path `messages`.
- `object_types`: GET `/object_types` - records path `types`.
- `reporting_webhooks`: GET `/reporting_webhooks` - records path `reporting_webhooks`.
- `sender_identities`: GET `/sender_identities` - records path `sender_identities`.
- `snippets`: GET `/snippets` - records path `snippets`.
- `subscription_channels`: GET `/subscription_channels` - records path `channels`.
- `subscription_topics`: GET `/subscription_topics` - records path `topics`.
- `workspaces`: GET `/workspaces` - records path `workspaces`.
- `collections`: GET `/collections` - records path `collections`.

## Write actions & risks

Overall write risk: external mutation of live Customer.io workspace config
(snippets/webhooks/segments) and live message sends (transactional email, broadcast triggers);
irreversible once delivered; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_snippet`: POST `/snippets` - kind `create`; body type `json`; required record fields
  `name`, `value`; accepted fields `name`, `value`; risk: external mutation; creates a reusable
  content snippet referenced by live messages/newsletters.
- `update_snippet`: PUT `/snippets` - kind `update`; body type `json`; required record fields
  `name`, `value`; accepted fields `name`, `value`; risk: external mutation; overwrites the content
  of a live snippet, changing every message/newsletter that references it.
- `delete_snippet`: DELETE `/snippets/{{ record.name }}` - kind `delete`; body type `none`; path
  fields `name`; required record fields `name`; accepted fields `name`; risk: external mutation;
  permanently removes a snippet; irreversible, breaks any message/newsletter still referencing it;
  approval required.
- `create_reporting_webhook`: POST `/reporting_webhooks` - kind `create`; body type `json`; required
  record fields `name`, `endpoint`, `events`; accepted fields `disabled`, `endpoint`, `events`,
  `full_resolution`, `name`, `with_content`; risk: external mutation; registers a new reporting
  webhook that will deliver live workspace event data to the given endpoint URL.
- `update_reporting_webhook`: PUT `/reporting_webhooks/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `name`, `endpoint`, `events`; accepted
  fields `disabled`, `endpoint`, `events`, `full_resolution`, `id`, `name`, `with_content`; risk:
  external mutation; changes a live reporting webhook's target endpoint/event selection or
  enables/disables delivery.
- `delete_reporting_webhook`: DELETE `/reporting_webhooks/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: external mutation; permanently removes a reporting
  webhook; event delivery to its target URL stops immediately; approval required.
- `create_manual_segment`: POST `/segments` - kind `create`; body type `json`; required record
  fields `segment`; accepted fields `segment`; risk: external mutation; creates a new manual segment
  in the live workspace.
- `delete_manual_segment`: DELETE `/segments/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: external mutation; permanently removes a manual segment;
  irreversible, any campaign/newsletter targeting it loses that audience slice immediately; approval
  required.
- `send_email`: POST `/send/email` - kind `create`; body type `json`; required record fields `to`;
  accepted fields `body`, `from`, `identifiers`, `message_data`, `subject`, `to`,
  `transactional_message_id`; risk: sends a live transactional email to the given recipient on the
  workspace's behalf; irreversible once delivered.
- `trigger_broadcast`: POST `/campaigns/{{ record.broadcast_id }}/triggers` - kind `custom`; body
  type `json`; path fields `broadcast_id`; required record fields `broadcast_id`; accepted fields
  `broadcast_id`, `data`, `email_add_duplicates`, `email_ignore_missing`, `id_ignore_missing`; risk:
  triggers a live API-triggered broadcast to its default audience; sends real messages to
  recipients, irreversible once delivered.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 16 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=67, out_of_scope=66.
- Client-side incremental filtering is used for: `campaigns`, `newsletters`, `segments`,
  `broadcasts`, `activities`.
