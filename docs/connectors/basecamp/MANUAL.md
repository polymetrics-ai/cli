# pm connectors inspect basecamp

```text
NAME
  pm connectors inspect basecamp - Basecamp connector manual

SYNOPSIS
  pm connectors inspect basecamp
  pm connectors inspect basecamp --json
  pm credentials add <name> --connector basecamp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Basecamp 3 projects, people, and account activity events through the Basecamp REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  base_url
  mode
  start_date (required)
  client_id (secret) (required)
  client_refresh_token_2 (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    cursor: updated_at
    fields: app_url(string), bookmark_url(string), created_at(string), description(string), id(integer), name(string), purpose(string), status(string), updated_at(string), url(string)
  people:
    primary key: id
    cursor: updated_at
    fields: admin(boolean), client(boolean), created_at(string), email_address(string), id(integer), name(string), owner(boolean), personable_type(string), time_zone(string), title(string), updated_at(string)
  events:
    primary key: id
    cursor: created_at
    fields: action(string), created_at(string), id(integer), kind(string), recording_id(integer), summary(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Basecamp API reads performed by the legacy connector via a Tier-2 hook
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
