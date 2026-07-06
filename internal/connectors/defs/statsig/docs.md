# Overview

Reads and manages Statsig feature gates, dynamic configs, experiments, segments, target apps, tags,
keys, holdouts, layers, users, audit logs, and environments through the Statsig Console API.

Readable streams: `feature_gates`, `dynamic_configs`, `experiments`, `segments`, `target_apps`,
`tags`, `keys`, `holdouts`, `layers`, `users`, `audit_logs`, `environments`.

Write actions: `create_gate`, `update_gate`, `delete_gate`, `create_dynamic_config`,
`update_dynamic_config`, `delete_dynamic_config`, `create_segment`, `delete_segment`, `create_tag`,
`update_tag`, `delete_tag`, `create_target_app`, `update_target_app`, `delete_target_app`,
`create_holdout`, `delete_holdout`, `create_layer`, `delete_layer`, `create_key`, `delete_key`.

Service API documentation: https://docs.statsig.com/console-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Statsig Console API key, sent as the STATSIG-API-KEY header.
  Never logged.
- `base_url` (optional, string); default `https://statsigapi.net/console/v1`; format `uri`; Statsig
  Console API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://statsigapi.net/console/v1`.

Authentication behavior:

- API key authentication in `STATSIG-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/gates` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `environments`; offset_limit: `feature_gates`, `dynamic_configs`,
`experiments`, `segments`, `target_apps`, `tags`, `keys`, `holdouts`, `layers`, `users`,
`audit_logs`.

- `feature_gates`: GET `/gates` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `dynamic_configs`: GET `/dynamic_configs` - records path `data`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `experiments`: GET `/experiments` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `segments`: GET `/segments` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `target_apps`: GET `/target_app` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `tags`: GET `/tags` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `keys`: GET `/keys` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `holdouts`: GET `/holdouts` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `layers`: GET `/layers` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `users`: GET `/users` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `audit_logs`: GET `/audit_logs` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `environments`: GET `/environments` - records path `data.environments`.

## Write actions & risks

Overall write risk: external mutation of Statsig feature gates, dynamic configs, segments, tags,
target apps, holdouts, layers, and API keys, including irreversible deletes and live-credential
creation/deletion.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_gate`: POST `/gates` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `description`, `idType`, `isEnabled`, `name`, `tags`; risk: external mutation;
  approval required.
- `update_gate`: PATCH `/gates/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `id`, `isEnabled`, `name`, `tags`;
  risk: external mutation; approval required.
- `delete_gate`: DELETE `/gates/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversible external deletion; approval required.
- `create_dynamic_config`: POST `/dynamic_configs` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `description`, `idType`, `isEnabled`, `name`, `tags`; risk:
  external mutation; approval required.
- `update_dynamic_config`: PATCH `/dynamic_configs/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `description`, `id`,
  `isEnabled`, `name`, `tags`; risk: external mutation; approval required.
- `delete_dynamic_config`: DELETE `/dynamic_configs/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversible external deletion; approval required.
- `create_segment`: POST `/segments` - kind `create`; body type `json`; required record fields
  `name`, `type`; accepted fields `description`, `idType`, `name`, `tags`, `type`; risk: external
  mutation; approval required.
- `delete_segment`: DELETE `/segments/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`,
  `description`; accepted fields `description`, `isCore`, `name`; risk: external mutation; approval
  required.
- `update_tag`: PATCH `/tags/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `id`, `isCore`, `name`; risk: external
  mutation; approval required.
- `delete_tag`: DELETE `/tags/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; risk: irreversible external deletion; approval required.
- `create_target_app`: POST `/target_app` - kind `create`; body type `json`; required record fields
  `name`, `description`; accepted fields `description`, `name`; risk: external mutation; approval
  required.
- `update_target_app`: PATCH `/target_app/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_target_app`: DELETE `/target_app/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_holdout`: POST `/holdouts` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `description`, `idType`, `name`, `teamID`; risk: external mutation;
  approval required.
- `delete_holdout`: DELETE `/holdouts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_layer`: POST `/layers` - kind `create`; body type `json`; required record fields `name`,
  `idType`; accepted fields `description`, `idType`, `name`; risk: external mutation; approval
  required.
- `delete_layer`: DELETE `/layers/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversible external deletion; approval required.
- `create_key`: POST `/keys` - kind `create`; body type `json`; required record fields
  `description`, `type`; accepted fields `description`, `environments`, `scopes`, `type`; risk:
  external mutation creating a live API credential; approval required.
- `delete_key`: DELETE `/keys/{{ record.key }}` - kind `delete`; body type `none`; path fields
  `key`; required record fields `key`; accepted fields `key`; missing records treated as success for
  status `404`; risk: irreversible external deletion of a live API credential; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 12 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=7, destructive_admin=51, duplicate_of=18, non_data_endpoint=3, out_of_scope=173,
  requires_elevated_scope=18.
