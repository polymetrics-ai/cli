# Overview

Reads deployments, projects, teams, domains, aliases, webhooks, log drains, and edge configs from
the Vercel REST API, and writes projects, deployments, domains, project environment variables,
webhooks, log drains, edge configs, and alias removal.

Readable streams: `deployments`, `projects`, `teams`, `domains`, `project_env_vars`, `aliases`,
`webhooks`, `log_drains`, `edge_configs`.

Write actions: `create_project`, `update_project`, `delete_project`, `create_deployment`,
`cancel_deployment`, `delete_deployment`, `add_project_domain`, `remove_project_domain`,
`create_project_env_var`, `delete_project_env_var`, `create_webhook`, `delete_webhook`,
`create_log_drain`, `delete_log_drain`, `create_edge_config`, `update_edge_config`,
`delete_edge_config`, `delete_alias`.

Service API documentation: https://vercel.com/docs/rest-api.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Vercel API bearer access token. Used only for Bearer
  auth; never logged.
- `base_url` (optional, string); default `https://api.vercel.com`; format `uri`; Vercel API base URL
  override for tests or proxies.
- `start_date` (optional, string); format `date-time`.
- `team_id` (optional, string); Optional Vercel Team ID sent as the 'teamId' query param on every
  request when set. Required by Vercel's own API for team-scoped access tokens; omitted entirely for
  personal-account tokens.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.vercel.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v6/deployments`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `projects`, `teams`, `domains`, `aliases`; none: `deployments`,
`project_env_vars`, `webhooks`, `log_drains`, `edge_configs`.

- `deployments`: GET `/v6/deployments` - records path `deployments`; query `from` from template `{{
  config.start_date }}`, omitted when absent; `teamId` from template `{{ config.team_id }}`, omitted
  when absent; computed output fields `created`, `id`, `name`, `state`.
- `projects`: GET `/v10/projects` - records path `projects`; query `teamId` from template `{{
  config.team_id }}`, omitted when absent; cursor pagination; cursor parameter `until`; next token
  from `pagination.next`.
- `teams`: GET `/v2/teams` - records path `teams`; cursor pagination; cursor parameter `until`; next
  token from `pagination.next`.
- `domains`: GET `/v5/domains` - records path `domains`; query `teamId` from template `{{
  config.team_id }}`, omitted when absent; cursor pagination; cursor parameter `until`; next token
  from `pagination.next`.
- `project_env_vars`: GET `/v10/projects/{{ fanout.id }}/env` - records path `envs`; fan-out; ids
  from request `/v10/projects`; id-list records path `projects`; id field `id`; id inserted into the
  request path; stamps `project_id`.
- `aliases`: GET `/v4/aliases` - records path `aliases`; query `teamId` from template `{{
  config.team_id }}`, omitted when absent; cursor pagination; cursor parameter `until`; next token
  from `pagination.next`.
- `webhooks`: GET `/v1/webhooks` - records path `.`; query `teamId` from template `{{ config.team_id
  }}`, omitted when absent.
- `log_drains`: GET `/v1/log-drains` - records path `.`; query `teamId` from template `{{
  config.team_id }}`, omitted when absent.
- `edge_configs`: GET `/v1/edge-config` - records path `.`; query `teamId` from template `{{
  config.team_id }}`, omitted when absent.

## Write actions & risks

Overall write risk: external mutation of Vercel projects, deployments, domains, environment
variables, webhooks, log drains, edge configs, and aliases; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_project`: POST `/v11/projects` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `buildCommand`, `framework`, `installCommand`, `name`, `outputDirectory`,
  `rootDirectory`; risk: external mutation; approval required.
- `update_project`: PATCH `/v9/projects/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `buildCommand`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_project`: DELETE `/v9/projects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; approval
  required.
- `create_deployment`: POST `/v13/deployments` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `name`, `project`, `target`; risk: external mutation; approval
  required.
- `cancel_deployment`: PATCH `/v12/deployments/{{ record.id }}/cancel` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; approval required.
- `delete_deployment`: DELETE `/v13/deployments/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation;
  approval required.
- `add_project_domain`: POST `/v10/projects/{{ record.project_id }}/domains` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`, `name`; accepted
  fields `name`, `project_id`; risk: external mutation; approval required.
- `remove_project_domain`: DELETE `/v9/projects/{{ record.project_id }}/domains/{{ record.domain }}`
  - kind `delete`; body type `none`; path fields `project_id`, `domain`; required record fields
  `project_id`, `domain`; accepted fields `domain`, `project_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; approval
  required.
- `create_project_env_var`: POST `/v10/projects/{{ record.project_id }}/env` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`, `key`, `value`,
  `type`; accepted fields `key`, `project_id`, `target`, `type`, `value`; risk: external mutation;
  approval required.
- `delete_project_env_var`: DELETE `/v9/projects/{{ record.project_id }}/env/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `project_id`, `id`; required record fields `project_id`,
  `id`; accepted fields `id`, `project_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external mutation; approval required.
- `create_webhook`: POST `/v1/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `events`; accepted fields `events`, `projectIds`, `url`; risk: external mutation; approval
  required.
- `delete_webhook`: DELETE `/v1/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; approval
  required.
- `create_log_drain`: POST `/v1/log-drains` - kind `create`; body type `json`; required record
  fields `deliveryFormat`, `url`, `sources`; accepted fields `deliveryFormat`, `environments`,
  `headers`, `name`, `projectIds`, `samplingRate`, `secret`, `sources`, `url`; risk: external
  mutation; approval required.
- `delete_log_drain`: DELETE `/v1/log-drains/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation;
  approval required.
- `create_edge_config`: POST `/v1/edge-config` - kind `create`; body type `json`; required record
  fields `slug`; accepted fields `items`, `slug`; risk: external mutation; approval required.
- `update_edge_config`: PUT `/v1/edge-config/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `slug`; accepted fields `id`, `slug`; risk:
  external mutation; approval required.
- `delete_edge_config`: DELETE `/v1/edge-config/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation;
  approval required.
- `delete_alias`: DELETE `/v2/aliases/{{ record.uid }}` - kind `delete`; body type `none`; path
  fields `uid`; required record fields `uid`; accepted fields `uid`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation (removes
  a deployment alias); approval required.

## Known limits

- API coverage includes 9 stream-backed endpoint group(s), 18 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=9, destructive_admin=8, duplicate_of=34, non_data_endpoint=8, out_of_scope=101,
  requires_elevated_scope=145.
