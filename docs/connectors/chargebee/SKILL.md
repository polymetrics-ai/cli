---
name: pm-chargebee
description: Chargebee connector knowledge and safe action guide.
---

# pm-chargebee

## Purpose

Reads and writes Chargebee subscription billing data (customers, subscriptions, invoices, plans, items, item prices, coupons, credit notes, transactions, orders, quotes, payment sources, events, and more) through the Chargebee v2 REST API.

## Icon

- id: chargebee
- asset: icons/chargebee.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://apidocs.chargebee.com/docs/api/versioning

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- max_pages
- mode
- page_size
- start_date
- site_api_key (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: auto_collection(string), company(string), created_at(integer), deleted(boolean), email(string), first_name(string), id(string), last_name(string), net_term_days(integer), phone(string), taxability(string), updated_at(integer)
- subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), currency_code(string), current_term_end(integer), current_term_start(integer), customer_id(string), deleted(boolean), id(string), plan_amount(integer), plan_id(string), plan_quantity(integer), started_at(integer), status(string), updated_at(integer)
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: amount_due(integer), amount_paid(integer), currency_code(string), customer_id(string), date(integer), deleted(boolean), due_date(integer), id(string), paid_at(integer), status(string), subscription_id(string), total(integer), updated_at(integer)
- plans:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), currency_code(string), id(string), invoice_name(string), name(string), period(integer), period_unit(string), price(integer), status(string), updated_at(integer)
- items:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), enabled_for_checkout(boolean), id(string), is_shippable(boolean), item_family_id(string), name(string), status(string), type(string), updated_at(integer)
- item_prices:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), currency_code(string), deleted(boolean), free_quantity(integer), id(string), is_taxable(boolean), item_family_id(string), item_id(string), item_type(string), name(string), period(integer), period_unit(string), price(integer), pricing_model(string), status(string), updated_at(integer)
- item_families:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), deleted(boolean), description(string), id(string), name(string), status(string), updated_at(integer)
- coupons:
  - primary key: id
  - cursor: updated_at
  - fields: apply_on(string), created_at(integer), currency_code(string), deleted(boolean), discount_amount(integer), discount_percentage(number), discount_type(string), duration_type(string), id(string), name(string), redemptions(integer), status(string), updated_at(integer), valid_till(integer)
- coupon_codes:
  - primary key: code
  - fields: code(string), coupon_id(string), coupon_set_id(string), coupon_set_name(string), status(string)
- coupon_sets:
  - primary key: id
  - fields: archived_count(integer), coupon_id(string), id(string), name(string), redeemed_count(integer), total_count(integer)
- credit_notes:
  - primary key: id
  - cursor: updated_at
  - fields: amount_allocated(integer), amount_available(integer), amount_refunded(integer), currency_code(string), customer_id(string), date(integer), deleted(boolean), id(string), reference_invoice_id(string), status(string), subscription_id(string), total(integer), type(string), updated_at(integer), voided_at(integer)
- transactions:
  - primary key: id
  - cursor: updated_at
  - fields: amount(integer), currency_code(string), customer_id(string), date(integer), deleted(boolean), gateway(string), id(string), payment_method(string), payment_source_id(string), status(string), subscription_id(string), type(string), updated_at(integer)
- orders:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), currency_code(string), customer_id(string), deleted(boolean), document_number(string), id(string), invoice_id(string), order_type(string), price_type(string), status(string), subscription_id(string), total(integer), updated_at(integer)
- quotes:
  - primary key: id
  - cursor: updated_at
  - fields: currency_code(string), customer_id(string), date(integer), id(string), invoice_id(string), name(string), operation_type(string), price_type(string), status(string), subscription_id(string), total(integer), updated_at(integer), valid_till(integer)
- payment_sources:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), customer_id(string), deleted(boolean), gateway(string), gateway_account_id(string), id(string), reference_id(string), status(string), type(string), updated_at(integer)
- events:
  - primary key: id
  - cursor: occurred_at
  - fields: api_version(string), event_type(string), id(string), occurred_at(integer), source(string)
- hosted_pages:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), expires_at(integer), id(string), state(string), type(string), updated_at(integer), url(string)
- virtual_bank_accounts:
  - primary key: id
  - cursor: updated_at
  - fields: account_number(string), bank_name(string), created_at(integer), customer_id(string), deleted(boolean), email(string), gateway(string), gateway_account_id(string), id(string), updated_at(integer)
- unbilled_charges:
  - primary key: id
  - fields: amount(integer), currency_code(string), customer_id(string), date_from(integer), date_to(integer), entity_id(string), entity_type(string), id(string), is_voided(boolean), subscription_id(string)
- ramps:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), deleted(boolean), description(string), effective_from(integer), id(string), status(string), subscription_id(string), updated_at(integer)
- gifts:
  - primary key: id
  - fields: auto_claim(boolean), claim_expiry_date(integer), id(string), no_expiry(boolean), scheduled_at(integer), status(string), updated_at(integer)
- alerts:
  - primary key: id
  - fields: created_at(integer), description(string), id(string), metered_feature_id(string), name(string), status(string), subscription_id(string), type(string), updated_at(integer)
- comments:
  - primary key: id
  - fields: added_by(string), created_at(integer), entity_id(string), entity_type(string), id(string), notes(string), type(string)
- promotional_credits:
  - primary key: id
  - fields: amount(integer), closing_balance(integer), created_at(integer), credit_type(string), currency_code(string), customer_id(string), description(string), id(string), type(string)
- features:
  - primary key: id
  - fields: created_at(integer), description(string), id(string), name(string), status(string), type(string), unit(string), updated_at(integer)
- entitlements:
  - primary key: id
  - fields: entity_id(string), entity_type(string), feature_id(string), feature_name(string), id(string), name(string), value(string)
