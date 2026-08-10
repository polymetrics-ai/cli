# pm connectors inspect insightful

```text
NAME
  pm connectors inspect insightful - Insightful connector manual

SYNOPSIS
  pm connectors inspect insightful
  pm connectors inspect insightful --json
  pm credentials add <name> --connector insightful [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Insightful workforce-analytics employees, teams, projects, and directory entries through the Insightful REST API.

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
  api_token (secret) (required)

ETL STREAMS
  employee:
    primary key: id
    cursor: updatedAt
    fields: createdAt(integer), email(string), id(string), modelName(string), name(string), updatedAt(integer)
  team:
    primary key: id
    fields: default(boolean), description(string), employees(array), id(string), modelName(string), name(string), projects(array)
  projects:
    primary key: id
    cursor: updatedAt
    fields: archived(boolean), billable(boolean), createdAt(integer), creatorId(string), employees(array), id(string), modelName(string), name(string), organizationId(string), updatedAt(integer)
  directory:
    primary key: id
    cursor: updatedAt
    fields: createdAt(integer), id(string), modelName(string), name(string), organizationId(string), updatedAt(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Insightful API read of workforce-analytics employees, teams, projects, and directory entries
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect insightful

  # Inspect as structured JSON
  pm connectors inspect insightful --json

AGENT WORKFLOW
  - Run pm connectors inspect insightful before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
