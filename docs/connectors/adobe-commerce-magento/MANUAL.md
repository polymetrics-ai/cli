# pm connectors inspect adobe-commerce-magento

```text
NAME
  pm connectors inspect adobe-commerce-magento - Adobe Commerce (Magento) connector manual

SYNOPSIS
  pm connectors inspect adobe-commerce-magento
  pm connectors inspect adobe-commerce-magento --json
  pm credentials add <name> --connector adobe-commerce-magento [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Adobe Commerce (Magento) products, orders, customers, categories, invoices, shipments, credit memos, customer groups, and store configuration through the Magento REST API, and writes product/category updates plus order cancellation.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  start_date
  api_key (secret)

ETL STREAMS
  products:
    primary key: id
    cursor: updated_at
    fields: attribute_set_id(), created_at(), id(), name(), price(), sku(), status(), type_id(), updated_at(), visibility(), weight()
  orders:
    primary key: entity_id
    cursor: updated_at
    fields: base_grand_total(), created_at(), customer_email(), customer_id(), entity_id(), grand_total(), increment_id(), order_currency_code(), state(), status(), updated_at()
  customers:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), firstname(), group_id(), id(), lastname(), store_id(), updated_at(), website_id()
  categories:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), is_active(), level(), name(), parent_id(), position(), product_count(), updated_at()
  invoices:
    primary key: entity_id
    cursor: updated_at
    fields: base_grand_total(), created_at(), entity_id(), grand_total(), increment_id(), order_id(), state(), store_id(), updated_at()
  shipments:
    primary key: entity_id
    cursor: created_at
    fields: created_at(), entity_id(), increment_id(), order_id(), shipment_status(), store_id(), total_qty(), updated_at()
  creditmemos:
    primary key: entity_id
    cursor: created_at
    fields: base_grand_total(), created_at(), entity_id(), grand_total(), increment_id(), order_id(), state(), store_id(), updated_at()
  customer_groups:
    primary key: id
    fields: code(), id(), tax_class_id(), tax_class_name()
  store_websites:
    primary key: id
    fields: code(), default_group_id(), id(), is_default(), name()
  store_views:
    primary key: id
    fields: code(), group_id(), id(), is_active(), name(), website_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  update_product:
    endpoint: PUT /products/{{ record.sku }}
    required fields: sku
    risk: external mutation; overwrites live Magento catalog product fields; approval required
  create_category:
    endpoint: POST /categories
    risk: external mutation; creates a live Magento catalog category; approval required
  update_category:
    endpoint: PUT /categories/{{ record.id }}
    required fields: id
    risk: external mutation; overwrites live Magento catalog category fields; approval required
  cancel_order:
    endpoint: POST /orders/{{ record.entity_id }}/cancel
    required fields: entity_id
    risk: external mutation; irreversibly cancels a live Magento sales order; approval required

SECURITY
  read risk: external Adobe Commerce (Magento) REST API read of catalog, order, and store-configuration data
  write risk: external mutation of live Magento catalog products/categories and cancellation of live sales orders; approval required for every write action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect adobe-commerce-magento

  # Inspect as structured JSON
  pm connectors inspect adobe-commerce-magento --json

AGENT WORKFLOW
  - Run pm connectors inspect adobe-commerce-magento before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
