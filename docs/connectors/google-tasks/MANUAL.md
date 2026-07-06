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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  records_limit
  api_key (secret)

ETL STREAMS
  tasklists:
    primary key: id
    cursor: updated
    fields: etag(), id(), kind(), self_link(), title(), updated()
  tasks:
    primary key: id
    cursor: updated
    fields: completed(), deleted(), due(), etag(), hidden(), id(), kind(), notes(), parent(), position(), self_link(), status(), tasklist_id(), title(), updated()

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
