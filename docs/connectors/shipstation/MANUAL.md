# pm connectors inspect shipstation

```text
NAME
  pm connectors inspect shipstation - ShipStation connector manual

SYNOPSIS
  pm connectors inspect shipstation
  pm connectors inspect shipstation --json
  pm credentials add <name> --connector shipstation [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ShipStation orders, shipments, products, and customers through the ShipStation REST API.

ICON
  id: shipstation
  asset: icons/shipstation.svg
  source: official
  review_status: official_verified
  review_url: https://www.shipstation.com/docs/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)
  api_secret (secret) (required)

ETL STREAMS
  orders:
    primary key: id
    cursor: modified_at
    fields: id(integer), modified_at(string), order_number(string), status(string)
  shipments:
    primary key: id
    cursor: modified_at
    fields: id(integer), modified_at(string), order_number(string), status(string)
  products:
    primary key: id
    cursor: modified_at
    fields: id(integer), modified_at(string), name(string), sku(string)
  customers:
    primary key: id
    cursor: modified_at
    fields: email(string), id(integer), modified_at(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external ShipStation API read of order, shipment, product, and customer data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shipstation

  # Inspect as structured JSON
  pm connectors inspect shipstation --json

AGENT WORKFLOW
  - Run pm connectors inspect shipstation before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
