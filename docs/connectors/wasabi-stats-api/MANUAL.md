# pm connectors inspect wasabi-stats-api

```text
NAME
  pm connectors inspect wasabi-stats-api - Wasabi Stats API connector manual

SYNOPSIS
  pm connectors inspect wasabi-stats-api
  pm connectors inspect wasabi-stats-api --json
  pm credentials add <name> --connector wasabi-stats-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Wasabi account and bucket storage statistics from the Wasabi Stats API.

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
  start_date
  api_key (secret)

ETL STREAMS
  bucket_stats:
    primary key: id
    cursor: date
    fields: bucket(), date(), id(), storage_bytes()
  account_stats:
    primary key: id
    cursor: date
    fields: date(), id(), object_count(), storage_bytes()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Wasabi Stats API read of account/bucket storage usage metrics
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect wasabi-stats-api

  # Inspect as structured JSON
  pm connectors inspect wasabi-stats-api --json

AGENT WORKFLOW
  - Run pm connectors inspect wasabi-stats-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
