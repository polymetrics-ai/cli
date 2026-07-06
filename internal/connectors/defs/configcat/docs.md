# Overview

Reads and writes ConfigCat feature-flag platform data: organizations, products, configs,
environments, settings/feature flags, deleted settings, SDK keys, segments, webhooks, permission
groups, integrations, proxy profiles, members, audit logs, stale flags, tags, and the authenticated
user's own profile through the ConfigCat Public Management API.

Readable streams: `organizations`, `products`, `configs`, `environments`, `tags`, `config`,
`environment`, `settings`, `setting`, `deleted_settings`, `sdk_keys`, `config_setting_values`,
`segments`, `segment`, `webhooks`, `webhook`, `permission_groups`, `permission_group`,
`integrations`, `integration`, `proxy_profiles`, `proxy_profile`, `members`, `audit_logs`,
`stale_flags`, `me`, `tag`.

Write actions: `create_config`, `update_config`, `delete_config`, `create_environment`,
`update_environment`, `delete_environment`, `create_flag`, `update_flag`, `delete_flag`,
`create_tag`, `update_tag`, `delete_tag`.

Service API documentation: https://api.configcat.com/docs/.

## Auth setup

Connection fields:

- `audit_log_config_id` (optional, string); Optional configId query filter for the audit_logs
  stream.
- `audit_log_environment_id` (optional, string); Optional environmentId query filter for the
  audit_logs stream.
- `base_url` (optional, string); default `https://api.configcat.com`; format `uri`; ConfigCat API
  base URL override for tests or proxies.
- `config_id` (optional, string); ConfigCat config id; required for the config, settings,
  deleted_settings, and setting/sdk_keys/config_setting_values detail streams (the latter two also
  require environment_id).
- `environment_id` (optional, string); ConfigCat environment id; required for the environment,
  sdk_keys, and config_setting_values detail streams.
- `integration_id` (optional, string); ConfigCat integration id; required for the integration detail
  stream.
- `mode` (optional, string).
- `organization_id` (optional, string); ConfigCat organization id; required for the proxy_profiles
  and members streams.
- `password` (required, secret, string); ConfigCat Public Management API Basic auth password. Used
  only for Basic auth; never logged.
- `permission_group_id` (optional, string); ConfigCat permission group id; required for the
  permission_group detail stream.
- `product_id` (optional, string); ConfigCat product id; required for the integrations and
  stale_flags detail streams (the
  configs/environments/tags/segments/webhooks/permission_groups/audit_logs streams instead fan out
  across every product automatically).
- `proxy_profile_id` (optional, string); ConfigCat proxy profile id; required for the proxy_profile
  detail stream.
- `segment_id` (optional, string); ConfigCat segment id; required for the segment detail stream.
- `setting_id` (optional, string); ConfigCat setting (feature flag) id; required for the setting
  detail stream.
- `tag_id` (optional, string); ConfigCat tag id; required for the tag detail stream.
- `username` (optional, string); ConfigCat Public Management API Basic auth username (not secret).
- `webhook_id` (optional, string); ConfigCat webhook id; required for the webhook detail stream.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.configcat.com`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password` when `{{ config.username
  }}`.
- HTTP Basic authentication using `secrets.username`, `secrets.password` when `{{ secrets.username
  }}`.
- HTTP Basic authentication using `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/organizations`.

## Streams notes

Default pagination: single request; no pagination.

- `organizations`: GET `/v1/organizations` - records at response root; computed output fields
  `organization_id`.
- `products`: GET `/v1/products` - records at response root; computed output fields
  `approve_required`, `organization_id`, `product_id`, `reason_required`.
- `configs`: GET `/v1/products/{{ fanout.id }}/configs` - records at response root; computed output
  fields `config_id`, `evaluation_version`; fan-out; ids from request `/v1/products`; id field
  `productId`; id inserted into the request path; stamps `product_id`.
- `environments`: GET `/v1/products/{{ fanout.id }}/environments` - records at response root;
  computed output fields `approve_required`, `environment_id`, `reason_required`; fan-out; ids from
  request `/v1/products`; id field `productId`; id inserted into the request path; stamps
  `product_id`.
- `tags`: GET `/v1/products/{{ fanout.id }}/tags` - records at response root; computed output fields
  `tag_id`; fan-out; ids from request `/v1/products`; id field `productId`; id inserted into the
  request path; stamps `product_id`.
- `config`: GET `/v1/configs/{{ config.config_id }}` - records at response root.
- `environment`: GET `/v1/environments/{{ config.environment_id }}` - records at response root.
- `settings`: GET `/v1/configs/{{ config.config_id }}/settings` - records at response root.
- `setting`: GET `/v1/settings/{{ config.setting_id }}` - records at response root.
- `deleted_settings`: GET `/v1/configs/{{ config.config_id }}/deleted-settings` - records at
  response root.
- `sdk_keys`: GET `/v1/configs/{{ config.config_id }}/environments/{{ config.environment_id }}` -
  records at response root.
- `config_setting_values`: GET `/v1/configs/{{ config.config_id }}/environments/{{
  config.environment_id }}/values` - records at response root.
- `segments`: GET `/v1/products/{{ fanout.id }}/segments` - records at response root; computed
  output fields `segment_id`; fan-out; ids from request `/v1/products`; id field `productId`; id
  inserted into the request path; stamps `product_id`.
- `segment`: GET `/v1/segments/{{ config.segment_id }}` - records at response root.
- `webhooks`: GET `/v1/products/{{ fanout.id }}/webhooks` - records at response root; fan-out; ids
  from request `/v1/products`; id field `productId`; id inserted into the request path; stamps
  `product_id`.
