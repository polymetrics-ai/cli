# Overview

Reads SparkPost recipient lists, templates, sending domains, transmissions, suppression list
records, IP pools, webhooks, subaccounts, tracking domains, inbound domains, relay webhooks, sending
IPs, and A/B tests; writes email sends, recipient
list/template/domain/suppression/IP-pool/webhook/subaccount/relay-webhook lifecycle mutations.

Readable streams: `recipient_lists`, `templates`, `sending_domains`, `transmissions`,
`suppression_list`, `ip_pools`, `webhooks`, `subaccounts`, `tracking_domains`, `inbound_domains`,
`relay_webhooks`, `sending_ips`, `ab_tests`, `account`.

Write actions: `update_account`, `create_transmission`, `create_recipient_list`, `create_template`,
`update_template`, `delete_template`, `create_sending_domain`, `update_sending_domain`,
`delete_sending_domain`, `create_or_update_suppression`, `delete_suppression`, `create_ip_pool`,
`update_ip_pool`, `delete_ip_pool`, `create_webhook`, `update_webhook`, `delete_webhook`,
`create_subaccount`, `update_subaccount`, `create_tracking_domain`, `delete_tracking_domain`,
`create_inbound_domain`, `delete_inbound_domain`, `create_relay_webhook`, `update_relay_webhook`,
`delete_relay_webhook`.

Service API documentation: https://developers.sparkpost.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); SparkPost API key, sent verbatim as the Authorization header
  value (no Bearer prefix). Never logged.
- `base_url` (optional, string); default `https://api.sparkpost.com/api/v1`; format `uri`; SparkPost
  API base URL. Defaults to the US endpoint; set explicitly to https://api.eu.sparkpost.com/api/v1
  for the EU region.
- `end_date` (optional, string); format `date-time`; Upper bound sent as the 'to' query parameter.
- `start_date` (optional, string); format `date-time`; Lower bound sent as the 'from' query
  parameter.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.sparkpost.com/api/v1`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/recipient-lists`.

## Streams notes

Default pagination: single request; no pagination.

- `recipient_lists`: GET `/recipient-lists` - records path `results`; query `from` from template `{{
  config.start_date }}`, omitted when absent; `to` from template `{{ config.end_date }}`, omitted
  when absent; emits passthrough records.
- `templates`: GET `/templates` - records path `results`; query `from` from template `{{
  config.start_date }}`, omitted when absent; `to` from template `{{ config.end_date }}`, omitted
  when absent; emits passthrough records.
- `sending_domains`: GET `/sending-domains` - records path `results`; query `from` from template `{{
  config.start_date }}`, omitted when absent; `to` from template `{{ config.end_date }}`, omitted
  when absent; emits passthrough records.
- `transmissions`: GET `/transmissions` - records path `results`; query `from` from template `{{
  config.start_date }}`, omitted when absent; `to` from template `{{ config.end_date }}`, omitted
  when absent; emits passthrough records.
- `suppression_list`: GET `/suppression-list` - records path `results`; query `from` from template
  `{{ config.start_date }}`, omitted when absent; `to` from template `{{ config.end_date }}`,
  omitted when absent; emits passthrough records.
- `ip_pools`: GET `/ip-pools` - records path `results`; emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `results`; emits passthrough records.
- `subaccounts`: GET `/subaccounts` - records path `results`; emits passthrough records.
- `tracking_domains`: GET `/tracking-domains` - records path `results`; emits passthrough records.
- `inbound_domains`: GET `/inbound-domains` - records path `results`; emits passthrough records.
- `relay_webhooks`: GET `/relay-webhooks` - records path `results`; emits passthrough records.
- `sending_ips`: GET `/sending-ips` - records path `results`; emits passthrough records.
- `ab_tests`: GET `/ab-test` - records path `results`; emits passthrough records.
- `account`: GET `/account` - single-object response; records path `results`; emits passthrough
  records.

## Write actions & risks

Overall write risk: external SparkPost API mutation including live email sends
(create_transmission), suppression/domain/webhook/subaccount lifecycle changes, and IP pool/relay
webhook configuration.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_account`: PUT `/account` - kind `update`; body type `json`; accepted fields
  `company_name`, `options`, `tfa_required`; risk: external mutation; changes account-wide settings
  (company name, two-factor requirement, default tracking/transactional options) affecting every
  sender on the account; approval required.
- `create_transmission`: POST `/transmissions` - kind `create`; body type `json`; required record
  fields `recipients`, `content`; accepted fields `campaign_id`, `content`, `description`,
  `metadata`, `options`, `recipients`, `return_path`, `substitution_data`; risk: external mutation;
  sends real email to every listed recipient through the connected SparkPost account; approval
  required.
- `create_recipient_list`: POST `/recipient-lists` - kind `create`; body type `json`; required
  record fields `recipients`; accepted fields `attributes`, `description`, `id`, `name`,
  `recipients`; risk: external mutation; creates a stored recipient list; approval required.
- `create_template`: POST `/templates` - kind `create`; body type `json`; required record fields
  `content`; accepted fields `content`, `description`, `id`, `name`, `options`, `published`,
  `shared_with_subaccounts`; risk: external mutation; creates a message template (as a draft unless
  published); approval required.
- `update_template`: PUT `/templates/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `content`, `description`, `id`, `name`,
  `options`, `published`; risk: external mutation; updates an existing message template's
  draft/published content; approval required.
