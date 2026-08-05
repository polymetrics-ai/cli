# pm connectors inspect katana

```text
NAME
  pm connectors inspect katana - Katana connector manual

SYNOPSIS
  pm connectors inspect katana
  pm connectors inspect katana --json
  pm credentials add <name> --connector katana [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Katana MRP (Cloud Inventory) products, materials, variants, sales orders, and customers through the Katana REST API.

ICON
  id: simple-icons-katana
  asset: icons/simple-icons/katana.svg
  title: Katana
  simple_icon_slug: katana
  simple_icon_hex: 000000
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Katana
  match: exact-name-or-slug
  matched_by: katana

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  products:
    primary key: id
    cursor: updated_at
    fields: additional_info(), category_name(), created_at(), default_supplier_id(), id(), is_producible(), is_purchasable(), is_sellable(), name(), uom(), updated_at()
  materials:
    primary key: id
    cursor: updated_at
    fields: additional_info(), category_name(), created_at(), default_supplier_id(), id(), is_sellable(), name(), uom(), updated_at()
  variants:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), material_id(), product_id(), purchase_price(), sales_price(), sku(), type(), updated_at()
  sales_orders:
    primary key: id
    cursor: updated_at
    fields: created_at(), currency(), customer_id(), delivery_date(), id(), order_created_date(), order_no(), status(), total(), total_in_base_currency(), updated_at()
  customers:
    primary key: id
    cursor: updated_at
    fields: category(), created_at(), currency(), email(), id(), name(), phone(), reference_id(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Katana MRP API read of inventory, sales, and customer data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

MECHANISM
  kind: official_api
  sanctioned_by_provider: true (official)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect katana

  # Inspect as structured JSON
  pm connectors inspect katana --json

AGENT WORKFLOW
  - Run pm connectors inspect katana before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
