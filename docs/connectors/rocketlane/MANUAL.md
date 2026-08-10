# pm connectors inspect rocketlane

```text
NAME
  pm connectors inspect rocketlane - Rocketlane connector manual

SYNOPSIS
  pm connectors inspect rocketlane
  pm connectors inspect rocketlane --json
  pm credentials add <name> --connector rocketlane [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Rocketlane projects, tasks, customers, users, and time entries through the REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  created_after
  mode
  project_id
  status
  updated_after
  api_key (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), id(string), name(string), status(string), stream(string), updated_at(string)
  tasks:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), project_id(string), status(string), stream(string), updated_at(string)
  customers:
    primary key: id
    cursor: updated_at
    fields: domain(string), id(string), name(string), stream(string), updated_at(string)
  users:
    primary key: id
    cursor: updated_at
    fields: email(string), id(string), name(string), status(string), stream(string), updated_at(string)
  time_entries:
    primary key: id
    cursor: updated_at
    fields: id(string), minutes(integer), project_id(string), stream(string), task_id(string), updated_at(string), user_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Rocketlane API read of project, task, customer, and time-entry data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect rocketlane

  # Inspect as structured JSON
  pm connectors inspect rocketlane --json

AGENT WORKFLOW
  - Run pm connectors inspect rocketlane before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
