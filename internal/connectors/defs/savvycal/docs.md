# Overview

Reads SavvyCal events, scheduling links, contacts, time zones, webhooks, and workflows, and writes
scheduling-link and webhook lifecycle mutations, through the SavvyCal API.

Readable streams: `events`, `links`, `contacts`, `time_zones`, `webhooks`, `workflows`,
`workflow_rules`.

Write actions: `create_personal_link`, `create_scope_link`, `update_link`, `delete_link`,
`duplicate_link`, `toggle_link`, `cancel_event`, `create_webhook`, `delete_webhook`.

Service API documentation: https://developers.savvycal.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); SavvyCal API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.savvycal.com`; format `uri`; SavvyCal API base
  URL override for tests or proxies.
- `page_size` (optional, string); default `100`.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.savvycal.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/events`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `events`, `links`, `contacts`, `webhooks`, `workflows`; none:
`time_zones`, `workflow_rules`.

- `events`: GET `/v1/events?page=1&per_page={{ config.page_size }}` - records path `data`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `links`: GET `/v1/links?page=1&per_page={{ config.page_size }}` - records path `data`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `contacts`: GET `/v1/contacts?page=1&per_page={{ config.page_size }}` - records path `data`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `time_zones`: GET `/v1/time_zones` - records path `data`; emits passthrough records.
- `webhooks`: GET `/v1/webhooks?page=1&per_page={{ config.page_size }}` - records path `data`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `workflows`: GET `/v1/workflows?page=1&per_page={{ config.page_size }}` - records path `data`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `workflow_rules`: GET `/v1/workflows/{{ fanout.id }}/rules` - records path `data`; fan-out; ids
  from request `/v1/workflows`; id-list records path `data`; id field `id`; id inserted into the
  request path; stamps `workflow_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external SavvyCal mutations: scheduling link
create/update/delete/duplicate/toggle, event cancellation, webhook subscription create/delete.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_personal_link`: POST `/v1/links` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `duration`, `name`, `slug`, `state`; risk: creates a new scheduling link
  in the authenticated user's personal scope; external mutation, approval required.
- `create_scope_link`: POST `/v1/scopes/{{ record.scope_slug }}/links` - kind `create`; body type
  `json`; path fields `scope_slug`; required record fields `scope_slug`, `name`; accepted fields
  `duration`, `name`, `scope_slug`, `slug`, `state`; risk: creates a new scheduling link under a
  specific team or individual scope; external mutation, approval required.
- `update_link`: PATCH `/v1/links/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `duration`, `id`, `name`, `slug`; risk:
  external mutation updating an existing scheduling link's name/slug/duration; approval required.
- `delete_link`: DELETE `/v1/links/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: destructive/irreversible: permanently deletes a scheduling link; approval
  required.
- `duplicate_link`: POST `/v1/links/{{ record.id }}/duplicate` - kind `create`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: creates a copy of an
  existing scheduling link; low-risk external mutation, no approval required.
- `toggle_link`: POST `/v1/links/{{ record.id }}/toggle` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: flips a scheduling link
  between active and disabled state, changing its public bookability; approval required.
- `cancel_event`: POST `/v1/events/{{ record.id }}/cancel` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: destructive/irreversible:
  cancels a scheduled event, notifying attendees; approval required.
- `create_webhook`: POST `/v1/webhooks` - kind `create`; body type `json`; required record fields
  `url`; accepted fields `events`, `url`; risk: creates a new webhook subscription that will POST
  event notifications to an external URL; approval required.
- `delete_webhook`: DELETE `/v1/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: destructive/irreversible: permanently deletes a webhook subscription;
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=4, non_data_endpoint=1, out_of_scope=2.
