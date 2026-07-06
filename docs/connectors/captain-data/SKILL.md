---
name: pm-captain-data
description: Captain Data connector knowledge and safe action guide.
---

# pm-captain-data

## Purpose

Reads Captain Data workspace, workflows, jobs, and job results, and writes a launch_workflow action to trigger a new workflow run, through the Captain Data v3 REST API.

## Icon

- asset: icons/captain-data.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.captaindata.co/api-documentation

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- job_uid
- mode
- project_uid
- workflow_uid
- api_key (secret)

## ETL Streams

- workspace:
  - primary key: uid
  - fields: created_at(), name(), plan(), uid()
- workflows:
  - primary key: uid
  - fields: created_at(), name(), status(), uid(), updated_at()
- jobs:
  - primary key: uid
  - fields: created_at(), ended_at(), status(), uid(), workflow_uid()
- job_results:
  - primary key: uid
  - fields: created_at(), data(), job_uid(), status(), uid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- launch_workflow:
  - endpoint: POST /workflows/{{ record.workflow_uid }}/launch
  - required fields: workflow_uid
  - risk: external mutation; launches a live Captain Data workflow run (a new job), consuming account credits and potentially performing external actions (scraping, enrichment, outreach) depending on the workflow's own configured steps; approval required

## Security

- read risk: external Captain Data API read of workspace, workflow, and job data
- write risk: external mutation; launches a live Captain Data workflow run, consuming account credits and potentially performing external side effects depending on the workflow's own configured steps; approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect captain-data
```

### Inspect as structured JSON

```bash
pm connectors inspect captain-data --json
```

## Agent Rules

- Run pm connectors inspect captain-data before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
