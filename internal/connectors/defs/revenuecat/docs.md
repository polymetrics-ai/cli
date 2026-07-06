# Overview

Reads and writes RevenueCat v2 project configuration, customer, product, offering, subscription,
purchase, paywall, virtual currency, integration, and metrics resources through the REST API.

Readable streams: `projects`, `apps`, `products`, `offerings`, `customers`, `app`,
`app_public_api_keys`, `app_store_kit_config`, `audit_logs`, `chart_data`, `chart_options`,
`collaborators`, `customer`, `customer_active_entitlements`, `customer_aliases`,
`customer_attributes`, `customer_center`, `customer_invoices`, `customer_purchases`,
`customer_subscriptions`, `customer_virtual_currencies`, `entitlements`, `entitlement`,
`entitlement_products`, `webhook_integrations`, `webhook_integration`, `overview_metrics`,
`revenue_metric`, `offering`, `offering_packages`, `package`, `package_products`, `paywalls`,
`paywall`, `product`, `purchases`, `purchase`, `purchase_entitlements`, `subscriptions`,
`subscription`, `subscription_authenticated_management_url`, `subscription_entitlements`,
`subscription_transactions`, `virtual_currencies`, `virtual_currency`.

Write actions: `create_project`, `create_app`, `update_app`, `delete_app`, `create_customer`,
`delete_customer`, `assign_customer_offering`, `grant_customer_entitlement`,
`restore_purchase_by_order_id`, `revoke_customer_granted_entitlement`, `transfer_customer_data`,
`set_customer_attributes`, `create_virtual_currencies_transaction`,
`update_virtual_currencies_balance`, `create_entitlement`, `update_entitlement`,
`delete_entitlement`, `archive_entitlement`, `unarchive_entitlement`,
`attach_products_to_entitlement`, `detach_products_from_entitlement`, `create_webhook_integration`,
`update_webhook_integration`, `delete_webhook_integration`, `create_offering`, `update_offering`,
`delete_offering`, `archive_offering`, `unarchive_offering`, `create_package`, `update_package`,
`delete_package`, `attach_products_to_package`, `detach_products_from_package`, `create_paywall`,
`update_paywall`, `delete_paywall`, `create_paywall_version`, `create_product`, `update_product`,
`delete_product`, `archive_product`, `unarchive_product`, `create_product_in_store`,
`refund_purchase`, `cancel_subscription`, `extend_subscription`, `refund_subscription`,
`refund_subscription_transaction`, `create_virtual_currency`, `update_virtual_currency`,
`delete_virtual_currency`, `archive_virtual_currency`, `unarchive_virtual_currency`.

Service API documentation: https://www.revenuecat.com/docs/api-v2.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); RevenueCat secret API key, sent as an Authorization: Bearer
  <api_key> header. Never logged.
- `app_id` (optional, string); RevenueCat app id for app detail, API-key, and StoreKit config
  streams.
- `base_url` (optional, string); default `https://api.revenuecat.com/v2`; format `uri`; RevenueCat
  API base URL override for tests or proxies.
- `chart_name` (optional, string); RevenueCat chart name for chart data and chart option streams.
- `created_after` (optional, string); Optional passthrough filter: only return records created at or
  after this value.
- `customer_id` (optional, string); RevenueCat customer id for customer detail and customer-scoped
  streams.
- `entitlement_id` (optional, string); RevenueCat entitlement id for entitlement detail/product
  streams.
- `invoice_id` (optional, string); RevenueCat invoice id for invoice-related operations.
- `offering_id` (optional, string); RevenueCat offering id for offering detail/package streams.
- `package_id` (optional, string); RevenueCat package id for package detail/product streams.
- `paywall_id` (optional, string); RevenueCat paywall id for paywall detail streams.
- `product_id` (optional, string); RevenueCat product id for product detail streams.
- `project_id` (optional, string); RevenueCat project id.
- `purchase_id` (optional, string); RevenueCat purchase id for purchase detail streams.
- `starting_after` (optional, string); Optional passthrough filter: cursor id to resume a
  customers/apps/products/offerings listing after.
