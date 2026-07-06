# pm connectors inspect adjust

```text
NAME
  pm connectors inspect adjust - Adjust connector manual

SYNOPSIS
  pm connectors inspect adjust
  pm connectors inspect adjust --json
  pm credentials add <name> --connector adjust [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Adjust report-service report rows for configured dimensions and metrics. Read-only.

ICON
  asset: icons/adjust.svg
  source: official
  review_status: official_verified
  review_url: https://dev.adjust.com/en/api/rs-api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  additional_metrics
  base_url
  dimensions
  end_date
  metrics
  mode
  start_date
  api_token (secret)

ETL STREAMS
  reports:
    fields: app(), clicks(), cost(), country(), date(), installs()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Adjust report-service read of configured dimensions/metrics
  approval: none; read-only reporting API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect adjust

  # Inspect as structured JSON
  pm connectors inspect adjust --json

AGENT WORKFLOW
  - Run pm connectors inspect adjust before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
