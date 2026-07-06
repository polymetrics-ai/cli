# Overview

Reads Pabbly customers, subscriptions, plans, and invoices, and writes customer/product/plan
mutations and subscription cancellations through the Pabbly Subscriptions Billing REST API.

Readable streams: `customers`, `subscriptions`, `plans`, `invoices`, `products`.

Write actions: `create_customer`, `update_customer`, `create_subscription`, `update_subscription`,
`create_coupon`, `create_payment_method`, `update_payment_method`, `create_addon`, `update_addon`,
`delete_addon`, `create_addon_category`, `update_addon_category`, `delete_addon_category`,
`create_license`, `update_license`, `cancel_subscription`, `create_product`, `update_product`,
`create_plan`, `update_plan`.

Service API documentation: https://www.pabbly.com/subscriptions/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://www.pabbly.com/subscriptions/api`; format `uri`;
  Pabbly Subscriptions Billing API base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-500).
- `password` (required, secret, string); Pabbly Subscriptions Billing account password. Sent via
  HTTP Basic auth; never logged.
- `username` (required, string); Pabbly Subscriptions Billing account username. Sent via HTTP Basic
  auth.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://www.pabbly.com/subscriptions/api`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next_page`.

Pagination by stream: cursor: `customers`, `subscriptions`, `plans`, `invoices`; none: `products`.

- `customers`: GET `/customers` - records path `data`; query `page`=`1`; `per_page` from template
  `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token
  from `next_page`.
- `subscriptions`: GET `/subscriptions` - records path `data`; query `page`=`1`; `per_page` from
  template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next
  token from `next_page`.
- `plans`: GET `/plans` - records path `data`; query `page`=`1`; `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `next_page`.
- `invoices`: GET `/invoices` - records path `data`; query `page`=`1`; `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `next_page`.
- `products`: GET `/products` - records path `data`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation; creates/updates billing customers, products, and plans, and
cancels live subscriptions (stops future billing).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customer` - kind `create`; body type `json`; required record fields
  `first_name`, `last_name`, `email_id`; accepted fields `billing_address`, `company_name`,
  `email_id`, `first_name`, `last_name`, `phone`, `shipping_address`, `website`; risk: external
  mutation; creates a billing customer record in Pabbly Subscriptions; approval required.
