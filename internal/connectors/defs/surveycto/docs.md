# Overview

Reads SurveyCTO form IDs, submissions, datasets (including case-management datasets), dataset
records, groups, roles, teams, and users, and writes dataset lifecycle mutations, dataset record
creation, and user lifecycle mutations, through the SurveyCTO Server API v2.

Readable streams: `datasets`, `dataset_records`, `submissions`, `groups`, `roles`, `users`.

Write actions: `create_dataset`, `update_dataset`, `delete_dataset`, `create_dataset_record`,
`create_user`, `update_user`, `delete_user`.

Service API documentation: https://developer.surveycto.com/.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; SurveyCTO API base URL, e.g.
  https://<server_name>.surveycto.com/api/v2.
- `form_id` (optional, string); SurveyCTO form ID; required for the submissions stream, which is
  scoped to a single form (the real API has no 'list all form ids' stream-friendly endpoint -- see
  docs.md Known limits).
- `mode` (optional, string).
- `password` (required, secret, string); SurveyCTO password or API key, sent as the HTTP Basic auth
  password. Never logged.
- `server_name` (optional, string); Bare SurveyCTO server name (no scheme/path), for
  documentation/reference only -- not wired into any template.
- `username` (required, secret, string); SurveyCTO account username/email, sent as the HTTP Basic
  auth username. Never logged.

Secret fields are redacted in logs and write previews: `password`, `username`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/datasets` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `nextCursor`.

- `datasets`: GET `/datasets` - records path `data`; query `limit`=`1000`; cursor pagination; cursor
  parameter `cursor`; next token from `nextCursor`.
- `dataset_records`: GET `/datasets/{{ fanout.id }}/records` - records path `data`; query
  `limit`=`1000`; cursor pagination; cursor parameter `cursor`; next token from `nextCursor`;
  fan-out; ids from request `/datasets`; id-list records path `data`; id field `id`; id inserted
  into the request path; stamps `dataset_id`.
- `submissions`: GET `/forms/{{ config.form_id }}/submissions` - records path `data`; query
  `limit`=`1000`; cursor pagination; cursor parameter `cursor`; next token from `nextCursor`;
  computed output fields `form_id`, `id`.
- `groups`: GET `/groups` - records path `data`; query `limit`=`1000`; cursor pagination; cursor
  parameter `cursor`; next token from `nextCursor`.
- `roles`: GET `/roles` - records path `data`; query `limit`=`1000`; cursor pagination; cursor
  parameter `cursor`; next token from `nextCursor`.
- `users`: GET `/users` - records path `data`; query `limit`=`1000`; cursor pagination; cursor
  parameter `cursor`; next token from `nextCursor`.

## Write actions & risks

Overall write risk: external SurveyCTO API mutation (dataset lifecycle, dataset record creation,
user lifecycle including password-setting).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_dataset`: POST `/datasets` - kind `create`; body type `json`; body fields `id`, `title`,
  `discriminator`, `uniqueRecordField`, `allowOfflineUpdates`; required record fields
  `discriminator`; accepted fields `allowOfflineUpdates`, `discriminator`, `id`, `title`,
  `uniqueRecordField`; risk: creates a new server dataset (a general-purpose, enumerator, or
  case-management dataset); low-risk external mutation, no approval required.
- `update_dataset`: PUT `/datasets/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `title`, `discriminator`, `uniqueRecordField`, `allowOfflineUpdates`; required
  record fields `id`, `discriminator`; accepted fields `allowOfflineUpdates`, `discriminator`, `id`,
  `title`, `uniqueRecordField`; risk: updates an existing dataset's metadata/configuration (the
  dataset type/discriminator itself cannot be changed after creation, per SurveyCTO's own API);
  external mutation, no approval required.
- `delete_dataset`: DELETE `/datasets/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversibly deletes a dataset and its records; approval required.
- `create_dataset_record`: POST `/datasets/{{ record.dataset_id }}/records` - kind `create`; body
  type `json`; path fields `dataset_id`; required record fields `dataset_id`; accepted fields
  `dataset_id`; risk: adds a new record to a dataset; the field name set is dataset-defined
  (SurveyCTO's own DatasetRecordFieldMap has no fixed schema), so record_schema only requires the
  routing field dataset_id -- every other record property is sent verbatim as the record's
  field-name/value map; low-risk external mutation, no approval required.
- `create_user`: POST `/users` - kind `create`; body type `json`; required record fields `username`,
  `roleId`, `password`; accepted fields `password`, `roleId`, `username`; risk: creates a new
  SurveyCTO server user AND sets their initial password in the same call; a credential-provisioning
  action, not an ordinary data mutation -- approval required.
- `update_user`: PUT `/users/{{ record.username }}` - kind `update`; body type `json`; path fields
  `username`; required record fields `username`; accepted fields `password`, `roleId`, `username`;
  risk: updates an existing user's password and/or role; a credential-provisioning action when
  password is set -- approval required.
- `delete_user`: DELETE `/users/{{ record.username }}` - kind `delete`; body type `none`; path
  fields `username`; required record fields `username`; accepted fields `username`; missing records
  treated as success for status `404`; risk: irreversibly deletes a server user and revokes their
  access; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=2, duplicate_of=4, non_data_endpoint=1, out_of_scope=2,
  requires_elevated_scope=12.
