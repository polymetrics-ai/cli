# pm connectors inspect basecamp

```text
NAME
  pm connectors inspect basecamp - Basecamp connector manual

SYNOPSIS
  pm connectors inspect basecamp
  pm connectors inspect basecamp --json
  pm credentials add <name> --connector basecamp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Basecamp 3 projects, people, and account events through fixed account-bound REST routes.

ICON
  id: simple-icons-basecamp
  asset: icons/simple-icons/basecamp.svg
  title: Basecamp
  simple_icon_slug: basecamp
  simple_icon_hex: 1D2D35
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Basecamp
  match: exact-name-or-slug
  matched_by: basecamp

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id (required)
  access_token (secret)
  client_id (secret)
  client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  projects:
    primary key: id
    cursor: updated_at
    fields: id(integer), updated_at(string)
  people:
    primary key: id
    cursor: updated_at
    fields: id(integer), updated_at(string)
  events:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded Basecamp account reads use a declared account path and bearer or refresh-token authentication.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect basecamp

  # Inspect as structured JSON
  pm connectors inspect basecamp --json

AGENT WORKFLOW
  - Run pm connectors inspect basecamp before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
