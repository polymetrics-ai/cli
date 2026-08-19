# pm connectors inspect revenuecat

```text
NAME
  pm connectors inspect revenuecat - RevenueCat connector manual

SYNOPSIS
  pm connectors inspect revenuecat
  pm connectors inspect revenuecat --json
  pm credentials add <name> --connector revenuecat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes RevenueCat v2 project configuration, customer, product, offering, subscription, purchase, paywall, virtual currency, integration, and metrics resources through the REST API.

ICON
  id: revenuecat
  asset: icons/revenuecat.svg
  source: official
  review_status: official_verified
  review_url: https://www.revenuecat.com/docs/api-v1

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_id
  base_url
  chart_name
  created_after
  customer_id
  entitlement_id
  invoice_id
  offering_id
  package_id
  paywall_id
  product_id
  project_id
  purchase_id
  starting_after
  store_purchase_identifier
  store_subscription_identifier
  subscription_id
  updated_after
  virtual_currency_code
  webhook_integration_id
  api_key (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    fields: created_at(string), icon_url(string), icon_url_large(string), id(string), name(string), object(string), stream(string)
  apps:
    primary key: id
    fields: created_at(string), id(string), name(string), object(string), project_id(string), stream(string), type(string)
  products:
    primary key: id
    fields: app_id(string), created_at(string), display_name(string), id(string), object(string), state(string), store_identifier(string), stream(string), type(string)
  offerings:
    primary key: id
    fields: created_at(string), description(string), display_name(string), id(string), identifier(string), lookup_key(string), object(string), project_id(string), state(string), stream(string)
  customers:
    primary key: id
    fields: app_user_id(string), created_at(string), first_seen_at(string), id(string), last_seen_at(string), object(string), project_id(string), stream(string), updated_at(string)
  app:
    primary key: id
    fields: created_at(string), id(string), name(string), object(string), project_id(string), stream(string), type(string)
  app_public_api_keys:
    primary key: id
    fields: app_id(string), created_at(string), environment(string), id(string), key(string), object(string), stream(string)
  app_store_kit_config:
    primary key: app_id
    fields: app_id(string), contents(string), object(string), stream(string)
  audit_logs:
    primary key: id
    fields: action_type(string), actor_identifier(string), actor_type(string), additional_data(string), id(string), object(string), occurred_at(string), project_id(string), stream(string), target_identifier(string), target_type(string)
  chart_data:
    primary key: chart_name
    fields: category(string), chart_name(string), description(string), display_name(string), display_type(string), object(string), resolution(string), stream(string), summary(string), values(string)
  chart_options:
    primary key: chart_name
    fields: chart_name(string), filters(string), object(string), resolutions(string), segments(string), stream(string), user_selectors(string)
  collaborators:
    primary key: id
    fields: accepted_at(string), email(string), has_mfa(string), id(string), name(string), object(string), role(string), stream(string)
  customer:
    primary key: id
    fields: active_entitlements(string), attributes(string), first_seen_at(string), id(string), last_seen_at(string), object(string), project_id(string), stream(string)
  customer_active_entitlements:
    primary key: entitlement_id
    fields: entitlement_id(string), expires_at(string), object(string), stream(string)
  customer_aliases:
    primary key: id
    fields: created_at(string), id(string), object(string), stream(string)
  customer_attributes:
    primary key: name
    fields: name(string), object(string), stream(string), updated_at(string), value(string)
  customer_center:
    primary key: customer_id
    fields: customer_center(string), customer_id(string), object(string), stream(string)
  customer_invoices:
    primary key: id
    fields: id(string), invoice_url(string), issued_at(string), line_items(string), object(string), paid_at(string), stream(string), total_amount(string)
  customer_purchases:
    primary key: id
    fields: customer_id(string), id(string), object(string), product_id(string), purchased_at(string), status(string), store_purchase_identifier(string), stream(string)
  customer_subscriptions:
    primary key: id
    fields: customer_id(string), id(string), object(string), product_id(string), starts_at(string), status(string), store_subscription_identifier(string), stream(string)
  customer_virtual_currencies:
    primary key: currency_code
    fields: balance(string), currency_code(string), description(string), name(string), object(string), stream(string)
  entitlements:
    primary key: id
    fields: created_at(string), display_name(string), id(string), lookup_key(string), object(string), products(string), project_id(string), state(string), stream(string)
  entitlement:
    primary key: id
    fields: created_at(string), display_name(string), id(string), lookup_key(string), object(string), products(string), project_id(string), state(string), stream(string)
  entitlement_products:
    primary key: id
    fields: app_id(string), created_at(string), display_name(string), id(string), object(string), state(string), store_identifier(string), stream(string), type(string)
  webhook_integrations:
    primary key: id
    fields: app_id(string), created_at(string), environment(string), event_types(string), id(string), name(string), object(string), project_id(string), stream(string), url(string)
  webhook_integration:
    primary key: id
    fields: app_id(string), created_at(string), environment(string), event_types(string), id(string), name(string), object(string), project_id(string), stream(string), url(string)
  overview_metrics:
    primary key: project_id
    fields: currency(string), metrics(string), object(string), project_id(string), stream(string)
  revenue_metric:
    primary key: project_id
    fields: currency(string), end_date(string), object(string), project_id(string), revenue_type(string), start_date(string), stream(string), value(string)
  offering:
    primary key: id
    fields: created_at(string), display_name(string), id(string), is_current(string), lookup_key(string), metadata(string), object(string), packages(string), paywall_id(string), project_id(string), stream(string)
  offering_packages:
    primary key: id
    fields: created_at(string), display_name(string), id(string), lookup_key(string), object(string), position(string), products(string), stream(string)
  package:
    primary key: id
    fields: created_at(string), display_name(string), id(string), lookup_key(string), object(string), position(string), products(string), stream(string)
  package_products:
    primary key: id
    fields: eligibility_criteria(string), id(string), object(string), product(string), stream(string)
  paywalls:
    primary key: id
    fields: automatically_scale_font_size(string), created_at(string), id(string), name(string), object(string), offering_id(string), published_at(string), stream(string)
  paywall:
    primary key: id
    fields: automatically_scale_font_size(string), components(string), created_at(string), id(string), name(string), object(string), offering_id(string), published_at(string), stream(string)
  product:
    primary key: id
    fields: app_id(string), created_at(string), display_name(string), id(string), object(string), state(string), store_identifier(string), stream(string), type(string)
  purchases:
    primary key: id
    fields: customer_id(string), id(string), object(string), product_id(string), purchased_at(string), status(string), store_purchase_identifier(string), stream(string)
  purchase:
    primary key: id
    fields: customer_id(string), id(string), object(string), product_id(string), purchased_at(string), status(string), store_purchase_identifier(string), stream(string)
  purchase_entitlements:
    primary key: id
    fields: created_at(string), display_name(string), id(string), lookup_key(string), object(string), project_id(string), state(string), stream(string)
  subscriptions:
    primary key: id
    fields: customer_id(string), id(string), object(string), product_id(string), starts_at(string), status(string), store_subscription_identifier(string), stream(string)
  subscription:
    primary key: id
    fields: customer_id(string), id(string), management_url(string), object(string), product_id(string), starts_at(string), status(string), store_subscription_identifier(string), stream(string)
  subscription_authenticated_management_url:
    primary key: subscription_id
    fields: management_url(string), object(string), stream(string), subscription_id(string)
  subscription_entitlements:
    primary key: id
    fields: created_at(string), display_name(string), id(string), lookup_key(string), object(string), project_id(string), state(string), stream(string)
  subscription_transactions:
    primary key: id
    fields: expiration_date(string), id(string), object(string), product_store_identifier(string), purchased_at(string), revenue_in_usd(string), stream(string)
  virtual_currencies:
    primary key: code
    fields: code(string), created_at(string), description(string), name(string), object(string), product_grants(string), project_id(string), state(string), stream(string)
  virtual_currency:
    primary key: code
    fields: code(string), created_at(string), description(string), name(string), object(string), product_grants(string), project_id(string), state(string), stream(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_project:
    endpoint: POST /projects
    required fields: name
    risk: creates a RevenueCat project
  create_app:
    endpoint: POST /projects/{{ config.project_id }}/apps
    required fields: name, type
    risk: creates a RevenueCat app in the configured project
  update_app:
    endpoint: POST /projects/{{ config.project_id }}/apps/{{ record.app_id }}
    required fields: app_id
    risk: updates an existing RevenueCat app
  delete_app:
    endpoint: DELETE /projects/{{ config.project_id }}/apps/{{ record.app_id }}
    required fields: app_id
    risk: permanently deletes a RevenueCat app and its configuration
  create_customer:
    endpoint: POST /projects/{{ config.project_id }}/customers
    required fields: id
    risk: creates a RevenueCat customer
  delete_customer:
    endpoint: DELETE /projects/{{ config.project_id }}/customers/{{ record.customer_id }}
    required fields: customer_id
    risk: permanently deletes a RevenueCat customer
  assign_customer_offering:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/assign_offering
    required fields: customer_id
    risk: assigns or clears a customer offering override
  grant_customer_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/grant_entitlement
    required fields: customer_id, entitlement_id
    risk: grants an entitlement to a customer
  restore_purchase_by_order_id:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/restore_purchase_by_order_id
    required fields: customer_id, order_id
    risk: restores a Google Play purchase by order id
  revoke_customer_granted_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/revoke_granted_entitlement
    required fields: customer_id, entitlement_id
    risk: revokes a granted customer entitlement
  transfer_customer_data:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/actions/transfer
    required fields: customer_id, destination_customer_id
    risk: transfers subscriptions and purchases to another customer
  set_customer_attributes:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/attributes
    required fields: customer_id, attributes
    risk: sets customer attributes
  create_virtual_currencies_transaction:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/virtual_currencies/transactions
    required fields: customer_id, currency_code, amount
    risk: creates a virtual currency transaction for a customer
  update_virtual_currencies_balance:
    endpoint: POST /projects/{{ config.project_id }}/customers/{{ record.customer_id }}/virtual_currencies/update_balance
    required fields: customer_id, currency_code, balance
    risk: updates a customer virtual currency balance without creating a transaction
  create_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/entitlements
    required fields: lookup_key, display_name
    risk: creates an entitlement
  update_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}
    required fields: entitlement_id
    risk: updates an entitlement
  delete_entitlement:
    endpoint: DELETE /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}
    required fields: entitlement_id
    risk: deletes an entitlement
  archive_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/archive
    required fields: entitlement_id
    risk: archives an entitlement
  unarchive_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/unarchive
    required fields: entitlement_id
    risk: unarchives an entitlement
  attach_products_to_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/attach_products
    required fields: entitlement_id, product_ids
    risk: attaches products to an entitlement
  detach_products_from_entitlement:
    endpoint: POST /projects/{{ config.project_id }}/entitlements/{{ record.entitlement_id }}/actions/detach_products
    required fields: entitlement_id, product_ids
    risk: detaches products from an entitlement
  create_webhook_integration:
    endpoint: POST /projects/{{ config.project_id }}/integrations/webhooks
    required fields: name, url
    risk: creates a webhook integration that sends events to the configured URL
  update_webhook_integration:
    endpoint: POST /projects/{{ config.project_id }}/integrations/webhooks/{{ record.webhook_integration_id }}
    required fields: webhook_integration_id
    risk: updates a webhook integration
  delete_webhook_integration:
    endpoint: DELETE /projects/{{ config.project_id }}/integrations/webhooks/{{ record.webhook_integration_id }}
    required fields: webhook_integration_id
    risk: deletes a webhook integration
  create_offering:
    endpoint: POST /projects/{{ config.project_id }}/offerings
    required fields: lookup_key, display_name
    risk: creates an offering
  update_offering:
    endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}
    required fields: offering_id
    risk: updates an offering
  delete_offering:
    endpoint: DELETE /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}
    required fields: offering_id
    risk: deletes an offering and attached packages
  archive_offering:
    endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}/actions/archive
    required fields: offering_id
    risk: archives an offering
  unarchive_offering:
    endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}/actions/unarchive
    required fields: offering_id
    risk: unarchives an offering
  create_package:
    endpoint: POST /projects/{{ config.project_id }}/offerings/{{ record.offering_id }}/packages
    required fields: offering_id, lookup_key, display_name
    risk: creates a package in an offering
  update_package:
    endpoint: POST /projects/{{ config.project_id }}/packages/{{ record.package_id }}
    required fields: package_id
    risk: updates a package
  delete_package:
    endpoint: DELETE /projects/{{ config.project_id }}/packages/{{ record.package_id }}
    required fields: package_id
    risk: deletes a package
  attach_products_to_package:
    endpoint: POST /projects/{{ config.project_id }}/packages/{{ record.package_id }}/actions/attach_products
    required fields: package_id, product_ids
    risk: attaches products to a package
  detach_products_from_package:
    endpoint: POST /projects/{{ config.project_id }}/packages/{{ record.package_id }}/actions/detach_products
    required fields: package_id, product_ids
    risk: detaches products from a package
  create_paywall:
    endpoint: POST /projects/{{ config.project_id }}/paywalls
    required fields: name, offering_id
    risk: creates a paywall
  update_paywall:
    endpoint: PATCH /projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}
    required fields: paywall_id
    risk: updates a paywall draft
  delete_paywall:
    endpoint: DELETE /projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}
    required fields: paywall_id
    risk: deletes a paywall
  create_paywall_version:
    endpoint: POST /projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}/versions
    required fields: paywall_id
    risk: creates a paywall version
  create_product:
    endpoint: POST /projects/{{ config.project_id }}/products
    required fields: store_identifier, type, app_id, display_name
    risk: creates a product
  update_product:
    endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}
    required fields: product_id
    risk: updates a product
  delete_product:
    endpoint: DELETE /projects/{{ config.project_id }}/products/{{ record.product_id }}
    required fields: product_id
    risk: deletes a product
  archive_product:
    endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}/actions/archive
    required fields: product_id
    risk: archives a product
  unarchive_product:
    endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}/actions/unarchive
    required fields: product_id
    risk: unarchives a product
  create_product_in_store:
    endpoint: POST /projects/{{ config.project_id }}/products/{{ record.product_id }}/create_in_store
    required fields: product_id
    risk: pushes a product to the configured store
  refund_purchase:
    endpoint: POST /projects/{{ config.project_id }}/purchases/{{ record.purchase_id }}/actions/refund
    required fields: purchase_id
    risk: refunds a Web Billing purchase
  cancel_subscription:
    endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/actions/cancel
    required fields: subscription_id
    risk: cancels an active Web Billing subscription
  extend_subscription:
    endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/actions/extend
    required fields: subscription_id
    risk: extends the current billing period of a subscription
  refund_subscription:
    endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/actions/refund
    required fields: subscription_id
    risk: refunds an active Web Billing subscription
  refund_subscription_transaction:
    endpoint: POST /projects/{{ config.project_id }}/subscriptions/{{ record.subscription_id }}/transactions/{{ record.transaction_id }}/actions/refund
    required fields: subscription_id, transaction_id
    risk: refunds a Play Store or Galaxy subscription transaction
  create_virtual_currency:
    endpoint: POST /projects/{{ config.project_id }}/virtual_currencies
    required fields: code, name
    risk: creates a virtual currency
  update_virtual_currency:
    endpoint: POST /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}
    required fields: virtual_currency_code
    risk: updates a virtual currency
  delete_virtual_currency:
    endpoint: DELETE /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}
    required fields: virtual_currency_code
    risk: deletes a virtual currency
  archive_virtual_currency:
    endpoint: POST /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}/actions/archive
    required fields: virtual_currency_code
    risk: archives a virtual currency
  unarchive_virtual_currency:
    endpoint: POST /projects/{{ config.project_id }}/virtual_currencies/{{ record.virtual_currency_code }}/actions/unarchive
    required fields: virtual_currency_code
    risk: unarchives a virtual currency

SECURITY
  read risk: external RevenueCat API reads of project, app, product, offering, customer, subscription, purchase, paywall, virtual currency, integration, metrics, and audit-log data
  write risk: external RevenueCat API mutations that create, update, archive, delete, refund, cancel, transfer, grant, revoke, or otherwise alter project and customer resources
  approval: reverse ETL writes require plan preview and approval token; destructive deletes/refunds/cancellations are flagged as destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect revenuecat

  # Inspect as structured JSON
  pm connectors inspect revenuecat --json

AGENT WORKFLOW
  - Run pm connectors inspect revenuecat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