- `store_purchase_identifier` (optional, string); Store purchase identifier used by the purchases
  search endpoint.
- `store_subscription_identifier` (optional, string); Store subscription identifier used by the
  subscriptions search endpoint.
- `subscription_id` (optional, string); RevenueCat subscription id for subscription detail streams.
- `updated_after` (optional, string); Optional passthrough filter: only return records updated at or
  after this value.
- `virtual_currency_code` (optional, string); RevenueCat virtual currency code for virtual currency
  detail streams.
- `webhook_integration_id` (optional, string); RevenueCat webhook integration id for webhook detail
  streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.revenuecat.com/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `projects`: GET `/projects` - records path `items`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `starting_after` from template `{{
  config.starting_after }}`, omitted when absent; `updated_after` from template `{{
  config.updated_after }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `id`, `stream`; emits
  passthrough records.
- `apps`: GET `/projects/{{ config.project_id }}/apps` - records path `items`; query `created_after`
  from template `{{ config.created_after }}`, omitted when absent; `starting_after` from template
  `{{ config.starting_after }}`, omitted when absent; `updated_after` from template `{{
  config.updated_after }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `id`, `stream`; emits
  passthrough records.
- `products`: GET `/projects/{{ config.project_id }}/products` - records path `items`; query
  `created_after` from template `{{ config.created_after }}`, omitted when absent; `starting_after`
  from template `{{ config.starting_after }}`, omitted when absent; `updated_after` from template
  `{{ config.updated_after }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; computed output fields `id`, `stream`; emits
  passthrough records.
- `offerings`: GET `/projects/{{ config.project_id }}/offerings` - records path `items`; query
  `created_after` from template `{{ config.created_after }}`, omitted when absent; `starting_after`
  from template `{{ config.starting_after }}`, omitted when absent; `updated_after` from template
  `{{ config.updated_after }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; computed output fields `id`, `stream`; emits
  passthrough records.
- `customers`: GET `/projects/{{ config.project_id }}/customers` - records path `items`; query
  `created_after` from template `{{ config.created_after }}`, omitted when absent; `starting_after`
  from template `{{ config.starting_after }}`, omitted when absent; `updated_after` from template
  `{{ config.updated_after }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; computed output fields `id`, `stream`; emits
  passthrough records.
- `app`: GET `/projects/{{ config.project_id }}/apps/{{ config.app_id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; computed output fields `stream`; emits passthrough records.
- `app_public_api_keys`: GET `/projects/{{ config.project_id }}/apps/{{ config.app_id
  }}/public_api_keys` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `app_store_kit_config`: GET `/projects/{{ config.project_id }}/apps/{{ config.app_id
  }}/store_kit_config` - single-object response; records path `.`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `app_id`, `stream`; emits passthrough records.
- `audit_logs`: GET `/projects/{{ config.project_id }}/audit_logs` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`; emits passthrough records.
- `chart_data`: GET `/projects/{{ config.project_id }}/charts/{{ config.chart_name }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `chart_name`, `stream`;
  emits passthrough records.
- `chart_options`: GET `/projects/{{ config.project_id }}/charts/{{ config.chart_name }}/options` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `chart_name`, `stream`;
  emits passthrough records.
- `collaborators`: GET `/projects/{{ config.project_id }}/collaborators` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`; emits passthrough records.
- `customer`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `customer_active_entitlements`: GET `/projects/{{ config.project_id }}/customers/{{
  config.customer_id }}/active_entitlements` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `stream`; emits passthrough records.
- `customer_aliases`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id
  }}/aliases` - records path `items`; page-number pagination; page parameter `page`; size parameter
  `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough records.
- `customer_attributes`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id
  }}/attributes` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `customer_center`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id
  }}/customer_center` - single-object response; records path `.`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `customer_id`, `stream`; emits passthrough records.
- `customer_invoices`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id
  }}/invoices` - records path `items`; page-number pagination; page parameter `page`; size parameter
  `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough records.
- `customer_purchases`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id
  }}/purchases` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `customer_subscriptions`: GET `/projects/{{ config.project_id }}/customers/{{ config.customer_id
  }}/subscriptions` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `customer_virtual_currencies`: GET `/projects/{{ config.project_id }}/customers/{{
  config.customer_id }}/virtual_currencies` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `stream`; emits passthrough records.
- `entitlements`: GET `/projects/{{ config.project_id }}/entitlements` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`; emits passthrough records.
- `entitlement`: GET `/projects/{{ config.project_id }}/entitlements/{{ config.entitlement_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `entitlement_products`: GET `/projects/{{ config.project_id }}/entitlements/{{
  config.entitlement_id }}/products` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits
  passthrough records.
- `webhook_integrations`: GET `/projects/{{ config.project_id }}/integrations/webhooks` - records
  path `items`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1;
  page size 100; computed output fields `stream`; emits passthrough records.
- `webhook_integration`: GET `/projects/{{ config.project_id }}/integrations/webhooks/{{
  config.webhook_integration_id }}` - single-object response; records path `.`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `stream`; emits passthrough records.
- `overview_metrics`: GET `/projects/{{ config.project_id }}/metrics/overview` - single-object
  response; records path `.`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; computed output fields `project_id`, `stream`; emits passthrough
  records.
- `revenue_metric`: GET `/projects/{{ config.project_id }}/metrics/revenue` - single-object
  response; records path `.`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; computed output fields `project_id`, `stream`; emits passthrough
  records.
- `offering`: GET `/projects/{{ config.project_id }}/offerings/{{ config.offering_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `offering_packages`: GET `/projects/{{ config.project_id }}/offerings/{{ config.offering_id
  }}/packages` - records path `items`; page-number pagination; page parameter `page`; size parameter
  `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough records.
- `package`: GET `/projects/{{ config.project_id }}/packages/{{ config.package_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `package_products`: GET `/projects/{{ config.project_id }}/packages/{{ config.package_id
  }}/products` - records path `items`; page-number pagination; page parameter `page`; size parameter
  `limit`; starts at 1; page size 100; computed output fields `id`, `stream`; emits passthrough
  records.
- `paywalls`: GET `/projects/{{ config.project_id }}/paywalls` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `stream`; emits passthrough records.
- `paywall`: GET `/projects/{{ config.project_id }}/paywalls/{{ config.paywall_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `product`: GET `/projects/{{ config.project_id }}/products/{{ config.product_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `purchases`: GET `/projects/{{ config.project_id }}/purchases` - records path `items`; query
  `store_purchase_identifier` from template `{{ config.store_purchase_identifier }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; computed output fields `stream`; emits passthrough records.
- `purchase`: GET `/projects/{{ config.project_id }}/purchases/{{ config.purchase_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `purchase_entitlements`: GET `/projects/{{ config.project_id }}/purchases/{{ config.purchase_id
  }}/entitlements` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `subscriptions`: GET `/projects/{{ config.project_id }}/subscriptions` - records path `items`;
  query `store_subscription_identifier` from template `{{ config.store_subscription_identifier }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 1; page size 100; computed output fields `stream`; emits passthrough records.
- `subscription`: GET `/projects/{{ config.project_id }}/subscriptions/{{ config.subscription_id }}`
  - single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `stream`; emits passthrough
  records.
- `subscription_authenticated_management_url`: GET `/projects/{{ config.project_id
  }}/subscriptions/{{ config.subscription_id }}/authenticated_management_url` - single-object
  response; records path `.`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; computed output fields `stream`, `subscription_id`; emits passthrough
  records.
- `subscription_entitlements`: GET `/projects/{{ config.project_id }}/subscriptions/{{
  config.subscription_id }}/entitlements` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `stream`; emits passthrough records.
- `subscription_transactions`: GET `/projects/{{ config.project_id }}/subscriptions/{{
  config.subscription_id }}/transactions` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `stream`; emits passthrough records.
- `virtual_currencies`: GET `/projects/{{ config.project_id }}/virtual_currencies` - records path
  `items`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; computed output fields `stream`; emits passthrough records.
- `virtual_currency`: GET `/projects/{{ config.project_id }}/virtual_currencies/{{
  config.virtual_currency_code }}` - single-object response; records path `.`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `stream`; emits passthrough records.

## Write actions & risks

Overall write risk: external RevenueCat API mutations that create, update, archive, delete, refund,
cancel, transfer, grant, revoke, or otherwise alter project and customer resources.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: creates a RevenueCat project.
- `create_app`: POST `/projects/{{ config.project_id }}/apps` - kind `create`; body type `json`;
  required record fields `name`, `type`; accepted fields `name`, `type`; risk: creates a RevenueCat
  app in the configured project.
- `update_app`: POST `/projects/{{ config.project_id }}/apps/{{ record.app_id }}` - kind `update`;
  body type `json`; path fields `app_id`; required record fields `app_id`; accepted fields `app_id`;
  risk: updates an existing RevenueCat app.
- `delete_app`: DELETE `/projects/{{ config.project_id }}/apps/{{ record.app_id }}` - kind `delete`;
  body type `none`; path fields `app_id`; required record fields `app_id`; accepted fields `app_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: permanently
  deletes a RevenueCat app and its configuration.
- `create_customer`: POST `/projects/{{ config.project_id }}/customers` - kind `create`; body type
  `json`; required record fields `id`; accepted fields `id`; risk: creates a RevenueCat customer.
- `delete_customer`: DELETE `/projects/{{ config.project_id }}/customers/{{ record.customer_id }}` -
  kind `delete`; body type `none`; path fields `customer_id`; required record fields `customer_id`;
  accepted fields `customer_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: permanently deletes a RevenueCat customer.
- `assign_customer_offering`: POST `/projects/{{ config.project_id }}/customers/{{
  record.customer_id }}/actions/assign_offering` - kind `update`; body type `json`; path fields
  `customer_id`; required record fields `customer_id`; accepted fields `customer_id`; risk: assigns
  or clears a customer offering override.
- `grant_customer_entitlement`: POST `/projects/{{ config.project_id }}/customers/{{
  record.customer_id }}/actions/grant_entitlement` - kind `create`; body type `json`; path fields
  `customer_id`; required record fields `customer_id`, `entitlement_id`; accepted fields
  `customer_id`, `entitlement_id`; risk: grants an entitlement to a customer.
- `restore_purchase_by_order_id`: POST `/projects/{{ config.project_id }}/customers/{{
  record.customer_id }}/actions/restore_purchase_by_order_id` - kind `create`; body type `json`;
  path fields `customer_id`; required record fields `customer_id`, `order_id`; accepted fields
  `customer_id`, `order_id`; risk: restores a Google Play purchase by order id.
- `revoke_customer_granted_entitlement`: POST `/projects/{{ config.project_id }}/customers/{{
  record.customer_id }}/actions/revoke_granted_entitlement` - kind `delete`; body type `none`; path
  fields `customer_id`; required record fields `customer_id`, `entitlement_id`; accepted fields
  `customer_id`, `entitlement_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: revokes a granted customer entitlement.
- `transfer_customer_data`: POST `/projects/{{ config.project_id }}/customers/{{ record.customer_id
  }}/actions/transfer` - kind `update`; body type `json`; path fields `customer_id`; required record
  fields `customer_id`, `destination_customer_id`; accepted fields `customer_id`,
  `destination_customer_id`; confirmation `destructive`; risk: transfers subscriptions and purchases
  to another customer.
- `set_customer_attributes`: POST `/projects/{{ config.project_id }}/customers/{{ record.customer_id
  }}/attributes` - kind `update`; body type `json`; path fields `customer_id`; required record
  fields `customer_id`, `attributes`; accepted fields `attributes`, `customer_id`; risk: sets
  customer attributes.
- `create_virtual_currencies_transaction`: POST `/projects/{{ config.project_id }}/customers/{{
  record.customer_id }}/virtual_currencies/transactions` - kind `create`; body type `json`; path
  fields `customer_id`; required record fields `customer_id`, `currency_code`, `amount`; accepted
  fields `amount`, `currency_code`, `customer_id`; risk: creates a virtual currency transaction for
  a customer.
- `update_virtual_currencies_balance`: POST `/projects/{{ config.project_id }}/customers/{{
  record.customer_id }}/virtual_currencies/update_balance` - kind `update`; body type `json`; path
  fields `customer_id`; required record fields `customer_id`, `currency_code`, `balance`; accepted
  fields `balance`, `currency_code`, `customer_id`; risk: updates a customer virtual currency
  balance without creating a transaction.
- `create_entitlement`: POST `/projects/{{ config.project_id }}/entitlements` - kind `create`; body
  type `json`; required record fields `lookup_key`, `display_name`; accepted fields `display_name`,
  `lookup_key`; risk: creates an entitlement.
- `update_entitlement`: POST `/projects/{{ config.project_id }}/entitlements/{{
  record.entitlement_id }}` - kind `update`; body type `json`; path fields `entitlement_id`;
  required record fields `entitlement_id`; accepted fields `entitlement_id`; risk: updates an
  entitlement.
- `delete_entitlement`: DELETE `/projects/{{ config.project_id }}/entitlements/{{
  record.entitlement_id }}` - kind `delete`; body type `none`; path fields `entitlement_id`;
  required record fields `entitlement_id`; accepted fields `entitlement_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: deletes an entitlement.
- `archive_entitlement`: POST `/projects/{{ config.project_id }}/entitlements/{{
  record.entitlement_id }}/actions/archive` - kind `update`; body type `json`; path fields
  `entitlement_id`; required record fields `entitlement_id`; accepted fields `entitlement_id`; risk:
  archives an entitlement.
- `unarchive_entitlement`: POST `/projects/{{ config.project_id }}/entitlements/{{
  record.entitlement_id }}/actions/unarchive` - kind `update`; body type `json`; path fields
  `entitlement_id`; required record fields `entitlement_id`; accepted fields `entitlement_id`; risk:
  unarchives an entitlement.
- `attach_products_to_entitlement`: POST `/projects/{{ config.project_id }}/entitlements/{{
  record.entitlement_id }}/actions/attach_products` - kind `update`; body type `json`; path fields
  `entitlement_id`; required record fields `entitlement_id`, `product_ids`; accepted fields
  `entitlement_id`, `product_ids`; risk: attaches products to an entitlement.
- `detach_products_from_entitlement`: POST `/projects/{{ config.project_id }}/entitlements/{{
  record.entitlement_id }}/actions/detach_products` - kind `update`; body type `json`; path fields
  `entitlement_id`; required record fields `entitlement_id`, `product_ids`; accepted fields
  `entitlement_id`, `product_ids`; risk: detaches products from an entitlement.
- `create_webhook_integration`: POST `/projects/{{ config.project_id }}/integrations/webhooks` -
  kind `create`; body type `json`; required record fields `name`, `url`; accepted fields `name`,
  `url`; risk: creates a webhook integration that sends events to the configured URL.
- `update_webhook_integration`: POST `/projects/{{ config.project_id }}/integrations/webhooks/{{
  record.webhook_integration_id }}` - kind `update`; body type `json`; path fields
  `webhook_integration_id`; required record fields `webhook_integration_id`; accepted fields
  `webhook_integration_id`; risk: updates a webhook integration.
- `delete_webhook_integration`: DELETE `/projects/{{ config.project_id }}/integrations/webhooks/{{
  record.webhook_integration_id }}` - kind `delete`; body type `none`; path fields
  `webhook_integration_id`; required record fields `webhook_integration_id`; accepted fields
  `webhook_integration_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a webhook integration.
- `create_offering`: POST `/projects/{{ config.project_id }}/offerings` - kind `create`; body type
  `json`; required record fields `lookup_key`, `display_name`; accepted fields `display_name`,
  `lookup_key`; risk: creates an offering.
- `update_offering`: POST `/projects/{{ config.project_id }}/offerings/{{ record.offering_id }}` -
  kind `update`; body type `json`; path fields `offering_id`; required record fields `offering_id`;
  accepted fields `offering_id`; risk: updates an offering.
- `delete_offering`: DELETE `/projects/{{ config.project_id }}/offerings/{{ record.offering_id }}` -
  kind `delete`; body type `none`; path fields `offering_id`; required record fields `offering_id`;
  accepted fields `offering_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes an offering and attached packages.
- `archive_offering`: POST `/projects/{{ config.project_id }}/offerings/{{ record.offering_id
  }}/actions/archive` - kind `update`; body type `json`; path fields `offering_id`; required record
  fields `offering_id`; accepted fields `offering_id`; risk: archives an offering.
- `unarchive_offering`: POST `/projects/{{ config.project_id }}/offerings/{{ record.offering_id
  }}/actions/unarchive` - kind `update`; body type `json`; path fields `offering_id`; required
  record fields `offering_id`; accepted fields `offering_id`; risk: unarchives an offering.
- `create_package`: POST `/projects/{{ config.project_id }}/offerings/{{ record.offering_id
  }}/packages` - kind `create`; body type `json`; path fields `offering_id`; required record fields
  `offering_id`, `lookup_key`, `display_name`; accepted fields `display_name`, `lookup_key`,
  `offering_id`; risk: creates a package in an offering.
- `update_package`: POST `/projects/{{ config.project_id }}/packages/{{ record.package_id }}` - kind
  `update`; body type `json`; path fields `package_id`; required record fields `package_id`;
  accepted fields `package_id`; risk: updates a package.
- `delete_package`: DELETE `/projects/{{ config.project_id }}/packages/{{ record.package_id }}` -
  kind `delete`; body type `none`; path fields `package_id`; required record fields `package_id`;
  accepted fields `package_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a package.
- `attach_products_to_package`: POST `/projects/{{ config.project_id }}/packages/{{
  record.package_id }}/actions/attach_products` - kind `update`; body type `json`; path fields
  `package_id`; required record fields `package_id`, `product_ids`; accepted fields `package_id`,
  `product_ids`; risk: attaches products to a package.
- `detach_products_from_package`: POST `/projects/{{ config.project_id }}/packages/{{
  record.package_id }}/actions/detach_products` - kind `update`; body type `json`; path fields
  `package_id`; required record fields `package_id`, `product_ids`; accepted fields `package_id`,
  `product_ids`; risk: detaches products from a package.
- `create_paywall`: POST `/projects/{{ config.project_id }}/paywalls` - kind `create`; body type
  `json`; required record fields `name`, `offering_id`; accepted fields `name`, `offering_id`; risk:
  creates a paywall.
- `update_paywall`: PATCH `/projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}` -
  kind `update`; body type `json`; path fields `paywall_id`; required record fields `paywall_id`;
  accepted fields `paywall_id`; risk: updates a paywall draft.
- `delete_paywall`: DELETE `/projects/{{ config.project_id }}/paywalls/{{ record.paywall_id }}` -
  kind `delete`; body type `none`; path fields `paywall_id`; required record fields `paywall_id`;
  accepted fields `paywall_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a paywall.
- `create_paywall_version`: POST `/projects/{{ config.project_id }}/paywalls/{{ record.paywall_id
  }}/versions` - kind `create`; body type `json`; path fields `paywall_id`; required record fields
  `paywall_id`; accepted fields `paywall_id`; risk: creates a paywall version.
- `create_product`: POST `/projects/{{ config.project_id }}/products` - kind `create`; body type
  `json`; required record fields `store_identifier`, `type`, `app_id`, `display_name`; accepted
  fields `app_id`, `display_name`, `store_identifier`, `type`; risk: creates a product.
- `update_product`: POST `/projects/{{ config.project_id }}/products/{{ record.product_id }}` - kind
  `update`; body type `json`; path fields `product_id`; required record fields `product_id`;
  accepted fields `product_id`; risk: updates a product.
- `delete_product`: DELETE `/projects/{{ config.project_id }}/products/{{ record.product_id }}` -
  kind `delete`; body type `none`; path fields `product_id`; required record fields `product_id`;
  accepted fields `product_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a product.
- `archive_product`: POST `/projects/{{ config.project_id }}/products/{{ record.product_id
  }}/actions/archive` - kind `update`; body type `json`; path fields `product_id`; required record
  fields `product_id`; accepted fields `product_id`; risk: archives a product.
- `unarchive_product`: POST `/projects/{{ config.project_id }}/products/{{ record.product_id
  }}/actions/unarchive` - kind `update`; body type `json`; path fields `product_id`; required record
  fields `product_id`; accepted fields `product_id`; risk: unarchives a product.
- `create_product_in_store`: POST `/projects/{{ config.project_id }}/products/{{ record.product_id
  }}/create_in_store` - kind `create`; body type `json`; path fields `product_id`; required record
  fields `product_id`; accepted fields `product_id`; risk: pushes a product to the configured store.
- `refund_purchase`: POST `/projects/{{ config.project_id }}/purchases/{{ record.purchase_id
  }}/actions/refund` - kind `delete`; body type `none`; path fields `purchase_id`; required record
  fields `purchase_id`; accepted fields `purchase_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: refunds a Web Billing purchase.
- `cancel_subscription`: POST `/projects/{{ config.project_id }}/subscriptions/{{
  record.subscription_id }}/actions/cancel` - kind `delete`; body type `none`; path fields
  `subscription_id`; required record fields `subscription_id`; accepted fields `subscription_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: cancels an
  active Web Billing subscription.
- `extend_subscription`: POST `/projects/{{ config.project_id }}/subscriptions/{{
  record.subscription_id }}/actions/extend` - kind `update`; body type `json`; path fields
  `subscription_id`; required record fields `subscription_id`; accepted fields `subscription_id`;
  risk: extends the current billing period of a subscription.
- `refund_subscription`: POST `/projects/{{ config.project_id }}/subscriptions/{{
  record.subscription_id }}/actions/refund` - kind `delete`; body type `none`; path fields
  `subscription_id`; required record fields `subscription_id`; accepted fields `subscription_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: refunds an
  active Web Billing subscription.
- `refund_subscription_transaction`: POST `/projects/{{ config.project_id }}/subscriptions/{{
  record.subscription_id }}/transactions/{{ record.transaction_id }}/actions/refund` - kind
  `delete`; body type `none`; path fields `subscription_id`, `transaction_id`; required record
  fields `subscription_id`, `transaction_id`; accepted fields `subscription_id`, `transaction_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: refunds a
  Play Store or Galaxy subscription transaction.
- `create_virtual_currency`: POST `/projects/{{ config.project_id }}/virtual_currencies` - kind
  `create`; body type `json`; required record fields `code`, `name`; accepted fields `code`, `name`;
  risk: creates a virtual currency.
- `update_virtual_currency`: POST `/projects/{{ config.project_id }}/virtual_currencies/{{
  record.virtual_currency_code }}` - kind `update`; body type `json`; path fields
  `virtual_currency_code`; required record fields `virtual_currency_code`; accepted fields
  `virtual_currency_code`; risk: updates a virtual currency.
- `delete_virtual_currency`: DELETE `/projects/{{ config.project_id }}/virtual_currencies/{{
  record.virtual_currency_code }}` - kind `delete`; body type `none`; path fields
  `virtual_currency_code`; required record fields `virtual_currency_code`; accepted fields
  `virtual_currency_code`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a virtual currency.
- `archive_virtual_currency`: POST `/projects/{{ config.project_id }}/virtual_currencies/{{
  record.virtual_currency_code }}/actions/archive` - kind `update`; body type `json`; path fields
  `virtual_currency_code`; required record fields `virtual_currency_code`; accepted fields
  `virtual_currency_code`; risk: archives a virtual currency.
- `unarchive_virtual_currency`: POST `/projects/{{ config.project_id }}/virtual_currencies/{{
  record.virtual_currency_code }}/actions/unarchive` - kind `update`; body type `json`; path fields
  `virtual_currency_code`; required record fields `virtual_currency_code`; accepted fields
  `virtual_currency_code`; risk: unarchives a virtual currency.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 45 stream-backed endpoint group(s), 54 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2.
