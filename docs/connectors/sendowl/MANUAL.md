# pm connectors inspect sendowl

```text
NAME
  pm connectors inspect sendowl - SendOwl connector manual

SYNOPSIS
  pm connectors inspect sendowl
  pm connectors inspect sendowl --json
  pm credentials add <name> --connector sendowl [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SendOwl orders, products, subscriptions, discounts, bundles, and licenses, and writes product/subscription/discount/bundle lifecycle mutations and order actions (refund, cancel subscription, resend email) through the SendOwl API.

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
  username
  password (secret)

ETL STREAMS
  orders:
    primary key: id
    cursor: created_at
    fields: buyer_email(), buyer_name(), created_at(), currency(), id(), name(), price(), state()
  products:
    primary key: id
    fields: created_at(), currency(), id(), name(), price(), product_type()
  subscriptions:
    primary key: id
    fields: buyer_email(), created_at(), id(), name(), state()
  discounts:
    primary key: id
    fields: code(), created_at(), currency_code(), current_uses(), discount_flat_rate(), discount_percentage(), end_at(), id(), max_uses(), start_at()
  bundles:
    primary key: id
    fields: created_at(), currency_code(), id(), name(), price(), self_hosted()
  licenses:
    primary key: id
    fields: created_at(), id(), key(), product_id(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_product:
    endpoint: POST /api/v1/products
    risk: creates a new sellable product (no file attachment; SendOwl's file-upload create path is a separate multipart-only endpoint this dialect cannot express, see docs.md Known limits); external mutation, approval required
  update_product:
    endpoint: PUT /api/v1/products/{{ record.id }}
    required fields: id
    risk: mutates an existing product's name/price/currency; affects future checkouts of this product
  delete_product:
    endpoint: DELETE /api/v1/products/{{ record.id }}
    required fields: id
    risk: permanently removes a product; breaks any existing order-fulfillment/download links referencing it
  create_subscription:
    endpoint: POST /api/v1/subscriptions
    risk: creates a new recurring-billing subscription product; external mutation, approval required
  update_subscription:
    endpoint: PUT /api/v1/subscriptions/{{ record.id }}
    required fields: id
    risk: mutates an existing subscription's name/price; affects future recurring charges for new subscribers to this plan
  delete_subscription:
    endpoint: DELETE /api/v1/subscriptions/{{ record.id }}
    required fields: id
    risk: permanently removes a subscription product; does not itself cancel any buyer's already-active recurring order (see cancel_order_subscription)
  create_discount:
    endpoint: POST /api/v1_2/discounts
    risk: creates a new discount code usable at checkout; external mutation, approval required
  update_discount:
    endpoint: PUT /api/v1_2/discounts/{{ record.id }}
    required fields: id
    risk: mutates an existing discount's percentage/usage cap/expiry; affects all buyers subsequently applying this code
  delete_discount:
    endpoint: DELETE /api/v1_2/discounts/{{ record.id }}
    required fields: id
    risk: permanently removes a discount code; any buyer with the code saved can no longer redeem it
  update_bundle:
    endpoint: PUT /api/v1/packages/{{ record.id }}
    required fields: id
    risk: mutates an existing bundle's name/price; affects future checkouts of this bundle
  delete_bundle:
    endpoint: DELETE /api/v1/packages/{{ record.id }}
    required fields: id
    risk: permanently removes a bundle; breaks any existing order-fulfillment links referencing it
  refund_order:
    endpoint: POST /api/v1/orders/{{ record.id }}/refund
    required fields: id
    optional fields: amount, cancel_subscription, revoke_access
    risk: issues a real financial refund against the buyer's original payment method; irreversible external money movement, approval required
  cancel_order_subscription:
    endpoint: PUT /api/v1/orders/{{ record.id }}/cancel_subscription
    required fields: id
    risk: cancels the buyer's active recurring subscription tied to this order; stops future recurring charges, irreversible without the buyer re-subscribing
  resend_order_email:
    endpoint: POST /api/v1/orders/{{ record.id }}/resend_email
    required fields: id
    optional fields: type
    risk: resends the order confirmation/receipt/download email to the buyer's address on file; low-risk external side effect, no approval required

SECURITY
  read risk: external SendOwl API read of order, product, subscription, discount, bundle, and license data
  write risk: external SendOwl API mutation, including a real financial refund action (refund_order)
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sendowl

  # Inspect as structured JSON
  pm connectors inspect sendowl --json

AGENT WORKFLOW
  - Run pm connectors inspect sendowl before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
