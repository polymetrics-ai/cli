# pm connectors inspect poplar

```text
NAME
  pm connectors inspect poplar - Poplar connector manual

SYNOPSIS
  pm connectors inspect poplar
  pm connectors inspect poplar --json
  pm credentials add <name> --connector poplar [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Poplar campaigns and orders through read-only REST list endpoints.

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
  api_token (secret)

ETL STREAMS
  campaigns:
    primary key: id
    fields: created_at(), id(), name(), status()
  orders:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Poplar API read of campaign and order data
  approval: none; read-only, no writes implemented by legacy
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect poplar

  # Inspect as structured JSON
  pm connectors inspect poplar --json

AGENT WORKFLOW
  - Run pm connectors inspect poplar before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
