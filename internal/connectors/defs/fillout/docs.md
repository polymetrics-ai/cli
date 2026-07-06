# Overview

Reads Fillout forms and manages webhooks/submission deletion through the Fillout REST API.

Readable streams: `forms`.

Write actions: `create_webhook`, `remove_webhook`, `delete_submission_by_id`.

Service API documentation: https://www.fillout.com/help/fillout-rest-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Fillout API key. Used only for Bearer auth (Authorization:
  Bearer <api_key>); never logged.
- `base_url` (optional, string); default `https://api.fillout.com/v1/api`; format `uri`; Fillout API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.fillout.com/v1/api`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms`.

## Streams notes

Default pagination: single request; no pagination.

- `forms`: GET `/forms` - records at response root; computed output fields `id`.

## Write actions & risks

Overall write risk: creates/removes outbound webhook subscriptions and deletes individual form
submissions; external mutation, approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_webhook`: POST `/webhook/create` - kind `create`; body type `json`; required record fields
  `formId`, `url`; accepted fields `formId`, `url`; risk: registers a new outbound webhook
  subscription that will POST live form-submission data to an external URL; external mutation,
  approval required.
- `remove_webhook`: POST `/webhook/delete` - kind `custom`; body type `json`; required record fields
  `webhookId`; accepted fields `webhookId`; confirmation `destructive`; risk: permanently removes a
  webhook subscription; event delivery to its target URL stops immediately.
- `delete_submission_by_id`: DELETE `/forms/{{ record.form_id }}/submissions/{{ record.submission_id
  }}` - kind `delete`; body type `none`; path fields `form_id`, `submission_id`; required record
  fields `form_id`, `submission_id`; accepted fields `form_id`, `submission_id`; missing records
  treated as success for status `404`; risk: permanently deletes a single form response;
  irreversible, approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 1 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=3.