- `update_customer`: PUT `/customer/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `billing_address`, `company_name`,
  `first_name`, `id`, `last_name`, `phone`, `shipping_address`, `website`; risk: external mutation;
  updates a billing customer record in Pabbly Subscriptions; approval required.
- `create_subscription`: POST `/subscription/{{ record.customer_id }}` - kind `create`; body type
  `json`; path fields `customer_id`; required record fields `customer_id`, `plan_id`,
  `gateway_type`, `gateway_id`, `method_id`; accepted fields `customer_id`, `gateway_id`,
  `gateway_type`, `method_id`, `plan_id`, `redirect_to`; risk: external mutation; creates a live
  billing subscription for an existing customer and starts recurring charges; approval required.
- `update_subscription`: PUT `/subscription/{{ record.id }}/update` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `plan_id`, `customer_id`; accepted fields
  `activated_at_val`, `customer_id`, `id`, `payment_term`, `plan_id`; risk: external mutation;
  changes an existing subscription's plan/payment terms, altering future customer billing; approval
  required.
- `create_coupon`: POST `/coupon/{{ record.product_id }}` - kind `create`; body type `json`; path
  fields `product_id`; required record fields `coupon_name`, `coupon_code`, `discount`,
  `discount_type`, `redemption_type`; accepted fields `apply_to`, `associate_plans`, `coupon_code`,
  `coupon_name`, `discount`, `discount_type`, `maximum_redemption`, `plans_array`, `product_id`,
  `redemption_cycle`, `redemption_type`, `status`, `valid_upto`; risk: external mutation; creates a
  discount coupon that lowers future billing amounts for any customer who redeems it; approval
  required.
- `create_payment_method`: POST `/paymentmethod/{{ record.customer_id }}` - kind `create`; body type
  `json`; path fields `customer_id`; required record fields `customer_id`, `gateway_type`,
  `first_name`, `last_name`, `email`, `card_number`, `month`, `year`, `cvv`; accepted fields
  `card_number`, `city`, `country`, `customer_id`, `cvv`, `email`, `first_name`, `gateway_type`,
  `last_name`, `month`, `state`, `street`, `year`, `zip_code`; risk: external mutation; stores a new
  payment card on file for an existing customer via the connected payment gateway; approval
  required.
- `update_payment_method`: PUT `/paymentmethod/{{ record.customer_id }}` - kind `update`; body type
  `json`; path fields `customer_id`; required record fields `customer_id`, `card_number`, `month`,
  `year`, `cvv`, `mid`; accepted fields `card_number`, `customer_id`, `cvv`, `mid`, `month`, `year`;
  risk: external mutation; replaces the card an existing customer's future recurring billing charges
  against; approval required.
- `create_addon`: POST `/addon/{{ record.product_id }}` - kind `create`; body type `json`; path
  fields `product_id`; required record fields `product_id`, `name`, `code`, `price`,
  `billing_cycle`, `billing_period`; accepted fields `associate_plans`, `billing_cycle`,
  `billing_period`, `billing_period_num`, `category_array`, `code`, `description`, `name`,
  `plans_array`, `price`, `product_id`, `status`; risk: external mutation; creates a sellable add-on
  that customers can attach to a plan, adding to their bill; approval required.
- `update_addon`: PUT `/addon/{{ record.addon_id }}` - kind `update`; body type `json`; path fields
  `addon_id`; required record fields `addon_id`, `product_id`; accepted fields `addon_id`,
  `associate_plans`, `billing_cycle`, `billing_period`, `billing_period_num`, `category_array`,
  `code`, `description`, `name`, `plans_array`, `price`, `product_id`, `status`; risk: external
  mutation; updates an existing sellable add-on's price/billing terms, changing future charges for
  customers who have it attached; approval required.
- `delete_addon`: DELETE `/addon/{{ record.addon_id }}` - kind `delete`; body type `none`; path
  fields `addon_id`; required record fields `addon_id`; accepted fields `addon_id`; confirmation
  `destructive`; risk: external mutation; permanently deletes a sellable add-on; approval required.
- `create_addon_category`: POST `/addoncategory` - kind `create`; body type `json`; required record
  fields `category_name`, `product_id`; accepted fields `category_name`, `product_id`; risk:
  external mutation; creates an add-on category used to organize sellable add-ons in Pabbly
  Subscriptions; approval required.
- `update_addon_category`: PUT `/addoncategory/{{ record.category_id }}` - kind `update`; body type
  `json`; path fields `category_id`; required record fields `category_id`, `category_name`; accepted
  fields `category_id`, `category_name`; risk: external mutation; renames an existing add-on
  category; approval required.
- `delete_addon_category`: DELETE `/addoncategory/{{ record.category_id }}` - kind `delete`; body
  type `none`; path fields `category_id`; required record fields `category_id`; accepted fields
  `category_id`; confirmation `destructive`; risk: external mutation; permanently deletes an add-on
  category; approval required.
- `create_license`: POST `/products/{{ record.product_id }}/licenses` - kind `create`; body type
  `json`; path fields `product_id`; required record fields `product_id`, `name`, `plan_id`,
  `method`, `license_codes`, `status`; accepted fields `license_codes`, `method`, `name`, `plan_id`,
  `product_id`, `status`; risk: external mutation; creates a license-key pool for a product's plan
  in Pabbly Subscriptions; approval required.
- `update_license`: PUT `/products/{{ record.product_id }}/licenses/{{ record.license_id }}` - kind
  `update`; body type `json`; path fields `product_id`, `license_id`; required record fields
  `product_id`, `license_id`; accepted fields `license_codes`, `license_id`, `method`, `name`,
  `plan_id`, `product_id`, `status`; risk: external mutation; updates an existing license-key pool
  (may add/replace codes) in Pabbly Subscriptions; approval required.
- `cancel_subscription`: POST `/subscription/{{ record.id }}/cancel` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `cancel_at_end`; accepted fields
  `cancel_at_end`, `id`; confirmation `destructive`; risk: external mutation; cancels a live billing
  subscription (immediately or at term end) and stops future customer billing; approval required.
- `create_product`: POST `/product/create` - kind `create`; body type `json`; required record fields
  `product_name`; accepted fields `description`, `product_name`, `redirect_url`; risk: external
  mutation; creates a sellable product in Pabbly Subscriptions; approval required.
- `update_product`: PUT `/product/update/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `product_name`,
  `redirect_url`; risk: external mutation; updates a sellable product in Pabbly Subscriptions;
  approval required.
- `create_plan`: POST `/plan/create` - kind `create`; body type `json`; required record fields
  `product_id`, `plan_name`, `plan_code`, `billing_cycle`, `price`, `billing_period`,
  `billing_period_num`, `plan_active`; accepted fields `billing_cycle`, `billing_cycle_num`,
  `billing_period`, `billing_period_num`, `currency_code`, `plan_active`, `plan_code`,
  `plan_description`, `plan_name`, `price`, `product_id`, `redirect_url`, `setup_fee`,
  `trial_period`; risk: external mutation; creates a billing plan under a product in Pabbly
  Subscriptions; approval required.
- `update_plan`: PUT `/plan/update/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `product_id`, `currency_code`; accepted fields `billing_cycle`,
  `billing_cycle_num`, `billing_period`, `billing_period_num`, `currency_code`, `id`, `plan_active`,
  `plan_code`, `plan_description`, `plan_name`, `price`, `product_id`, `redirect_url`, `setup_fee`;
  risk: external mutation; updates a billing plan in Pabbly Subscriptions; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=9, non_data_endpoint=3, out_of_scope=2, requires_elevated_scope=10.
