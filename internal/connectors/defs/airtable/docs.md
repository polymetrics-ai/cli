# Overview

Reads Airtable bases, tables, records, webhooks, and record comments, and writes
record/table/field/comment/webhook mutations, through the Airtable Web API.

Readable streams: `bases`, `tables`, `records`, `webhooks`, `comments`.

Write actions: `create_record`, `update_record`, `delete_record`, `create_table`, `update_table`,
`create_field`, `update_field`, `create_comment`, `update_comment`, `delete_comment`,
`create_webhook`, `delete_webhook`.

Service API documentation: https://airtable.com/developers/web/api/introduction.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Airtable OAuth2 access token, sent as a Bearer token
  when api_key is unset. Never logged.
- `api_key` (optional, secret, string); Airtable Personal Access Token, sent as a Bearer token.
  Preferred over access_token when both are set. Never logged.
- `base_id` (optional, string); Airtable base id (required for the 'tables' and 'records' streams).
- `base_url` (optional, string); default `https://api.airtable.com/v0`; format `uri`; Airtable API
  base URL override for tests or proxies.
- `page_size` (optional, string); default `100`; Records per page for the 'records' stream (1-100).
- `table_id` (optional, string); Airtable table id or name (required for the 'records' stream).

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Default configuration values: `base_url=https://api.airtable.com/v0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key` when `{{ secrets.api_key }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/meta/bases`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `offset`; next token from `offset`.

- `bases`: GET `/meta/bases` - records path `bases`; cursor pagination; cursor parameter `offset`;
  next token from `offset`.
- `tables`: GET `/meta/bases/{{ config.base_id }}/tables` - records path `tables`; cursor
  pagination; cursor parameter `offset`; next token from `offset`.
- `records`: GET `/{{ config.base_id }}/{{ config.table_id }}` - records path `records`; query
  `pageSize`=`{{ config.page_size }}`; cursor pagination; cursor parameter `offset`; next token from
  `offset`.
- `webhooks`: GET `/bases/{{ config.base_id }}/webhooks` - records path `webhooks`; cursor
  pagination; cursor parameter `offset`; next token from `offset`.
- `comments`: GET `/{{ config.base_id }}/{{ config.table_id }}/{{ fanout.id }}/comments` - records
  path `comments`; cursor pagination; cursor parameter `offset`; next token from `offset`; fan-out;
  ids from request `/{{ config.base_id }}/{{ config.table_id }}`; id-list records path `records`; id
  field `id`; id inserted into the request path; stamps `record_id`.

## Write actions & risks

Overall write risk: external Airtable API mutation of records, table/field schema, comments, and
webhooks; schema mutations (create_table/update_table/create_field/update_field) are visible to
every base collaborator, approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_record`: POST `/{{ config.base_id }}/{{ config.table_id }}` - kind `create`; body type
  `json`; required record fields `fields`; accepted fields `fields`; risk: creates a new record in
  the configured base/table; low-risk external mutation, no approval required.
- `update_record`: PATCH `/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `fields`; accepted
  fields `fields`, `id`; risk: mutates only the field values included in the request
  (non-destructive PATCH); unincluded cell values are left unchanged.
- `delete_record`: DELETE `/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently removes a record from the
  base/table; irreversible.
- `create_table`: POST `/meta/bases/{{ config.base_id }}/tables` - kind `create`; body type `json`;
  required record fields `name`, `fields`; accepted fields `description`, `fields`, `name`; risk:
  creates a new table (schema mutation) in the configured base; low-risk but changes the base's
  structure, visible to every collaborator.
- `update_table`: PATCH `/meta/bases/{{ config.base_id }}/tables/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `description`,
  `id`, `name`; risk: renames or redescribes an existing table; a visible schema change for every
  collaborator on the base.
- `create_field`: POST `/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields` - kind
  `create`; body type `json`; path fields `table_id`; required record fields `table_id`, `name`,
  `type`; accepted fields `description`, `name`, `table_id`, `type`; risk: creates a new column
  (schema mutation) in the target table; low-risk but changes the table's structure, visible to
  every collaborator.
- `update_field`: PATCH `/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields/{{
  record.id }}` - kind `update`; body type `json`; path fields `table_id`, `id`; required record
  fields `table_id`, `id`; accepted fields `description`, `id`, `name`, `table_id`; risk: renames or
  redescribes an existing column; a visible schema change for every collaborator on the base.
- `create_comment`: POST `/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id
  }}/comments` - kind `create`; body type `json`; path fields `record_id`; required record fields
  `record_id`, `text`; accepted fields `parentCommentId`, `record_id`, `text`; risk: adds a visible
  comment to a record; every base collaborator with record access can see it, no external side
  effect.
- `update_comment`: PATCH `/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id
  }}/comments/{{ record.id }}` - kind `update`; body type `json`; path fields `record_id`, `id`;
  required record fields `record_id`, `id`, `text`; accepted fields `id`, `record_id`, `text`; risk:
  edits the text of an existing comment; visible to every base collaborator with record access.
- `delete_comment`: DELETE `/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id
  }}/comments/{{ record.id }}` - kind `delete`; body type `none`; path fields `record_id`, `id`;
  required record fields `record_id`, `id`; accepted fields `id`, `record_id`; missing records
  treated as success for status `404`; risk: permanently removes a comment from a record;
  irreversible.
- `create_webhook`: POST `/bases/{{ config.base_id }}/webhooks` - kind `create`; body type `json`;
  required record fields `notificationUrl`, `specification`; accepted fields `notificationUrl`,
  `specification`; risk: registers a new outbound webhook that will POST live base-change
  notifications to an external URL of the caller's choosing; verify the target endpoint before
  enabling.
- `delete_webhook`: DELETE `/bases/{{ config.base_id }}/webhooks/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: permanently removes a webhook subscription;
  notification delivery to its target URL stops immediately.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=5, non_data_endpoint=1, out_of_scope=4,
  requires_elevated_scope=1.
