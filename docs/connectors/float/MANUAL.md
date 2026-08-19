# pm connectors inspect float

```text
NAME
  pm connectors inspect float - Float connector manual

SYNOPSIS
  pm connectors inspect float
  pm connectors inspect float --json
  pm credentials add <name> --connector float [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Float people, projects, clients, tasks, and departments through the Float v3 REST API.

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
  mode
  page_size
  access_token (secret) (required)

ETL STREAMS
  people:
    primary key: people_id
    fields: active(integer), created(string), department_id(integer), email(string), employee_type(integer), job_title(string), modified(string), name(string), people_id(integer), people_type_id(integer), role_id(integer), start_date(string)
  projects:
    primary key: project_id
    fields: active(integer), budget_total(number), budget_type(integer), client_id(integer), color(string), created(string), default_hourly_rate(number), modified(string), name(string), non_billable(integer), notes(string), project_id(integer), project_manager(integer), tags(array)
  clients:
    primary key: client_id
    fields: client_id(integer), created(string), modified(string), name(string)
  tasks:
    primary key: task_id
    fields: billable(integer), created(string), modified(string), name(string), project_id(integer), task_id(integer), task_meta_id(integer)
  departments:
    primary key: department_id
    fields: department_id(integer), name(string), parent_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Float API read of resource-planning and staffing data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect float

  # Inspect as structured JSON
  pm connectors inspect float --json

AGENT WORKFLOW
  - Run pm connectors inspect float before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
