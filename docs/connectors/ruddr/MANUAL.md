# pm connectors inspect ruddr

```text
NAME
  pm connectors inspect ruddr - Ruddr connector manual

SYNOPSIS
  pm connectors inspect ruddr
  pm connectors inspect ruddr --json
  pm credentials add <name> --connector ruddr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Ruddr clients, projects, and time entries through the Ruddr API. Read-only.

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
  workspace_id
  api_key (secret)

ETL STREAMS
  clients:
    primary key: id
    fields: id(), name(), stream()
  projects:
    primary key: id
    fields: id(), name(), project_id(), stream()
  time_entries:
    primary key: id
    fields: hours(), id(), name(), project_id(), stream()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Ruddr API read of client, project, and time-entry data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ruddr

  # Inspect as structured JSON
  pm connectors inspect ruddr --json

AGENT WORKFLOW
  - Run pm connectors inspect ruddr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
