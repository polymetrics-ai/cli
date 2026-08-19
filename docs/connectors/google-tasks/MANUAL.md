# pm connectors inspect google-tasks

```text
NAME
  pm connectors inspect google-tasks - Google Tasks connector manual

SYNOPSIS
  pm connectors inspect google-tasks
  pm connectors inspect google-tasks --json
  pm credentials add <name> --connector google-tasks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google task lists and tasks through the Google Tasks REST API.

ICON
  id: simple-icons-googletasks
  asset: icons/simple-icons/googletasks.svg
  title: Google Tasks
  simple_icon_slug: googletasks
  simple_icon_hex: 2684FC
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Google%20Tasks
  match: exact-name-or-slug
  matched_by: google-tasks

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  records_limit
  api_key (secret) (required)

ETL STREAMS
  tasklists:
    primary key: id
    cursor: updated
    fields: etag(string), id(string), kind(string), self_link(string), title(string), updated(string)
  tasks:
    primary key: id
    cursor: updated
    fields: completed(string), deleted(boolean), due(string), etag(string), hidden(boolean), id(string), kind(string), notes(string), parent(string), position(string), self_link(string), status(string), tasklist_id(string), title(string), updated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Google Tasks API read of the authenticated user's task lists and tasks
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-tasks

  # Inspect as structured JSON
  pm connectors inspect google-tasks --json

AGENT WORKFLOW
  - Run pm connectors inspect google-tasks before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
