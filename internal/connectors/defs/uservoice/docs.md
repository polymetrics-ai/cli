# Overview

Reads suggestions, forums, users, categories, statuses, labels, comments, notes, and teams from the
UserVoice Admin API, and writes suggestion/comment/label/note lifecycle mutations.

Readable streams: `suggestions`, `forums`, `users`, `categories`, `statuses`, `labels`, `comments`,
`notes`, `teams`.

Write actions: `create_suggestion`, `update_suggestion`, `approve_suggestion`, `delete_suggestion`,
`create_comment`, `create_label`, `update_label`, `create_note`.

Service API documentation: https://developer.uservoice.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); UserVoice API bearer token. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://api.uservoice.com`; format `uri`; UserVoice API
  base URL override for tests or proxies.
- `start_date` (optional, string); format `date-time`.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.uservoice.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v2/suggestions`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `suggestions`; page_number: `forums`, `users`, `categories`, `statuses`,
`labels`, `comments`, `notes`, `teams`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `suggestions`: GET `/api/v2/suggestions` - records path `suggestions`; query `start_date` from
  template `{{ config.start_date }}`, omitted when absent; computed output fields `created_at`,
  `id`, `state`, `title`.
- `forums`: GET `/api/v2/admin/forums` - records path `forums`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; sent as `updated_after`; formatted as `rfc3339`.
- `users`: GET `/api/v2/admin/users` - records path `users`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor `updated_at`;
  sent as `updated_after`; formatted as `rfc3339`.
- `categories`: GET `/api/v2/admin/categories` - records path `categories`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; sent as `updated_after`; formatted as `rfc3339`.
- `statuses`: GET `/api/v2/admin/statuses` - records path `statuses`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; sent as `updated_after`; formatted as `rfc3339`.
- `labels`: GET `/api/v2/admin/labels` - records path `labels`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; sent as `updated_after`; formatted as `rfc3339`.
- `comments`: GET `/api/v2/admin/comments` - records path `comments`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor
  `updated_at`; sent as `updated_after`; formatted as `rfc3339`.
- `notes`: GET `/api/v2/admin/notes` - records path `notes`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30; incremental cursor `updated_at`;
  sent as `updated_after`; formatted as `rfc3339`.
- `teams`: GET `/api/v2/admin/teams` - records path `teams`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 30.

## Write actions & risks

Overall write risk: external mutation of UserVoice suggestions (create/update/approve/delete),
comments, labels, and internal notes; suggestion delete is a soft moderation action, not permanent
data loss.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_suggestion`: POST `/api/v2/admin/suggestions` - kind `create`; body type `json`; required
  record fields `title`, `links`; accepted fields `body`, `links`, `title`; risk: creates a new
  customer suggestion (idea); low-risk external mutation, no approval required.
- `update_suggestion`: PUT `/api/v2/admin/suggestions/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `body`, `id`, `title`;
  risk: updates an existing suggestion's title/body; external mutation, no approval required.
- `approve_suggestion`: PUT `/api/v2/admin/suggestions/{{ record.id }}/approve` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  approves (publishes) a pending suggestion, making it publicly visible; no approval required.
- `delete_suggestion`: PUT `/api/v2/admin/suggestions/{{ record.id }}/delete` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: soft-deletes (moderates) a suggestion; UserVoice's own
  API keeps a matching restore endpoint (not modeled here) so this is a reversible moderation
  action, not permanent data loss, but is still marked destructive-shaped for operator awareness.
- `create_comment`: POST `/api/v2/admin/comments` - kind `create`; body type `json`; required record
  fields `body`, `links`; accepted fields `body`, `channel`, `links`; risk: posts a new comment on
  an existing suggestion; low-risk external mutation, no approval required.
- `create_label`: POST `/api/v2/admin/labels` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `can_recommend`, `name`; risk: creates a new label for tagging
  suggestions; low-risk external mutation, no approval required.
- `update_label`: PUT `/api/v2/admin/labels/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `can_recommend`, `id`, `name`; risk:
  updates an existing label's name/settings; external mutation, no approval required.
- `create_note`: POST `/api/v2/admin/notes` - kind `create`; body type `json`; required record
  fields `body`, `links`; accepted fields `body`, `links`; risk: creates an internal (non-public)
  note on a suggestion; low-risk external mutation, no approval required.

## Known limits

- API coverage includes 9 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=19, duplicate_of=39, non_data_endpoint=10, out_of_scope=88,
  requires_elevated_scope=12.
