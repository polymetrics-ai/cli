# Overview

Reads and manages Amplitude behavioral cohorts, chart annotations, annotation categories, event
lists, and the governed taxonomy (event/category definitions) through the Amplitude Analytics REST
API.

Readable streams: `cohorts`, `cohorts_usage`, `annotations`, `annotation_categories`, `events_list`,
`taxonomy_categories`, `taxonomy_events`, `taxonomy_event_properties`, `taxonomy_user_properties`,
`taxonomy_group_properties`.

Write actions: `create_annotation`, `update_annotation`, `delete_annotation`,
`create_annotation_category`, `update_annotation_category`, `delete_annotation_category`,
`create_taxonomy_category`, `update_taxonomy_category`, `delete_taxonomy_category`,
`create_taxonomy_event`, `update_taxonomy_event`, `delete_taxonomy_event`.

Service API documentation: https://www.docs.developers.amplitude.com/analytics/apis/http-v2-api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Amplitude project API key. Used as the HTTP Basic auth
  username; never logged.
- `base_url` (optional, string); default `https://amplitude.com`; format `uri`; Amplitude Analytics
  REST API base URL. Defaults to the Standard-server host; set explicitly to
  https://analytics.eu.amplitude.com for EU-residency projects (see docs.md Known limits).
- `mode` (optional, string).
- `secret_key` (required, secret, string); Amplitude project secret key. Used as the HTTP Basic auth
  password; never logged.
- `taxonomy_show_deleted` (optional, string); Optional 'true'/'false' string forwarded as the
  showDeleted query parameter on the taxonomy_events and taxonomy_user_properties streams. Left
  unset, the parameter is omitted and Amplitude applies its own default (deleted items excluded).

Secret fields are redacted in logs and write previews: `api_key`, `secret_key`.

Default configuration values: `base_url=https://amplitude.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.secret_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/3/cohorts`.

## Streams notes

Default pagination: single request; no pagination.

- `cohorts`: GET `/api/3/cohorts` - records path `cohorts`.
- `cohorts_usage`: GET `/api/3/cohorts/usage` - records at response root; computed output fields
  `limit`, `resets_at`, `usage`.
- `annotations`: GET `/api/3/annotations` - records path `data`.
- `annotation_categories`: GET `/api/3/annotation-categories` - records path `data`.
- `events_list`: GET `/api/2/events/list` - records path `data`.
- `taxonomy_categories`: GET `/api/2/taxonomy/category` - records path `data`.
- `taxonomy_events`: GET `/api/2/taxonomy/event` - records path `data`; query `showDeleted` from
  template `{{ config.taxonomy_show_deleted }}`, omitted when absent.
- `taxonomy_event_properties`: GET `/api/2/taxonomy/event-property` - records path `data`.
- `taxonomy_user_properties`: GET `/api/2/taxonomy/user-property` - records path `data`; query
  `showDeleted` from template `{{ config.taxonomy_show_deleted }}`, omitted when absent.
- `taxonomy_group_properties`: GET `/api/2/taxonomy/group-property` - records path `data`.

## Write actions & risks

Overall write risk: external Amplitude API mutation of chart annotations, annotation categories, and
governed taxonomy event/category definitions - never behavioral event data itself.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_annotation`: POST `/api/3/annotations` - kind `create`; body type `json`; required record
  fields `label`, `start`; accepted fields `category`, `chart_id`, `details`, `end`, `label`,
  `start`; risk: creates a chart annotation visible to every Amplitude project user.
- `update_annotation`: PUT `/api/3/annotations/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `category`, `chart_id`, `details`,
  `end`, `id`, `label`, `start`; risk: mutates an existing chart annotation visible to every
  Amplitude project user.
- `delete_annotation`: DELETE `/api/3/annotations/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes a chart annotation.
- `create_annotation_category`: POST `/api/3/annotation-categories` - kind `create`; body type
  `json`; required record fields `category`; accepted fields `category`; risk: creates a new
  annotation category shared across the Amplitude project.
- `update_annotation_category`: PUT `/api/3/annotation-categories/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `category`; accepted fields
  `category`, `id`; risk: renames an existing annotation category shared across the Amplitude
  project.
- `delete_annotation_category`: DELETE `/api/3/annotation-categories/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently deletes an annotation
  category shared across the Amplitude project.
- `create_taxonomy_category`: POST `/api/2/taxonomy/category` - kind `create`; body type `json`;
  required record fields `category_name`; accepted fields `category_name`; risk: creates a new event
  category in the Amplitude project's governed taxonomy.
- `update_taxonomy_category`: PUT `/api/2/taxonomy/category/{{ record.category_id }}` - kind
  `update`; body type `json`; path fields `category_id`; required record fields `category_id`,
  `category_name`; accepted fields `category_id`, `category_name`; risk: renames an existing event
  category in the Amplitude project's governed taxonomy.
- `delete_taxonomy_category`: DELETE `/api/2/taxonomy/category/{{ record.category_id }}` - kind
  `delete`; body type `none`; path fields `category_id`; required record fields `category_id`;
  accepted fields `category_id`; missing records treated as success for status `404`; risk:
  permanently deletes an event category from the Amplitude project's governed taxonomy.
- `create_taxonomy_event`: POST `/api/2/taxonomy/event` - kind `create`; body type `json`; required
  record fields `event_type`; accepted fields `category`, `description`, `event_type`, `is_active`,
  `owner`, `tags`; risk: registers a new governed event type in the Amplitude project's taxonomy.
- `update_taxonomy_event`: PUT `/api/2/taxonomy/event/{{ record.event_type }}` - kind `update`; body
  type `json`; path fields `event_type`; required record fields `event_type`; accepted fields
  `category`, `description`, `display_name`, `event_type`, `is_active`, `new_event_type`, `owner`,
  `tags`; risk: mutates an existing governed event type's taxonomy metadata.
- `delete_taxonomy_event`: DELETE `/api/2/taxonomy/event/{{ record.event_type }}` - kind `delete`;
  body type `none`; path fields `event_type`; required record fields `event_type`; accepted fields
  `event_type`; missing records treated as success for status `404`; risk: soft-deletes a governed
  event type from the Amplitude project's taxonomy (recoverable via the restore endpoint, not
  modeled as a separate write action).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, duplicate_of=7, non_data_endpoint=1, out_of_scope=14, requires_elevated_scope=2.
