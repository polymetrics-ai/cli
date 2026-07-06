# Overview

Reads and writes Dremio catalog entries, reflections, sources, users, and roles through the Dremio
REST API.

Readable streams: `catalog`, `reflections`, `sources`, `users`, `roles`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_role`, `update_role`,
`delete_role`, `update_reflection`, `refresh_reflection`, `delete_reflection`,
`create_personal_access_token`, `delete_personal_access_token`.

Service API documentation: https://docs.dremio.com/software/rest-api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Dremio Personal Access Token (PAT), sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.dremio.cloud/v0`; format `uri`; Dremio REST
  API base URL override for Dremio Software/self-hosted, EU-region, tests, or proxies.
- `page_size` (optional, integer); default `100`; Page size (1-500) sent as the 'maxResults' query
  parameter on each list request.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.dremio.cloud/v0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/catalog`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`.

- `catalog`: GET `/catalog` - records path `data`; query `maxResults`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `reflections`: GET `/reflections` - records path `data`; query `maxResults`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `sources`: GET `/source` - records path `data`; query `maxResults`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `users`: GET `/user` - records path `data`; query `maxResults`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `pageToken`; next token from `nextPageToken`.
- `roles`: GET `/role` - records path `data`; query `maxResults`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `pageToken`; next token from `nextPageToken`.

## Write actions & risks

Overall write risk: external Dremio API mutation of user/role/reflection/PAT lifecycle objects;
several actions are destructive (delete_user, delete_role, delete_reflection,
delete_personal_access_token) and require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/user` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `email`, `firstName`, `lastName`, `name`; risk: creates a new Dremio user account
  with instance-wide access, scoped by whatever role assignment follows; external mutation, approval
  required.
- `update_user`: PUT `/user/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `active`, `email`, `firstName`, `id`, `lastName`,
  `name`; risk: mutates an existing Dremio user account, including its active flag which can lock
  the user out; external mutation, approval required.
- `delete_user`: DELETE `/user/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: permanently removes a Dremio user account and revokes its
  access; destructive, approval required.
- `create_role`: POST `/role` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `description`, `name`; risk: creates a new Dremio role; low external mutation risk
  on its own until members/grants are attached.
- `update_role`: PUT `/role/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `id`, `name`; risk: mutates an
  existing Dremio role's name/description; external mutation, approval required.
- `delete_role`: DELETE `/role/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: permanently removes a Dremio role and its grants for
  every member; destructive, approval required.
- `update_reflection`: PUT `/reflection/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `enabled`, `id`, `name`, `tag`; risk:
  mutates an existing reflection's definition (name/enabled/tag); disabling a reflection removes its
  query-acceleration benefit until re-enabled and rebuilt; external mutation, approval required.
- `refresh_reflection`: POST `/reflection/{{ record.id }}/refresh` - kind `custom`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: forces an
  immediate reflection rebuild, consuming cluster compute; low external-mutation risk, no data loss,
  no approval required.
- `delete_reflection`: DELETE `/reflection/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: permanently removes a reflection definition;
  any query relying on it for acceleration falls back to the raw dataset; destructive, approval
  required.
- `create_personal_access_token`: POST `/user/{{ record.user_id }}/token` - kind `create`; body type
  `json`; path fields `user_id`; body fields `label`, `millisToExpire`; required record fields
  `user_id`, `label`; accepted fields `label`, `millisToExpire`, `user_id`; risk: mints a new
  long-lived Personal Access Token credential for the named user; the response body carries the
  plaintext token exactly once and must never be logged; external mutation, approval required.
- `delete_personal_access_token`: DELETE `/user/{{ record.user_id }}/token/{{ record.token_id }}` -
  kind `delete`; body type `none`; path fields `user_id`, `token_id`; required record fields
  `user_id`, `token_id`; accepted fields `token_id`, `user_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: revokes a single Personal Access Token,
  immediately invalidating any client still using it; destructive to that credential's holders,
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=4, duplicate_of=9, non_data_endpoint=2, out_of_scope=26,
  requires_elevated_scope=6.
