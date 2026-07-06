# Overview

Reads SonarCloud issues, components, projects, hotspots, rules, metrics, languages, quality gates,
measures, webhooks, and project analyses through the Web API; writes webhook lifecycle, issue
comment/assign/tag/transition, and project-tag mutations.

Readable streams: `issues`, `components`, `quality_gates`, `measures`, `projects`, `hotspots`,
`languages`, `metrics`, `rules`, `webhooks`, `project_analyses`.

Write actions: `create_webhook`, `update_webhook`, `delete_webhook`, `add_issue_comment`,
`assign_issue`, `set_issue_tags`, `do_issue_transition`, `set_project_tags`.

Service API documentation: https://sonarcloud.io/web_api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://sonarcloud.io`; format `uri`; SonarCloud API base
  URL override for tests or proxies.
- `component_keys` (optional, string); Comma-separated SonarCloud component keys. Sent as
  'componentKeys' on issues/quality-gates/measures streams; only the first key is sent as
  'component' on the components stream.
- `end_date` (optional, string); format `date-time`; Upper bound sent as the 'createdBefore' query
  parameter (issues search).
- `mode` (optional, string).
- `organization` (optional, string); SonarCloud organization key; sent as the 'organization' query
  parameter on every request when set.
- `page_size` (optional, string); default `100`; Records per page (1-500), sent as the 'ps' query
  parameter.
- `start_date` (optional, string); format `date-time`; Lower bound sent as the 'createdAfter' query
  parameter (issues search).
- `user_token` (required, secret, string); SonarCloud user token, sent as a Bearer token
  (Authorization: Bearer <user_token>). Never logged.

Secret fields are redacted in logs and write previews: `user_token`.

Default configuration values: `base_url=https://sonarcloud.io`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.user_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/issues/search`.

## Streams notes

Default pagination: single request; no pagination.

- `issues`: GET `/api/issues/search` - records path `issues`; query `componentKeys` from template
  `{{ config.component_keys }}`, omitted when absent; `createdAfter` from template `{{
  config.start_date }}`, omitted when absent; `createdBefore` from template `{{ config.end_date }}`,
  omitted when absent; `organization` from template `{{ config.organization }}`, omitted when
  absent; `p`=`1`; `ps` from template `{{ config.page_size }}`, default `100`; emits passthrough
  records.
- `components`: GET `/api/components/search` - records path `components`; query `component` from
  template `{{ config.component_keys }}`, omitted when absent; `createdAfter` from template `{{
  config.start_date }}`, omitted when absent; `createdBefore` from template `{{ config.end_date }}`,
  omitted when absent; `organization` from template `{{ config.organization }}`, omitted when
  absent; `p`=`1`; `ps` from template `{{ config.page_size }}`, default `100`; emits passthrough
  records.
- `quality_gates`: GET `/api/qualitygates/list` - records path `qualitygates`; query `componentKeys`
  from template `{{ config.component_keys }}`, omitted when absent; `createdAfter` from template `{{
  config.start_date }}`, omitted when absent; `createdBefore` from template `{{ config.end_date }}`,
  omitted when absent; `organization` from template `{{ config.organization }}`, omitted when
  absent; `p`=`1`; `ps` from template `{{ config.page_size }}`, default `100`; emits passthrough
  records.
- `measures`: GET `/api/measures/search` - records path `measures`; query `componentKeys` from
  template `{{ config.component_keys }}`, omitted when absent; `createdAfter` from template `{{
  config.start_date }}`, omitted when absent; `createdBefore` from template `{{ config.end_date }}`,
  omitted when absent; `organization` from template `{{ config.organization }}`, omitted when
  absent; `p`=`1`; `ps` from template `{{ config.page_size }}`, default `100`; emits passthrough
  records.
- `projects`: GET `/api/projects/search` - records path `components`; query `organization` from
  template `{{ config.organization }}`, omitted when absent; `p`=`1`; `ps` from template `{{
  config.page_size }}`, default `100`; `q` from template `{{ config.component_keys }}`, omitted when
  absent; emits passthrough records.
- `hotspots`: GET `/api/hotspots/search` - records path `hotspots`; query `p`=`1`; `projectKey` from
  template `{{ config.component_keys }}`, omitted when absent; `ps` from template `{{
  config.page_size }}`, default `100`; emits passthrough records.
- `languages`: GET `/api/languages/list` - records path `languages`; emits passthrough records.
- `metrics`: GET `/api/metrics/search` - records path `metrics`; query `p`=`1`; `ps` from template
  `{{ config.page_size }}`, default `100`; emits passthrough records.
- `rules`: GET `/api/rules/search` - records path `rules`; query `organization` from template `{{
  config.organization }}`, omitted when absent; `p`=`1`; `ps` from template `{{ config.page_size
  }}`, default `100`; emits passthrough records.
- `webhooks`: GET `/api/webhooks/list` - records path `webhooks`; query `organization` from template
  `{{ config.organization }}`, omitted when absent; `project` from template `{{
  config.component_keys }}`, omitted when absent; emits passthrough records.
- `project_analyses`: GET `/api/project_analyses/search` - records path `analyses`; query `from`
  from template `{{ config.start_date }}`, omitted when absent; `p`=`1`; `project` from template `{{
  config.component_keys }}`; `ps` from template `{{ config.page_size }}`, default `100`; `to` from
  template `{{ config.end_date }}`, omitted when absent; emits passthrough records.

## Write actions & risks

Overall write risk: external SonarCloud API mutation of webhooks (create/update/delete), issue
comments/assignment/tags/workflow transitions, and project tags.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_webhook`: POST `/api/webhooks/create` - kind `create`; body type `form`; required record
  fields `name`, `organization`, `url`; accepted fields `name`, `organization`, `project`, `secret`,
  `url`; risk: external mutation; creates a project or organization webhook that will receive
  analysis-completion callbacks; approval required.
