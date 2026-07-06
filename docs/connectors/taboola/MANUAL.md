# pm connectors inspect taboola

```text
NAME
  pm connectors inspect taboola - Taboola connector manual

SYNOPSIS
  pm connectors inspect taboola
  pm connectors inspect taboola --json
  pm credentials add <name> --connector taboola [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Taboola campaigns through the Backstage API. Read-only.

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
  account_id
  base_url
  max_pages
  mode
  page_size
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  campaigns:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Taboola Backstage API read of campaign data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect taboola

  # Inspect as structured JSON
  pm connectors inspect taboola --json

AGENT WORKFLOW
  - Run pm connectors inspect taboola before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
