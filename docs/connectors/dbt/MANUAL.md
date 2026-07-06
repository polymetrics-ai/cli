# pm connectors inspect dbt

```text
NAME
  pm connectors inspect dbt - dbt Cloud connector manual

SYNOPSIS
  pm connectors inspect dbt
  pm connectors inspect dbt --json
  pm credentials add <name> --connector dbt [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads dbt Cloud projects, runs, repositories, users, environments, jobs, invites, licenses, notifications, and SSH tunnels, and writes job/notification/SSH-tunnel mutations and run-control actions (trigger/retry/cancel), through the dbt Cloud Administrative API v2.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  mode
  api_key_2 (secret)

ETL STREAMS
  projects:
    primary key: id
    fields: account_id(), connection_id(), created_at(), dbt_project_subdirectory(), description(), id(), name(), repository_id(), state(), updated_at()
  runs:
    primary key: id
    fields: account_id(), created_at(), environment_id(), finished_at(), id(), is_cancelled(), is_complete(), is_error(), job_definition_id(), project_id(), started_at(), status(), status_humanized(), updated_at()
  repositories:
    primary key: id
    fields: account_id(), created_at(), git_clone_strategy(), id(), project_id(), remote_backend(), remote_url(), state(), updated_at()
  users:
    primary key: id
    fields: account_id(), created_at(), email(), first_name(), fullname(), id(), is_active(), last_name()
  environments:
    primary key: id
    fields: account_id(), created_at(), custom_branch(), dbt_version(), id(), name(), project_id(), state(), type(), updated_at(), use_custom_branch()
  jobs:
    primary key: id
    cursor: updated_at
    fields: account_id(), created_at(), dbt_version(), description(), environment_id(), execute_steps(), generate_docs(), id(), job_type(), name(), project_id(), run_generate_sources(), state(), triggers_on_draft_pr(), updated_at()
  invites:
    primary key: id
    cursor: created_at
    fields: account_id(), created_at(), email_address(), group_ids(), id(), license_type(), redeemed_at(), status(), type()
  licenses:
    primary key: account_id
    fields: account_id(), analyst(), developer(), explorer(), it(), read_only()
  notifications:
    primary key: id
    cursor: updated_at
    fields: account_id(), created_at(), external_email(), id(), on_cancel(), on_failure(), on_success(), on_warning(), slack_channel_id(), slack_channel_name(), state(), updated_at(), user_id()
  ssh_tunnels:
    primary key: id
    fields: account_id(), connection_id(), hostname(), id(), port(), public_key(), state(), username()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_job:
    endpoint: POST /accounts/{{ config.account_id }}/jobs/
    risk: creates a new scheduled/triggerable dbt Cloud job definition; low-risk until triggered, no approval required
  update_job:
    endpoint: POST /accounts/{{ config.account_id }}/jobs/{{ record.id }}/
    required fields: id
    risk: mutates an existing dbt Cloud job's definition (steps, schedule, environment); a changed schema/target affects the next triggered run, external mutation, approval required
  delete_job:
    endpoint: DELETE /accounts/{{ config.account_id }}/jobs/{{ record.id }}/
    required fields: id
    risk: irreversible removal of a job definition (its schedule/trigger and run history reference); approval required
  trigger_job_run:
    endpoint: POST /accounts/{{ config.account_id }}/jobs/{{ record.job_id }}/run/
    required fields: job_id
    risk: kicks off a real dbt Cloud job run against the configured warehouse connection (builds/materializes models, can run arbitrary project SQL); external mutation with warehouse side effects, approval required
  retry_failed_job:
    endpoint: POST /accounts/{{ config.account_id }}/jobs/{{ record.job_id }}/rerun/
    required fields: job_id
    risk: retries a job's most recent failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required
  cancel_run:
    endpoint: POST /accounts/{{ config.account_id }}/runs/{{ record.run_id }}/cancel/
    required fields: run_id
    risk: cancels an in-progress dbt Cloud run; stops warehouse queries mid-execution, external mutation, approval required
  retry_run:
    endpoint: POST /accounts/{{ config.account_id }}/runs/{{ record.run_id }}/retry/
    required fields: run_id
    risk: retries a specific failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required
  create_notification:
    endpoint: POST /accounts/{{ config.account_id }}/notifications/
    risk: registers an outbound job-status notification (email or Slack channel of the caller's choosing); low-risk external mutation, no approval required
  update_notification:
    endpoint: POST /accounts/{{ config.account_id }}/notifications/{{ record.id }}/
    required fields: id
    risk: repoints or reconfigures an existing job-status notification's destination (email/Slack channel); external mutation, approval required for a changed destination
  delete_notification:
    endpoint: DELETE /accounts/{{ config.account_id }}/notifications/{{ record.id }}/
    required fields: id
    risk: removes an existing job-status notification configuration; approval required
  create_ssh_tunnel:
    endpoint: POST /accounts/{{ config.account_id }}/encryptions/
    risk: creates an SSH tunnel encrypting traffic for a warehouse connection; may carry a private key in the request body, external mutation, approval required
  update_ssh_tunnel:
    endpoint: POST /accounts/{{ config.account_id }}/encryptions/{{ record.id }}/
    required fields: id
    risk: mutates an existing SSH tunnel's connection details; may carry a private key in the request body, external mutation, approval required
  delete_ssh_tunnel:
    endpoint: DELETE /accounts/{{ config.account_id }}/encryptions/{{ record.id }}/
    required fields: id
    risk: removes an SSH tunnel; the associated warehouse connection falls back to unencrypted/direct connectivity, approval required

SECURITY
  read risk: external dbt Cloud API read of account projects, runs, repositories, users, environments, jobs, invites, licenses, notifications, and SSH tunnel configuration
  write risk: external mutation of dbt Cloud job/notification/SSH-tunnel definitions and job/run control actions; trigger_job_run/retry_failed_job/retry_run run real warehouse queries and cancel_run stops one mid-execution, so every write ships an explicit per-action risk string
  approval: required for delete_job/delete_notification/delete_ssh_tunnel (irreversible) and for trigger_job_run/retry_failed_job/retry_run/cancel_run (real warehouse side effects); create_job/create_notification are low-risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect dbt

  # Inspect as structured JSON
  pm connectors inspect dbt --json

AGENT WORKFLOW
  - Run pm connectors inspect dbt before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
