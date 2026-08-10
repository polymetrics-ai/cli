---
name: pm-dbt
description: dbt Cloud connector knowledge and safe action guide.
---

# pm-dbt

## Purpose

Reads dbt Cloud projects, runs, repositories, users, environments, jobs, invites, licenses, notifications, and SSH tunnels, and writes job/notification/SSH-tunnel mutations and run-control actions (trigger/retry/cancel), through the dbt Cloud Administrative API v2.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- mode
- api_key_2 (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: account_id(integer), connection_id(integer), created_at(string), dbt_project_subdirectory(string), description(string), id(integer), name(string), repository_id(integer), state(integer), updated_at(string)
- runs:
  - primary key: id
  - fields: account_id(integer), created_at(string), environment_id(integer), finished_at(string), id(integer), is_cancelled(boolean), is_complete(boolean), is_error(boolean), job_definition_id(integer), project_id(integer), started_at(string), status(integer), status_humanized(string), updated_at(string)
- repositories:
  - primary key: id
  - fields: account_id(integer), created_at(string), git_clone_strategy(string), id(integer), project_id(integer), remote_backend(string), remote_url(string), state(integer), updated_at(string)
- users:
  - primary key: id
  - fields: account_id(integer), created_at(string), email(string), first_name(string), fullname(string), id(integer), is_active(boolean), last_name(string)
- environments:
  - primary key: id
  - fields: account_id(integer), created_at(string), custom_branch(string), dbt_version(string), id(integer), name(string), project_id(integer), state(integer), type(string), updated_at(string), use_custom_branch(boolean)
- jobs:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(integer), created_at(string), dbt_version(string), description(string), environment_id(integer), execute_steps(array), generate_docs(boolean), id(integer), job_type(string), name(string), project_id(integer), run_generate_sources(boolean), state(integer), triggers_on_draft_pr(boolean), updated_at(string)
- invites:
  - primary key: id
  - cursor: created_at
  - fields: account_id(integer), created_at(string), email_address(string), group_ids(array), id(integer), license_type(string), redeemed_at(string), status(integer), type(string)
- licenses:
  - primary key: account_id
  - fields: account_id(integer), analyst(object), developer(object), explorer(object), it(object), read_only(object)
- notifications:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(integer), created_at(string), external_email(string), id(integer), on_cancel(array), on_failure(array), on_success(array), on_warning(array), slack_channel_id(string), slack_channel_name(string), state(integer), updated_at(string), user_id(integer)
- ssh_tunnels:
  - primary key: id
  - fields: account_id(integer), connection_id(integer), hostname(string), id(integer), port(integer), public_key(string), state(integer), username(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_job:
  - endpoint: POST /accounts/{{ config.account_id }}/jobs/
  - required fields: project_id, environment_id, name, execute_steps
  - risk: creates a new scheduled/triggerable dbt Cloud job definition; low-risk until triggered, no approval required
- update_job:
  - endpoint: POST /accounts/{{ config.account_id }}/jobs/{{ record.id }}/
  - required fields: id
  - risk: mutates an existing dbt Cloud job's definition (steps, schedule, environment); a changed schema/target affects the next triggered run, external mutation, approval required
- delete_job:
  - endpoint: DELETE /accounts/{{ config.account_id }}/jobs/{{ record.id }}/
  - required fields: id
  - risk: irreversible removal of a job definition (its schedule/trigger and run history reference); approval required
- trigger_job_run:
  - endpoint: POST /accounts/{{ config.account_id }}/jobs/{{ record.job_id }}/run/
  - required fields: job_id, cause
  - risk: kicks off a real dbt Cloud job run against the configured warehouse connection (builds/materializes models, can run arbitrary project SQL); external mutation with warehouse side effects, approval required
- retry_failed_job:
  - endpoint: POST /accounts/{{ config.account_id }}/jobs/{{ record.job_id }}/rerun/
  - required fields: job_id
  - risk: retries a job's most recent failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required
- cancel_run:
  - endpoint: POST /accounts/{{ config.account_id }}/runs/{{ record.run_id }}/cancel/
  - required fields: run_id
  - risk: cancels an in-progress dbt Cloud run; stops warehouse queries mid-execution, external mutation, approval required
- retry_run:
  - endpoint: POST /accounts/{{ config.account_id }}/runs/{{ record.run_id }}/retry/
  - required fields: run_id
  - risk: retries a specific failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required
- create_notification:
  - endpoint: POST /accounts/{{ config.account_id }}/notifications/
  - required fields: user_id, on_cancel, on_failure, on_success, on_warning, state
  - risk: registers an outbound job-status notification (email or Slack channel of the caller's choosing); low-risk external mutation, no approval required
- update_notification:
  - endpoint: POST /accounts/{{ config.account_id }}/notifications/{{ record.id }}/
  - required fields: id
  - risk: repoints or reconfigures an existing job-status notification's destination (email/Slack channel); external mutation, approval required for a changed destination
- delete_notification:
  - endpoint: DELETE /accounts/{{ config.account_id }}/notifications/{{ record.id }}/
  - required fields: id
  - risk: removes an existing job-status notification configuration; approval required
- create_ssh_tunnel:
  - endpoint: POST /accounts/{{ config.account_id }}/encryptions/
  - required fields: connection_id, username, port, hostname, state
  - risk: creates an SSH tunnel encrypting traffic for a warehouse connection; may carry a private key in the request body, external mutation, approval required
- update_ssh_tunnel:
  - endpoint: POST /accounts/{{ config.account_id }}/encryptions/{{ record.id }}/
  - required fields: id
  - risk: mutates an existing SSH tunnel's connection details; may carry a private key in the request body, external mutation, approval required
- delete_ssh_tunnel:
  - endpoint: DELETE /accounts/{{ config.account_id }}/encryptions/{{ record.id }}/
  - required fields: id
  - risk: removes an SSH tunnel; the associated warehouse connection falls back to unencrypted/direct connectivity, approval required

## Security

- read risk: external dbt Cloud API read of account projects, runs, repositories, users, environments, jobs, invites, licenses, notifications, and SSH tunnel configuration
- write risk: external mutation of dbt Cloud job/notification/SSH-tunnel definitions and job/run control actions; trigger_job_run/retry_failed_job/retry_run run real warehouse queries and cancel_run stops one mid-execution, so every write ships an explicit per-action risk string
- approval: required for delete_job/delete_notification/delete_ssh_tunnel (irreversible) and for trigger_job_run/retry_failed_job/retry_run/cancel_run (real warehouse side effects); create_job/create_notification are low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run dbt Cloud's declared streams and reverse-ETL actions.
- Usage: pm dbt <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete api v2 accounts account-id environments id - Documented DELETE /api/v2/accounts/{account_id}/environments/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.delete.api-v2-accounts-account-id-environments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v2 accounts account-id projects id - Documented DELETE /api/v2/accounts/{account_id}/projects/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.delete.api-v2-accounts-account-id-projects-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v2 accounts account-id repositories id - Documented DELETE /api/v2/accounts/{account_id}/repositories/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.delete.api-v2-accounts-account-id-repositories-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get api v2 accounts - Documented GET /api/v2/accounts/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id - Documented GET /api/v2/accounts/{account_id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id encryptions id - Documented GET /api/v2/accounts/{account_id}/encryptions/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-encryptions-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id environments id - Documented GET /api/v2/accounts/{account_id}/environments/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-environments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id invites id - Documented GET /api/v2/accounts/{account_id}/invites/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-invites-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id jobs id - Documented GET /api/v2/accounts/{account_id}/jobs/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-jobs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id jobs job-id artifacts remainder - Documented GET /api/v2/accounts/{account_id}/jobs/{job_id}/artifacts/{remainder} (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-jobs-job-id-artifacts-remainder]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id notifications id - Documented GET /api/v2/accounts/{account_id}/notifications/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-notifications-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id projects id - Documented GET /api/v2/accounts/{account_id}/projects/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-projects-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id repositories id - Documented GET /api/v2/accounts/{account_id}/repositories/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-repositories-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id runs id - Documented GET /api/v2/accounts/{account_id}/runs/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-runs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id runs run-id artifacts - Documented GET /api/v2/accounts/{account_id}/runs/{run_id}/artifacts/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-runs-run-id-artifacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id runs run-id artifacts remainder - Documented GET /api/v2/accounts/{account_id}/runs/{run_id}/artifacts/{remainder} (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-runs-run-id-artifacts-remainder]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id runs run-id retry - Documented GET /api/v2/accounts/{account_id}/runs/{run_id}/retry/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-runs-run-id-retry]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 accounts account-id steps id - Documented GET /api/v2/accounts/{account_id}/steps/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-accounts-account-id-steps-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v2 users id - Documented GET /api/v2/users/{id}/ (not implemented) [intent=direct_read availability=not_implemented operation=dbt.get.api-v2-users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch api v2 accounts account-id - Documented PATCH /api/v2/accounts/{account_id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.patch.api-v2-accounts-account-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id - Documented POST /api/v2/accounts/{account_id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id environments - Documented POST /api/v2/accounts/{account_id}/environments/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id-environments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id environments id - Documented POST /api/v2/accounts/{account_id}/environments/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id-environments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id permissions id - Documented POST /api/v2/accounts/{account_id}/permissions/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id-permissions-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id projects - Documented POST /api/v2/accounts/{account_id}/projects/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id-projects]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id projects id - Documented POST /api/v2/accounts/{account_id}/projects/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id-projects-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 accounts account-id repositories - Documented POST /api/v2/accounts/{account_id}/repositories/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-accounts-account-id-repositories]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 notifications unsubscribe - Documented POST /api/v2/notifications/unsubscribe/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-notifications-unsubscribe]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v2 users id - Documented POST /api/v2/users/{id}/ (not implemented) [intent=direct_write availability=not_implemented operation=dbt.post.api-v2-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - cancel run apply - Plan and execute the cancel run reverse-ETL action [intent=reverse_etl availability=implemented write=cancel_run]; approval: requires plan, preview, approval, and execute; risk: cancels an in-progress dbt Cloud run; stops warehouse queries mid-execution, external mutation, approval required; flags: --run_id (required)
  - create job apply - Plan and execute the create job reverse-ETL action [intent=reverse_etl availability=implemented write=create_job]; approval: requires plan, preview, approval, and execute; risk: creates a new scheduled/triggerable dbt Cloud job definition; low-risk until triggered, no approval required; flags: --environment_id (required), --execute_steps (required), --name (required), --project_id (required)
  - create notification apply - Plan and execute the create notification reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_notification]; approval: requires plan, preview, approval, and execute; risk: registers an outbound job-status notification (email or Slack channel of the caller's choosing); low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create ssh tunnel apply - Plan and execute the create ssh tunnel reverse-ETL action [intent=reverse_etl availability=implemented write=create_ssh_tunnel]; approval: requires plan, preview, approval, and execute; risk: creates an SSH tunnel encrypting traffic for a warehouse connection; may carry a private key in the request body, external mutation, approval required; flags: --connection_id (required), --hostname (required), --port (required), --state (required), --username (required)
  - delete job apply - Plan and execute the delete job reverse-ETL action [intent=reverse_etl availability=implemented write=delete_job]; approval: requires plan, preview, approval, and execute; risk: irreversible removal of a job definition (its schedule/trigger and run history reference); approval required; flags: --id (required)
  - delete notification apply - Plan and execute the delete notification reverse-ETL action [intent=reverse_etl availability=implemented write=delete_notification]; approval: requires plan, preview, approval, and execute; risk: removes an existing job-status notification configuration; approval required; flags: --id (required)
  - delete ssh tunnel apply - Plan and execute the delete ssh tunnel reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ssh_tunnel]; approval: requires plan, preview, approval, and execute; risk: removes an SSH tunnel; the associated warehouse connection falls back to unencrypted/direct connectivity, approval required; flags: --id (required)
  - environments list - Run the environments ETL stream [intent=etl availability=implemented stream=environments]
  - invites list - Run the invites ETL stream [intent=etl availability=implemented stream=invites]
  - jobs list - Run the jobs ETL stream [intent=etl availability=implemented stream=jobs]
  - licenses list - Run the licenses ETL stream [intent=etl availability=implemented stream=licenses]
  - notifications list - Run the notifications ETL stream [intent=etl availability=implemented stream=notifications]
  - projects list - Run the projects ETL stream [intent=etl availability=implemented stream=projects]
  - repositories list - Run the repositories ETL stream [intent=etl availability=implemented stream=repositories]
  - retry failed job apply - Plan and execute the retry failed job reverse-ETL action [intent=reverse_etl availability=implemented write=retry_failed_job]; approval: requires plan, preview, approval, and execute; risk: retries a job's most recent failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required; flags: --job_id (required)
  - retry run apply - Plan and execute the retry run reverse-ETL action [intent=reverse_etl availability=implemented write=retry_run]; approval: requires plan, preview, approval, and execute; risk: retries a specific failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required; flags: --run_id (required)
  - runs list - Run the runs ETL stream [intent=etl availability=implemented stream=runs]
  - ssh tunnels list - Run the ssh tunnels ETL stream [intent=etl availability=implemented stream=ssh_tunnels]
  - trigger job run apply - Plan and execute the trigger job run reverse-ETL action [intent=reverse_etl availability=implemented write=trigger_job_run]; approval: requires plan, preview, approval, and execute; risk: kicks off a real dbt Cloud job run against the configured warehouse connection (builds/materializes models, can run arbitrary project SQL); external mutation with warehouse side effects, approval required; flags: --cause (required), --job_id (required)
  - update job apply - Plan and execute the update job reverse-ETL action [intent=reverse_etl availability=implemented write=update_job]; approval: requires plan, preview, approval, and execute; risk: mutates an existing dbt Cloud job's definition (steps, schedule, environment); a changed schema/target affects the next triggered run, external mutation, approval required; flags: --id (required)
  - update notification apply - Plan and execute the update notification reverse-ETL action [intent=reverse_etl availability=implemented write=update_notification]; approval: requires plan, preview, approval, and execute; risk: repoints or reconfigures an existing job-status notification's destination (email/Slack channel); external mutation, approval required for a changed destination; flags: --id (required)
  - update ssh tunnel apply - Plan and execute the update ssh tunnel reverse-ETL action [intent=reverse_etl availability=implemented write=update_ssh_tunnel]; approval: requires plan, preview, approval, and execute; risk: mutates an existing SSH tunnel's connection details; may carry a private key in the request body, external mutation, approval required; flags: --id (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

## Commands

### Inspect as a manual

```bash
pm connectors inspect dbt
```

### Inspect as structured JSON

```bash
pm connectors inspect dbt --json
```

## Agent Rules

- Run pm connectors inspect dbt before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
