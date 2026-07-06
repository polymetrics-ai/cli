# Overview

Reads and writes WorkRamp users and groups, and reads guides, resources, and SCORM courses, through
the real WorkRamp Employee Learning Cloud API (app.workramp.com/api/v1).

Readable streams: `users`, `groups`, `courses`, `resources`, `scorm_courses`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_group`, `update_group`.

Service API documentation: https://developers.workramp.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); WorkRamp API key (Integrations page -> API), sent as a
  Bearer token on every request. Belongs to a specific admin user; actions are attributed to that
  user. Never logged.
- `base_url` (optional, string); default `https://app.workramp.com`; format `uri`; WorkRamp API base
  URL. Defaults to the real production host (developers.workramp.com's Getting Started guide); EU
  customers should override to https://app.eu.workramp.com.
- `max_pages` (optional, string); default `1`.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-200 for users; the real API's
  per_page/limit parameter name varies by resource, see docs.md).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.workramp.com`, `max_pages=1`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/users` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
0; page size 100.

Pagination by stream: none: `groups`, `resources`; page_number: `users`, `courses`, `scorm_courses`.

- `users`: GET `/api/v1/users` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 100; emits passthrough records.
- `groups`: GET `/api/v1/groups` - records at response root; emits passthrough records.
- `courses`: GET `/api/v1/guides` - records path `data.guides`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 20; emits passthrough records.
- `resources`: GET `/api/v1/resources` - records path `resources`; emits passthrough records.
- `scorm_courses`: GET `/api/v1/scorms` - records path `data.scorms`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 20; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of WorkRamp users and groups (create/update/delete); actions
are attributed to the API key's owning admin user; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/api/v1/users` - kind `create`; body type `json`; required record fields
  `email`; accepted fields `email`, `groups`, `isAdmin`, `managers`, `name`; risk: creates a
  WorkRamp user account; approval required.
- `update_user`: POST `/api/v1/users/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `email`, `id`, `isAdmin`, `isDeleted`,
  `managerEmails`, `name`, `overwriteManagerEmails`; risk: updates a WorkRamp user account's
  attributes; approval required.
- `delete_user`: DELETE `/api/v1/users/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: permanently deletes a
  WorkRamp user account; approval required.
- `create_group`: POST `/api/v1/groups` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `description`, `name`; risk: creates a WorkRamp group; approval required.
- `update_group`: POST `/api/v1/groups/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  updates a WorkRamp group's attributes; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 5 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=43, requires_elevated_scope=1.
