# Overview

Reads Tally.so forms, form-scoped submissions, webhooks, and workspaces, and writes
form/webhook/workspace mutations through the Tally REST API.

Readable streams: `forms`, `workspaces`, `webhooks`, `submissions`.

Write actions: `create_webhook`, `update_webhook`, `delete_webhook`, `create_form`, `update_form`,
`delete_form`, `delete_submission`, `create_workspace`.

Service API documentation: https://developers.tally.so/api-reference/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Tally personal access token, sent as a Bearer token.
  Generate one from Tally account settings. Never logged.
- `base_url` (optional, string); default `https://api.tally.so`; format `uri`; Tally API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page for forms/workspaces/submissions
  (1-500 per Tally's documented max).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only submissions
  submitted at or after this time are read.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.tally.so`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `forms`: GET `/forms` - records path `items`; query `limit`=`{{ config.page_size }}`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `workspaces`: GET `/workspaces` - records path `items`; query `limit`=`{{ config.page_size }}`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `webhooks`: GET `/webhooks` - records path `webhooks`; query `limit`=`{{ config.page_size }}`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
  Note: the webhooks endpoint caps the page size at 100, while forms and submissions accept up to
  500.
- `submissions`: GET `/forms/{{ fanout.id }}/submissions` - records path `submissions`; query
  `limit`=`{{ config.page_size }}`; page-number pagination; page parameter `page`; size parameter
  `limit`; starts at 1; page size 100; incremental cursor `submitted_at`; sent as `startDate`;
  formatted as `rfc3339`; initial lower bound from `start_date`; computed output fields
  `submitted_at`; fan-out; ids from request `/forms`; id-list records path `items`; id field `id`;
  id inserted into the request path; stamps `form_id`.

## Write actions & risks

Overall write risk: external Tally API mutation (form/webhook/workspace create-update-delete,
submission delete).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `formId`, `url`, `eventTypes`; accepted fields `eventTypes`, `externalSubscriber`, `formId`,
  `httpHeaders`, `signingSecret`, `url`; risk: registers an external endpoint to receive form
  submission events.
- `update_webhook`: PATCH `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `formId`, `url`, `eventTypes`, `isEnabled`; accepted fields
  `eventTypes`, `formId`, `httpHeaders`, `id`, `isEnabled`, `signingSecret`, `url`; risk: changes
  where and whether an existing webhook delivers form submission events.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  stops delivery of form submission events to the webhook's registered endpoint; if this is the
  form's last webhook, the webhooks integration is also marked deleted.
- `create_form`: POST `/forms` - kind `create`; body type `json`; required record fields `blocks`,
  `status`; accepted fields `blocks`, `settings`, `status`, `templateId`, `workspaceId`; risk:
  creates a new live form in the Tally account.
- `update_form`: PATCH `/forms/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `blocks`, `id`, `name`, `settings`, `status`;
  requires at least one changed field alongside the form `id` — requests with no changes are
  rejected; risk: changes a live form's name, status, blocks, or settings.
- `delete_form`: DELETE `/forms/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk: moves a
  form to the trash, stopping new submissions.
- `delete_submission`: DELETE `/forms/{{ record.form_id }}/submissions/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `form_id`, `id`; required record fields `form_id`, `id`;
  accepted fields `form_id`, `id`; confirmation `destructive`; risk: permanently removes a
  respondent's submission and its answers from Tally.
- `create_workspace`: POST `/workspaces` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: creates a new workspace; requires the account to have a Pro
  subscription — Free-tier accounts receive a 403.

## Known limits

- Published rate limit metadata: requests_per_minute=100.
- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=3, duplicate_of=3, non_data_endpoint=1, out_of_scope=17,
  requires_elevated_scope=2.
