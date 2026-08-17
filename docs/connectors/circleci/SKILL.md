---
name: pm-circleci
description: CircleCI connector knowledge and safe action guide.
---

# pm-circleci

## Purpose

Reads and writes CircleCI projects, pipelines, workflows, jobs, contexts, schedules, environment variables, checkout keys, and workflow insights through the CircleCI v2 REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- org
- pipeline_id
- repo
- vcs_type
- workflow_id
- api_key (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: default_branch(), id(), name(), organization_id(), organization_name(), organization_slug(), slug(), vcs_url()
- pipelines:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), number(), project_slug(), state(), updated_at()
- workflows:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name(), pipeline_id(), pipeline_number(), project_slug(), status(), stopped_at()
- jobs:
  - primary key: id
  - cursor: started_at
  - fields: id(), job_number(), name(), project_slug(), started_at(), status(), stopped_at(), type()
- contexts:
  - primary key: id
  - fields: created_at(), id(), name()
- schedules:
  - primary key: id
  - cursor: updated-at
  - fields: actor(), created-at(), description(), id(), name(), parameters(), project-slug(), timetable(), updated-at()
- checkout_keys:
  - primary key: fingerprint
  - fields: created-at(), fingerprint(), preferred(), public-key(), type()
- environment_variables:
  - primary key: name
  - fields: created-at(), name(), value()
- insights_workflow_summary:
  - primary key: name
  - fields: metrics(), name(), project_id(), window_end(), window_start()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_schedule:
  - endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule
  - risk: external mutation; creates a new scheduled-pipeline trigger for this project
- update_schedule:
  - endpoint: PATCH /schedule/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing scheduled-pipeline trigger's timetable or parameters
- delete_schedule:
  - endpoint: DELETE /schedule/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a scheduled-pipeline trigger; approval required
- create_environment_variable:
  - endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar
  - risk: external mutation; creates or overwrites a project environment variable used by every future CI run
- delete_environment_variable:
  - endpoint: DELETE /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar/{{ record.name }}
  - required fields: name
  - risk: irreversible external deletion of a project environment variable; may break future CI runs that depend on it; approval required
- create_checkout_key:
  - endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key
  - risk: external mutation; creates a new deploy/checkout SSH key with repository access
- delete_checkout_key:
  - endpoint: DELETE /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key/{{ record.fingerprint }}
  - required fields: fingerprint
  - risk: irreversible external revocation of a deploy/checkout SSH key; may break future CI checkouts that depend on it; approval required

## Security

- read risk: external CircleCI API read of CI project, pipeline, workflow, job, context, schedule, environment-variable, checkout-key, and workflow-insight metadata
- write risk: external mutation of CircleCI project configuration: schedule/environment-variable/checkout-key create and delete; never triggers, cancels, or approves a live CI run
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect circleci
```

### Inspect as structured JSON

```bash
pm connectors inspect circleci --json
```

## Agent Rules

- Run pm connectors inspect circleci before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
