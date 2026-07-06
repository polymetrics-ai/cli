# Overview

Reads Box folder files and per-file detail metadata, and writes file rename/description updates,
through the Box REST API using the OAuth2 client-credentials grant.

Readable streams: `files`, `file_details`.

Write actions: `update_file`.

Service API documentation: https://developer.box.com/reference/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.box.com/2.0`; format `uri`; Box API base URL
  override for tests or proxies.
- `box_folder_id` (optional, string); default `0`; Box folder id for the files stream
  (folders/{box_folder_id}/items); defaults to the root folder (0).
- `box_subject_id` (optional, string); Box token-scoping subject id: the enterprise id
  (box_subject_type=enterprise) or user id (box_subject_type=user). Sent as the box_subject_id
  token-request form param.
- `box_subject_type` (optional, string); default `enterprise`; Box token-scoping subject type:
  'enterprise' (application service account) or 'user'. Sent as the box_subject_type token-request
  form param.
- `client_id` (required, secret, string); Box OAuth2 client-credentials client_id. Used only for the
  token exchange; never logged.
- `client_secret` (required, secret, string); Box OAuth2 client-credentials client_secret. Used only
  for the token exchange; never logged.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page.
- `token_url` (optional, string); default `https://api.box.com/oauth2/token`; format `uri`; Box
  OAuth2 token endpoint override for tests or proxies.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.box.com/2.0`, `box_folder_id=0`,
`box_subject_type=enterprise`, `page_size=100`, `token_url=https://api.box.com/oauth2/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication with extra token parameters `box_subject_id`,
  `box_subject_type` using `config.token_url`, `secrets.client_id`, `secrets.client_secret`,
  `config.box_subject_type`, `config.box_subject_id`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/folders/{{ config.box_folder_id }}/items` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `files`: GET `/folders/{{ config.box_folder_id }}/items` - records path `entries`; query
  `limit`=`{{ config.page_size }}`; offset/limit pagination; offset parameter `offset`; limit
  parameter `limit`; page size 100; emits passthrough records.
- `file_details`: GET `/files/{{ fanout.id }}` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; fan-out; ids from request
  `/folders/{{ config.box_folder_id }}/items`; id-list records path `entries`; id field `id`; id
  inserted into the request path; stamps `file_id`.

## Write actions & risks

Overall write risk: external mutation renaming or updating the description of a Box file.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_file`: PUT `/files/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `id`, `name`; risk: external mutation;
  renames or updates the description of a Box file; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, non_data_endpoint=1, out_of_scope=4.
