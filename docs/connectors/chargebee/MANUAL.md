# pm connectors inspect chargebee

```text
NAME
  pm connectors inspect chargebee - Chargebee connector manual

SYNOPSIS
  pm connectors inspect chargebee
  pm connectors inspect chargebee --json
  pm credentials add <name> --connector chargebee [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Chargebee subscription billing data (customers, subscriptions, invoices, plans, items, item prices, coupons, credit notes, transactions, orders, quotes, payment sources, events, and more) through the Chargebee v2 REST API.

ICON
  asset: icons/chargebee.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.chargebee.com/docs/api/versioning

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  start_date
  site_api_key (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: auto_collection(), company(), created_at(), deleted(), email(), first_name(), id(), last_name(), net_term_days(), phone(), taxability(), updated_at()
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: created_at(), currency_code(), current_term_end(), current_term_start(), customer_id(), deleted(), id(), plan_amount(), plan_id(), plan_quantity(), started_at(), status(), updated_at()
  invoices:
    primary key: id
    cursor: updated_at
    fields: amount_due(), amount_paid(), currency_code(), customer_id(), date(), deleted(), due_date(), id(), paid_at(), status(), subscription_id(), total(), updated_at()
  plans:
    primary key: id
    cursor: updated_at
    fields: created_at(), currency_code(), id(), invoice_name(), name(), period(), period_unit(), price(), status(), updated_at()
  items:
    primary key: id
    cursor: updated_at
    fields: created_at(), enabled_for_checkout(), id(), is_shippable(), item_family_id(), name(), status(), type(), updated_at()
  item_prices:
    primary key: id
    cursor: updated_at
    fields: created_at(), currency_code(), deleted(), free_quantity(), id(), is_taxable(), item_family_id(), item_id(), item_type(), name(), period(), period_unit(), price(), pricing_model(), status(), updated_at()
  item_families:
    primary key: id
    cursor: updated_at
    fields: created_at(), deleted(), description(), id(), name(), status(), updated_at()
  coupons:
    primary key: id
    fields: apply_on(), created_at(), currency_code(), deleted(), discount_amount(), discount_percentage(), discount_type(), duration_type(), id(), name(), redemptions(), status(), updated_at(), valid_till()
  coupon_codes:
    primary key: code
    fields: code(), coupon_id(), coupon_set_id(), coupon_set_name(), status()
  coupon_sets:
    primary key: id
    fields: archived_count(), coupon_id(), id(), name(), redeemed_count(), total_count()
  credit_notes:
    primary key: id
    cursor: updated_at
    fields: amount_allocated(), amount_available(), amount_refunded(), currency_code(), customer_id(), date(), deleted(), id(), reference_invoice_id(), status(), subscription_id(), total(), type(), updated_at(), voided_at()
  transactions:
    primary key: id
    cursor: updated_at
    fields: amount(), currency_code(), customer_id(), date(), deleted(), gateway(), id(), payment_method(), payment_source_id(), status(), subscription_id(), type(), updated_at()
  orders:
    primary key: id
    cursor: updated_at
    fields: created_at(), currency_code(), customer_id(), deleted(), document_number(), id(), invoice_id(), order_type(), price_type(), status(), subscription_id(), total(), updated_at()
  quotes:
    primary key: id
    cursor: updated_at
    fields: currency_code(), customer_id(), date(), id(), invoice_id(), name(), operation_type(), price_type(), status(), subscription_id(), total(), updated_at(), valid_till()
  payment_sources:
    primary key: id
    cursor: updated_at
    fields: created_at(), customer_id(), deleted(), gateway(), gateway_account_id(), id(), reference_id(), status(), type(), updated_at()
  events:
    primary key: id
    cursor: occurred_at
    fields: api_version(), event_type(), id(), occurred_at(), source()
  hosted_pages:
    primary key: id
    cursor: updated_at
    fields: created_at(), expires_at(), id(), state(), type(), updated_at(), url()
  virtual_bank_accounts:
    primary key: id
    cursor: updated_at
    fields: account_number(), bank_name(), created_at(), customer_id(), deleted(), email(), gateway(), gateway_account_id(), id(), updated_at()
  unbilled_charges:
    primary key: id
    fields: amount(), currency_code(), customer_id(), date_from(), date_to(), entity_id(), entity_type(), id(), is_voided(), subscription_id()
  ramps:
    primary key: id
    cursor: updated_at
    fields: created_at(), deleted(), description(), effective_from(), id(), status(), subscription_id(), updated_at()
  gifts:
    primary key: id
    fields: auto_claim(), claim_expiry_date(), id(), no_expiry(), scheduled_at(), status(), updated_at()
  alerts:
    primary key: id
    fields: created_at(), description(), id(), metered_feature_id(), name(), status(), subscription_id(), type(), updated_at()
  comments:
    primary key: id
    fields: added_by(), created_at(), entity_id(), entity_type(), id(), notes(), type()
  promotional_credits:
    primary key: id
    fields: amount(), closing_balance(), created_at(), credit_type(), currency_code(), customer_id(), description(), id(), type()
  features:
    primary key: id
    fields: created_at(), description(), id(), name(), status(), type(), unit(), updated_at()
  entitlements:
    primary key: id
    fields: entity_id(), entity_type(), feature_id(), feature_name(), id(), name(), value()
  differential_prices:
    primary key: id
    fields: created_at(), currency_code(), deleted(), id(), item_price_id(), parent_item_id(), price(), status(), updated_at()
  price_variants:
    primary key: id
    cursor: updated_at
    fields: created_at(), deleted(), description(), id(), name(), status(), updated_at(), variant_group()
  products:
    primary key: id
    cursor: updated_at
    fields: created_at(), deleted(), description(), external_name(), has_variant(), id(), name(), shippable(), sku(), status(), updated_at()
  webhook_endpoints:
    primary key: id
    fields: api_version(), disabled(), id(), name(), primary_url(), url()
  ledger_operations:
    primary key: id
    fields: amount(), created_at(), end_balance(), id(), modified_at(), start_balance(), subscription_id(), type(), unit_id(), unit_type()
  ledger_account_balances:
    primary key: subscription_id, unit_id, unit_type
    fields: modified_at(), subscription_id(), unit_id(), unit_type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: external mutation; approval required
  update_customer:
    endpoint: POST /customers/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_customer:
    endpoint: POST /customers/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_item:
    endpoint: POST /items
    risk: external mutation; approval required
  update_item:
    endpoint: POST /items/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_item:
    endpoint: POST /items/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_item_price:
    endpoint: POST /item_prices
    risk: external mutation; approval required
  update_item_price:
    endpoint: POST /item_prices/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_item_price:
    endpoint: POST /item_prices/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_item_family:
    endpoint: POST /item_families
    risk: external mutation; approval required
  update_item_family:
    endpoint: POST /item_families/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_item_family:
    endpoint: POST /item_families/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_subscription:
    endpoint: POST /customers/{{ record.customer_id }}/subscription_for_items
    required fields: customer_id
    risk: external mutation with billing side effects; approval required
  update_subscription:
    endpoint: POST /subscriptions/{{ record.id }}/update_for_items
    required fields: id
    risk: external mutation with billing side effects; approval required
  cancel_subscription:
    endpoint: POST /subscriptions/{{ record.id }}/cancel_for_items
    required fields: id
    risk: irreversible external mutation (subscription cancellation) with billing side effects; approval required
  create_credit_note:
    endpoint: POST /credit_notes
    risk: external mutation with accounting/billing side effects; approval required
  void_credit_note:
    endpoint: POST /credit_notes/{{ record.id }}/void
    required fields: id
    risk: irreversible external mutation; approval required
  create_coupon:
    endpoint: POST /coupons/create_for_items
    risk: external mutation with billing/discount side effects; approval required
  update_coupon:
    endpoint: POST /coupons/{{ record.id }}/update_for_items
    required fields: id
    risk: external mutation with billing/discount side effects; approval required
  delete_coupon:
    endpoint: POST /coupons/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_order:
    endpoint: POST /orders
    risk: external mutation; approval required
  update_order:
    endpoint: POST /orders/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  cancel_order:
    endpoint: POST /orders/{{ record.id }}/cancel
    required fields: id
    risk: irreversible external mutation (order cancellation); approval required
  void_invoice:
    endpoint: POST /invoices/{{ record.id }}/void
    required fields: id
    risk: irreversible external mutation with accounting side effects; approval required
  collect_payment_for_invoice:
    endpoint: POST /invoices/{{ record.id }}/collect_payment
    required fields: id
    risk: external mutation that attempts to charge a payment method; approval required
  create_webhook_endpoint:
    endpoint: POST /webhook_endpoints
    risk: external mutation exposing business data to a third-party URL; approval required
  update_webhook_endpoint:
    endpoint: POST /webhook_endpoints/{{ record.id }}
    required fields: id
    risk: external mutation exposing business data to a third-party URL; approval required
  delete_webhook_endpoint:
    endpoint: POST /webhook_endpoints/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_comment:
    endpoint: POST /comments
    risk: external mutation; approval required
  delete_comment:
    endpoint: POST /comments/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  add_promotional_credit:
    endpoint: POST /promotional_credits/add
    risk: external mutation with a direct billing-credit financial effect; approval required
  deduct_promotional_credit:
    endpoint: POST /promotional_credits/deduct
    risk: external mutation with a direct billing-credit financial effect; approval required
  create_virtual_bank_account:
    endpoint: POST /virtual_bank_accounts
    risk: external mutation; approval required
  delete_virtual_bank_account:
    endpoint: POST /virtual_bank_accounts/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required
  create_card_payment_source:
    endpoint: POST /payment_sources/create_card
    risk: external mutation carrying raw payment-card data; approval required
  delete_payment_source:
    endpoint: POST /payment_sources/{{ record.id }}/delete
    required fields: id
    risk: irreversible external deletion; approval required

SECURITY
  read risk: external Chargebee API read of customer and billing data
  write risk: external mutation of Chargebee billing data (customers, subscriptions, invoices, credit notes, orders, coupons, payment sources); several actions have direct financial/billing side effects and require approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chargebee

  # Inspect as structured JSON
  pm connectors inspect chargebee --json

AGENT WORKFLOW
  - Run pm connectors inspect chargebee before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
