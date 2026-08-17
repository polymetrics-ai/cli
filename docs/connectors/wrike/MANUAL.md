# pm connectors inspect wrike

```text
NAME
  pm connectors inspect wrike - Wrike connector manual

SYNOPSIS
  pm connectors inspect wrike
  pm connectors inspect wrike --json
  pm credentials add <name> --connector wrike [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Wrike tasks, folders, and contacts through the Wrike REST API. Read-only.

ICON
  asset: icons/wrike.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.wrike.com/api/v4/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  access_token (secret)

ETL STREAMS
  tasks:
    primary key: id
    cursor: updatedDate
    fields: id(), title(), updatedDate()
  folders:
    primary key: id
    cursor: updatedDate
    fields: id(), title(), updatedDate()
  contacts:
    primary key: id
    fields: firstName(), id(), lastName()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Wrike API read of task, folder, and contact data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect wrike

  # Inspect as structured JSON
  pm connectors inspect wrike --json

AGENT WORKFLOW
  - Run pm connectors inspect wrike before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
