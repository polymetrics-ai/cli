# Overview

Reads Codefresh projects, pipelines, builds, runner agents, shared contexts, container images,
registries, triggers, and annotations, and can create/update/delete/run projects, pipelines,
contexts, and agents through the Codefresh REST API.

Readable streams: `projects`, `pipelines`, `agents`, `contexts`, `builds`, `images`, `registries`,
`triggers`, `trigger_events`, `annotations`.

Write actions: `create_project`, `delete_project`, `create_pipeline`, `update_pipeline`,
`delete_pipeline`, `run_pipeline`, `create_context`, `delete_context`, `create_agent`,
`delete_agent`.

Service API documentation: https://g.codefresh.io/api/.

## Auth setup

Connection fields:

- `account_id` (optional, string); Optional Codefresh account id; sent as the X-Access-Token header
  for account scoping. Omitted entirely when unset.
- `api_key` (required, secret, string); Codefresh API key. Sent verbatim (no Bearer prefix) as the
  Authorization header; used only for auth and never logged.
- `base_url` (optional, string); default `https://g.codefresh.io/api`; format `uri`; Codefresh API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://g.codefresh.io/api`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 50.

Pagination by stream: none: `registries`, `triggers`, `trigger_events`, `annotations`; offset_limit:
`images`; page_number: `projects`, `pipelines`, `agents`, `contexts`, `builds`.

- `projects`: GET `/projects` - records path `projects`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; computed output fields `favorite`,
  `id`, `pipelines_number`, `project_name`, `updated_at`.
- `pipelines`: GET `/pipelines` - records path `docs`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; computed output fields `created_at`,
  `id`, `is_public`, `name`, `project`, `updated_at`.
- `agents`: GET `/agents` - records at response root; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 50; computed output fields `created_at`, `id`,
  `name`, `status`, `version`.
- `contexts`: GET `/contexts` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; computed output fields `id`, `name`,
  `owner`, `type`.
- `builds`: GET `/workflow` - records path `workflows.docs`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; computed output fields `branch_name`,
  `commit_message`, `committer`, `created`, `finished`, `id`, `pipeline_name`, `progress`,
  `project`, `project_id`, `provider`, `repo_name`, `repo_owner`, `revision`, `status`, and 3 more.
- `images`: GET `/images` - records path `docs`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 50; computed output fields `branch`, `commit`, `commit_url`,
  `created`, `id`, `image_display_name`, `image_name`, `repo`, `sha`, `size`.
- `registries`: GET `/registries` - records at response root; computed output fields
  `behind_firewall`, `default`, `domain`, `id`, `internal`, `kind`, `name`, `primary`, `provider`.
- `triggers`: GET `/hermes/triggers` - records at response root; computed output fields `event`,
  `event_description`, `event_status`, `event_type`, `filter_tag`, `pipeline`.
- `trigger_events`: GET `/hermes/events` - records at response root; computed output fields
  `account`, `description`, `endpoint`, `kind`, `status`, `type`, `uri`.
- `annotations`: GET `/annotations` - records at response root; computed output fields `account_id`,
  `entity_id`, `entity_type`, `id`, `key`, `type`, `value`.

## Write actions & risks

Overall write risk: external mutation of Codefresh projects, pipelines, contexts, and runner agents,
including irreversible deletes and triggering real pipeline runs (consumes build minutes/resources).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `projectName`; accepted fields `projectName`, `tags`, `variables`; risk: external mutation;
  creates a new Codefresh project; approval required.
- `delete_project`: DELETE `/projects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive; irreversible deletion of a
  Codefresh project; approval required.
- `create_pipeline`: POST `/pipelines` - kind `create`; body type `json`; required record fields
  `metadata`; accepted fields `metadata`, `spec`; risk: external mutation; creates a new Codefresh
  pipeline; approval required.
- `update_pipeline`: PUT `/pipelines/{{ record.name }}` - kind `update`; body type `json`; path
  fields `name`; required record fields `name`, `metadata`; accepted fields `metadata`, `name`,
  `spec`; risk: external mutation; replaces an existing Codefresh pipeline's spec; approval
  required.
- `delete_pipeline`: DELETE `/pipelines/{{ record.name }}` - kind `delete`; body type `none`; path
  fields `name`; required record fields `name`; accepted fields `name`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive; irreversible deletion of
  a Codefresh pipeline; approval required.
- `run_pipeline`: POST `/pipelines/run/{{ record.name }}` - kind `custom`; body type `json`; path
  fields `name`; required record fields `name`; accepted fields `branch`, `name`, `options`,
  `variables`; risk: external mutation; triggers a real Codefresh pipeline run (build
  minutes/resources consumed); approval required.
- `create_context`: POST `/contexts` - kind `create`; body type `json`; required record fields
  `metadata`, `spec`; accepted fields `metadata`, `spec`, `type`; risk: external mutation; creates a
  new Codefresh shared context (may hold configuration values); approval required.
- `delete_context`: DELETE `/contexts/{{ record.name }}` - kind `delete`; body type `none`; path
  fields `name`; required record fields `name`; accepted fields `name`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive; irreversible deletion of
  a Codefresh shared context; approval required.
- `create_agent`: POST `/agents` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`, `runtimes`; risk: external mutation; registers a new Codefresh runner
  agent; approval required.
- `delete_agent`: DELETE `/agent/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive; irreversible deregistration of a
  Codefresh runner agent; approval required.

## Known limits

- Batch defaults: read_page_size=2.
- API coverage includes 10 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=8, non_data_endpoint=14, out_of_scope=238, requires_elevated_scope=139.