- differential_prices:
  - primary key: id
  - fields: created_at(integer), currency_code(string), deleted(boolean), id(string), item_price_id(string), parent_item_id(string), price(integer), status(string), updated_at(integer)
- price_variants:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), deleted(boolean), description(string), id(string), name(string), status(string), updated_at(integer), variant_group(string)
- products:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(integer), deleted(boolean), description(string), external_name(string), has_variant(boolean), id(string), name(string), shippable(boolean), sku(string), status(string), updated_at(integer)
- webhook_endpoints:
  - primary key: id
  - fields: api_version(string), disabled(boolean), id(string), name(string), primary_url(boolean), url(string)
- ledger_operations:
  - primary key: id
  - fields: amount(string), created_at(integer), end_balance(string), id(string), modified_at(integer), start_balance(string), subscription_id(string), type(string), unit_id(string), unit_type(string)
- ledger_account_balances:
  - primary key: subscription_id, unit_id, unit_type
  - fields: modified_at(integer), subscription_id(string), unit_id(string), unit_type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation; approval required
- update_customer:
  - endpoint: POST /customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_customer:
  - endpoint: POST /customers/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_item:
  - endpoint: POST /items
  - required fields: id, name, type, item_family_id
  - risk: external mutation; approval required
- update_item:
  - endpoint: POST /items/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_item:
  - endpoint: POST /items/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_item_price:
  - endpoint: POST /item_prices
  - required fields: id, item_id, name
  - risk: external mutation; approval required
- update_item_price:
  - endpoint: POST /item_prices/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_item_price:
  - endpoint: POST /item_prices/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_item_family:
  - endpoint: POST /item_families
  - required fields: id, name
  - risk: external mutation; approval required
- update_item_family:
  - endpoint: POST /item_families/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_item_family:
  - endpoint: POST /item_families/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_subscription:
  - endpoint: POST /customers/{{ record.customer_id }}/subscription_for_items
  - required fields: customer_id
  - risk: external mutation with billing side effects; approval required
- update_subscription:
  - endpoint: POST /subscriptions/{{ record.id }}/update_for_items
  - required fields: id
  - risk: external mutation with billing side effects; approval required
- cancel_subscription:
  - endpoint: POST /subscriptions/{{ record.id }}/cancel_for_items
  - required fields: id
  - risk: irreversible external mutation (subscription cancellation) with billing side effects; approval required
- create_credit_note:
  - endpoint: POST /credit_notes
  - required fields: type
  - risk: external mutation with accounting/billing side effects; approval required
- void_credit_note:
  - endpoint: POST /credit_notes/{{ record.id }}/void
  - required fields: id
  - risk: irreversible external mutation; approval required
- create_coupon:
  - endpoint: POST /coupons/create_for_items
  - required fields: id, name, apply_on
  - risk: external mutation with billing/discount side effects; approval required
- update_coupon:
  - endpoint: POST /coupons/{{ record.id }}/update_for_items
  - required fields: id
  - risk: external mutation with billing/discount side effects; approval required
- delete_coupon:
  - endpoint: POST /coupons/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_order:
  - endpoint: POST /orders
  - required fields: invoice_id
  - risk: external mutation; approval required
- update_order:
  - endpoint: POST /orders/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- cancel_order:
  - endpoint: POST /orders/{{ record.id }}/cancel
  - required fields: id, cancellation_reason
  - risk: irreversible external mutation (order cancellation); approval required
- void_invoice:
  - endpoint: POST /invoices/{{ record.id }}/void
  - required fields: id
  - risk: irreversible external mutation with accounting side effects; approval required
- collect_payment_for_invoice:
  - endpoint: POST /invoices/{{ record.id }}/collect_payment
  - required fields: id
  - risk: external mutation that attempts to charge a payment method; approval required
- create_webhook_endpoint:
  - endpoint: POST /webhook_endpoints
  - required fields: name, url
  - risk: external mutation exposing business data to a third-party URL; approval required
- update_webhook_endpoint:
  - endpoint: POST /webhook_endpoints/{{ record.id }}
  - required fields: id
  - risk: external mutation exposing business data to a third-party URL; approval required
- delete_webhook_endpoint:
  - endpoint: POST /webhook_endpoints/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_comment:
  - endpoint: POST /comments
  - required fields: entity_type, entity_id, notes
  - risk: external mutation; approval required
- delete_comment:
  - endpoint: POST /comments/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- add_promotional_credit:
  - endpoint: POST /promotional_credits/add
  - required fields: customer_id, description
  - risk: external mutation with a direct billing-credit financial effect; approval required
- deduct_promotional_credit:
  - endpoint: POST /promotional_credits/deduct
  - required fields: customer_id, description
  - risk: external mutation with a direct billing-credit financial effect; approval required
- create_virtual_bank_account:
  - endpoint: POST /virtual_bank_accounts
  - required fields: customer_id
  - risk: external mutation; approval required
- delete_virtual_bank_account:
  - endpoint: POST /virtual_bank_accounts/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_card_payment_source:
  - endpoint: POST /payment_sources/create_card
  - required fields: customer_id
  - risk: external mutation carrying raw payment-card data; approval required
- delete_payment_source:
  - endpoint: POST /payment_sources/{{ record.id }}/delete
  - required fields: id
  - risk: irreversible external deletion; approval required

## Security

- read risk: external Chargebee API read of customer and billing data
- write risk: external mutation of Chargebee billing data (customers, subscriptions, invoices, credit notes, orders, coupons, payment sources); several actions have direct financial/billing side effects and require approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect chargebee
```

### Inspect as structured JSON

```bash
pm connectors inspect chargebee --json
```

## Agent Rules

- Run pm connectors inspect chargebee before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
