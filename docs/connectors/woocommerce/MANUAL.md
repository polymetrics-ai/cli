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
  id: woocommerce
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
  base_url (required)
  max_pages
  page_size
  start_date
  api_key (secret) (required)
  api_secret (secret) (required)

ETL STREAMS
  orders:
    primary key: id
    cursor: date_modified_gmt
    fields: currency(string), customer_id(integer), date_created(string), date_created_gmt(string), date_modified(string), date_modified_gmt(string), date_paid(string), id(integer), number(string), payment_method(string), status(string), total(string), total_tax(string)
  products:
    primary key: id
    cursor: date_modified_gmt
    fields: date_created_gmt(string), date_modified_gmt(string), id(integer), name(string), price(string), regular_price(string), sale_price(string), sku(string), slug(string), status(string), stock_quantity(integer), stock_status(string), total_sales(integer), type(string)
  customers:
    primary key: id
    cursor: date_modified_gmt
    fields: date_created(string), date_created_gmt(string), date_modified(string), date_modified_gmt(string), email(string), first_name(string), id(integer), is_paying_customer(boolean), last_name(string), role(string), username(string)
  coupons:
    primary key: id
    cursor: date_modified_gmt
    fields: amount(string), code(string), date_created(string), date_created_gmt(string), date_expires(string), date_modified(string), date_modified_gmt(string), discount_type(string), id(integer), usage_count(integer), usage_limit(integer)

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
