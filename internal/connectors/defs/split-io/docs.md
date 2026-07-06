# Overview

Reads Split.io workspaces, environments, feature flags, segments, groups, traffic types, and users,
and writes feature-flag kill/restore/archive/unarchive and segment-key mutations through the Split
Admin API.

Readable streams: `workspaces`, `environments`, `splits`, `segments`, `groups`, `traffic_types`,
`users`.

Write actions: `kill_feature_flag_in_environment`, `restore_feature_flag_in_environment`,
`archive_feature_flag`, `unarchive_feature_flag`, `add_segment_keys_in_environment`,
`remove_segment_keys_from_environment`.

Service API documentation: https://docs.split.io/reference/api-overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Split.io Admin API key, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.split.io`; format `uri`; Split.io API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `workspace_id` (optional, string); Split.io workspace ID; required by the environments, splits,
  and segments streams to substitute the {workspace_id} path segment.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.split.io`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/internal/api/v2/workspaces` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `users`; none: `groups`, `traffic_types`; offset_limit: `workspaces`,
`environments`, `splits`, `segments`.

- `workspaces`: GET `/internal/api/v2/workspaces` - records path `objects`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `environments`: GET `/internal/api/v2/environments/ws/{{ config.workspace_id }}` - records path
  `objects`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  100.
- `splits`: GET `/internal/api/v2/splits/ws/{{ config.workspace_id }}` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `segments`: GET `/internal/api/v2/segments/ws/{{ config.workspace_id }}` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `groups`: GET `/internal/api/v2/groups` - records path `objects`.
- `traffic_types`: GET `/internal/api/v2/trafficTypes/ws/{{ config.workspace_id }}` - records path
  None.
- `users`: GET `/internal/api/v2/users` - records path `data`; cursor pagination; cursor parameter
  `after`; next token from `nextMarker`.

## Write actions & risks

Overall write risk: external Split.io API mutation that reshapes live feature-flag targeting/rollout
state or segment membership for every SDK evaluating it.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `kill_feature_flag_in_environment`: PUT `/internal/api/v2/splits/ws/{{ record.workspace_id }}/{{
  record.feature_flag_name }}/environments/{{ record.environment_id }}/kill` - kind `update`; body
  type `none`; path fields `workspace_id`, `feature_flag_name`, `environment_id`; required record
  fields `workspace_id`, `feature_flag_name`, `environment_id`; accepted fields `environment_id`,
  `feature_flag_name`, `workspace_id`; risk: immediately forces every SDK evaluating this feature
  flag in the given environment onto its off/default treatment; high-impact production
  traffic-shaping mutation, approval required.
- `restore_feature_flag_in_environment`: PUT `/internal/api/v2/splits/ws/{{ record.workspace_id
  }}/{{ record.feature_flag_name }}/environments/{{ record.environment_id }}/restore` - kind
  `update`; body type `none`; path fields `workspace_id`, `feature_flag_name`, `environment_id`;
  required record fields `workspace_id`, `feature_flag_name`, `environment_id`; accepted fields
  `environment_id`, `feature_flag_name`, `workspace_id`; risk: reverts a previously-killed feature
  flag in the given environment back to its configured targeting rules; production traffic-shaping
  mutation, approval required.
- `archive_feature_flag`: PUT `/internal/api/v2/splits/ws/{{ record.workspace_id }}/{{
  record.feature_flag_name }}/archive` - kind `update`; body type `json`; path fields
  `workspace_id`, `feature_flag_name`; required record fields `workspace_id`, `feature_flag_name`;
  accepted fields `comment`, `feature_flag_name`, `title`, `workspace_id`; risk: archives a feature
  flag account-wide (all SDKs calling it return control); approval required.
- `unarchive_feature_flag`: PUT `/internal/api/v2/splits/ws/{{ record.workspace_id }}/{{
  record.feature_flag_name }}/unarchive` - kind `update`; body type `json`; path fields
  `workspace_id`, `feature_flag_name`; required record fields `workspace_id`, `feature_flag_name`;
  accepted fields `comment`, `feature_flag_name`, `title`, `workspace_id`; risk: restores an
  archived feature flag to active use account-wide; approval required.
- `add_segment_keys_in_environment`: PUT `/internal/api/v2/segments/{{ record.environment_id }}/{{
  record.segment_name }}/uploadKeys` - kind `update`; body type `json`; path fields
  `environment_id`, `segment_name`; required record fields `environment_id`, `segment_name`, `keys`;
  accepted fields `comment`, `environment_id`, `keys`, `segment_name`; risk: adds member keys to a
  segment in the given environment, changing which end-users match segment-based targeting rules for
  every feature flag using it; production traffic-shaping mutation, approval required.
- `remove_segment_keys_from_environment`: PUT `/internal/api/v2/segments/{{ record.environment_id
  }}/{{ record.segment_name }}/removeKeys` - kind `update`; body type `json`; path fields
  `environment_id`, `segment_name`; required record fields `environment_id`, `segment_name`, `keys`;
  accepted fields `comment`, `environment_id`, `keys`, `segment_name`; risk: removes member keys
  from a segment in the given environment, changing which end-users match segment-based targeting
  rules for every feature flag using it; production traffic-shaping mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=11, duplicate_of=4, out_of_scope=38,
  requires_elevated_scope=2.
