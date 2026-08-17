---
name: pm-revenuecat
description: RevenueCat connector knowledge and safe action guide.
---

# pm-revenuecat

## Purpose

Reads and writes RevenueCat v2 project configuration, customer, product, offering, subscription, purchase, paywall, virtual currency, integration, and metrics resources through the REST API.

## Icon

- asset: icons/revenuecat.svg
- source: official
- review_status: official_verified
- review_url: https://www.revenuecat.com/docs/api-v1

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- app_id
- base_url
- chart_name
- created_after
- customer_id
- entitlement_id
- invoice_id
- offering_id
- package_id
- paywall_id
- product_id
- project_id
- purchase_id
- starting_after
- store_purchase_identifier
- store_subscription_identifier
- subscription_id
- updated_after
- virtual_currency_code
- webhook_integration_id
- api_key (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: created_at(), icon_url(), icon_url_large(), id(), name(), object(), stream()
- apps:
  - primary key: id
  - fields: created_at(), id(), name(), object(), project_id(), stream(), type()
- products:
  - primary key: id
  - fields: app_id(), created_at(), display_name(), id(), object(), state(), store_identifier(), stream(), type()
- offerings:
  - primary key: id
  - fields: created_at(), description(), display_name(), id(), identifier(), lookup_key(), object(), project_id(), state(), stream()
- customers:
  - primary key: id
  - fields: app_user_id(), created_at(), first_seen_at(), id(), last_seen_at(), object(), project_id(), stream(), updated_at()
- app:
  - primary key: id
  - fields: created_at(), id(), name(), object(), project_id(), stream(), type()
- app_public_api_keys:
  - primary key: id
  - fields: app_id(), created_at(), environment(), id(), key(), object(), stream()
- app_store_kit_config:
  - primary key: app_id
  - fields: app_id(), contents(), object(), stream()
- audit_logs:
  - primary key: id
  - fields: action_type(), actor_identifier(), actor_type(), additional_data(), id(), object(), occurred_at(), project_id(), stream(), target_identifier(), target_type()
- chart_data:
  - primary key: chart_name
  - fields: category(), chart_name(), description(), display_name(), display_type(), object(), resolution(), stream(), summary(), values()
- chart_options:
  - primary key: chart_name
  - fields: chart_name(), filters(), object(), resolutions(), segments(), stream(), user_selectors()
- collaborators:
  - primary key: id
  - fields: accepted_at(), email(), has_mfa(), id(), name(), object(), role(), stream()
- customer:
  - primary key: id
  - fields: active_entitlements(), attributes(), first_seen_at(), id(), last_seen_at(), object(), project_id(), stream()
- customer_active_entitlements:
  - primary key: entitlement_id
  - fields: entitlement_id(), expires_at(), object(), stream()
- customer_aliases:
  - primary key: id
  - fields: created_at(), id(), object(), stream()
- customer_attributes:
  - primary key: name
  - fields: name(), object(), stream(), updated_at(), value()
- customer_center:
  - primary key: customer_id
  - fields: customer_center(), customer_id(), object(), stream()
- customer_invoices:
  - primary key: id
  - fields: id(), invoice_url(), issued_at(), line_items(), object(), paid_at(), stream(), total_amount()
- customer_purchases:
  - primary key: id
  - fields: customer_id(), id(), object(), product_id(), purchased_at(), status(), store_purchase_identifier(), stream()
- customer_subscriptions:
  - primary key: id
  - fields: customer_id(), id(), object(), product_id(), starts_at(), status(), store_subscription_identifier(), stream()
- customer_virtual_currencies:
  - primary key: currency_code
  - fields: balance(), currency_code(), description(), name(), object(), stream()
- entitlements:
  - primary key: id
  - fields: created_at(), display_name(), id(), lookup_key(), object(), products(), project_id(), state(), stream()
- entitlement:
  - primary key: id
  - fields: created_at(), display_name(), id(), lookup_key(), object(), products(), project_id(), state(), stream()
- entitlement_products:
  - primary key: id
  - fields: app_id(), created_at(), display_name(), id(), object(), state(), store_identifier(), stream(), type()
- webhook_integrations:
  - primary key: id
  - fields: app_id(), created_at(), environment(), event_types(), id(), name(), object(), project_id(), stream(), url()
- webhook_integration:
  - primary key: id
  - fields: app_id(), created_at(), environment(), event_types(), id(), name(), object(), project_id(), stream(), url()
- overview_metrics:
  - primary key: project_id
  - fields: currency(), metrics(), object(), project_id(), stream()
- revenue_metric:
  - primary key: project_id
  - fields: currency(), end_date(), object(), project_id(), revenue_type(), start_date(), stream(), value()
- offering:
  - primary key: id
  - fields: created_at(), display_name(), id(), is_current(), lookup_key(), metadata(), object(), packages(), paywall_id(), project_id(), stream()
- offering_packages:
  - primary key: id
  - fields: created_at(), display_name(), id(), lookup_key(), object(), position(), products(), stream()
- package:
  - primary key: id
  - fields: created_at(), display_name(), id(), lookup_key(), object(), position(), products(), stream()
- package_products:
  - primary key: id
  - fields: eligibility_criteria(), id(), object(), product(), stream()
- paywalls:
  - primary key: id
  - fields: automatically_scale_font_size(), created_at(), id(), name(), object(), offering_id(), published_at(), stream()
- paywall:
  - primary key: id
  - fields: automatically_scale_font_size(), components(), created_at(), id(), name(), object(), offering_id(), published_at(), stream()
- product:
  - primary key: id
  - fields: app_id(), created_at(), display_name(), id(), object(), state(), store_identifier(), stream(), type()
- purchases:
  - primary key: id
  - fields: customer_id(), id(), object(), product_id(), purchased_at(), status(), store_purchase_identifier(), stream()
- purchase:
  - primary key: id
  - fields: customer_id(), id(), object(), product_id(), purchased_at(), status(), store_purchase_identifier(), stream()
- purchase_entitlements:
  - primary key: id
  - fields: created_at(), display_name(), id(), lookup_key(), object(), project_id(), state(), stream()
- subscriptions:
  - primary key: id
  - fields: customer_id(), id(), object(), product_id(), starts_at(), status(), store_subscription_identifier(), stream()
- subscription:
  - primary key: id
  - fields: customer_id(), id(), management_url(), object(), product_id(), starts_at(), status(), store_subscription_identifier(), stream()
- subscription_authenticated_management_url:
  - primary key: subscription_id
  - fields: management_url(), object(), stream(), subscription_id()
- subscription_entitlements:
  - primary key: id
  - fields: created_at(), display_name(), id(), lookup_key(), object(), project_id(), state(), stream()
- subscription_transactions:
  - primary key: id
  - fields: expiration_date(), id(), object(), product_store_identifier(), purchased_at(), revenue_in_usd(), stream()
- virtual_currencies:
  - primary key: code
  - fields: code(), created_at(), description(), name(), object(), product_grants(), project_id(), state(), stream()
- virtual_currency:
  - primary key: code
  - fields: code(), created_at(), description(), name(), object(), product_grants(), project_id(), state(), stream()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_project:
  - endpoint: POST /projects
  - risk: creates a RevenueCat project
- create_app:
  - endpoint: POST /projects/{{ config.project_id }}/apps
  - risk: creates a RevenueCat app in the configured project
- update_app:
  - endpoint: POST /projects/{{ config.project_id }}/apps/{{ record.app_id }}
  - required fields: app_id
  - risk: updates an existing RevenueCat app
- delete_app:
  - endpoint: DELETE /projects/{{ config.project_id }}/apps/{{ record.app_id }}
  - required fields: app_id
  - risk: permanently deletes a RevenueCat app and its configuration
- create_customer:
  - endpoint: POST /projects/{{ config.project_id }}/customers
  - risk: creates a RevenueCat customer
- delete_customer:
  - endpoint: DELETE /projects/{{ config.project_id }}/customers/{{ record.customer_id }}
  - required fields: customer_id
  - risk: permanently deletes a RevenueCat customer
- assign_customer_offering:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/assign_offering
  - required fields: customer_id
  - risk: assigns or clears a customer offering override
- grant_customer_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/grant_entitlement
  - required fields: customer_id
  - risk: grants an entitlement to a customer
- restore_purchase_by_order_id:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/restore_purchase_by_order_id
  - required fields: customer_id
  - risk: restores a Google Play purchase by order id
- revoke_customer_granted_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/revoke_granted_entitlement
  - required fields: customer_id
  - risk: revokes a granted customer entitlement
- transfer_customer_data:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/transfer
  - required fields: customer_id
  - risk: transfers subscriptions and purchases to another customer
- set_customer_attributes:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/attributes
  - required fields: customer_id
  - risk: sets customer attributes
- create_virtual_currencies_transaction:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/virtual_currencies/transactions
  - required fields: customer_id
  - risk: creates a virtual currency transaction for a customer
- update_virtual_currencies_balance:
  - endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/virtual_currencies/update_balance
  - required fields: customer_id
  - risk: updates a customer virtual currency balance without creating a transaction
- create_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/entitlements
  - risk: creates an entitlement
- update_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}
  - required fields: entitlement_id
  - risk: updates an entitlement
