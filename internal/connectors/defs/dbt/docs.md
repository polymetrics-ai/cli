# Overview

Reads dbt Cloud projects, runs, repositories, users, environments, jobs, invites, licenses,
notifications, and SSH tunnels, and writes job/notification/SSH-tunnel mutations and run-control
actions (trigger/retry/cancel), through the dbt Cloud Administrative API v2.

Readable streams: `projects`, `runs`, `repositories`, `users`, `environments`, `jobs`, `invites`,
`licenses`, `notifications`, `ssh_tunnels`.

Write actions: `create_job`, `update_job`, `delete_job`, `trigger_job_run`, `retry_failed_job`,
`cancel_run`, `retry_run`, `create_notification`, `update_notification`, `delete_notification`,
`create_ssh_tunnel`, `update_ssh_tunnel`, `delete_ssh_tunnel`.

Service API documentation: https://docs.getdbt.com/dbt-cloud/api-v2.

## Auth setup

Connection fields:

- `account_id` (required, string); dbt Cloud account ID; every stream is scoped under
  /accounts/{account_id}/.
- `api_key_2` (required, secret, string); dbt Cloud service token, sent as the Authorization header
  in the form 'Token <api_key_2>'. Never logged.
- `base_url` (optional, string); default `https://cloud.getdbt.com/api/v2`; format `uri`; dbt Cloud
  Administrative API v2 base URL override for tests or self-hosted instances.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key_2`.

Default configuration values: `base_url=https://cloud.getdbt.com/api/v2`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token` using `secrets.api_key_2`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts/{{ config.account_id }}/projects/` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `licenses`; offset_limit: `projects`, `runs`, `repositories`, `users`,
`environments`, `jobs`, `invites`, `notifications`, `ssh_tunnels`.

- `projects`: GET `/accounts/{{ config.account_id }}/projects/` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `runs`: GET `/accounts/{{ config.account_id }}/runs/` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `repositories`: GET `/accounts/{{ config.account_id }}/repositories/` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `users`: GET `/accounts/{{ config.account_id }}/users/` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `environments`: GET `/accounts/{{ config.account_id }}/environments/` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `jobs`: GET `/accounts/{{ config.account_id }}/jobs/` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `invites`: GET `/accounts/{{ config.account_id }}/invites/` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `licenses`: GET `/accounts/{{ config.account_id }}/licenses/` - records path `data`.
- `notifications`: GET `/accounts/{{ config.account_id }}/notifications/` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `ssh_tunnels`: GET `/accounts/{{ config.account_id }}/encryptions/` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

