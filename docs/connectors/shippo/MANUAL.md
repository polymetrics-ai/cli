# pm connectors inspect shippo

```text
NAME
  pm connectors inspect shippo - Shippo connector manual

SYNOPSIS
  pm connectors inspect shippo
  pm connectors inspect shippo --json
  pm credentials add <name> --connector shippo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Shippo addresses, parcels, shipments, and transactions through the Shippo REST API.

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
  addresses:
    primary key: id
    cursor: updated_at
    fields: email(), id(), name(), updated_at()
  parcels:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), updated_at()
  shipments:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), updated_at()
  transactions:
    primary key: id
    cursor: updated_at
    fields: id(), name(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Shippo API read of address, parcel, shipment, and transaction data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shippo

  # Inspect as structured JSON
  pm connectors inspect shippo --json

AGENT WORKFLOW
  - Run pm connectors inspect shippo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
