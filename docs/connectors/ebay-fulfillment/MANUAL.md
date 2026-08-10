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
  id: simple-icons-ebay-fulfillment
  asset: icons/simple-icons/ebay-fulfillment.svg
  title: eBay
  simple_icon_slug: ebay
  simple_icon_hex: E53238
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=eBay
  match: curated-alias
  matched_by: ebay

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
  refresh_token (secret) (required)
  username (secret)

ETL STREAMS
  orders:
    primary key: order_id
    cursor: creation_date
    fields: buyer_username(string), creation_date(string), last_modified_date(string), legacy_order_id(string), line_item_count(integer), order_fulfillment_status(string), order_id(string), order_payment_status(string), sales_record_reference(string), seller_id(string), total_currency(string), total_value(string)
  order_line_items:
    primary key: line_item_id
    cursor: creation_date
    fields: creation_date(string), legacy_item_id(string), line_item_fulfillment_status(string), line_item_id(string), order_id(string), quantity(integer), sku(string), title(string), total_currency(string), total_value(string)
  shipping_fulfillments:
    primary key: order_id
    cursor: creation_date
    fields: creation_date(string), legacy_order_id(string), order_fulfillment_status(string), order_id(string), ship_to_city(string), ship_to_country_code(string), ship_to_name(string), ship_to_postal_code(string), ship_to_state_or_province(string), shipping_step(string)
  payment_disputes:
    primary key: payment_dispute_id
    cursor: open_date
    fields: amount_currency(string), amount_value(string), buyer_username(string), dispute_state(string), dispute_status(string), open_date(string), order_id(string), payment_dispute_id(string), reason(string)

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
