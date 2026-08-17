# pm connectors inspect lightspeed-retail

```text
NAME
  pm connectors inspect lightspeed-retail - Lightspeed Retail connector manual

SYNOPSIS
  pm connectors inspect lightspeed-retail
  pm connectors inspect lightspeed-retail --json
  pm credentials add <name> --connector lightspeed-retail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Lightspeed Retail (X-Series) products, customers, sales, outlets, and registers through the Lightspeed REST API. Read-only.

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
  mode
  subdomain
  api_key (secret)

ETL STREAMS
  products:
    primary key: id
    cursor: version
    fields: brand_id(), created_at(), description(), handle(), has_variants(), id(), is_active(), is_composite(), name(), price_excluding_tax(), price_including_tax(), product_category(), sku(), supplier_id(), supply_price(), updated_at(), version()
  customers:
    primary key: id
    cursor: version
    fields: balance(), created_at(), customer_code(), customer_group_id(), do_not_email(), enable_loyalty(), id(), loyalty_balance(), updated_at(), version(), year_to_date()
  sales:
    primary key: id
    cursor: version
    fields: created_at(), customer_id(), id(), invoice_number(), register_id(), sale_date(), status(), total_price(), total_tax(), updated_at(), user_id(), version()
  outlets:
    primary key: id
    cursor: version
    fields: currency(), currency_symbol(), default_tax_id(), display_prices(), id(), name(), time_zone(), version()
  registers:
    primary key: id
    cursor: version
    fields: email_receipt(), id(), invoice_prefix(), invoice_sequence(), is_open(), name(), outlet_id(), print_receipt(), version()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Lightspeed Retail API read of product, customer, and sales data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect lightspeed-retail

  # Inspect as structured JSON
  pm connectors inspect lightspeed-retail --json

AGENT WORKFLOW
  - Run pm connectors inspect lightspeed-retail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
