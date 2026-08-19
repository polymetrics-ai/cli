# pm connectors inspect railz

```text
NAME
  pm connectors inspect railz - Railz connector manual

SYNOPSIS
  pm connectors inspect railz
  pm connectors inspect railz --json
  pm credentials add <name> --connector railz [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Railz businesses, connections, customers, invoices, and bills through the Railz REST API. Read-only.

ICON
  id: railz
  asset: icons/railz.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_token (secret)
  api_key (secret)

ETL STREAMS
  businesses:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), name(string), status(string)
  connections:
    primary key: id
    cursor: created_at
    fields: business_id(string), created_at(string), id(string), status(string)
  customers:
    primary key: id
    fields: business_id(string), email(string), id(string), name(string)
  invoices:
    primary key: id
    fields: business_id(string), customer_id(string), id(string), status(string), total_amount(number), vendor_id(string)
  bills:
    primary key: id
    fields: business_id(string), id(string), status(string), total_amount(number), vendor_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Railz API read of connected-business accounting data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect railz

  # Inspect as structured JSON
  pm connectors inspect railz --json

AGENT WORKFLOW
  - Run pm connectors inspect railz before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
