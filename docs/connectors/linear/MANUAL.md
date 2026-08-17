# pm connectors inspect linear

```text
NAME
  pm connectors inspect linear - Linear connector manual

SYNOPSIS
  pm connectors inspect linear
  pm connectors inspect linear --json
  pm credentials add <name> --connector linear [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Linear issues, teams, projects, and users through the Linear GraphQL API. Read-only.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  auth_type
  base_url
  max_pages
  page_size
  access_token (secret)
  api_key (secret)

ETL STREAMS
  issues:
    primary key: id
    cursor: updated_at
    fields: assignee_email(), assignee_id(), branch_name(), canceled_at(), completed_at(), createdAt(), created_at(), creator_id(), description(), estimate(), id(), identifier(), priority(), state_id(), state_name(), state_type(), team_id(), team_key(), title(), updatedAt(), updated_at(), url()
  teams:
    primary key: id
    cursor: updated_at
    fields: createdAt(), created_at(), description(), id(), key(), name(), private(), updatedAt(), updated_at()
  projects:
    primary key: id
    cursor: updated_at
    fields: canceled_at(), completed_at(), createdAt(), created_at(), description(), id(), name(), progress(), started_at(), state(), updatedAt(), updated_at(), url()
  users:
    primary key: id
    cursor: updated_at
    fields: active(), admin(), createdAt(), created_at(), display_name(), email(), id(), name(), updatedAt(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Linear GraphQL API read of issues/teams/projects/users
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect linear

  # Inspect as structured JSON
  pm connectors inspect linear --json

AGENT WORKFLOW
  - Run pm connectors inspect linear before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
