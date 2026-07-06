# pm connectors inspect cart

```text
NAME
  pm connectors inspect cart - Cart.com connector manual

SYNOPSIS
  pm connectors inspect cart
  pm connectors inspect cart --json
  pm credentials add <name> --connector cart [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Cart.com orders, customers, products, and inventory through a read-only REST API.

ICON
  asset: icons/cart.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.cart.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  access_token (secret)

ETL STREAMS
  orders:
    primary key: id
    fields: id(), order_number(), updated_at()
  customers:
    primary key: id
    fields: id(), order_number(), updated_at()
  products:
    primary key: id
    fields: id(), order_number(), updated_at()
  inventory:
    primary key: id
    fields: id(), product_id(), quantity(), sku(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Cart.com API read of order, customer, product, and inventory data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect cart

  # Inspect as structured JSON
  pm connectors inspect cart --json

AGENT WORKFLOW
  - Run pm connectors inspect cart before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
