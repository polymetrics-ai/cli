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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  mode
  start_date
  api_key (secret) (required)

ETL STREAMS
  products:
    primary key: id
    cursor: updated_at
    fields: attribute_set_id(integer), created_at(string), id(integer), name(string), price(number), sku(string), status(integer), type_id(string), updated_at(string), visibility(integer), weight(number)
  orders:
    primary key: entity_id
    cursor: updated_at
    fields: base_grand_total(number), created_at(string), customer_email(string), customer_id(integer), entity_id(integer), grand_total(number), increment_id(string), order_currency_code(string), state(string), status(string), updated_at(string)
  customers:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), firstname(string), group_id(integer), id(integer), lastname(string), store_id(integer), updated_at(string), website_id(integer)
  categories:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(integer), is_active(boolean), level(integer), name(string), parent_id(integer), position(integer), product_count(integer), updated_at(string)
  invoices:
    primary key: entity_id
    cursor: updated_at
    fields: base_grand_total(number), created_at(string), entity_id(integer), grand_total(number), increment_id(string), order_id(integer), state(integer), store_id(integer), updated_at(string)
  shipments:
    primary key: entity_id
    cursor: created_at
    fields: created_at(string), entity_id(integer), increment_id(string), order_id(integer), shipment_status(integer), store_id(integer), total_qty(number), updated_at(string)
  creditmemos:
    primary key: entity_id
    cursor: created_at
    fields: base_grand_total(number), created_at(string), entity_id(integer), grand_total(number), increment_id(string), order_id(integer), state(integer), store_id(integer), updated_at(string)
  customer_groups:
    primary key: id
    fields: code(string), id(integer), tax_class_id(integer), tax_class_name(string)
  store_websites:
    primary key: id
    fields: code(string), default_group_id(integer), id(integer), is_default(boolean), name(string)
  store_views:
    primary key: id
    fields: code(string), group_id(integer), id(integer), is_active(boolean), name(string), website_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  update_product:
    endpoint: PUT /products/{{ record.sku }}
    required fields: sku
    risk: external mutation; overwrites live Magento catalog product fields; approval required
  create_category:
    endpoint: POST /categories
    required fields: name, parent_id
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
