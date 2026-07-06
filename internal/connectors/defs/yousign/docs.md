# Overview

Reads and writes Yousign signature requests, contacts, documents, webhooks, templates, users, and
workflow sessions through the Yousign REST API.

Readable streams: `signature_requests`, `contacts`, `documents`, `webhooks`, `templates`, `users`,
`workflow_sessions`.

Write actions: `create_signature_request`, `activate_signature_request`, `cancel_signature_request`,
`create_contact`, `update_contact`, `delete_contact`, `create_webhook`, `delete_webhook`.

Service API documentation: https://developers.yousign.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Yousign API key. Used only for Bearer auth (Authorization:
  Bearer <api_key>); never logged.
- `base_url` (optional, string); default `https://api.yousign.app/v3`; format `uri`; Yousign API
  root. Also usable as a base URL override for tests/proxies.
- `limit` (optional, string); Optional per-request record limit, sent as the 'limit' query parameter
  when set. Omitted entirely when unset.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.yousign.app/v3`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/signature_requests`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `templates`, `users`, `workflow_sessions`; none: `signature_requests`,
`contacts`, `documents`, `webhooks`.

- `signature_requests`: GET `/signature_requests` - records path `data`; query `limit` from template
  `{{ config.limit }}`, omitted when absent; computed output fields `updated_at`; emits passthrough
  records.
- `contacts`: GET `/contacts` - records path `data`; query `limit` from template `{{ config.limit
  }}`, omitted when absent; computed output fields `name`, `updated_at`; emits passthrough records.
- `documents`: GET `/documents` - records path `data`; query `limit` from template `{{ config.limit
  }}`, omitted when absent; computed output fields `name`, `updated_at`; emits passthrough records.
- `webhooks`: GET `/webhooks` - records at response root; computed output fields `name`.
- `templates`: GET `/templates` - records path `data`; cursor pagination; cursor parameter `after`;
  next token from `meta.next_cursor`.
- `users`: GET `/users` - records path `data`; cursor pagination; cursor parameter `after`; next
  token from `meta.next_cursor`; computed output fields `name`.
- `workflow_sessions`: GET `/workflow_sessions` - records path `data`; cursor pagination; cursor
  parameter `after`; next token from `meta.next_cursor`.

## Write actions & risks

Overall write risk: external mutation of e-signature-critical Yousign records: signature request
creation/activation (may immediately notify signers by email) and cancellation (irreversible),
contact create/update/delete, and webhook subscription create/delete.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_signature_request`: POST `/signature_requests` - kind `create`; body type `json`; required
  record fields `name`, `delivery_mode`; accepted fields `delivery_mode`, `expiration_date`,
  `external_id`, `name`, `ordered_signers`, `signers_allowed_to_decline`, `template_id`, `timezone`;
  risk: creates a new draft signature request (no documents/signers attached yet); external
  mutation, approval required.
- `activate_signature_request`: POST `/signature_requests/{{ record.id }}/activate` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  activates a draft signature request, taking it out of draft status; if delivery_mode is not none
  this immediately notifies approvers/signers/followers by email; external mutation, approval
  required.
- `cancel_signature_request`: POST `/signature_requests/{{ record.id }}/cancel` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `reason`; accepted fields
  `custom_note`, `id`, `reason`; confirmation `destructive`; risk: irreversibly cancels a signature
  request in approval or ongoing status; external mutation, approval required.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `first_name`, `last_name`, `email`, `locale`; accepted fields `company_name`, `email`,
  `first_name`, `job_title`, `last_name`, `locale`, `phone_number`, `workspace_id`; risk: creates a
  new saved contact profile; external mutation, approval required.
- `update_contact`: PATCH `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `company_name`, `email`, `first_name`, `id`,
  `job_title`, `last_name`, `phone_number`; risk: mutates an existing contact's profile fields;
  external mutation, approval required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: irreversibly deletes a saved contact profile;
  external mutation, approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `endpoint`, `subscribed_events`, `scopes`, `sandbox`, `auto_retry`, `enabled`; accepted fields
  `auto_retry`, `description`, `enabled`, `endpoint`, `sandbox`, `scopes`, `subscribed_events`,
  `workspaces`; risk: registers a new webhook subscription that will receive real-time event
  notifications at an external endpoint; external mutation, approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: irreversibly deletes a registered webhook
  subscription, silently stopping the caller's own event delivery; external mutation, approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=15, deprecated=1, destructive_admin=21, duplicate_of=16, non_data_endpoint=2,
  out_of_scope=60, requires_elevated_scope=16.