- `update_webhook`: POST `/api/webhooks/update` - kind `update`; body type `form`; required record
  fields `webhook`, `name`, `url`; accepted fields `name`, `secret`, `url`, `webhook`; risk:
  external mutation; changes an existing webhook's callback URL/secret; approval required.
- `delete_webhook`: POST `/api/webhooks/delete` - kind `delete`; body type `form`; required record
  fields `webhook`; accepted fields `webhook`; risk: external mutation; permanently removes a
  webhook; approval required.
- `add_issue_comment`: POST `/api/issues/add_comment` - kind `create`; body type `form`; required
  record fields `issue`, `text`; accepted fields `isFeedback`, `issue`, `text`; risk: external
  mutation; adds a permanent comment to an issue; approval required.
- `assign_issue`: POST `/api/issues/assign` - kind `update`; body type `form`; required record
  fields `issue`; accepted fields `assignee`, `issue`; risk: external mutation; assigns or unassigns
  (empty assignee) an issue; approval required.
- `set_issue_tags`: POST `/api/issues/set_tags` - kind `update`; body type `form`; required record
  fields `issue`; accepted fields `issue`, `tags`; risk: external mutation; replaces an issue's full
  tag set (empty tags clears them); approval required.
- `do_issue_transition`: POST `/api/issues/do_transition` - kind `update`; body type `form`;
  required record fields `issue`, `transition`; accepted fields `comment`, `issue`, `transition`;
  risk: external mutation; moves an issue through its workflow (e.g. resolve, wontfix,
  falsepositive); some transitions require elevated project permissions on the live API; approval
  required.
- `set_project_tags`: POST `/api/project_tags/set` - kind `update`; body type `form`; required
  record fields `project`, `tags`; accepted fields `project`, `tags`; risk: external mutation;
  replaces a project's full tag set; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 11 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, deprecated=28, destructive_admin=31, duplicate_of=9, non_data_endpoint=6,
  out_of_scope=32, requires_elevated_scope=26.
