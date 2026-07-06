# pm connectors inspect imagga

```text
NAME
  pm connectors inspect imagga - Imagga connector manual

SYNOPSIS
  pm connectors inspect imagga
  pm connectors inspect imagga --json
  pm credentials add <name> --connector imagga [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Imagga account API usage and per-image tags/categories via the Imagga REST API. Read-only. The colors and faces_detections detection streams are not yet ported — see docs.md Known limits.

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
  image_urls
  api_key (secret)
  api_secret (secret)

ETL STREAMS
  usage:
    primary key: period
    fields: daily_processed(), monthly_limit(), monthly_processed(), period(), requests()
  tags:
    primary key: image_url, tag
    fields: confidence(), image_url(), tag()
  categories:
    primary key: image_url, category
    fields: category(), confidence(), image_url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Imagga API read of account usage data and per-image tags/categories
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect imagga

  # Inspect as structured JSON
  pm connectors inspect imagga --json

AGENT WORKFLOW
  - Run pm connectors inspect imagga before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
