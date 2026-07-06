# Overview

Reads ConvertKit (Kit) subscribers, forms, sequences, tags, broadcasts, custom fields, and
purchases, and writes subscriber/tag/form/sequence/broadcast/custom-field/purchase/webhook
mutations, through the ConvertKit v3 REST API.

Readable streams: `subscribers`, `forms`, `sequences`, `tags`, `broadcasts`, `custom_fields`,
`purchases`.

Write actions: `update_subscriber`, `create_tag`, `tag_subscriber`, `remove_tag_from_subscriber`,
`subscribe_to_form`, `subscribe_to_sequence`, `create_broadcast`, `update_broadcast`,
`delete_broadcast`, `create_custom_field`, `update_custom_field`, `create_purchase`,
`create_webhook`, `delete_webhook`.

Service API documentation: https://developers.kit.com/api-reference/v3/overview.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string).
- `api_key` (required, secret, string); ConvertKit v3 API secret, sent as the api_secret query
  parameter on every request. Never logged.
- `api_secret` (optional, secret, string).
- `base_url` (optional, string); default `https://api.convertkit.com/v3`; format `uri`; ConvertKit
  API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`, `api_key`, `api_secret`.

Default configuration values: `base_url=https://api.convertkit.com/v3`.

Authentication behavior:

- API key authentication in query parameter `api_secret` using `secrets.api_key` when `{{
  secrets.api_key }}`.
- API key authentication in query parameter `api_secret` using `secrets.access_token` when `{{
  secrets.access_token }}`.
- API key authentication in query parameter `api_secret` using `secrets.api_secret` when `{{
  secrets.api_secret }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `forms`, `sequences`, `tags`, `custom_fields`; page_number:
`subscribers`, `broadcasts`, `purchases`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `subscribers`: GET `/subscribers` - records path `subscribers`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 50; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `forms`: GET `/forms` - records path `forms`; incremental cursor `created_at`; formatted as
  `rfc3339`.
- `sequences`: GET `/sequences` - records path `sequences`; incremental cursor `created_at`;
  formatted as `rfc3339`.
- `tags`: GET `/tags` - records path `tags`; incremental cursor `created_at`; formatted as
  `rfc3339`.
- `broadcasts`: GET `/broadcasts` - records path `broadcasts`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 50; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `custom_fields`: GET `/custom_fields` - records path `custom_fields`.
- `purchases`: GET `/purchases` - records path `purchases`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 50; incremental cursor `transaction_time`;
  formatted as `rfc3339`.

## Write actions & risks

Overall write risk: external mutation: creates/updates subscribers, tags, forms/sequences
subscriptions, broadcasts (including scheduling live sends), custom fields, purchase records, and
webhooks; deletes are limited to broadcasts/webhooks/tag-removal (no
subscriber/custom-field/global-unsubscribe deletes).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_subscriber`: PUT `/subscribers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `email_address`, `fields`, `first_name`,
  `id`; risk: mutates an existing subscriber's name/email/custom-field values; external mutation, no
  approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a new tag on the account; low-risk external mutation, no
  approval required.
- `tag_subscriber`: POST `/tags/{{ record.tag_id }}/subscribe` - kind `update`; body type `json`;
  path fields `tag_id`; required record fields `tag_id`, `email`; accepted fields `email`, `fields`,
  `first_name`, `tag_id`, `tags`; risk: applies a tag to a subscriber (creating the subscriber if
  the email is new); external mutation, no approval required.
- `remove_tag_from_subscriber`: DELETE `/subscribers/{{ record.subscriber_id }}/tags/{{
  record.tag_id }}` - kind `delete`; body type `none`; path fields `subscriber_id`, `tag_id`;
  required record fields `subscriber_id`, `tag_id`; accepted fields `subscriber_id`, `tag_id`;
  missing records treated as success for status `404`; risk: removes a tag from a subscriber;
  external mutation, no approval required.
- `subscribe_to_form`: POST `/forms/{{ record.form_id }}/subscribe` - kind `update`; body type
  `json`; path fields `form_id`; required record fields `form_id`, `email`; accepted fields `email`,
  `fields`, `first_name`, `form_id`, `tags`; risk: subscribes an email address to a form (creating
  the subscriber if the email is new); external mutation, no approval required.
- `subscribe_to_sequence`: POST `/sequences/{{ record.sequence_id }}/subscribe` - kind `update`;
  body type `json`; path fields `sequence_id`; required record fields `sequence_id`, `email`;
  accepted fields `email`, `fields`, `first_name`, `sequence_id`, `tags`; risk: subscribes an email
  address to a sequence (creating the subscriber if the email is new); external mutation, no
  approval required.
- `create_broadcast`: POST `/broadcasts` - kind `create`; body type `json`; required record fields
  `subject`, `content`; accepted fields `content`, `description`, `email_address`,
  `email_layout_template`, `public`, `published_at`, `send_at`, `subject`, `thumbnail_alt`,
  `thumbnail_url`; risk: creates a draft or scheduled email broadcast; a scheduled broadcast
  (send_at/published_at set) will send to the account's live subscriber list, external mutation,
  approval required.
- `update_broadcast`: PUT `/broadcasts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `content`, `description`, `id`,
  `public`, `published_at`, `send_at`, `subject`; risk: mutates a draft or scheduled broadcast's
  content/send time; external mutation, approval required.
- `delete_broadcast`: DELETE `/broadcasts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently deletes a draft or scheduled broadcast record; irreversible,
  approval required.
- `create_custom_field`: POST `/custom_fields` - kind `create`; body type `json`; required record
  fields `label`; accepted fields `label`; risk: creates a new custom subscriber field on the
  account (up to 140 total); low-risk external mutation, no approval required.
- `update_custom_field`: PUT `/custom_fields/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `label`; accepted fields `id`, `label`; risk:
  renames a custom field's label (the underlying key is unchanged per Kit's own docs); external
  mutation, no approval required.
- `create_purchase`: POST `/purchases` - kind `create`; body type `json`; required record fields
  `purchase`; accepted fields `integration`, `integration_key`, `purchase`; risk: records a new
  purchase-tracking transaction for a subscriber; external mutation, no approval required.
- `create_webhook`: POST `/automations/hooks` - kind `create`; body type `json`; required record
  fields `target_url`, `event`; accepted fields `event`, `target_url`; risk: creates a webhook that
  POSTs subscriber-event payloads to an external URL the caller controls; external mutation,
  approval required.
- `delete_webhook`: DELETE `/automations/hooks/{{ record.rule_id }}` - kind `delete`; body type
  `none`; path fields `rule_id`; required record fields `rule_id`; accepted fields `rule_id`;
  missing records treated as success for status `404`; risk: permanently deletes a webhook
  automation; irreversible, approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 7 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=7, non_data_endpoint=2, out_of_scope=1.
