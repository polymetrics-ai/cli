# pm connectors inspect pabbly-subscriptions-billing

```text
NAME
  pm connectors inspect pabbly-subscriptions-billing - Pabbly Subscriptions Billing connector manual

SYNOPSIS
  pm connectors inspect pabbly-subscriptions-billing
  pm connectors inspect pabbly-subscriptions-billing --json
  pm credentials add <name> --connector pabbly-subscriptions-billing [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pabbly customers, subscriptions, plans, and invoices, and writes customer/product/plan mutations and subscription cancellations through the Pabbly Subscriptions Billing REST API.

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
  page_size
  username
  password (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), customer_id(), email(), id(), plan_id(), status()
  subscriptions:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), customer_id(), email(), id(), plan_id(), status()
  plans:
    primary key: id
    fields: amount(), created_at(), currency(), customer_id(), email(), id(), plan_id(), status()
  invoices:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), customer_id(), email(), id(), plan_id(), status()
  products:
    primary key: id
    fields: createdAt(), description(), id(), notification_email(), product_name(), redirect_url(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customer
    risk: external mutation; creates a billing customer record in Pabbly Subscriptions; approval required
  update_customer:
    endpoint: PUT /customer/{{ record.id }}
    required fields: id
    risk: external mutation; updates a billing customer record in Pabbly Subscriptions; approval required
  create_subscription:
    endpoint: POST /subscription/{{ record.customer_id }}
    required fields: customer_id
    risk: external mutation; creates a live billing subscription for an existing customer and starts recurring charges; approval required
  update_subscription:
    endpoint: PUT /subscription/{{ record.id }}/update
    required fields: id
    risk: external mutation; changes an existing subscription's plan/payment terms, altering future customer billing; approval required
  create_coupon:
    endpoint: POST /coupon/{{ record.product_id }}
    required fields: product_id
    risk: external mutation; creates a discount coupon that lowers future billing amounts for any customer who redeems it; approval required
  create_payment_method:
    endpoint: POST /paymentmethod/{{ record.customer_id }}
    required fields: customer_id
    risk: external mutation; stores a new payment card on file for an existing customer via the connected payment gateway; approval required
  update_payment_method:
    endpoint: PUT /paymentmethod/{{ record.customer_id }}
    required fields: customer_id
    risk: external mutation; replaces the card an existing customer's future recurring billing charges against; approval required
  create_addon:
    endpoint: POST /addon/{{ record.product_id }}
    required fields: product_id
    risk: external mutation; creates a sellable add-on that customers can attach to a plan, adding to their bill; approval required
  update_addon:
    endpoint: PUT /addon/{{ record.addon_id }}
    required fields: addon_id
    risk: external mutation; updates an existing sellable add-on's price/billing terms, changing future charges for customers who have it attached; approval required
  delete_addon:
    endpoint: DELETE /addon/{{ record.addon_id }}
    required fields: addon_id
    risk: external mutation; permanently deletes a sellable add-on; approval required
  create_addon_category:
    endpoint: POST /addoncategory
    risk: external mutation; creates an add-on category used to organize sellable add-ons in Pabbly Subscriptions; approval required
  update_addon_category:
    endpoint: PUT /addoncategory/{{ record.category_id }}
    required fields: category_id
    risk: external mutation; renames an existing add-on category; approval required
  delete_addon_category:
    endpoint: DELETE /addoncategory/{{ record.category_id }}
    required fields: category_id
    risk: external mutation; permanently deletes an add-on category; approval required
  create_license:
    endpoint: POST /products/{{ record.product_id }}/licenses
    required fields: product_id
    risk: external mutation; creates a license-key pool for a product's plan in Pabbly Subscriptions; approval required
  update_license:
    endpoint: PUT /products/{{ record.product_id }}/licenses/{{ record.license_id }}
    required fields: product_id, license_id
    risk: external mutation; updates an existing license-key pool (may add/replace codes) in Pabbly Subscriptions; approval required
  cancel_subscription:
    endpoint: POST /subscription/{{ record.id }}/cancel
    required fields: id
    risk: external mutation; cancels a live billing subscription (immediately or at term end) and stops future customer billing; approval required
  create_product:
    endpoint: POST /product/create
    risk: external mutation; creates a sellable product in Pabbly Subscriptions; approval required
  update_product:
    endpoint: PUT /product/update/{{ record.id }}
    required fields: id
    risk: external mutation; updates a sellable product in Pabbly Subscriptions; approval required
  create_plan:
    endpoint: POST /plan/create
    risk: external mutation; creates a billing plan under a product in Pabbly Subscriptions; approval required
  update_plan:
    endpoint: PUT /plan/update/{{ record.id }}
    required fields: id
    risk: external mutation; updates a billing plan in Pabbly Subscriptions; approval required

SECURITY
  read risk: external Pabbly Subscriptions Billing API read of customer and billing data
  write risk: external mutation; creates/updates billing customers, products, and plans, and cancels live subscriptions (stops future billing)
  approval: approval required before writes; cancel_subscription is a destructive action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pabbly-subscriptions-billing

  # Inspect as structured JSON
  pm connectors inspect pabbly-subscriptions-billing --json

AGENT WORKFLOW
  - Run pm connectors inspect pabbly-subscriptions-billing before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