Overall write risk: external mutation of dbt Cloud job/notification/SSH-tunnel definitions and
job/run control actions; trigger_job_run/retry_failed_job/retry_run run real warehouse queries and
cancel_run stops one mid-execution, so every write ships an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_job`: POST `/accounts/{{ config.account_id }}/jobs/` - kind `create`; body type `json`;
  required record fields `project_id`, `environment_id`, `name`, `execute_steps`; accepted fields
  `dbt_version`, `description`, `environment_id`, `execute_steps`, `generate_docs`, `name`,
  `project_id`, `run_generate_sources`, `settings`, `triggers`; risk: creates a new
  scheduled/triggerable dbt Cloud job definition; low-risk until triggered, no approval required.
- `update_job`: POST `/accounts/{{ config.account_id }}/jobs/{{ record.id }}/` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `dbt_version`,
  `description`, `execute_steps`, `generate_docs`, `id`, `name`, `settings`, `state`, `triggers`;
  risk: mutates an existing dbt Cloud job's definition (steps, schedule, environment); a changed
  schema/target affects the next triggered run, external mutation, approval required.
- `delete_job`: DELETE `/accounts/{{ config.account_id }}/jobs/{{ record.id }}/` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: irreversible removal of a job definition (its
  schedule/trigger and run history reference); approval required.
- `trigger_job_run`: POST `/accounts/{{ config.account_id }}/jobs/{{ record.job_id }}/run/` - kind
  `custom`; body type `json`; path fields `job_id`; required record fields `job_id`, `cause`;
  accepted fields `cause`, `dbt_version_override`, `generate_docs_override`, `git_branch`,
  `git_sha`, `job_id`, `schema_override`, `steps_override`, `target_name_override`,
  `threads_override`; risk: kicks off a real dbt Cloud job run against the configured warehouse
  connection (builds/materializes models, can run arbitrary project SQL); external mutation with
  warehouse side effects, approval required.
- `retry_failed_job`: POST `/accounts/{{ config.account_id }}/jobs/{{ record.job_id }}/rerun/` -
  kind `custom`; body type `none`; path fields `job_id`; required record fields `job_id`; accepted
  fields `job_id`; risk: retries a job's most recent failed run from the point of failure; runs real
  warehouse queries, external mutation with warehouse side effects, approval required.
- `cancel_run`: POST `/accounts/{{ config.account_id }}/runs/{{ record.run_id }}/cancel/` - kind
  `custom`; body type `none`; path fields `run_id`; required record fields `run_id`; accepted fields
  `run_id`; risk: cancels an in-progress dbt Cloud run; stops warehouse queries mid-execution,
  external mutation, approval required.
- `retry_run`: POST `/accounts/{{ config.account_id }}/runs/{{ record.run_id }}/retry/` - kind
  `custom`; body type `none`; path fields `run_id`; required record fields `run_id`; accepted fields
  `run_id`; risk: retries a specific failed run from the point of failure; runs real warehouse
  queries, external mutation with warehouse side effects, approval required.
- `create_notification`: POST `/accounts/{{ config.account_id }}/notifications/` - kind `create`;
  body type `json`; required record fields `user_id`, `on_cancel`, `on_failure`, `on_success`,
  `on_warning`, `state`; accepted fields `external_email`, `on_cancel`, `on_failure`, `on_success`,
  `on_warning`, `slack_channel_id`, `slack_channel_name`, `state`, `user_id`; risk: registers an
  outbound job-status notification (email or Slack channel of the caller's choosing); low-risk
  external mutation, no approval required.
- `update_notification`: POST `/accounts/{{ config.account_id }}/notifications/{{ record.id }}/` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `external_email`, `id`, `on_cancel`, `on_failure`, `on_success`, `on_warning`, `slack_channel_id`,
  `slack_channel_name`, `state`; risk: repoints or reconfigures an existing job-status
  notification's destination (email/Slack channel); external mutation, approval required for a
  changed destination.
- `delete_notification`: DELETE `/accounts/{{ config.account_id }}/notifications/{{ record.id }}/` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: removes an existing job-status
  notification configuration; approval required.
- `create_ssh_tunnel`: POST `/accounts/{{ config.account_id }}/encryptions/` - kind `create`; body
  type `json`; required record fields `connection_id`, `username`, `port`, `hostname`, `state`;
  accepted fields `connection_id`, `hostname`, `port`, `private_key`, `public_key`, `state`,
  `username`; risk: creates an SSH tunnel encrypting traffic for a warehouse connection; may carry a
  private key in the request body, external mutation, approval required.
- `update_ssh_tunnel`: POST `/accounts/{{ config.account_id }}/encryptions/{{ record.id }}/` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `hostname`, `id`, `port`, `private_key`, `public_key`, `state`, `username`; risk: mutates an
  existing SSH tunnel's connection details; may carry a private key in the request body, external
  mutation, approval required.
- `delete_ssh_tunnel`: DELETE `/accounts/{{ config.account_id }}/encryptions/{{ record.id }}/` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: removes an SSH tunnel; the
  associated warehouse connection falls back to unencrypted/direct connectivity, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, deprecated=9, duplicate_of=10, non_data_endpoint=1, out_of_scope=1,
  requires_elevated_scope=5.
