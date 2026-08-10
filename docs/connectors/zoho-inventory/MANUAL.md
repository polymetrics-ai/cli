# pm connectors inspect zoho-inventory

```text
NAME
  pm connectors inspect zoho-inventory - Zoho Inventory connector manual

SYNOPSIS
  pm connectors inspect zoho-inventory
  pm connectors inspect zoho-inventory --json
  pm credentials add <name> --connector zoho-inventory [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Zoho Inventory contacts, items, and sales orders through the Zoho Inventory REST API.

ICON
  id: simple-icons-zoho-inventory
  asset: icons/simple-icons/zoho-inventory.svg
  title: Zoho
  simple_icon_slug: zoho
  simple_icon_hex: E42527
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Zoho
  match: curated-alias
  matched_by: zoho

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  organization_id
  page_size
  access_token (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: company_name(string), contact_id(string), contact_name(string), contact_type(string), created_time(string), currency_code(string), email(string), id(string), last_modified_time(string), outstanding_receivable_amount(number), phone(string), status(string), updated_at(string)
  items:
    primary key: id
    cursor: updated_at
    fields: created_time(string), description(string), id(string), item_id(string), item_name(string), last_modified_time(string), name(string), rate(number), sku(string), status(string), unit(string), updated_at(string)
  salesorders:
    primary key: id
    cursor: updated_at
    fields: balance(number), created_time(string), currency_code(string), customer_id(string), customer_name(string), date(string), id(string), last_modified_time(string), salesorder_id(string), salesorder_number(string), shipment_date(string), status(string), total(number), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Zoho Inventory API read of contact/item/sales-order data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-inventory

  # Inspect as structured JSON
  pm connectors inspect zoho-inventory --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-inventory before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
