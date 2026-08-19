# pm connectors inspect captain-data

```text
NAME
  pm connectors inspect captain-data - Captain Data connector manual

SYNOPSIS
  pm connectors inspect captain-data
  pm connectors inspect captain-data --json
  pm credentials add <name> --connector captain-data [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Captain Data workspace, workflows, jobs, and job results, and writes a launch_workflow action to trigger a new workflow run, through the Captain Data v3 REST API.

ICON
  id: captain-data
  asset: icons/captain-data.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.captaindata.co/api-documentation

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  job_uid
  mode
  project_uid (required)
  workflow_uid
  api_key (secret) (required)

ETL STREAMS
  workspace:
    primary key: uid
    fields: created_at(string), name(string), plan(string), uid(string)
  workflows:
    primary key: uid
    fields: created_at(string), name(string), status(string), uid(string), updated_at(string)
  jobs:
    primary key: uid
    fields: created_at(string), ended_at(string), status(string), uid(string), workflow_uid(string)
  job_results:
    primary key: uid
    fields: created_at(string), data(object), job_uid(string), status(string), uid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  launch_workflow:
    endpoint: POST /workflows/{{ record.workflow_uid }}/launch
    required fields: workflow_uid
    risk: external mutation; launches a live Captain Data workflow run (a new job), consuming account credits and potentially performing external actions (scraping, enrichment, outreach) depending on the workflow's own configured steps; approval required

SECURITY
  read risk: external Captain Data API read of workspace, workflow, and job data
  write risk: external mutation; launches a live Captain Data workflow run, consuming account credits and potentially performing external side effects depending on the workflow's own configured steps; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect captain-data

  # Inspect as structured JSON
  pm connectors inspect captain-data --json

AGENT WORKFLOW
  - Run pm connectors inspect captain-data before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