- `webhook`: GET `/v1/webhooks/{{ config.webhook_id }}` - records at response root.
- `permission_groups`: GET `/v1/products/{{ fanout.id }}/permissions` - records at response root;
  fan-out; ids from request `/v1/products`; id field `productId`; id inserted into the request path;
  stamps `product_id`.
- `permission_group`: GET `/v1/permissions/{{ config.permission_group_id }}` - records at response
  root.
- `integrations`: GET `/v1/products/{{ config.product_id }}/integrations` - records path
  `integrations`.
- `integration`: GET `/v1/integrations/{{ config.integration_id }}` - records at response root.
- `proxy_profiles`: GET `/v1/organizations/{{ config.organization_id }}/proxy-profiles` - records
  path `profiles`.
- `proxy_profile`: GET `/v1/proxy-profiles/{{ config.proxy_profile_id }}` - records at response
  root.
- `members`: GET `/v1/organizations/{{ config.organization_id }}/members` - records at response
  root.
- `audit_logs`: GET `/v1/products/{{ fanout.id }}/auditlogs` - records at response root; query
  `configId` from template `{{ config.audit_log_config_id }}`, omitted when absent; `environmentId`
  from template `{{ config.audit_log_environment_id }}`, omitted when absent; fan-out; ids from
  request `/v1/products`; id field `productId`; id inserted into the request path; stamps
  `product_id`.
- `stale_flags`: GET `/v1/products/{{ config.product_id }}/staleflags` - records at response root.
- `me`: GET `/v1/me` - records at response root.
- `tag`: GET `/v1/tags/{{ config.tag_id }}` - records at response root.

## Write actions & risks

Overall write risk: external mutation of ConfigCat configs, environments, feature flags/settings,
and tags (create/update/delete); does not change a feature flag's evaluated VALUE in any environment
(see docs.md).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_config`: POST `/v1/products/{{ config.product_id }}/configs` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `description`, `evaluationVersion`, `name`,
  `order`; risk: creates a new ConfigCat config within the configured product; low risk, no data
  destruction.
- `update_config`: PUT `/v1/configs/{{ record.configId }}` - kind `update`; body type `json`; path
  fields `configId`; required record fields `configId`; accepted fields `configId`, `description`,
  `name`, `order`; risk: renames/reorders an existing ConfigCat config; may affect SDK-visible
  dashboard organization.
- `delete_config`: DELETE `/v1/configs/{{ record.configId }}` - kind `delete`; body type `none`;
  path fields `configId`; required record fields `configId`; accepted fields `configId`; missing
  records treated as success for status `404`; risk: permanently deletes a ConfigCat config and
  every feature flag/setting defined in it; destructive, external mutation; approval required.
- `create_environment`: POST `/v1/products/{{ config.product_id }}/environments` - kind `create`;
  body type `json`; required record fields `name`; accepted fields `color`, `description`, `name`,
  `order`; risk: creates a new ConfigCat environment within the configured product; low risk, no
  data destruction.
- `update_environment`: PUT `/v1/environments/{{ record.environmentId }}` - kind `update`; body type
  `json`; path fields `environmentId`; required record fields `environmentId`; accepted fields
  `color`, `description`, `environmentId`, `name`, `order`; risk: renames/recolors an existing
  ConfigCat environment; may affect dashboard organization visible to other users.
- `delete_environment`: DELETE `/v1/environments/{{ record.environmentId }}` - kind `delete`; body
  type `none`; path fields `environmentId`; required record fields `environmentId`; accepted fields
  `environmentId`; missing records treated as success for status `404`; risk: permanently deletes a
  ConfigCat environment and every feature flag value/SDK key scoped to it; destructive, external
  mutation; approval required.
- `create_flag`: POST `/v1/configs/{{ config.config_id }}/settings` - kind `create`; body type
  `json`; required record fields `key`, `name`, `settingType`; accepted fields `hint`, `isJson`,
  `key`, `name`, `order`, `settingType`, `tags`; risk: creates a new ConfigCat feature flag/setting
  within the configured config; low risk, no data destruction.
- `update_flag`: PUT `/v1/settings/{{ record.settingId }}` - kind `update`; body type `json`; path
  fields `settingId`; required record fields `settingId`, `name`; accepted fields `hint`, `name`,
  `order`, `settingId`, `tags`; risk: replaces an existing ConfigCat feature flag/setting's metadata
  (name/hint/tags); does not itself change the flag's evaluated VALUE in any environment.
- `delete_flag`: DELETE `/v1/settings/{{ record.settingId }}` - kind `delete`; body type `none`;
  path fields `settingId`; required record fields `settingId`; accepted fields `settingId`; missing
  records treated as success for status `404`; risk: permanently deletes a ConfigCat feature
  flag/setting and its values in every environment; destructive, external mutation; approval
  required.
- `create_tag`: POST `/v1/products/{{ config.product_id }}/tags` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `color`, `name`; risk: creates a new ConfigCat tag
  within the configured product; low risk, no data destruction.
- `update_tag`: PUT `/v1/tags/{{ record.tagId }}` - kind `update`; body type `json`; path fields
  `tagId`; required record fields `tagId`; accepted fields `color`, `name`, `tagId`; risk:
  renames/recolors an existing ConfigCat tag; affects every feature flag tagged with it.
- `delete_tag`: DELETE `/v1/tags/{{ record.tagId }}` - kind `delete`; body type `none`; path fields
  `tagId`; required record fields `tagId`; accepted fields `tagId`; missing records treated as
  success for status `404`; risk: permanently deletes a ConfigCat tag and untags every feature flag
  that used it; destructive, external mutation; approval required.

## Known limits

- Batch defaults: read_page_size=0.
- API coverage includes 27 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=19, out_of_scope=27, requires_elevated_scope=21.
