# Overview

Reads Smaily campaigns, segments, contacts, templates, automations, and organization users;
creates/updates subscribers and segments, unsubscribes recipients, sends messages, and triggers
automation workflows.

Readable streams: `campaigns`, `segments`, `subscribers`, `templates`, `automations`,
`segment_rules`, `segment_subscribers`, `ab_tests`, `organization_users`.

Write actions: `create_or_update_subscriber`, `create_or_update_segment`, `unsubscribe_recipient`,
`send_message`, `trigger_automation_workflow`, `launch_ab_test`.

Service API documentation: https://smaily.com/help/api/.

## Auth setup

Connection fields:

- `api_password` (required, secret, string); Smaily API password, sent as the HTTP Basic auth
  password. Never logged.
- `api_username` (required, string); Smaily API username, sent as the HTTP Basic auth username.
- `base_url` (required, string); format `uri`; Your Smaily account's API base URL (e.g.
  https://<subdomain>.sendsmaily.net).
- `segment_id` (optional, string); Optional Smaily segment id used to scope the segment_subscribers
  stream to one segment's members (sent as the list query parameter on GET api/contact.php). When
  unset, segment_subscribers reads the account's full contact list unfiltered.

Secret fields are redacted in logs and write previews: `api_password`.

Authentication behavior:

- HTTP Basic authentication using `config.api_username`, `secrets.api_password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `api/campaign.php`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `campaigns`, `segments`, `subscribers`, `templates`, `automations`,
`segment_rules`, `segment_subscribers`, `ab_tests`; page_number: `organization_users`.

- `campaigns`: GET `api/campaign.php` - records at response root; emits passthrough records.
- `segments`: GET `api/segment.php` - records at response root; emits passthrough records.
- `subscribers`: GET `api/contact.php` - records at response root; emits passthrough records.
- `templates`: GET `api/template.php` - records at response root; emits passthrough records.
- `automations`: GET `api/autoresponder.php` - records at response root; emits passthrough records.
- `segment_rules`: GET `api/list.php` - records at response root; emits passthrough records.
- `segment_subscribers`: GET `api/contact.php` - records at response root; query `list` from
  template `{{ config.segment_id }}`, omitted when absent; emits passthrough records.
- `ab_tests`: GET `api/split.php` - records at response root; emits passthrough records.
- `organization_users`: GET `api/organizations/users.php` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 0; page size 2; emits
  passthrough records.

## Write actions & risks

Overall write risk: creates/updates subscribers and segments, unsubscribes a recipient from a
campaign, sends an individually-templated outbound email, and triggers an automation workflow for
real subscribers.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_or_update_subscriber`: POST `api/contact.php` - kind `upsert`; body type `json`; required
  record fields `email`; accepted fields `email`, `is_deleted`, `is_unsubscribed`, `name`; risk:
  external mutation; creates or updates a subscriber (matched by email) on the connected Smaily
  account; does not trigger automation workflows; approval required.
- `create_or_update_segment`: POST `api/list.php` - kind `upsert`; body type `json`; required record
  fields `name`, `filter_type`, `filter_data`; accepted fields `filter_data`, `filter_type`, `id`,
  `name`; risk: external mutation; creates a new segment or, when id is set, overwrites an existing
  segment's filter definition on the connected Smaily account; approval required.
- `unsubscribe_recipient`: POST `api/unsubscribe.php` - kind `update`; body type `json`; required
  record fields `email`, `campaign_id`; accepted fields `campaign_id`, `email`; risk: external
  mutation; unsubscribes a recipient from a specific campaign (reflected in that campaign's
  statistics); approval required.
- `send_message`: POST `api/message/send.php` - kind `create`; body type `json`; required record
  fields `autoresponder_id`, `to`; accepted fields `attachments`, `autoresponder_id`, `context`,
  `to`; risk: external mutation; sends a real, individually-templated outbound email to real
  recipients using an automation workflow's template (without triggering the workflow itself);
  approval required.
- `trigger_automation_workflow`: POST `api/autoresponder.php` - kind `custom`; body type `json`;
  required record fields `autoresponder`, `addresses`; accepted fields `addresses`, `autoresponder`;
  risk: external mutation; opts in subscribers and triggers a 'form submitted' automation workflow
  for them, updating subscriber data before any scheduled messages send; approval required.
- `launch_ab_test`: POST `api/split.php` - kind `create`; body type `json`; required record fields
  `splits`, `list`, `size`, `win_at`; accepted fields `condition`, `list`, `save_as_draft`, `size`,
  `splits`, `win_at`; risk: external mutation; creates and, unless save_as_draft is set, immediately
  launches a real A/B test campaign send to a percentage of a real subscriber list, with the winning
  variant auto-sent to the remainder at win_at; approval required.

## Known limits

- API coverage includes 9 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, out_of_scope=4.
