# Overview

Reads SmartEngage avatars, tags, custom fields, sequences, and subscribers; creates/updates
subscribers, tags, custom fields, and sequence enrollments.

Readable streams: `avatars`, `tags`, `custom_fields`, `sequences`, `subscribers`.

Write actions: `add_subscriber`, `update_subscriber`, `create_tag`, `add_tag_to_subscriber`,
`remove_tag_from_subscriber`, `create_custom_field`, `set_custom_field_value`,
`add_subscriber_to_sequence`, `remove_subscriber_from_sequence`.

Service API documentation: https://smartengage.com/docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); SmartEngage API token, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `avatar_id` (optional, string).
- `base_url` (optional, string); default `https://api.smartengage.com`; format `uri`; SmartEngage
  API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.smartengage.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `avatars/list`.

## Streams notes

Default pagination: single request; no pagination.

- `avatars`: GET `avatars/list` - records at response root; query `avatar_id` from template `{{
  config.avatar_id }}`, omitted when absent; emits passthrough records.
- `tags`: GET `tags/list/` - records at response root; query `avatar_id` from template `{{
  config.avatar_id }}`, omitted when absent; emits passthrough records.
- `custom_fields`: GET `customfields/list/` - records at response root; query `avatar_id` from
  template `{{ config.avatar_id }}`, omitted when absent; emits passthrough records.
- `sequences`: GET `sequences/list/` - records at response root; query `avatar_id` from template `{{
  config.avatar_id }}`, omitted when absent; emits passthrough records.
- `subscribers`: GET `subscribers/list/` - records at response root; query `avatar_id` from template
  `{{ config.avatar_id }}`, omitted when absent; emits passthrough records.

## Write actions & risks

Overall write risk: creates/updates subscribers and custom-field values, creates tags and
attaches/detaches them from subscribers, and enrolls/unenrolls subscribers in automation sequences
(which triggers or stops scheduled outbound messages).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_subscriber`: POST `subscribers/add` - kind `create`; body type `form`; required record fields
  `avatar_id`; accepted fields `avatar_id`, `email`, `facebook_id`, `first_name`, `full_name`,
  `last_name`, `push_id`; risk: external mutation; creates a new subscriber on the connected
  SmartEngage account; approval required.
- `update_subscriber`: POST `subscribers/update` - kind `update`; body type `form`; required record
  fields `avatar_id`, `subscriber_id`; accepted fields `avatar_id`, `email`, `facebook_id`,
  `first_name`, `full_name`, `last_name`, `subscriber_id`; risk: external mutation; overwrites
  subscriber fields on the connected SmartEngage account (fields omitted from the record remain
  unchanged); approval required.
- `create_tag`: POST `tags/create` - kind `create`; body type `form`; required record fields
  `avatar_id`, `name`; accepted fields `avatar_id`, `name`; risk: external mutation; creates a new
  tag on the connected SmartEngage account; approval required.
- `add_tag_to_subscriber`: POST `tags/add` - kind `custom`; body type `form`; required record fields
  `avatar_id`, `subscriber_id`, `tag`; accepted fields `avatar_id`, `subscriber_id`, `tag`; risk:
  external mutation; attaches an existing tag to a subscriber; approval required.
- `remove_tag_from_subscriber`: POST `tags/delete` - kind `custom`; body type `form`; required
  record fields `avatar_id`, `subscriber_id`, `tag`; accepted fields `avatar_id`, `subscriber_id`,
  `tag`; risk: external mutation; detaches a tag from a subscriber; approval required.
- `create_custom_field`: POST `customfields/create` - kind `create`; body type `form`; required
  record fields `avatar_id`, `custom_field_name`, `custom_field_type`; accepted fields `avatar_id`,
  `custom_field_desc`, `custom_field_name`, `custom_field_title`, `custom_field_type`; risk:
  external mutation; creates a new custom field definition on the connected SmartEngage account;
  approval required.
- `set_custom_field_value`: POST `customfields/update` - kind `update`; body type `form`; required
  record fields `avatar_id`, `subscriber_id`, `field`, `value`; accepted fields `avatar_id`,
  `field`, `subscriber_id`, `value`; risk: external mutation; sets a custom field value on a
  subscriber; approval required.
- `add_subscriber_to_sequence`: POST `sequences/add` - kind `custom`; body type `form`; required
  record fields `avatar_id`, `subscriber_id`, `sequence`; accepted fields `avatar_id`, `sequence`,
  `subscriber_id`; risk: external mutation; enrolls a subscriber into an automation sequence,
  triggering scheduled messages; approval required.
- `remove_subscriber_from_sequence`: POST `sequences/remove` - kind `custom`; body type `form`;
  required record fields `avatar_id`, `subscriber_id`, `sequence`; accepted fields `avatar_id`,
  `sequence`, `subscriber_id`; risk: external mutation; unenrolls a subscriber from an automation
  sequence, stopping scheduled messages; approval required.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