- `delete_template`: DELETE `/templates/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a message template; approval required.
- `create_sending_domain`: POST `/sending-domains` - kind `create`; body type `json`; required
  record fields `domain`; accepted fields `dkim`, `dkim_key_length`, `domain`, `generate_dkim`,
  `shared_with_subaccounts`, `tracking_domain`; risk: external mutation; registers a new sending
  domain pending DNS verification; approval required.
- `update_sending_domain`: PUT `/sending-domains/{{ record.domain }}` - kind `update`; body type
  `json`; path fields `domain`; required record fields `domain`; accepted fields `dkim`, `domain`,
  `is_default_bounce_domain`, `shared_with_subaccounts`, `tracking_domain`; risk: external mutation;
  changes an existing sending domain's DKIM/tracking/bounce configuration; approval required.
- `delete_sending_domain`: DELETE `/sending-domains/{{ record.domain }}` - kind `delete`; body type
  `none`; path fields `domain`; required record fields `domain`; accepted fields `domain`; risk:
  external mutation; permanently removes a sending domain; approval required.
- `create_or_update_suppression`: PUT `/suppression-list/{{ record.recipient }}` - kind `upsert`;
  body type `json`; path fields `recipient`; required record fields `recipient`, `type`; accepted
  fields `description`, `list_id`, `recipient`, `type`; risk: external mutation; adds or updates a
  recipient's suppression (opt-out) entry, affecting future deliverability to that address; approval
  required.
- `delete_suppression`: DELETE `/suppression-list/{{ record.recipient }}` - kind `delete`; body type
  `none`; path fields `recipient`; required record fields `recipient`; accepted fields `recipient`;
  risk: external mutation; removes a recipient's suppression entry, re-enabling delivery to that
  address; approval required.
- `create_ip_pool`: POST `/ip-pools` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `auto_warmup_overflow_pool`, `fbl_signing_domain`, `name`,
  `signing_domain`; risk: external mutation; creates a dedicated IP pool; approval required.
- `update_ip_pool`: PUT `/ip-pools/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`; accepted fields `auto_warmup_overflow_pool`,
  `fbl_signing_domain`, `id`, `name`, `signing_domain`; risk: external mutation; changes an IP
  pool's DKIM signing domain / auto-warmup overflow configuration; approval required.
- `delete_ip_pool`: DELETE `/ip-pools/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes an IP pool; approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `name`, `target`, `events`; accepted fields `active`, `auth_credentials`, `auth_request_details`,
  `auth_type`, `custom_headers`, `events`, `name`, `target`; risk: external mutation; creates a
  webhook that will POST live event batches to an externally-supplied URL; a test POST is sent to
  target immediately; approval required.
- `update_webhook`: PUT `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `active`, `auth_type`, `custom_headers`,
  `events`, `id`, `name`, `target`; risk: external mutation; changes an existing webhook's
  target/events/auth; a test POST is sent to a new target immediately; approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a webhook; approval required.
- `create_subaccount`: POST `/subaccounts` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `ip_pool`, `key_grants`, `key_label`, `key_valid_ips`, `name`, `options`,
  `setup_api_key`; risk: external mutation; provisions a new subaccount, optionally with a live API
  key; approval required.
- `update_subaccount`: PUT `/subaccounts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `id`, `ip_pool`, `name`, `options`,
  `status`; risk: external mutation; changes a subaccount's name/status/ip_pool -- status
  transitions (e.g. to suspended/terminated) directly affect that subaccount's ability to send mail;
  approval required.
- `create_tracking_domain`: POST `/tracking-domains` - kind `create`; body type `json`; required
  record fields `domain`; accepted fields `domain`, `secure`; risk: external mutation; registers a
  new tracking domain pending DNS verification; approval required.
- `delete_tracking_domain`: DELETE `/tracking-domains/{{ record.domain }}` - kind `delete`; body
  type `none`; path fields `domain`; required record fields `domain`; accepted fields `domain`;
  risk: external mutation; permanently removes a tracking domain; approval required.
- `create_inbound_domain`: POST `/inbound-domains` - kind `create`; body type `json`; required
  record fields `domain`; accepted fields `domain`; risk: external mutation; registers a new inbound
  (receiving) domain; approval required.
- `delete_inbound_domain`: DELETE `/inbound-domains/{{ record.domain }}` - kind `delete`; body type
  `none`; path fields `domain`; required record fields `domain`; accepted fields `domain`; risk:
  external mutation; permanently removes an inbound domain, stopping inbound relay of mail addressed
  to it; approval required.
- `create_relay_webhook`: POST `/relay-webhooks` - kind `create`; body type `json`; required record
  fields `target`, `match`; accepted fields `auth_request_details`, `auth_token`, `auth_type`,
  `custom_headers`, `match`, `name`, `target`; risk: external mutation; creates a relay webhook that
  will POST live inbound-mail batches to an externally-supplied URL; approval required.
- `update_relay_webhook`: PUT `/relay-webhooks/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `auth_type`, `custom_headers`,
  `id`, `match`, `name`, `target`; risk: external mutation; changes an existing relay webhook's
  target/match/auth; approval required.
- `delete_relay_webhook`: DELETE `/relay-webhooks/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; permanently deletes a relay webhook; approval required.

## Known limits

- API coverage includes 14 stream-backed endpoint group(s), 26 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=6, duplicate_of=13, non_data_endpoint=11, out_of_scope=77, requires_elevated_scope=3.
