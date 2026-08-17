# pm connectors inspect planhat

```text
NAME
  pm connectors inspect planhat - Planhat connector manual

SYNOPSIS
  pm connectors inspect planhat
  pm connectors inspect planhat --json
  pm credentials add <name> --connector planhat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Planhat companies, end users, and licenses through the Planhat REST API.

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
  max_pages
  mode
  page_size
  api_token (secret)

ETL STREAMS
  companies:
    primary key: id
    cursor: updated_at
    fields: email(), id(), name(), phase(), updated_at()
  endusers:
    primary key: id
    cursor: updated_at
    fields: email(), id(), name(), phase(), updated_at()
  licenses:
    primary key: id
    cursor: updated_at
    fields: id(), name(), phase(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Planhat API read of customer success data
  approval: none; read-only customer success platform API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect planhat

  # Inspect as structured JSON
  pm connectors inspect planhat --json

AGENT WORKFLOW
  - Run pm connectors inspect planhat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
