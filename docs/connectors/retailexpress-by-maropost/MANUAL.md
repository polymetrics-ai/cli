# pm connectors inspect retailexpress-by-maropost

```text
NAME
  pm connectors inspect retailexpress-by-maropost - Retail Express by Maropost connector manual

SYNOPSIS
  pm connectors inspect retailexpress-by-maropost
  pm connectors inspect retailexpress-by-maropost --json
  pm credentials add <name> --connector retailexpress-by-maropost [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Retail Express products, customers, orders, stock levels, and stores through the Maropost API. Read-only.

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
  created_after
  status
  store_id
  updated_after
  access_token (secret)
  api_key (secret)

ETL STREAMS
  products:
    primary key: id
    cursor: updated_at
    fields: id(), name(), sku(), status(), stream(), updated_at()
  customers:
    primary key: id
    cursor: updated_at
    fields: email(), first_name(), id(), last_name(), stream(), updated_at()
  orders:
    primary key: id
    cursor: updated_at
    fields: customer_id(), id(), order_number(), status(), stream(), total(), updated_at()
  stock_levels:
    primary key: id
    cursor: updated_at
    fields: id(), product_id(), quantity(), store_id(), stream(), updated_at()
  stores:
    primary key: id
    fields: code(), id(), name(), status(), stream()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Retail Express by Maropost API read of product, customer, order, and stock data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect retailexpress-by-maropost

  # Inspect as structured JSON
  pm connectors inspect retailexpress-by-maropost --json

AGENT WORKFLOW
  - Run pm connectors inspect retailexpress-by-maropost before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
