# pm connectors inspect gridly

```text
NAME
  pm connectors inspect gridly - Gridly connector manual

SYNOPSIS
  pm connectors inspect gridly
  pm connectors inspect gridly --json
  pm credentials add <name> --connector gridly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Gridly views, per-view records (with flattened column cells), and per-view branches through the Gridly REST API.

ICON
  asset: icons/gridly.svg
  source: official
  review_status: official_verified
  review_url: https://www.gridly.com/docs/api/

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
  view_ids
  api_key (secret)

ETL STREAMS
  views:
    primary key: id
    fields: id(), name()
  records:
    primary key: view_id, id
    fields: cells(), id(), path(), view_id()
  branches:
    primary key: view_id, id
    fields: id(), name(), view_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Gridly API read of view/grid content
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gridly

  # Inspect as structured JSON
  pm connectors inspect gridly --json

AGENT WORKFLOW
  - Run pm connectors inspect gridly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
