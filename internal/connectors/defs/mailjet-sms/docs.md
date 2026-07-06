# Overview

Reads Mailjet SMS messages, message counts, and export job status; writes SMS send and
export-request actions.

Readable streams: `sms`, `sms_count`, `sms_message`, `sms_export`.

Write actions: `send_sms`, `request_sms_export`.

Service API documentation: https://dev.mailjet.com/sms/reference/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.mailjet.com/v4`; format `uri`; Mailjet SMS API
  base URL override for tests or proxies.
- `end_date` (optional, string); Unix-seconds upper bound sent as the ToTS query param on list/count
  streams.
- `export_job_id` (optional, string); Mailjet export job id used by the sms_export detail stream.
- `mode` (optional, string).
- `recipient` (optional, string); Optional To filter for message list/count requests.
- `sms_id` (optional, string); SMS message id used by the sms_message detail stream.
- `sms_ids` (optional, string); Optional IDs filter for the sms list stream; use Mailjet's
  comma-separated ID format.
- `start_date` (optional, string); Unix-seconds lower bound sent as the FromTS query param on
  list/count streams.
- `status_code` (optional, string); Optional Mailjet SMS StatusCode filter for message list/count
  requests; use the API's comma-separated code format when filtering multiple statuses.
- `token` (required, secret, string); Mailjet SMS API Bearer token. Used only for Bearer auth; never
  logged.

Secret fields are redacted in logs and write previews: `token`.

Default configuration values: `base_url=https://api.mailjet.com/v4`.

Authentication behavior:

- Bearer token authentication using `secrets.token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sms/count`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `sms_count`, `sms_message`, `sms_export`; offset_limit: `sms`.

- `sms`: GET `/sms` - records path `Data`; query `FromTS` from template `{{ config.start_date }}`,
  omitted when absent; `IDs` from template `{{ config.sms_ids }}`, omitted when absent; `StatusCode`
  from template `{{ config.status_code }}`, omitted when absent; `To` from template `{{
  config.recipient }}`, omitted when absent; `ToTS` from template `{{ config.end_date }}`, omitted
  when absent; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page
  size 100; computed output fields `cost_currency`, `cost_value`, `status_code`,
  `status_description`, `status_name`.
- `sms_count`: GET `/sms/count` - single-object response; records path `.`; query `FromTS` from
  template `{{ config.start_date }}`, omitted when absent; `StatusCode` from template `{{
  config.status_code }}`, omitted when absent; `To` from template `{{ config.recipient }}`, omitted
  when absent; `ToTS` from template `{{ config.end_date }}`, omitted when absent.
- `sms_message`: GET `/sms/{{ config.sms_id }}` - records path `Data`; computed output fields
  `cost_currency`, `cost_value`, `status_code`, `status_description`, `status_name`.
- `sms_export`: GET `/sms/export/{{ config.export_job_id }}` - single-object response; records path
  `.`; computed output fields `status_code`, `status_description`, `status_name`.

## Write actions & risks

Overall write risk: external Mailjet SMS API mutation; may send SMS messages or request asynchronous
SMS exports.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `send_sms`: POST `/sms-send` - kind `create`; body type `json`; required record fields `From`,
  `To`, `Text`; accepted fields `From`, `Text`, `To`; risk: external mutation; sends an SMS message;
  approval required.
- `request_sms_export`: POST `/sms/export` - kind `create`; body type `json`; required record fields
  `FromTS`, `ToTS`; accepted fields `FromTS`, `ToTS`; risk: external mutation; creates an
  asynchronous SMS export job; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=100.
- API coverage includes 4 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
