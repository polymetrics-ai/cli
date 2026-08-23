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

- account_id (required)
- base_url
- mode
- api_key_2 (secret) (required)

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
