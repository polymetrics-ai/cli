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
  id: simple-icons-linear
  asset: icons/simple-icons/linear.svg
  title: Linear
  simple_icon_slug: linear
  simple_icon_hex: 5E6AD2
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Linear
  match: exact-name-or-slug
  matched_by: linear

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
    fields: assignee_email(string), assignee_id(string), branch_name(string), canceled_at(string), completed_at(string), createdAt(string), created_at(string), creator_id(string), description(string), estimate(number), id(string), identifier(string), priority(integer), state_id(string), state_name(string), state_type(string), team_id(string), team_key(string), title(string), updatedAt(string), updated_at(string), url(string)
  teams:
    primary key: id
    cursor: updated_at
    fields: createdAt(string), created_at(string), description(string), id(string), key(string), name(string), private(boolean), updatedAt(string), updated_at(string)
  projects:
    primary key: id
    cursor: updated_at
    fields: canceled_at(string), completed_at(string), createdAt(string), created_at(string), description(string), id(string), name(string), progress(number), started_at(string), state(string), updatedAt(string), updated_at(string), url(string)
  users:
    primary key: id
    cursor: updated_at
    fields: active(boolean), admin(boolean), createdAt(string), created_at(string), display_name(string), email(string), id(string), name(string), updatedAt(string), updated_at(string)

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
