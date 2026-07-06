# Overview

Reads Box users, groups, collections, folder items, webhooks, retention policies, legal hold
policies, storage policies, sign requests, terms of services, metadata templates, and pending
collaborations, and writes group/webhook/collaboration lifecycle mutations, through the Box REST API
using the OAuth2 client-credentials grant.

Readable streams: `users`, `groups`, `collections`, `folder_items`, `webhooks`,
`retention_policies`, `legal_hold_policies`, `storage_policies`, `sign_requests`,
`terms_of_services`, `metadata_templates`, `pending_collaborations`.

Write actions: `create_group`, `update_group`, `delete_group`, `create_webhook`, `update_webhook`,
`delete_webhook`, `create_collaboration`, `update_collaboration`, `delete_collaboration`.

Service API documentation: https://developer.box.com/reference/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.box.com/2.0`; format `uri`; Box API base URL
  override for tests or proxies.
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
- `folder_id` (optional, string); default `0`; Box folder id for the folder_items stream
  (folders/{folder_id}/items); defaults to the root folder (0).
- `mode` (optional, string).
- `token_url` (optional, string); default `https://api.box.com/oauth2/token`; format `uri`; Box
  OAuth2 token endpoint override for tests or proxies.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.box.com/2.0`, `box_subject_type=enterprise`,
`folder_id=0`, `token_url=https://api.box.com/oauth2/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication with extra token parameters `box_subject_id`,
  `box_subject_type` using `config.token_url`, `secrets.client_id`, `secrets.client_secret`,
  `config.box_subject_type`, `config.box_subject_id`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: cursor: `webhooks`, `retention_policies`, `legal_hold_policies`,
`storage_policies`, `sign_requests`, `metadata_templates`; none: `terms_of_services`; offset_limit:
`users`, `groups`, `collections`, `folder_items`, `pending_collaborations`.

- `users`: GET `/users` - records path `entries`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `groups`: GET `/groups` - records path `entries`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `collections`: GET `/collections` - records path `entries`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `folder_items`: GET `/folders/{{ config.folder_id }}/items` - records path `entries`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `webhooks`: GET `/webhooks` - records path `entries`; cursor pagination; cursor parameter
  `marker`; next token from `next_marker`.
- `retention_policies`: GET `/retention_policies` - records path `entries`; cursor pagination;
  cursor parameter `marker`; next token from `next_marker`.
- `legal_hold_policies`: GET `/legal_hold_policies` - records path `entries`; cursor pagination;
  cursor parameter `marker`; next token from `next_marker`.
- `storage_policies`: GET `/storage_policies` - records path `entries`; cursor pagination; cursor
  parameter `marker`; next token from `next_marker`.
- `sign_requests`: GET `/sign_requests` - records path `entries`; cursor pagination; cursor
  parameter `marker`; next token from `next_marker`.
- `terms_of_services`: GET `/terms_of_services` - records path `entries`.
- `metadata_templates`: GET `/metadata_templates/enterprise` - records path `entries`; cursor
  pagination; cursor parameter `marker`; next token from `next_marker`; computed output fields
  `copy_instance_on_item_copy`, `display_name`, `template_key`.
- `pending_collaborations`: GET `/collaborations` - records path `entries`; query
  `status`=`pending`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
  page size 100.

## Write actions & risks

Overall write risk: external mutation of Box enterprise groups, webhook subscriptions, and
file/folder collaborations (access grants); includes 3 destructive (irreversible-effect) actions
(delete_group, delete_webhook, delete_collaboration).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_group`: POST `/groups` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `description`, `external_sync_identifier`, `invitability_level`,
  `member_viewability_level`, `name`, `provenance`; risk: external mutation; creates a new Box
  enterprise group; approval required.
- `update_group`: PUT `/groups/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `external_sync_identifier`, `id`,
  `invitability_level`, `member_viewability_level`, `name`, `provenance`; risk: external mutation;
  updates an existing Box enterprise group's settings; approval required.
- `delete_group`: DELETE `/groups/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; permanently deletes
  a Box enterprise group; approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `target`, `address`, `triggers`; accepted fields `address`, `target`, `triggers`; risk: external
  mutation; creates a new Box webhook subscription that will POST event payloads to an external
  address; approval required.
- `update_webhook`: PUT `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `address`, `id`, `target`, `triggers`; risk:
  external mutation; updates an existing Box webhook's target/address/triggers; approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; permanently
  deletes a Box webhook subscription; approval required.
- `create_collaboration`: POST `/collaborations` - kind `create`; body type `json`; required record
  fields `item`, `accessible_by`, `role`; accepted fields `accessible_by`, `can_view_path`,
  `is_access_only`, `item`, `role`; risk: external mutation; grants a user or group access to a Box
  file/folder; approval required.
- `update_collaboration`: PUT `/collaborations/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `can_view_path`, `id`, `role`,
  `status`; risk: external mutation; changes an existing Box collaboration's role, or
  accepts/rejects a pending invitation; approval required.
- `delete_collaboration`: DELETE `/collaborations/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; permanently revokes a user or group's access to a Box file/folder; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 12 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=14, destructive_admin=16, duplicate_of=33, non_data_endpoint=5, out_of_scope=205.
