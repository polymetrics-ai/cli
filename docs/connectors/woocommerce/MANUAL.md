# pm connectors inspect woocommerce

```text
NAME
  pm connectors inspect woocommerce - WooCommerce connector manual

SYNOPSIS
  pm connectors inspect woocommerce
  pm connectors inspect woocommerce --json
  pm credentials add <name> --connector woocommerce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads WooCommerce orders, products, customers, and coupons through the WooCommerce REST API (wc/v3).

ICON
  asset: icons/woocommerce.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://woocommerce.github.io/woocommerce-rest-api-docs/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  page_size
  start_date
  api_key (secret)
  api_secret (secret)

ETL STREAMS
  orders:
    primary key: id
    cursor: date_modified_gmt
    fields: currency(), customer_id(), date_created(), date_created_gmt(), date_modified(), date_modified_gmt(), date_paid(), id(), number(), payment_method(), status(), total(), total_tax()
  products:
    primary key: id
    cursor: date_modified_gmt
    fields: date_created_gmt(), date_modified_gmt(), id(), name(), price(), regular_price(), sale_price(), sku(), slug(), status(), stock_quantity(), stock_status(), total_sales(), type()
  customers:
    primary key: id
    cursor: date_modified_gmt
    fields: date_created(), date_created_gmt(), date_modified(), date_modified_gmt(), email(), first_name(), id(), is_paying_customer(), last_name(), role(), username()
  coupons:
    primary key: id
    cursor: date_modified_gmt
    fields: amount(), code(), date_created(), date_created_gmt(), date_expires(), date_modified(), date_modified_gmt(), discount_type(), id(), usage_count(), usage_limit()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external WooCommerce store read of orders, products, customers, and coupons
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect woocommerce

  # Inspect as structured JSON
  pm connectors inspect woocommerce --json

AGENT WORKFLOW
  - Run pm connectors inspect woocommerce before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
