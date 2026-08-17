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
  base_url
  domain
  subdomain
  username
  api_key (secret)
  password (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: country(), created_at(), email(), first_name(), id(), last_name(), organization(), phone(), reference(), updated_at()
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: balance_in_cents(), created_at(), current_period_ends_at(), current_period_started_at(), customer_id(), id(), product_id(), state(), total_revenue_in_cents(), updated_at()
  products:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), handle(), id(), interval(), interval_unit(), name(), price_in_cents(), product_family_id(), updated_at()
  coupons:
    primary key: id
    cursor: updated_at
    fields: amount_in_cents(), code(), created_at(), description(), id(), name(), percentage(), product_family_id(), updated_at()
  transactions:
    primary key: id
    cursor: created_at
    fields: amount_in_cents(), created_at(), customer_id(), id(), kind(), product_id(), subscription_id(), success(), transaction_type()
  product_families:
    primary key: id
    cursor: updated_at
    fields: accounting_code(), created_at(), description(), handle(), id(), name(), updated_at()
  invoices:
    primary key: id
    cursor: updated_at
    fields: created_at(), currency(), customer_id(), due_amount(), due_date(), id(), issue_date(), number(), paid_amount(), state(), subscription_id(), total_amount(), updated_at()
  payment_profiles:
    primary key: id
    fields: card_type(), created_at(), current_vault(), customer_id(), expiration_month(), expiration_year(), id(), last_four(), payment_type(), updated_at()
  events:
    primary key: id
    cursor: created_at
    fields: created_at(), customer_id(), id(), key(), message(), subscription_id()
  statements:
    primary key: id
    fields: closing_balance_in_cents(), created_at(), customer_id(), id(), settlement_date(), subscription_id(), uid()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers.json
    risk: external mutation; approval required
  update_customer:
    endpoint: PUT /customers/{{ record.id }}.json
    required fields: id
    risk: external mutation; approval required
  create_subscription:
    endpoint: POST /subscriptions.json
    risk: external mutation with billing side effects; approval required
  update_subscription:
    endpoint: PUT /subscriptions/{{ record.id }}.json
    required fields: id
    risk: external mutation with billing side effects; approval required
  cancel_subscription:
    endpoint: POST /subscriptions/{{ record.id }}/cancel.json
    required fields: id
    risk: external mutation with billing side effects; approval required
  create_product_family:
    endpoint: POST /product_families.json
    risk: external mutation; approval required
  create_product:
    endpoint: POST /product_families/{{ record.product_family_id }}/products.json
    required fields: product_family_id
    risk: external mutation; approval required
  update_product:
    endpoint: PUT /products/{{ record.id }}.json
    required fields: id
    risk: external mutation; approval required
  create_coupon:
    endpoint: POST /product_families/{{ record.product_family_id }}/coupons.json
    required fields: product_family_id
    risk: external mutation; approval required
  update_coupon:
    endpoint: PUT /coupons/{{ record.id }}.json
    required fields: id
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
