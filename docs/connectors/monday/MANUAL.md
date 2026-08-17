# pm connectors inspect monday

```text
NAME
  pm connectors inspect monday - Monday connector manual

SYNOPSIS
  pm connectors inspect monday
  pm connectors inspect monday --json
  pm credentials add <name> --connector monday [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads monday.com boards, items, users, teams, and tags through the monday.com GraphQL API. Read-only.

ICON
  asset: icons/monday.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.monday.com/api-reference/docs

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_version
  base_url
  max_pages
  page_size
  access_token (secret)
  api_token (secret)

ETL STREAMS
  boards:
    primary key: id
    cursor: updated_at
    fields: board_kind(), description(), id(), name(), state(), type(), updated_at(), workspace_id()
  items:
    primary key: id
    cursor: updated_at
    fields: board_id(), board_name(), created_at(), group_id(), group_title(), id(), name(), state(), updated_at()
  users:
    primary key: id
    fields: created_at(), email(), enabled(), id(), is_admin(), is_guest(), is_pending(), name()
  teams:
    primary key: id
    fields: id(), name(), picture_url()
  tags:
    primary key: id
    fields: color(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external monday.com GraphQL API read of boards/items/users/teams/tags
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect monday

  # Inspect as structured JSON
  pm connectors inspect monday --json

AGENT WORKFLOW
  - Run pm connectors inspect monday before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
