# pm connectors inspect paperform

```text
NAME
  pm connectors inspect paperform - Paperform connector manual

SYNOPSIS
  pm connectors inspect paperform
  pm connectors inspect paperform --json
  pm credentials add <name> --connector paperform [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Paperform forms and form submissions through the Paperform REST API.

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
  form_id
  api_key (secret)

ETL STREAMS
  forms:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), slug(), title(), updated_at()
  submissions:
    primary key: id
    cursor: created_at
    fields: created_at(), data(), form_id(), id(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Paperform API read of form and submission data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect paperform

  # Inspect as structured JSON
  pm connectors inspect paperform --json

AGENT WORKFLOW
  - Run pm connectors inspect paperform before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
