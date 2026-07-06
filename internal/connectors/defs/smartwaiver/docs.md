# Overview

Reads Smartwaiver waivers, checkins, templates, published keys, user info, and account settings;
sends prefill/SMS/webhook mutations through the Smartwaiver API.

Readable streams: `waivers`, `checkins`, `templates`, `published_keys`, `user_info`, `settings`.

Write actions: `set_webhook_config`, `resend_webhook`, `send_sms`, `prefill_template`.

Service API documentation: https://api.smartwaiver.com/docs/v4/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Smartwaiver API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.smartwaiver.com`; format `uri`; Smartwaiver
  API base URL override for tests or proxies.
- `end_date` (optional, string); Optional toDts query filter passed through verbatim.
- `page_size` (optional, string); default `100`; Records per page (limit query param), 1-100.
- `start_date` (optional, string); Optional fromDts query filter passed through verbatim.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.smartwaiver.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v4/me`.

## Streams notes

Default pagination: single request; no pagination.

- `waivers`: GET `/v4/waivers` - records path `waivers.waivers`; query `fromDts` from template `{{
  config.start_date }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `100`; `offset`=`0`; `toDts` from template `{{ config.end_date }}`, omitted when absent;
  emits passthrough records.
- `checkins`: GET `/v4/checkins` - records path `checkins.checkins`; query `fromDts` from template
  `{{ config.start_date }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `100`; `offset`=`0`; `toDts` from template `{{ config.end_date }}`, omitted when absent;
  emits passthrough records.
- `templates`: GET `/v4/templates` - records path `templates.templates`; query `fromDts` from
  template `{{ config.start_date }}`, omitted when absent; `limit` from template `{{
  config.page_size }}`, default `100`; `offset`=`0`; `toDts` from template `{{ config.end_date }}`,
  omitted when absent; emits passthrough records.
- `published_keys`: GET `/v4/keys/published` - records path `published_keys.keys`; query `fromDts`
  from template `{{ config.start_date }}`, omitted when absent; `limit` from template `{{
  config.page_size }}`, default `100`; `offset`=`0`; `toDts` from template `{{ config.end_date }}`,
  omitted when absent; emits passthrough records.
- `user_info`: GET `/v4/info` - records path `.`; query `fromDts` from template `{{
  config.start_date }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `100`; `offset`=`0`; `toDts` from template `{{ config.end_date }}`, omitted when absent;
  emits passthrough records.
- `settings`: GET `/v4/settings` - records path `.`; emits passthrough records.

## Write actions & risks

Overall write risk: configures the account's webhook delivery endpoint, resends a waiver's webhook
notification, sends an outbound SMS waiver-signing link to a real phone number, and generates a
prefilled waiver-signing link carrying participant PII.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `set_webhook_config`: PUT `/v4/webhooks/configure` - kind `update`; body type `json`; required
  record fields `endpoint`; accepted fields `checkinWebhook`, `create`, `emailValidationRequired`,
  `endpoint`, `searchWebhook`, `webhookNumber`; risk: changes where the account's near-real-time
  waiver-signed webhook notifications are delivered; approval required.
- `resend_webhook`: PUT `/v4/webhooks/resend/{{ record.waiver_id }}` - kind `custom`; body type
  `none`; path fields `waiver_id`; required record fields `waiver_id`; accepted fields `waiver_id`;
  risk: re-triggers the new-waiver webhook delivery for a specific waiver (testing aid, heavily rate
  limited by Smartwaiver at 2/minute); approval required.
- `send_sms`: POST `/v4/sms` - kind `create`; body type `json`; required record fields `templateId`,
  `number`; accepted fields `number`, `templateId`; risk: sends an outbound SMS with a
  waiver-signing link to a real phone number (rate limited daily by Smartwaiver for anti-spam);
  approval required.
- `prefill_template`: POST `/v4/templates/{{ record.template_id }}/prefill` - kind `create`; body
  type `json`; path fields `template_id`; required record fields `template_id`; accepted fields
  `addressCity`, `addressLineOne`, `addressState`, `addressZip`, `adult`, `customWaiverFields`,
  `email`, `expiration`, `guardian`, `lockdownPrefill`, `participants`, `template_id`; risk:
  generates a prefilled waiver-signing link carrying real participant PII (name/DOB/address/custom
  fields); approval required.

## Known limits

- API coverage includes 6 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, duplicate_of=3, non_data_endpoint=1, out_of_scope=9.
