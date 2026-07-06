# pm connectors inspect veeqo

```text
NAME
  pm connectors inspect veeqo - Veeqo connector manual

SYNOPSIS
  pm connectors inspect veeqo
  pm connectors inspect veeqo --json
  pm credentials add <name> --connector veeqo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads orders, products, customers, warehouses, suppliers, purchase orders, sales channels, delivery methods, and tags from the Veeqo API, and writes orders, products, customers, suppliers, warehouses, delivery methods, tags, sales channels, product properties, payments, and shipments.

ICON
  asset: icons/veeqo.svg
  source: official
  review_status: official_verified
  review_url: https://developers.veeqo.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  api_key (secret)

ETL STREAMS
  orders:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), number(), status()
  products:
    primary key: id
    fields: created_at(), id(), notes(), title(), updated_at()
  customers:
    primary key: id
    fields: email(), id(), notes(), phone()
  warehouses:
    primary key: id
    fields: city(), country(), id(), name(), post_code()
  suppliers:
    primary key: id
    fields: created_at(), currency_code(), id(), name(), updated_at()
  purchase_orders:
    primary key: id
    fields: completed_at(), created_at(), destination_warehouse_id(), id()
  channels:
    primary key: id
    fields: id(), name(), state(), type_code()
  delivery_methods:
    primary key: id
    fields: cost(), created_at(), id(), name()
  tags:
    primary key: id
    fields: colour(), id(), name(), taggings_count()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_supplier:
    endpoint: POST /suppliers
    risk: external mutation; approval required
  update_supplier:
    endpoint: PUT /suppliers/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_supplier:
    endpoint: DELETE /suppliers/{{ record.id }}
    required fields: id
    risk: destructive external mutation; approval required
  create_warehouse:
    endpoint: POST /warehouses
    risk: external mutation; approval required
  update_warehouse:
    endpoint: PUT /warehouses/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  create_delivery_method:
    endpoint: POST /delivery_methods
    risk: external mutation; approval required
  update_delivery_method:
    endpoint: PUT /delivery_methods/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_delivery_method:
    endpoint: DELETE /delivery_methods/{{ record.id }}
    required fields: id
    risk: destructive external mutation; approval required
  create_tag:
    endpoint: POST /tags
    risk: external mutation; approval required
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: destructive external mutation; approval required
  create_channel:
    endpoint: POST /channels
    risk: external mutation; approval required
  update_channel:
    endpoint: PUT /channels/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_channel:
    endpoint: DELETE /channels/{{ record.id }}
    required fields: id
    risk: destructive external mutation; approval required
  create_product_property:
    endpoint: POST /product_properties
    risk: external mutation; approval required
  create_customer:
    endpoint: POST /customers
    risk: external mutation; approval required
  update_customer:
    endpoint: PUT /customers/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  create_product:
    endpoint: POST /products
    risk: external mutation; approval required
  update_product:
    endpoint: PUT /products/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_product:
    endpoint: DELETE /products/{{ record.id }}
    required fields: id
    risk: destructive external mutation; approval required
  create_order:
    endpoint: POST /orders
    risk: external mutation; approval required
  update_order:
    endpoint: PUT /orders/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  cancel_order:
    endpoint: PUT /orders/{{ record.id }}/cancel
    required fields: id
    risk: external mutation (cancels an order); approval required
  create_payment:
    endpoint: POST /payments
    risk: external mutation; approval required
  create_shipment:
    endpoint: POST /shipments
    risk: external mutation; approval required

SECURITY
  read risk: external Veeqo API read of order, product, customer, and inventory data
  write risk: external mutation of Veeqo orders, products, customers, suppliers, warehouses, delivery methods, tags, sales channels, product properties, payments, and shipments; approval required
  approval: read: none; write: required for every action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect veeqo

  # Inspect as structured JSON
  pm connectors inspect veeqo --json

AGENT WORKFLOW
  - Run pm connectors inspect veeqo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
