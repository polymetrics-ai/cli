# Overview

Reads DataScope locations, form answers, lists, notifications, task assignments, tickets (findings),
and generated files, and writes location/list/task-assignment/form-answer mutations, through the
DataScope external REST API (full-refresh).

Readable streams: `locations`, `answers`, `lists`, `notifications`, `task_assigns`, `findings`,
`files`.

Write actions: `create_location`, `update_location`, `assign_task`, `create_metadata_object`,
`update_metadata_object`, `bulk_update_metadata_objects`, `create_metadata_type`,
`update_metadata_type`, `change_form_answer`.

Service API documentation: https://app.mydatascope.com/api/external/docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); DataScope API key, sent as the raw Authorization header
  value (no Bearer/Basic prefix). Never logged.
- `base_url` (optional, string); default `https://www.mydatascope.com/api/external`; format `uri`;
  DataScope API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://www.mydatascope.com/api/external`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/locations`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 200.

Pagination by stream: none: `files`; offset_limit: `locations`, `answers`, `lists`, `notifications`,
`task_assigns`, `findings`.

- `locations`: GET `/locations` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 200.
- `answers`: GET `/v2/answers` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 200.
- `lists`: GET `/metadata_objects` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 200.
- `notifications`: GET `/notifications` - records path `.`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 200.
- `task_assigns`: GET `/task_assigns` - records path `task_assigns`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 200.
- `findings`: GET `/findings/list` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 200.
- `files`: GET `/files` - records path `.`.

## Write actions & risks

Overall write risk: external mutation of DataScope locations, lists (metadata objects/types), task
assignments, and previously-submitted form answers; change_form_answer rewrites collected field data
after the fact and bulk_update_metadata_objects affects many list elements in one call, so every
write ships an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_location`: POST `/locations` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `address`, `city`, `code`, `company_code`, `company_name`, `country`,
  `description`, `email`, `latitude`, `longitude`, `name`, `phone`; risk: creates a new
  field-data-collection location record; low-risk external mutation, no approval required.
- `update_location`: POST `/locations/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address`, `city`, `code`,
  `company_code`, `company_name`, `country`, `description`, `email`, `id`, `latitude`, `longitude`,
  `name`, `phone`; risk: mutates an existing location's address/contact metadata; external mutation,
  approval required.
- `assign_task`: POST `/assign_task` - kind `create`; body type `json`; required record fields
  `form_id`, `user_id`, `date`; accepted fields `c_code`, `c_name`, `code`, `date`, `form_id`,
  `gap`, `l_code`, `l_email`, `l_phone`, `latitude`, `location_address`, `location_name`,
  `longitude`, `task_instruction`, `user_id`; risk: assigns a new field task/inspection to a user
  for a scheduled date; low-risk external mutation, no approval required.
- `create_metadata_object`: POST `/metadata_object` - kind `create`; body type `json`; required
  record fields `metadata_type`, `name`; accepted fields `attribute1`, `attribute2`, `code`,
  `description`, `metadata_type`, `name`; risk: creates a new list (metadata object) element;
  low-risk external mutation, no approval required.
- `update_metadata_object`: POST `/metadata_object/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `attribute1`, `attribute2`,
  `code`, `description`, `id`, `name`; risk: mutates an existing list element's fields; external
  mutation, approval required.
- `bulk_update_metadata_objects`: POST `/metadata_objects/bulk_update` - kind `update`; body type
  `json`; required record fields `metadata_type`, `list_objects`; accepted fields `list_objects`,
  `metadata_type`, `name`; risk: replaces/updates many list elements of one metadata_type in a
  single call; higher blast radius than a single-object update, approval required.
- `create_metadata_type`: POST `/metadata_types` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `code`, `description`, `list_type`, `name`; risk: creates a new
  empty list (metadata type/category); low-risk external mutation, no approval required.
- `update_metadata_type`: POST `/metadata_types/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `code`, `description`, `id`,
  `list_type`, `name`; risk: renames/reconfigures an existing list definition; every list element
  under it is affected, external mutation, approval required.
- `change_form_answer`: POST `/change_form_answer` - kind `update`; body type `json`; required
  record fields `form_name`, `form_code`, `question_name`, `question_value`; accepted fields
  `form_code`, `form_name`, `question_name`, `question_value`, `subform_index`; risk: overwrites a
  previously-submitted form answer's value in place, rewriting collected field data after the fact;
  external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 7 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=4, non_data_endpoint=1.
