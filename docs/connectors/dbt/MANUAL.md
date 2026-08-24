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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id (required)
  base_url
  mode
  api_key_2 (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    fields: account_id(integer), connection_id(integer), created_at(string), dbt_project_subdirectory(string), description(string), id(integer), name(string), repository_id(integer), state(integer), updated_at(string)
  runs:
    primary key: id
    fields: account_id(integer), created_at(string), environment_id(integer), finished_at(string), id(integer), is_cancelled(boolean), is_complete(boolean), is_error(boolean), job_definition_id(integer), project_id(integer), started_at(string), status(integer), status_humanized(string), updated_at(string)
  repositories:
    primary key: id
    fields: account_id(integer), created_at(string), git_clone_strategy(string), id(integer), project_id(integer), remote_backend(string), remote_url(string), state(integer), updated_at(string)
  users:
    primary key: id
    fields: account_id(integer), created_at(string), email(string), first_name(string), fullname(string), id(integer), is_active(boolean), last_name(string)
  environments:
    primary key: id
    fields: account_id(integer), created_at(string), custom_branch(string), dbt_version(string), id(integer), name(string), project_id(integer), state(integer), type(string), updated_at(string), use_custom_branch(boolean)
  jobs:
    primary key: id
    cursor: updated_at
    fields: account_id(integer), created_at(string), dbt_version(string), description(string), environment_id(integer), execute_steps(array), generate_docs(boolean), id(integer), job_type(string), name(string), project_id(integer), run_generate_sources(boolean), state(integer), triggers_on_draft_pr(boolean), updated_at(string)
  invites:
    primary key: id
    cursor: created_at
    fields: account_id(integer), created_at(string), email_address(string), group_ids(array), id(integer), license_type(string), redeemed_at(string), status(integer), type(string)
  licenses:
    primary key: account_id
    fields: account_id(integer), analyst(object), developer(object), explorer(object), it(object), read_only(object)
  notifications:
    primary key: id
    cursor: updated_at
    fields: account_id(integer), created_at(string), external_email(string), id(integer), on_cancel(array), on_failure(array), on_success(array), on_warning(array), slack_channel_id(string), slack_channel_name(string), state(integer), updated_at(string), user_id(integer)
  ssh_tunnels:
    primary key: id
    fields: account_id(integer), connection_id(integer), hostname(string), id(integer), port(integer), public_key(string), state(integer), username(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_job:
    endpoint: POST /accounts/{{ config.account_id }}/jobs/
    required fields: project_id, environment_id, name, execute_steps
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
    required fields: job_id, cause
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
    required fields: user_id, on_cancel, on_failure, on_success, on_warning, state
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
    required fields: connection_id, username, port, hostname, state
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

COMMAND SURFACE
  Run dbt Cloud's declared typed write actions.
  Usage: pm dbt <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    cancel run apply - Typed action cancel_run [intent=reverse_etl availability=partial write=cancel_run]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: cancels an in-progress dbt Cloud run; stops warehouse queries mid-execution, external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --run-id (required)
    create job apply - Typed action create_job [intent=reverse_etl availability=partial write=create_job]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: creates a new scheduled/triggerable dbt Cloud job definition; low-risk until triggered, no approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --environment-id (required), --execute-steps (required), --name (required), --project-id (required)
    create notification apply - Typed action create_notification [intent=reverse_etl availability=partial write=create_notification]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: registers an outbound job-status notification (email or Slack channel of the caller's choosing); low-risk external mutation, no approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --on-cancel (required), --on-failure (required), --on-success (required), --on-warning (required), --state (required), --user-id (required)
    create ssh tunnel apply - Typed action create_ssh_tunnel [intent=reverse_etl availability=partial write=create_ssh_tunnel]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: creates an SSH tunnel encrypting traffic for a warehouse connection; may carry a private key in the request body, external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --connection-id (required), --hostname (required), --port (required), --state (required), --username (required)
    delete job apply - Typed action delete_job [intent=reverse_etl availability=partial write=delete_job]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: irreversible removal of a job definition (its schedule/trigger and run history reference); approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --id (required)
    delete notification apply - Typed action delete_notification [intent=reverse_etl availability=partial write=delete_notification]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: removes an existing job-status notification configuration; approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --id (required)
    delete ssh tunnel apply - Typed action delete_ssh_tunnel [intent=reverse_etl availability=partial write=delete_ssh_tunnel]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: removes an SSH tunnel; the associated warehouse connection falls back to unencrypted/direct connectivity, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --id (required)
    retry failed job apply - Typed action retry_failed_job [intent=reverse_etl availability=partial write=retry_failed_job]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: retries a job's most recent failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --job-id (required)
    retry run apply - Typed action retry_run [intent=reverse_etl availability=partial write=retry_run]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: retries a specific failed run from the point of failure; runs real warehouse queries, external mutation with warehouse side effects, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --run-id (required)
    trigger job run apply - Typed action trigger_job_run [intent=reverse_etl availability=partial write=trigger_job_run]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: kicks off a real dbt Cloud job run against the configured warehouse connection (builds/materializes models, can run arbitrary project SQL); external mutation with warehouse side effects, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --cause (required), --job-id (required)
    update job apply - Typed action update_job [intent=reverse_etl availability=partial write=update_job]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: mutates an existing dbt Cloud job's definition (steps, schedule, environment); a changed schema/target affects the next triggered run, external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --id (required)
    update notification apply - Typed action update_notification [intent=reverse_etl availability=partial write=update_notification]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: repoints or reconfigures an existing job-status notification's destination (email/Slack channel); external mutation, approval required for a changed destination; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --id (required)
    update ssh tunnel apply - Typed action update_ssh_tunnel [intent=reverse_etl availability=partial write=update_ssh_tunnel]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; risk: mutates an existing SSH tunnel's connection details; may carry a private key in the request body, external mutation, approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.account_id }}.; flags: --id (required)

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

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
