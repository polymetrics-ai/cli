# pm connectors inspect kissmetrics

```text
NAME
  pm connectors inspect kissmetrics - Kissmetrics connector manual

SYNOPSIS
  pm connectors inspect kissmetrics
  pm connectors inspect kissmetrics --json
  pm credentials add <name> --connector kissmetrics [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Kissmetrics products, reports, events, and properties through the Kissmetrics query API using HTTP Basic authentication.

ICON
  asset: icons/kissmetrics.svg
  source: official
  review_status: official_verified
  review_url: https://support.kissmetrics.io/reference

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
  product_id
  username
  password (secret)

ETL STREAMS
  products:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  reports:
    primary key: id
    fields: created_at(), id(), name(), product_id(), type(), updated_at()
  events:
    primary key: id
    fields: created_at(), display_name(), id(), name(), product_id()
  properties:
    primary key: id
    fields: created_at(), display_name(), id(), name(), product_id(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Kissmetrics query API read of product analytics metadata
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect kissmetrics

  # Inspect as structured JSON
  pm connectors inspect kissmetrics --json

AGENT WORKFLOW
  - Run pm connectors inspect kissmetrics before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
