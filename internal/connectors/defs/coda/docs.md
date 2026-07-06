# Overview

Reads Coda docs and doc-scoped tables, rows, columns, pages, formulas, and controls, and writes
rows/pages, through the Coda REST API v1.

Readable streams: `docs`, `tables`, `pages`, `formulas`, `controls`, `columns`, `rows`.

Write actions: `upsert_rows`, `update_row`, `delete_row`, `delete_rows`, `push_button`,
`create_page`, `update_page`, `delete_page`.

Service API documentation: https://coda.io/developers/apis/v1.

## Auth setup

Connection fields:

- `auth_token` (required, secret, string); Coda API token. Sent only as Authorization: Bearer
  <auth_token>; never logged.
- `base_url` (optional, string); default `https://coda.io/apis/v1`; format `uri`; Coda API base URL
  override for tests or proxies.
- `doc_id` (optional, string); Coda doc id. Required for the doc-scoped streams (tables, pages,
  formulas, controls); ignored by the workspace-level docs stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `25`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `auth_token`.

Default configuration values: `base_url=https://coda.io/apis/v1`, `page_size=25`.

Authentication behavior:

- Bearer token authentication using `secrets.auth_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/docs`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

- `docs`: GET `/docs` - records path `items`; query `limit` from template `{{ config.page_size }}`,
  default `25`; cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `tables`: GET `/docs/{{ config.doc_id }}/tables` - records path `items`; query `limit` from
  template `{{ config.page_size }}`, default `25`; cursor pagination; cursor parameter `pageToken`;
  next token from `nextPageToken`; computed output fields `doc_id`.
- `pages`: GET `/docs/{{ config.doc_id }}/pages` - records path `items`; query `limit` from template
  `{{ config.page_size }}`, default `25`; cursor pagination; cursor parameter `pageToken`; next
  token from `nextPageToken`; computed output fields `doc_id`.
- `formulas`: GET `/docs/{{ config.doc_id }}/formulas` - records path `items`; query `limit` from
  template `{{ config.page_size }}`, default `25`; cursor pagination; cursor parameter `pageToken`;
  next token from `nextPageToken`; computed output fields `doc_id`.
- `controls`: GET `/docs/{{ config.doc_id }}/controls` - records path `items`; query `limit` from
  template `{{ config.page_size }}`, default `25`; cursor pagination; cursor parameter `pageToken`;
  next token from `nextPageToken`; computed output fields `doc_id`.
- `columns`: GET `/docs/{{ config.doc_id }}/tables/{{ fanout.id }}/columns` - records path `items`;
  query `limit` from template `{{ config.page_size }}`, default `25`; cursor pagination; cursor
  parameter `pageToken`; next token from `nextPageToken`; computed output fields `doc_id`; fan-out;
  ids from request `/docs/{{ config.doc_id }}/tables`; id-list records path `items`; id field `id`;
  id inserted into the request path; stamps `table_id`.
- `rows`: GET `/docs/{{ config.doc_id }}/tables/{{ fanout.id }}/rows` - records path `items`; query
  `limit` from template `{{ config.page_size }}`, default `25`; `valueFormat`=`simpleWithArrays`;
  cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`; computed output
  fields `doc_id`; fan-out; ids from request `/docs/{{ config.doc_id }}/tables`; id-list records
  path `items`; id field `id`; id inserted into the request path; stamps `table_id`.

## Write actions & risks

Overall write risk: external mutation of Coda table rows and doc pages (insert/upsert/update/delete
rows, push a row button, create/update/delete a page); push_button and delete actions are
approval-gated per writes.json risk text.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `upsert_rows`: POST `/docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows` - kind `create`;
  body type `json`; path fields `table_id`; body fields `rows`, `keyColumns`; required record fields
  `table_id`, `rows`; accepted fields `keyColumns`, `rows`, `table_id`; risk: inserts new rows, or
  upserts existing ones when keyColumns is set, into a Coda table; queued for async processing (202)
  and generally applied within seconds.
- `update_row`: PUT `/docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id
  }}` - kind `update`; body type `json`; path fields `table_id`, `row_id`; body fields `row`;
  required record fields `table_id`, `row_id`, `row`; accepted fields `row`, `row_id`, `table_id`;
  risk: overwrites cell values on an existing row; queued for async processing (202) and generally
  applied within seconds.
- `delete_row`: DELETE `/docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id
  }}` - kind `delete`; body type `none`; path fields `table_id`, `row_id`; required record fields
  `table_id`, `row_id`; accepted fields `row_id`, `table_id`; missing records treated as success for
  status `404`; risk: permanently removes a row from a Coda table; irreversible, queued for async
  processing (202).
- `delete_rows`: DELETE `/docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows` - kind
  `delete`; body type `json`; path fields `table_id`; body fields `rowIds`; required record fields
  `table_id`, `rowIds`; accepted fields `rowIds`, `table_id`; missing records treated as success for
  status `404`; risk: permanently removes multiple rows from a Coda table in one request;
  irreversible, queued for async processing (202).
- `push_button`: POST `/docs/{{ config.doc_id }}/tables/{{ record.table_id }}/rows/{{ record.row_id
  }}/buttons/{{ record.column_id }}` - kind `update`; body type `none`; path fields `table_id`,
  `row_id`, `column_id`; required record fields `table_id`, `row_id`, `column_id`; accepted fields
  `column_id`, `row_id`, `table_id`; risk: pushes a button on a row; the underlying button can
  perform ANY action the doc's formulas define, including writes to other tables and Pack actions
  outside this connector's declared surface - high blast-radius, approval required.
- `create_page`: POST `/docs/{{ config.doc_id }}/pages` - kind `create`; body type `json`; accepted
  fields `iconName`, `imageUrl`, `name`, `parentPageId`, `subtitle`; risk: creates a new page in the
  configured doc; requires Doc Maker access in the workspace, queued for async processing (202).
- `update_page`: PUT `/docs/{{ config.doc_id }}/pages/{{ record.page_id }}` - kind `update`; body
  type `json`; path fields `page_id`; required record fields `page_id`; accepted fields `iconName`,
  `imageUrl`, `isHidden`, `name`, `page_id`, `subtitle`; risk: renames, hides, or restyles an
  existing page; renaming/re-iconing requires Doc Maker access in the workspace, queued for async
  processing (202).
- `delete_page`: DELETE `/docs/{{ config.doc_id }}/pages/{{ record.page_id }}` - kind `delete`; body
  type `none`; path fields `page_id`; required record fields `page_id`; accepted fields `page_id`;
  missing records treated as success for status `404`; risk: permanently removes a page (and its
  subpages/content) from the doc; irreversible, queued for async processing (202).

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 7 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, duplicate_of=7, non_data_endpoint=4, out_of_scope=2,
  requires_elevated_scope=15.
