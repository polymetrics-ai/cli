---
name: pm-buildkite
description: Buildkite connector knowledge and safe action guide.
---

# pm-buildkite

## Purpose

Reads and writes Buildkite organizations, pipelines, builds, agents, teams, and clusters through the Buildkite REST API v2.

## Icon

- id: simple-icons-buildkite
- asset: icons/simple-icons/buildkite.svg
- title: Buildkite
- simple_icon_slug: buildkite
- simple_icon_hex: 14CC80
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Buildkite
- match: exact-name-or-slug
- matched_by: buildkite

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- organization
- start_date
- api_key (secret) (required)

## ETL Streams

- organizations:
  - primary key: id
  - cursor: created_at
  - fields: agents_url(string), created_at(string), graphql_id(string), id(string), name(string), pipelines_url(string), slug(string), url(string), web_url(string)
- pipelines:
  - primary key: id
  - cursor: created_at
  - fields: archived_at(string), builds_url(string), created_at(string), default_branch(string), description(string), graphql_id(string), id(string), name(string), repository(string), slug(string), url(string), visibility(string), web_url(string)
- builds:
  - primary key: id
  - cursor: created_at
  - fields: blocked(boolean), branch(string), commit(string), created_at(string), finished_at(string), graphql_id(string), id(string), message(string), number(integer), scheduled_at(string), source(string), started_at(string), state(string), url(string), web_url(string)
- agents:
  - primary key: id
  - cursor: created_at
  - fields: connection_state(string), created_at(string), graphql_id(string), hostname(string), id(string), ip_address(string), last_job_finished_at(string), name(string), priority(integer), url(string), user_agent(string), version(string), web_url(string)
- teams:
  - primary key: id
  - fields: created_at(string), default(boolean), description(string), graphql_id(string), id(string), name(string), privacy(string), slug(string)
- clusters:
  - primary key: id
  - fields: color(string), created_at(string), default_queue_id(string), description(string), emoji(string), graphql_id(string), id(string), name(string), url(string), web_url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_pipeline:
  - endpoint: POST /organizations/{{ config.organization }}/pipelines
  - required fields: name, cluster_id, repository
  - risk: creates a new CI/CD pipeline scoped to a cluster and repository; low-risk external mutation, no approval required
- update_pipeline:
  - endpoint: PATCH /organizations/{{ config.organization }}/pipelines/{{ record.slug }}
  - required fields: slug
  - risk: mutates an existing pipeline's repository, configuration, or visibility; a changed configuration/repository affects every future build
- archive_pipeline:
  - endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.slug }}/archive
  - required fields: slug
  - risk: archives a pipeline, hiding it from the default pipeline list and blocking new builds until unarchived
- unarchive_pipeline:
  - endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.slug }}/unarchive
  - required fields: slug
  - risk: restores a previously archived pipeline to active/buildable status
- delete_pipeline:
  - endpoint: DELETE /organizations/{{ config.organization }}/pipelines/{{ record.slug }}
  - required fields: slug
  - risk: permanently deletes a pipeline and its build history; irreversible
- create_build:
  - endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds
  - required fields: pipeline_slug, commit, branch
  - risk: immediately triggers a new CI/CD build on the target pipeline/branch; consumes agent capacity and may run arbitrary pipeline-defined commands
- cancel_build:
  - endpoint: PUT /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds/{{ record.number }}/cancel
  - required fields: pipeline_slug, number
  - risk: cancels a running or scheduled build; any in-progress jobs are terminated immediately
- rebuild_build:
  - endpoint: PUT /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds/{{ record.number }}/rebuild
  - required fields: pipeline_slug, number
  - risk: triggers a full re-run of a completed build on new agent capacity; may run arbitrary pipeline-defined commands again
- create_annotation:
  - endpoint: POST /organizations/{{ config.organization }}/pipelines/{{ record.pipeline_slug }}/builds/{{ record.build_number }}/annotations
  - required fields: pipeline_slug, build_number, body
  - risk: posts a visible HTML/Markdown annotation onto a build's detail page; low-risk external mutation, no approval required
- retry_job:
  - endpoint: PUT /organizations/{{ config.organization }}/jobs/{{ record.job_id }}/retry
  - required fields: job_id
  - risk: re-runs a single failed/finished job on new agent capacity, without re-running the rest of the build
- unblock_job:
  - endpoint: PUT /organizations/{{ config.organization }}/jobs/{{ record.job_id }}/unblock
  - required fields: job_id
  - optional fields: fields, unblocker
  - risk: releases a manual 'block' pipeline step, allowing the build to continue past it immediately
- stop_agent:
  - endpoint: PUT /organizations/{{ config.organization }}/agents/{{ record.id }}/stop
  - required fields: id
  - optional fields: force
  - risk: stops an agent; force=true cancels any job it is currently processing
- pause_agent:
  - endpoint: PUT /organizations/{{ config.organization }}/agents/{{ record.id }}/pause
  - required fields: id
  - optional fields: note, timeout_in_minutes
  - risk: pauses an agent so it stops picking up new jobs until resumed or the timeout elapses
- resume_agent:
  - endpoint: PUT /organizations/{{ config.organization }}/agents/{{ record.id }}/resume
  - required fields: id
  - risk: resumes a previously paused agent so it can pick up new jobs again
- create_team:
  - endpoint: POST /organizations/{{ config.organization }}/teams
  - required fields: name
  - risk: creates a new team; low-risk external mutation, no approval required
- update_team:
  - endpoint: PATCH /organizations/{{ config.organization }}/teams/{{ record.id }}
  - required fields: id
  - risk: mutates an existing team's name, privacy, or default permissions; a privacy change from visible to secret hides membership immediately
- delete_team:
  - endpoint: DELETE /organizations/{{ config.organization }}/teams/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a team and its pipeline/member associations; irreversible

## Security

- read risk: external Buildkite API read of organization, pipeline, build, agent, team, and cluster data
- write risk: external mutation of pipeline lifecycle, build triggering/cancellation, job control, agent lifecycle, and team management; create_build/rebuild_build run arbitrary pipeline-defined commands on real agent capacity
- approval: required for all write actions; each action's per-record risk string in writes.json is the authoritative summary
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect buildkite
```

### Inspect as structured JSON

```bash
pm connectors inspect buildkite --json
```

## Agent Rules

- Run pm connectors inspect buildkite before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