- delete_entitlement:
  - endpoint: DELETE /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}
  - required fields: entitlement_id
  - risk: deletes an entitlement
- archive_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/archive
  - required fields: entitlement_id
  - risk: archives an entitlement
- unarchive_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/unarchive
  - required fields: entitlement_id
  - risk: unarchives an entitlement
- attach_products_to_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/attach_products
  - required fields: entitlement_id
  - risk: attaches products to an entitlement
- detach_products_from_entitlement:
  - endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/detach_products
  - required fields: entitlement_id
  - risk: detaches products from an entitlement
- create_webhook_integration:
  - endpoint: POST /projects/{{ config.project_id }}/integrations/webhooks
  - risk: creates a webhook integration that sends events to the configured URL
- update_webhook_integration:
  - endpoint: POST /projects/{{ config.project_id }}/integrations/webhooks/{{ record.webhook_integration_id }}
  - required fields: webhook_integration_id
  - risk: updates a webhook integration
- delete_webhook_integration:
  - endpoint: DELETE /projects/{{ config.project_id }}/integrations/webhooks/{{ record.webhook_integration_id }}
  - required fields: webhook_integration_id
  - risk: deletes a webhook integration
- create_offering:
  - endpoint: POST /projects/{{ config.project_id }}/offerings
  - risk: creates an offering
