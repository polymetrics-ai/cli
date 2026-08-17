# pm connectors inspect buildkite

```text
NAME
  pm connectors inspect buildkite - Buildkite connector manual

SYNOPSIS
  pm connectors inspect buildkite
  pm connectors inspect buildkite --json
  pm credentials add <name> --connector buildkite [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Buildkite organizations, pipelines, builds, agents, teams, and clusters through the Buildkite REST API v2.

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
  base_url
  mode
  organization
  start_date
  api_key (secret)

ETL STREAMS
  organizations:
    primary key: id
    cursor: created_at
    fields: agents_url(), created_at(), graphql_id(), id(), name(), pipelines_url(), slug(), url(), web_url()
  pipelines:
    primary key: id
    cursor: created_at
    fields: archived_at(), builds_url(), created_at(), default_branch(), description(), graphql_id(), id(), name(), repository(), slug(), url(), visibility(), web_url()
  builds:
    primary key: id
    cursor: created_at
    fields: blocked(), branch(), commit(), created_at(), finished_at(), graphql_id(), id(), message(), number(), scheduled_at(), source(), started_at(), state(), url(), web_url()
  agents:
    primary key: id
    cursor: created_at
    fields: connection_state(), created_at(), graphql_id(), hostname(), id(), ip_address(), last_job_finished_at(), name(), priority(), url(), user_agent(), version(), web_url()
  teams:
    primary key: id
    fields: created_at(), default(), description(), graphql_id(), id(), name(), privacy(), slug()
  clusters:
    primary key: id
    fields: color(), created_at(), default_queue_id(), description(), emoji(), graphql_id(), id(), name(), url(), web_url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_pipeline:
    endpoint: POST /organizations/{{ config.organization }}/pipelines
    risk: creates a new CI/CD pipeline scoped to a cluster and repository; low-risk external mutation, no approval required
  update_pipeline:
    endpoint: PATCH /organizations/{{ config.organization }}/pipelines/{{ record.slug }}
    required fields: slug
    risk: mutates an existing pipeline's repository, configuration, or visibility; a changed configuration/repository affects every future build
  archive_pipeline:
    endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.slug }}/archive
    required fields: slug
    risk: archives a pipeline, hiding it from the default pipeline list and blocking new builds until unarchived
  unarchive_pipeline:
    endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.slug }}/unarchive
    required fields: slug
    risk: restores a previously archived pipeline to active/buildable status
  delete_pipeline:
    endpoint: DELETE /organizations/{{ config.organization }}/pipelines/{{ record.slug }}
    required fields: slug
    risk: permanently deletes a pipeline and its build history; irreversible
  create_build:
    endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds
    required fields: pipeline_slug
    risk: immediately triggers a new CI/CD build on the target pipeline/branch; consumes agent capacity and may run arbitrary pipeline-defined commands
  cancel_build:
    endpoint: PUT /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds/{{ record.number }}/cancel
    required fields: pipeline_slug, number
    risk: cancels a running or scheduled build; any in-progress jobs are terminated immediately
  rebuild_build:
    endpoint: PUT /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds/{{ record.number }}/rebuild
    required fields: pipeline_slug, number
    risk: triggers a full re-run of a completed build on new agent capacity; may run arbitrary pipeline-defined commands again
  create_annotation:
    endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds/{{ record.build_number }}/annotations
    required fields: pipeline_slug, build_number
    risk: posts a visible HTML/Markdown annotation onto a build's detail page; low-risk external mutation, no approval required
  retry_job:
    endpoint: PUT /organizations/{{ config.organization }}/jobs/{{ record.job_id }}/retry
    required fields: job_id
    risk: re-runs a single failed/finished job on new agent capacity, without re-running the rest of the build
  unblock_job:
    endpoint: PUT /organizations/{{ config.organization }}/jobs/{{ record.job_id }}/unblock
    required fields: job_id
    optional fields: fields, unblocker
    risk: releases a manual 'block' pipeline step, allowing the build to continue past it immediately
  stop_agent:
    endpoint: PUT /organizations/{{ config.organization }}/agents/{{ record.id }}/stop
    required fields: id
    optional fields: force
    risk: stops an agent; force=true cancels any job it is currently processing
  pause_agent:
    endpoint: PUT /organizations/{{ config.organization }}/agents/{{ record.id }}/pause
    required fields: id
    optional fields: note, timeout_in_minutes
    risk: pauses an agent so it stops picking up new jobs until resumed or the timeout elapses
  resume_agent:
    endpoint: PUT /organizations/{{ config.organization }}/agents/{{ record.id }}/resume
    required fields: id
    risk: resumes a previously paused agent so it can pick up new jobs again
  create_team:
    endpoint: POST /organizations/{{ config.organization }}/teams
    risk: creates a new team; low-risk external mutation, no approval required
  update_team:
    endpoint: PATCH /organizations/{{ config.organization }}/teams/{{ record.id }}
    required fields: id
    risk: mutates an existing team's name, privacy, or default permissions; a privacy change from visible to secret hides membership immediately
  delete_team:
    endpoint: DELETE /organizations/{{ config.organization }}/teams/{{ record.id }}
    required fields: id
    risk: permanently deletes a team and its pipeline/member associations; irreversible

SECURITY
  read risk: external Buildkite API read of organization, pipeline, build, agent, team, and cluster data
  write risk: external mutation of pipeline lifecycle, build triggering/cancellation, job control, agent lifecycle, and team management; create_build/rebuild_build run arbitrary pipeline-defined commands on real agent capacity
  approval: required for all write actions; each action's per-record risk string in writes.json is the authoritative summary
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect buildkite

  # Inspect as structured JSON
  pm connectors inspect buildkite --json

AGENT WORKFLOW
  - Run pm connectors inspect buildkite before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
