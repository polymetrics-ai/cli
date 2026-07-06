# Overview

Reads Coassemble courses, screen types, collections, clients, users, learner tracking, and
translations, and writes course/collection/client/user/translation lifecycle actions, through the
Coassemble headless REST API.

Readable streams: `courses`, `screen_types`, `trackings`, `collections`, `clients`, `users`,
`user_trackings`, `collection_trackings`, `translations`.

Write actions: `publish_course`, `duplicate_course`, `delete_course`, `restore_course`,
`delete_tracking`, `create_collection`, `delete_collection`, `restore_collection`, `update_client`,
`delete_client`, `update_user`, `delete_user`, `translate_course`, `set_default_translation`,
`sync_translation`, `delete_translation`.

Service API documentation: https://developers.coassemble.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.coassemble.com`; format `uri`; Coassemble API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `user_id` (required, secret, string); Coassemble headless API user id. Sent only inside the
  Authorization header; never logged.
- `user_token` (required, secret, string); Coassemble headless API user token. Sent only inside the
  Authorization header; never logged.

Secret fields are redacted in logs and write previews: `user_id`, `user_token`.

Default configuration values: `base_url=https://api.coassemble.com`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.user_id`, `secrets.user_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/headless/courses`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `length`; starts
at 1; page size 20.

Pagination by stream: none: `screen_types`, `translations`; page_number: `courses`, `trackings`,
`collections`, `clients`, `users`, `user_trackings`, `collection_trackings`.

- `courses`: GET `/api/v1/headless/courses` - records at response root; page-number pagination; page
  parameter `page`; size parameter `length`; starts at 1; page size 20.
- `screen_types`: GET `/api/v1/headless/screen/types` - records at response root.
- `trackings`: GET `/api/v1/headless/trackings` - records at response root; page-number pagination;
  page parameter `page`; size parameter `length`; starts at 1; page size 20.
- `collections`: GET `/api/v1/headless/collections` - records at response root; page-number
  pagination; page parameter `page`; size parameter `length`; starts at 0; page size 100.
- `clients`: GET `/api/v1/headless/clients` - records at response root; page-number pagination; page
  parameter `page`; size parameter `length`; starts at 0; page size 100.
- `users`: GET `/api/v1/headless/users` - records at response root; page-number pagination; page
  parameter `page`; size parameter `length`; starts at 0; page size 100.
- `user_trackings`: GET `/api/v1/headless/user/trackings` - records at response root; page-number
  pagination; page parameter `page`; size parameter `length`; starts at 0; page size 100; fan-out;
  ids from request `/api/v1/headless/users`; id field `identifier`; id sent as query parameter
  `identifier`; stamps `identifier`.
- `collection_trackings`: GET `/api/v1/headless/collection/trackings` - records at response root;
  page-number pagination; page parameter `page`; size parameter `length`; starts at 0; page size
  100; fan-out; ids from request `/api/v1/headless/collections`; id field `id`; id sent as query
  parameter `id`; stamps `collection_id`.
- `translations`: GET `/api/v1/headless/translations/{{ fanout.id }}` - records at response root;
  fan-out; ids from request `/api/v1/headless/courses`; id field `id`; id inserted into the request
  path; stamps `course_id`.

## Write actions & risks

Overall write risk: external mutation of Coassemble courses, collections, clients, users, and
translations (publish/duplicate/restore/delete a course; delete a tracking record;
create/delete/restore a collection; update/delete a client; update/delete a user;
translate/set-default/sync/delete a course translation).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `publish_course`: POST `/api/v1/headless/course/{{ record.id }}/publish` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: publishes
  the current draft of a course, making it live for learners; no approval required.
- `duplicate_course`: POST `/api/v1/headless/course/{{ record.id }}/duplicate` - kind `create`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `clientIdentifier`,
  `id`, `identifier`; risk: creates a full copy of an existing course; low-risk external mutation,
  no approval required.
- `delete_course`: DELETE `/api/v1/headless/course/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: soft-deletes a course (recoverable via restore_course
  within Coassemble's retention window); approval required.
- `restore_course`: POST `/api/v1/headless/course/{{ record.id }}/restore` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: restores a
  previously soft-deleted course; no approval required.
- `delete_tracking`: DELETE `/api/v1/headless/tracking` - kind `delete`; body type `json`; body
  fields `id`, `identifier`; required record fields `id`, `identifier`; accepted fields `id`,
  `identifier`; missing records treated as success for status `404`; risk: permanently erases one
  learner's tracking/progress record for a course; irreversible, approval required.
- `create_collection`: POST `/api/v1/headless/collection` - kind `create`; body type `json`;
  required record fields `title`; accepted fields `clientIdentifier`, `courseIds`, `description`,
  `identifier`, `themeId`, `title`; risk: creates a new collection of courses; low-risk external
  mutation, no approval required.
- `delete_collection`: DELETE `/api/v1/headless/collection/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: soft-deletes a collection (recoverable via
  restore_collection); approval required.
- `restore_collection`: POST `/api/v1/headless/collection/{{ record.id }}/restore` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  restores a previously soft-deleted collection; no approval required.
- `update_client`: PUT `/api/v1/headless/client/{{ record.clientIdentifier }}` - kind `update`; body
  type `json`; path fields `clientIdentifier`; body fields `metadata`; required record fields
  `clientIdentifier`; accepted fields `clientIdentifier`, `metadata`; risk: overwrites a client's
  arbitrary metadata bag; no approval required.
- `delete_client`: DELETE `/api/v1/headless/client/{{ record.clientIdentifier }}` - kind `delete`;
  body type `none`; path fields `clientIdentifier`; required record fields `clientIdentifier`;
  accepted fields `clientIdentifier`; missing records treated as success for status `404`; risk:
  irreversibly removes a client (multi-tenant sub-account) and its documented on-delete effects on
  associated users; approval required.
- `update_user`: PUT `/api/v1/headless/user/{{ record.identifier }}` - kind `update`; body type
  `json`; path fields `identifier`; body fields `clientIdentifier`, `metadata`, `name`, `avatar`;
  required record fields `identifier`; accepted fields `avatar`, `clientIdentifier`, `identifier`,
  `metadata`, `name`; risk: overwrites a learner's profile fields (name/avatar/metadata) or
  reassigns their client; no approval required.
- `delete_user`: DELETE `/api/v1/headless/user/{{ record.identifier }}` - kind `delete`; body type
  `none`; path fields `identifier`; required record fields `identifier`; accepted fields
  `identifier`; missing records treated as success for status `404`; risk: irreversibly removes a
  learner identity, applying Coassemble's server-side DEFAULT handling for that identity's course
  progress (the real endpoint also accepts optional
  action=reallocate|delete|ignore/reallocateTo/clientIdentifier query params to control that
  handling explicitly, and Coassemble's own docs do not fully specify their exact semantics beyond
  "choose what to do with any courses associated with this identifier" - this action deliberately
  does not expose them, since the write-action path/query dialect has no way to send an optional
  record field only when present, and silently defaulting an ambiguous, irreversible
  per-learner-data-retention choice would be worse than declaring it out of scope; approval
  required.
- `translate_course`: POST `/api/v1/headless/translation/translate/{{ record.course_id }}` - kind
  `create`; body type `json`; path fields `course_id`; body fields `language`; required record
  fields `course_id`, `language`; accepted fields `course_id`, `language`; risk: kicks off machine
  translation of a course into a new BCP-47 language variant; low-risk external mutation, no
  approval required.
- `set_default_translation`: POST `/api/v1/headless/translation/default/{{ record.course_id }}/{{
  record.language }}` - kind `update`; body type `none`; path fields `course_id`, `language`;
  required record fields `course_id`, `language`; accepted fields `course_id`, `language`; risk:
  changes which language variant learners see by default for this course; no approval required.
- `sync_translation`: POST `/api/v1/headless/translation/sync/{{ record.course_id }}/{{
  record.language }}` - kind `update`; body type `none`; path fields `course_id`, `language`;
  required record fields `course_id`, `language`; accepted fields `course_id`, `language`; risk:
  re-syncs a translated variant's content with upstream changes to the source-language course, which
  can overwrite manual edits made directly in the translated variant; no approval required.
- `delete_translation`: DELETE `/api/v1/headless/translation/{{ record.course_id }}/{{
  record.language }}` - kind `delete`; body type `none`; path fields `course_id`, `language`;
  required record fields `course_id`, `language`; accepted fields `course_id`, `language`; missing
  records treated as success for status `404`; risk: permanently removes a language variant of a
  course; irreversible, approval required.

## Known limits

- Batch defaults: read_page_size=20.
- API coverage includes 9 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, deprecated=3, duplicate_of=6, out_of_scope=3.
