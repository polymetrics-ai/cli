# pm connectors inspect ebay-fulfillment

```text
NAME
  pm connectors inspect ebay-fulfillment - eBay Fulfillment connector manual

SYNOPSIS
  pm connectors inspect ebay-fulfillment
  pm connectors inspect ebay-fulfillment --json
  pm credentials add <name> --connector ebay-fulfillment [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads eBay seller orders, exploded line items, shipping fulfillments, and payment disputes through the eBay Sell Fulfillment REST API.

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
  mode
  page_size
  refresh_token_endpoint
  scope
  start_date
  password (secret)
  refresh_token (secret)
  username (secret)

ETL STREAMS
  orders:
    primary key: order_id
    cursor: creation_date
    fields: buyer_username(), creation_date(), last_modified_date(), legacy_order_id(), line_item_count(), order_fulfillment_status(), order_id(), order_payment_status(), sales_record_reference(), seller_id(), total_currency(), total_value()
  order_line_items:
    primary key: line_item_id
    cursor: creation_date
    fields: creation_date(), legacy_item_id(), line_item_fulfillment_status(), line_item_id(), order_id(), quantity(), sku(), title(), total_currency(), total_value()
  shipping_fulfillments:
    primary key: order_id
    cursor: creation_date
    fields: creation_date(), legacy_order_id(), order_fulfillment_status(), order_id(), ship_to_city(), ship_to_country_code(), ship_to_name(), ship_to_postal_code(), ship_to_state_or_province(), shipping_step()
  payment_disputes:
    primary key: payment_dispute_id
    cursor: open_date
    fields: amount_currency(), amount_value(), buyer_username(), dispute_state(), dispute_status(), open_date(), order_id(), payment_dispute_id(), reason()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external eBay Sell Fulfillment API read of a seller's order, shipment, and dispute data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ebay-fulfillment

  # Inspect as structured JSON
  pm connectors inspect ebay-fulfillment --json

AGENT WORKFLOW
  - Run pm connectors inspect ebay-fulfillment before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
