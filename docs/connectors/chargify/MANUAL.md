# pm connectors inspect chargify

```text
NAME
  pm connectors inspect chargify - Chargify connector manual

SYNOPSIS
  pm connectors inspect chargify
  pm connectors inspect chargify --json
  pm credentials add <name> --connector chargify [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Chargify (Maxio Advanced Billing) customers, subscriptions, products, product families, coupons, transactions, invoices, payment profiles, events, and statements through the Chargify REST API.

ICON
  id: chargify
  asset: icons/chargify.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.chargify.com/docs/api-docs/YXBpOjE0MTA4MjYx-chargify-api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  domain
  subdomain
  username
  api_key (secret)
  password (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: country(string), created_at(string), email(string), first_name(string), id(integer), last_name(string), organization(string), phone(string), reference(string), updated_at(string)
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: balance_in_cents(integer), created_at(string), current_period_ends_at(string), current_period_started_at(string), customer_id(integer), id(integer), product_id(integer), state(string), total_revenue_in_cents(integer), updated_at(string)
  products:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), handle(string), id(integer), interval(integer), interval_unit(string), name(string), price_in_cents(integer), product_family_id(integer), updated_at(string)
  coupons:
    primary key: id
    cursor: updated_at
    fields: amount_in_cents(integer), code(string), created_at(string), description(string), id(integer), name(string), percentage(string), product_family_id(integer), updated_at(string)
  transactions:
    primary key: id
    cursor: created_at
    fields: amount_in_cents(integer), created_at(string), customer_id(integer), id(integer), kind(string), product_id(integer), subscription_id(integer), success(boolean), transaction_type(string)
  product_families:
    primary key: id
    cursor: updated_at
    fields: accounting_code(string), created_at(string), description(string), handle(string), id(integer), name(string), updated_at(string)
  invoices:
    primary key: id
    cursor: updated_at
    fields: created_at(string), currency(string), customer_id(integer), due_amount(string), due_date(string), id(string), issue_date(string), number(string), paid_amount(string), state(string), subscription_id(integer), total_amount(string), updated_at(string)
  payment_profiles:
    primary key: id
    fields: card_type(string), created_at(string), current_vault(string), customer_id(integer), expiration_month(integer), expiration_year(integer), id(integer), last_four(string), payment_type(string), updated_at(string)
  events:
    primary key: id
    cursor: created_at
    fields: created_at(string), customer_id(integer), id(integer), key(string), message(string), subscription_id(integer)
  statements:
    primary key: id
    fields: closing_balance_in_cents(integer), created_at(string), customer_id(integer), id(integer), settlement_date(string), subscription_id(integer), uid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers.json
    required fields: customer
    risk: external mutation; approval required
  update_customer:
    endpoint: PUT /customers/{{ record.id }}.json
    required fields: id, customer
    risk: external mutation; approval required
  create_subscription:
    endpoint: POST /subscriptions.json
    required fields: subscription
    risk: external mutation with billing side effects; approval required
  update_subscription:
    endpoint: PUT /subscriptions/{{ record.id }}.json
    required fields: id, subscription
    risk: external mutation with billing side effects; approval required
  cancel_subscription:
    endpoint: POST /subscriptions/{{ record.id }}/cancel.json
    required fields: id
    risk: external mutation with billing side effects; approval required
  create_product_family:
    endpoint: POST /product_families.json
    required fields: product_family
    risk: external mutation; approval required
  create_product:
    endpoint: POST /product_families/{{ record.product_family_id }}/products.json
    required fields: product_family_id, product
    risk: external mutation; approval required
  update_product:
    endpoint: PUT /products/{{ record.id }}.json
    required fields: id, product
    risk: external mutation; approval required
  create_coupon:
    endpoint: POST /product_families/{{ record.product_family_id }}/coupons.json
    required fields: product_family_id, coupon
    risk: external mutation; approval required
  update_coupon:
    endpoint: PUT /coupons/{{ record.id }}.json
    required fields: id, coupon
    risk: external mutation; approval required

SECURITY
  read risk: external Chargify API read of customer and billing data
  write risk: external mutation of Chargify billing data (customers, subscriptions, product catalog, coupons); subscription create/update/cancel actions have direct billing side effects and require approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chargify

  # Inspect as structured JSON
  pm connectors inspect chargify --json

AGENT WORKFLOW
  - Run pm connectors inspect chargify before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
