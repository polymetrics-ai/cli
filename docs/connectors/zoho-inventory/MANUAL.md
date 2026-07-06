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
  max_pages
  mode
  organization_id
  page_size
  access_token (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: company_name(), contact_id(), contact_name(), contact_type(), created_time(), currency_code(), email(), id(), last_modified_time(), outstanding_receivable_amount(), phone(), status(), updated_at()
  items:
    primary key: id
    cursor: updated_at
    fields: created_time(), description(), id(), item_id(), item_name(), last_modified_time(), name(), rate(), sku(), status(), unit(), updated_at()
  salesorders:
    primary key: id
    cursor: updated_at
    fields: balance(), created_time(), currency_code(), customer_id(), customer_name(), date(), id(), last_modified_time(), salesorder_id(), salesorder_number(), shipment_date(), status(), total(), updated_at()

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
