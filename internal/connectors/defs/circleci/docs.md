# Overview

Reads and writes CircleCI projects, pipelines, workflows, jobs, contexts, schedules, environment
variables, checkout keys, and workflow insights through the CircleCI v2 REST API.

Readable streams: `projects`, `pipelines`, `workflows`, `jobs`, `contexts`, `schedules`,
`checkout_keys`, `environment_variables`, `insights_workflow_summary`.

Write actions: `create_schedule`, `update_schedule`, `delete_schedule`,
`create_environment_variable`, `delete_environment_variable`, `create_checkout_key`,
`delete_checkout_key`.

Service API documentation: https://circleci.com/docs/api/v2/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); CircleCI personal API token, sent as the Circle-Token
  header. Never logged.
- `base_url` (optional, string); default `https://circleci.com/api/v2`; format `uri`; CircleCI API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `org` (optional, string); CircleCI organization/account segment of the project slug (e.g. acme).
  Required for the projects, pipelines, schedules, checkout_keys, and environment_variables streams,
  and doubles as the second segment of the contexts stream's derived owner-slug.
- `pipeline_id` (optional, string); CircleCI pipeline ID. Required for the workflows stream.
- `repo` (optional, string); CircleCI repository segment of the project slug (e.g. widgets).
  Required for the projects, pipelines, schedules, checkout_keys, environment_variables, and
  insights_workflow_summary streams.
- `vcs_type` (optional, string); CircleCI VCS type segment of the project slug (e.g. gh, bb).
  Required for the projects, pipelines, schedules, checkout_keys, and environment_variables streams,
  and doubles as the first segment of the contexts stream's derived owner-slug.
- `workflow_id` (optional, string); CircleCI workflow ID. Required for the jobs stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://circleci.com/api/v2`.

Authentication behavior:

- API key authentication in `Circle-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/me`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `pipelines`, `workflows`, `jobs`, `contexts`, `schedules`,
`checkout_keys`, `environment_variables`, `insights_workflow_summary`; none: `projects`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `projects`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}` - records path
  None; computed output fields `default_branch`, `vcs_url`.
- `pipelines`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/pipeline` -
  records path `items`; cursor pagination; cursor parameter `page-token`; next token from
  `next_page_token`; incremental cursor `created_at`; formatted as `rfc3339`.
- `workflows`: GET `/pipeline/{{ config.pipeline_id }}/workflow` - records path `items`; cursor
  pagination; cursor parameter `page-token`; next token from `next_page_token`; incremental cursor
  `created_at`; formatted as `rfc3339`.
- `jobs`: GET `/workflow/{{ config.workflow_id }}/job` - records path `items`; cursor pagination;
  cursor parameter `page-token`; next token from `next_page_token`; incremental cursor `started_at`;
  formatted as `rfc3339`.
- `contexts`: GET `/context` - records path `items`; query `owner-slug`=`{{ config.vcs_type }}/{{
  config.org }}`; cursor pagination; cursor parameter `page-token`; next token from
  `next_page_token`.
- `schedules`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule` -
  records path `items`; cursor pagination; cursor parameter `page-token`; next token from
  `next_page_token`; incremental cursor `updated-at`; formatted as `rfc3339`.
- `checkout_keys`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/checkout-key` - records path `items`; cursor pagination; cursor parameter `page-token`; next
  token from `next_page_token`.
- `environment_variables`: GET `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/envvar` - records path `items`; cursor pagination; cursor parameter `page-token`; next token
  from `next_page_token`.
- `insights_workflow_summary`: GET `/insights/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/workflows` - records path `items`; cursor pagination; cursor parameter `page-token`; next token
  from `next_page_token`.

## Write actions & risks

Overall write risk: external mutation of CircleCI project configuration:
schedule/environment-variable/checkout-key create and delete; never triggers, cancels, or approves a
live CI run.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_schedule`: POST `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/schedule` - kind `create`; body type `json`; required record fields `name`, `timetable`,
  `attribution-actor`, `parameters`; accepted fields `attribution-actor`, `description`, `name`,
  `parameters`, `timetable`; risk: external mutation; creates a new scheduled-pipeline trigger for
  this project.
- `update_schedule`: PATCH `/schedule/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `attribution-actor`, `description`,
  `id`, `name`, `parameters`, `timetable`; risk: external mutation; updates an existing
  scheduled-pipeline trigger's timetable or parameters.
- `delete_schedule`: DELETE `/schedule/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion of a scheduled-pipeline trigger; approval
  required.
- `create_environment_variable`: POST `/project/{{ config.vcs_type }}/{{ config.org }}/{{
  config.repo }}/envvar` - kind `create`; body type `json`; required record fields `name`, `value`;
  accepted fields `name`, `value`; risk: external mutation; creates or overwrites a project
  environment variable used by every future CI run.
- `delete_environment_variable`: DELETE `/project/{{ config.vcs_type }}/{{ config.org }}/{{
  config.repo }}/envvar/{{ record.name }}` - kind `delete`; body type `none`; path fields `name`;
  required record fields `name`; accepted fields `name`; missing records treated as success for
  status `404`; risk: irreversible external deletion of a project environment variable; may break
  future CI runs that depend on it; approval required.
- `create_checkout_key`: POST `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/checkout-key` - kind `create`; body type `json`; required record fields `type`; accepted fields
  `type`; risk: external mutation; creates a new deploy/checkout SSH key with repository access.
- `delete_checkout_key`: DELETE `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo
  }}/checkout-key/{{ record.fingerprint }}` - kind `delete`; body type `none`; path fields
  `fingerprint`; required record fields `fingerprint`; accepted fields `fingerprint`; missing
  records treated as success for status `404`; risk: irreversible external revocation of a
  deploy/checkout SSH key; may break future CI checkouts that depend on it; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 9 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=5, duplicate_of=22, non_data_endpoint=1, out_of_scope=39,
  requires_elevated_scope=28.