- update_offering:
  - endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}
  - required fields: offering_id
  - risk: updates an offering
- delete_offering:
  - endpoint: DELETE /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}
  - required fields: offering_id
  - risk: deletes an offering and attached packages
- archive_offering:
  - endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}/actions/archive
  - required fields: offering_id
  - risk: archives an offering
- unarchive_offering:
  - endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}/actions/unarchive
  - required fields: offering_id
  - risk: unarchives an offering
- create_package:
  - endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}/packages
  - required fields: offering_id
  - risk: creates a package in an offering
- update_package:
  - endpoint: POST /projects/{{ config.project_id }}/packages/{{ record.package_id }}
  - required fields: package_id
  - risk: updates a package
- delete_package:
  - endpoint: DELETE /projects/{{ config.project_id }}/packages/{{ record.package_id }}
  - required fields: package_id
  - risk: deletes a package
- attach_products_to_package:
  - endpoint: POST /projects/{{ config.project_id }}/packages/{{ record.package_id }}/actions/attach_products
  - required fields: package_id
  - risk: attaches products to a package
- detach_products_from_package:
  - endpoint: POST /projects/{{ config.project_id }}/packages/{{ record.package_id }}/actions/detach_products
  - required fields: package_id
  - risk: detaches products from a package
- create_paywall:
  - endpoint: POST /projects/{{ config.project_id }}/paywalls
  - risk: creates a paywall
- update_paywall:
  - endpoint: PATCH /projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}
  - required fields: paywall_id
  - risk: updates a paywall draft
- delete_paywall:
  - endpoint: DELETE /projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}
  - required fields: paywall_id
  - risk: deletes a paywall
- create_paywall_version:
  - endpoint: POST /projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}/versions
  - required fields: paywall_id
  - risk: creates a paywall version
- create_product:
  - endpoint: POST /projects/{{ config.project_id }}/products
  - risk: creates a product
- update_product:
  - endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}
  - required fields: product_id
  - risk: updates a product
- delete_product:
  - endpoint: DELETE /projects/{{ config.project_id }}/products/{{ record.product_id }}
  - required fields: product_id
  - risk: deletes a product
- archive_product:
  - endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}/actions/archive
  - required fields: product_id
  - risk: archives a product
- unarchive_product:
  - endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}/actions/unarchive
  - required fields: product_id
  - risk: unarchives a product
- create_product_in_store:
  - endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}/create_in_store
  - required fields: product_id
  - risk: pushes a product to the configured store
- refund_purchase:
  - endpoint: POST /projects/{{ config.project_id }}/purchases/{{ record.purchase_id }}/actions/refund
  - required fields: purchase_id
  - risk: refunds a Web Billing purchase
- cancel_subscription:
  - endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/actions/cancel
  - required fields: subscription_id
  - risk: cancels an active Web Billing subscription
- extend_subscription:
  - endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/actions/extend
  - required fields: subscription_id
  - risk: extends the current billing period of a subscription
- refund_subscription:
  - endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/actions/refund
  - required fields: subscription_id
  - risk: refunds an active Web Billing subscription
- refund_subscription_transaction:
  - endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/transactions/{{ record.transaction_id }}/actions/refund
  - required fields: subscription_id, transaction_id
  - risk: refunds a Play Store or Galaxy subscription transaction
- create_virtual_currency:
  - endpoint: POST /projects/{{ config.project_id }}/virtual_currencies
  - risk: creates a virtual currency
- update_virtual_currency:
  - endpoint: POST /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}
  - required fields: virtual_currency_code
  - risk: updates a virtual currency
- delete_virtual_currency:
  - endpoint: DELETE /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}
  - required fields: virtual_currency_code
  - risk: deletes a virtual currency
- archive_virtual_currency:
  - endpoint: POST /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}/actions/archive
  - required fields: virtual_currency_code
  - risk: archives a virtual currency
- unarchive_virtual_currency:
  - endpoint: POST /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}/actions/unarchive
  - required fields: virtual_currency_code
  - risk: unarchives a virtual currency

## Security

- read risk: external RevenueCat API reads of project, app, product, offering, customer, subscription, purchase, paywall, virtual currency, integration, metrics, and audit-log data
- write risk: external RevenueCat API mutations that create, update, archive, delete, refund, cancel, transfer, grant, revoke, or otherwise alter project and customer resources
- approval: reverse ETL writes require plan preview and approval token; destructive deletes/refunds/cancellations are flagged as destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect revenuecat
```

### Inspect as structured JSON

```bash
pm connectors inspect revenuecat --json
```

## Agent Rules

- Run pm connectors inspect revenuecat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
