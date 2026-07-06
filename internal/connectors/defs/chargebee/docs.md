# Overview

Reads and writes Chargebee subscription billing data (customers, subscriptions, invoices, plans,
items, item prices, coupons, credit notes, transactions, orders, quotes, payment sources, events,
and more) through the Chargebee v2 REST API.

Readable streams: `customers`, `subscriptions`, `invoices`, `plans`, `items`, `item_prices`,
`item_families`, `coupons`, `coupon_codes`, `coupon_sets`, `credit_notes`, `transactions`, `orders`,
`quotes`, `payment_sources`, `events`, `hosted_pages`, `virtual_bank_accounts`, `unbilled_charges`,
`ramps`, `gifts`, `alerts`, `comments`, `promotional_credits`, `features`, `entitlements`,
`differential_prices`, `price_variants`, `products`, `webhook_endpoints`, `ledger_operations`,
`ledger_account_balances`.

Write actions: `create_customer`, `update_customer`, `delete_customer`, `create_item`,
`update_item`, `delete_item`, `create_item_price`, `update_item_price`, `delete_item_price`,
`create_item_family`, `update_item_family`, `delete_item_family`, `create_subscription`,
`update_subscription`, `cancel_subscription`, `create_credit_note`, `void_credit_note`,
`create_coupon`, `update_coupon`, `delete_coupon`, `create_order`, `update_order`, `cancel_order`,
`void_invoice`, `collect_payment_for_invoice`, `create_webhook_endpoint`, `update_webhook_endpoint`,
`delete_webhook_endpoint`, `create_comment`, `delete_comment`, `add_promotional_credit`,
`deduct_promotional_credit`, `create_virtual_bank_account`, `delete_virtual_bank_account`,
`create_card_payment_source`, `delete_payment_source`.

Service API documentation: https://apidocs.chargebee.com/docs/api.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Chargebee API base URL, e.g.
  https://{site}.chargebee.com/api/v2.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `site_api_key` (required, secret, string); Chargebee site API key. Sent as the HTTP Basic username
  with an empty password; never logged.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects updated at
  or after this time are read.

Secret fields are redacted in logs and write previews: `site_api_key`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.site_api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `offset`; next token from `next_offset`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers` - records path `list`; query `limit`=`100`; `sort_by[asc]` from
  template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `auto_collection`, `company`, `created_at`,
  `deleted`, `email`, `first_name`, `id`, `last_name`, `net_term_days`, `phone`, `taxability`,
  `updated_at`.
- `subscriptions`: GET `/subscriptions` - records path `list`; query `limit`=`100`; `sort_by[asc]`
  from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `created_at`, `currency_code`, `current_term_end`,
  `current_term_start`, `customer_id`, `deleted`, `id`, `plan_amount`, `plan_id`, `plan_quantity`,
  `started_at`, `status`, `updated_at`.
- `invoices`: GET `/invoices` - records path `list`; query `limit`=`100`; `sort_by[asc]` from
  template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `amount_due`, `amount_paid`, `currency_code`,
  `customer_id`, `date`, `deleted`, `due_date`, `id`, `paid_at`, `status`, `subscription_id`,
  `total`, `updated_at`.
- `plans`: GET `/plans` - records path `list`; query `limit`=`100`; `sort_by[asc]` from template `{{
  incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `created_at`, `currency_code`, `id`, `invoice_name`, `name`, `period`,
  `period_unit`, `price`, `status`, `updated_at`.
- `items`: GET `/items` - records path `list`; query `limit`=`100`; `sort_by[asc]` from template `{{
  incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `created_at`, `enabled_for_checkout`, `id`, `is_shippable`,
  `item_family_id`, `name`, `status`, `type`, `updated_at`.
- `item_prices`: GET `/item_prices` - records path `list`; query `limit`=`100`; `sort_by[asc]` from
  template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `created_at`, `currency_code`, `deleted`,
  `free_quantity`, `id`, `is_taxable`, `item_family_id`, `item_id`, `item_type`, `name`, `period`,
  `period_unit`, `price`, `pricing_model`, `status`, and 1 more.
- `item_families`: GET `/item_families` - records path `list`; query `limit`=`100`; `sort_by[asc]`
  from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `created_at`, `deleted`, `description`, `id`,
  `name`, `status`, `updated_at`.
- `coupons`: GET `/coupons` - records path `list`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `apply_on`, `created_at`, `currency_code`, `deleted`, `discount_amount`,
  `discount_percentage`, `discount_type`, `duration_type`, `id`, `name`, `redemptions`, `status`,
  `updated_at`, `valid_till`.
- `coupon_codes`: GET `/coupon_codes` - records path `list`; query `limit`=`100`; cursor pagination;
  cursor parameter `offset`; next token from `next_offset`; computed output fields `code`,
  `coupon_id`, `coupon_set_id`, `coupon_set_name`, `status`.
- `coupon_sets`: GET `/coupon_sets` - records path `list`; query `limit`=`100`; cursor pagination;
  cursor parameter `offset`; next token from `next_offset`; computed output fields `archived_count`,
  `coupon_id`, `id`, `name`, `redeemed_count`, `total_count`.
- `credit_notes`: GET `/credit_notes` - records path `list`; query `limit`=`100`; cursor pagination;
  cursor parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `amount_allocated`, `amount_available`, `amount_refunded`, `currency_code`,
  `customer_id`, `date`, `deleted`, `id`, `reference_invoice_id`, `status`, `subscription_id`,
  `total`, `type`, `updated_at`, `voided_at`.
- `transactions`: GET `/transactions` - records path `list`; query `limit`=`100`; `sort_by[asc]`
  from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `amount`, `currency_code`, `customer_id`, `date`,
  `deleted`, `gateway`, `id`, `payment_method`, `payment_source_id`, `status`, `subscription_id`,
  `type`, `updated_at`.
- `orders`: GET `/orders` - records path `list`; query `limit`=`100`; `sort_by[asc]` from template
  `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `created_at`, `currency_code`, `customer_id`, `deleted`, `document_number`,
  `id`, `invoice_id`, `order_type`, `price_type`, `status`, `subscription_id`, `total`,
  `updated_at`.
- `quotes`: GET `/quotes` - records path `list`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `currency_code`, `customer_id`, `date`, `id`, `invoice_id`, `name`,
  `operation_type`, `price_type`, `status`, `subscription_id`, `total`, `updated_at`, `valid_till`.
- `payment_sources`: GET `/payment_sources` - records path `list`; query `limit`=`100`;
  `sort_by[asc]` from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when
  absent; cursor pagination; cursor parameter `offset`; next token from `next_offset`; incremental
  cursor `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`; computed output fields `created_at`, `customer_id`, `deleted`,
  `gateway`, `gateway_account_id`, `id`, `reference_id`, `status`, `type`, `updated_at`.
- `events`: GET `/events` - records path `list`; query `limit`=`100`; `sort_by[asc]` from template
  `{{ incremental.lower_bound | const:occurred_at }}`, omitted when absent; cursor pagination;
  cursor parameter `offset`; next token from `next_offset`; incremental cursor `occurred_at`; sent
  as `occurred_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from
  `start_date`; computed output fields `api_version`, `event_type`, `id`, `occurred_at`, `source`.
- `hosted_pages`: GET `/hosted_pages` - records path `list`; query `limit`=`100`; cursor pagination;
  cursor parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `created_at`, `expires_at`, `id`, `state`, `type`, `updated_at`, `url`.
- `virtual_bank_accounts`: GET `/virtual_bank_accounts` - records path `list`; query `limit`=`100`;
  cursor pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `account_number`, `bank_name`, `created_at`,
  `customer_id`, `deleted`, `email`, `gateway`, `gateway_account_id`, `id`, `updated_at`.
- `unbilled_charges`: GET `/unbilled_charges` - records path `list`; query `limit`=`100`; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; computed output fields
  `amount`, `currency_code`, `customer_id`, `date_from`, `date_to`, `entity_id`, `entity_type`,
  `id`, `is_voided`, `subscription_id`.
- `ramps`: GET `/ramps` - records path `list`; query `limit`=`100`; `sort_by[asc]` from template `{{
  incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; incremental cursor `updated_at`; sent as
  `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower bound from `start_date`;
  computed output fields `created_at`, `deleted`, `description`, `effective_from`, `id`, `status`,
  `subscription_id`, `updated_at`.
- `gifts`: GET `/gifts` - records path `list`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; computed output fields `auto_claim`,
  `claim_expiry_date`, `id`, `no_expiry`, `scheduled_at`, `status`, `updated_at`.
- `alerts`: GET `/alerts` - records path `list`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; computed output fields `created_at`,
  `description`, `id`, `metered_feature_id`, `name`, `status`, `subscription_id`, `type`,
  `updated_at`.
- `comments`: GET `/comments` - records path `list`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; computed output fields `added_by`,
  `created_at`, `entity_id`, `entity_type`, `id`, `notes`, `type`.
- `promotional_credits`: GET `/promotional_credits` - records path `list`; query `limit`=`100`;
  cursor pagination; cursor parameter `offset`; next token from `next_offset`; computed output
  fields `amount`, `closing_balance`, `created_at`, `credit_type`, `currency_code`, `customer_id`,
  `description`, `id`, `type`.
- `features`: GET `/features` - records path `list`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next_offset`; computed output fields `created_at`,
  `description`, `id`, `name`, `status`, `type`, `unit`, `updated_at`.
- `entitlements`: GET `/entitlements` - records path `list`; query `limit`=`100`; cursor pagination;
  cursor parameter `offset`; next token from `next_offset`; computed output fields `entity_id`,
  `entity_type`, `feature_id`, `feature_name`, `id`, `name`, `value`.
- `differential_prices`: GET `/differential_prices` - records path `list`; query `limit`=`100`;
  cursor pagination; cursor parameter `offset`; next token from `next_offset`; computed output
  fields `created_at`, `currency_code`, `deleted`, `id`, `item_price_id`, `parent_item_id`, `price`,
  `status`, `updated_at`.
- `price_variants`: GET `/price_variants` - records path `list`; query `limit`=`100`; `sort_by[asc]`
  from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `created_at`, `deleted`, `description`, `id`,
  `name`, `status`, `updated_at`, `variant_group`.
- `products`: GET `/products` - records path `list`; query `limit`=`100`; `sort_by[asc]` from
  template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; incremental cursor
  `updated_at`; sent as `updated_at[after]`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`; computed output fields `created_at`, `deleted`, `description`,
  `external_name`, `has_variant`, `id`, `name`, `shippable`, `sku`, `status`, `updated_at`.
- `webhook_endpoints`: GET `/webhook_endpoints` - records path `list`; query `limit`=`100`; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; computed output fields
  `api_version`, `disabled`, `id`, `name`, `primary_url`, `url`.
- `ledger_operations`: GET `/ledger_operations` - records path `list`; query `limit`=`100`; cursor
  pagination; cursor parameter `offset`; next token from `next_offset`; computed output fields
  `amount`, `created_at`, `end_balance`, `id`, `modified_at`, `start_balance`, `subscription_id`,
  `type`, `unit_id`, `unit_type`.
- `ledger_account_balances`: GET `/ledger_account_balances` - records path `list`; query
  `limit`=`100`; cursor pagination; cursor parameter `offset`; next token from `next_offset`;
  computed output fields `modified_at`, `subscription_id`, `unit_id`, `unit_type`.

## Write actions & risks

Overall write risk: external mutation of Chargebee billing data (customers, subscriptions, invoices,
credit notes, orders, coupons, payment sources); several actions have direct financial/billing side
effects and require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `form`; accepted fields `company`,
  `email`, `first_name`, `last_name`, `phone`; risk: external mutation; approval required.
- `update_customer`: POST `/customers/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `company`, `email`, `first_name`, `id`,
  `last_name`, `phone`; risk: external mutation; approval required.
- `delete_customer`: POST `/customers/{{ record.id }}/delete` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: irreversible external deletion; approval required.
- `create_item`: POST `/items` - kind `create`; body type `form`; required record fields `id`,
  `name`, `type`, `item_family_id`; accepted fields `description`, `id`, `item_family_id`, `name`,
  `type`, `unit`; risk: external mutation; approval required.
- `update_item`: POST `/items/{{ record.id }}` - kind `update`; body type `form`; path fields `id`;
  required record fields `id`; accepted fields `description`, `id`, `name`, `status`; risk: external
  mutation; approval required.
- `delete_item`: POST `/items/{{ record.id }}/delete` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversible external deletion; approval required.
- `create_item_price`: POST `/item_prices` - kind `create`; body type `form`; required record fields
  `id`, `item_id`, `name`; accepted fields `currency_code`, `description`, `id`, `item_id`, `name`,
  `period`, `period_unit`, `price`; risk: external mutation; approval required.
- `update_item_price`: POST `/item_prices/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`, `price`,
  `status`; risk: external mutation; approval required.
- `delete_item_price`: POST `/item_prices/{{ record.id }}/delete` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: irreversible external deletion; approval required.
- `create_item_family`: POST `/item_families` - kind `create`; body type `form`; required record
  fields `id`, `name`; accepted fields `description`, `id`, `name`; risk: external mutation;
  approval required.
- `update_item_family`: POST `/item_families/{{ record.id }}` - kind `update`; body type `form`;
  path fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_item_family`: POST `/item_families/{{ record.id }}/delete` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversible external deletion; approval required.
- `create_subscription`: POST `/customers/{{ record.customer_id }}/subscription_for_items` - kind
  `create`; body type `form`; path fields `customer_id`; required record fields `customer_id`;
  accepted fields `auto_collection`, `customer_id`, `id`, `po_number`, `start_date`; risk: external
  mutation with billing side effects; approval required.
- `update_subscription`: POST `/subscriptions/{{ record.id }}/update_for_items` - kind `update`;
  body type `form`; path fields `id`; required record fields `id`; accepted fields
  `auto_collection`, `id`, `invoice_date`, `po_number`; risk: external mutation with billing side
  effects; approval required.
- `cancel_subscription`: POST `/subscriptions/{{ record.id }}/cancel_for_items` - kind `update`;
  body type `form`; path fields `id`; required record fields `id`; accepted fields `cancel_option`,
  `cancel_reason_code`, `end_of_term`, `id`; risk: irreversible external mutation (subscription
  cancellation) with billing side effects; approval required.
- `create_credit_note`: POST `/credit_notes` - kind `create`; body type `form`; required record
  fields `type`; accepted fields `currency_code`, `customer_id`, `reason_code`,
  `reference_invoice_id`, `total`, `type`; risk: external mutation with accounting/billing side
  effects; approval required.
- `void_credit_note`: POST `/credit_notes/{{ record.id }}/void` - kind `update`; body type `form`;
  path fields `id`; required record fields `id`; accepted fields `comment`, `id`; risk: irreversible
  external mutation; approval required.
- `create_coupon`: POST `/coupons/create_for_items` - kind `create`; body type `form`; required
  record fields `id`, `name`, `apply_on`; accepted fields `apply_on`, `discount_amount`,
  `discount_percentage`, `discount_type`, `duration_type`, `id`, `name`; risk: external mutation
  with billing/discount side effects; approval required.
- `update_coupon`: POST `/coupons/{{ record.id }}/update_for_items` - kind `update`; body type
  `form`; path fields `id`; required record fields `id`; accepted fields `id`, `name`, `status`,
  `valid_till`; risk: external mutation with billing/discount side effects; approval required.
- `delete_coupon`: POST `/coupons/{{ record.id }}/delete` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_order`: POST `/orders` - kind `create`; body type `form`; required record fields
  `invoice_id`; accepted fields `id`, `invoice_id`, `reference_id`, `status`, `tracking_id`; risk:
  external mutation; approval required.
- `update_order`: POST `/orders/{{ record.id }}` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `id`, `note`, `status`, `tracking_id`,
  `tracking_url`; risk: external mutation; approval required.
- `cancel_order`: POST `/orders/{{ record.id }}/cancel` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`, `cancellation_reason`; accepted fields
  `cancellation_reason`, `comment`, `id`; risk: irreversible external mutation (order cancellation);
  approval required.
- `void_invoice`: POST `/invoices/{{ record.id }}/void` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `comment`, `id`; risk: irreversible
  external mutation with accounting side effects; approval required.
- `collect_payment_for_invoice`: POST `/invoices/{{ record.id }}/collect_payment` - kind `update`;
  body type `form`; path fields `id`; required record fields `id`; accepted fields `amount`,
  `comment`, `id`, `payment_source_id`; risk: external mutation that attempts to charge a payment
  method; approval required.
- `create_webhook_endpoint`: POST `/webhook_endpoints` - kind `create`; body type `form`; required
  record fields `name`, `url`; accepted fields `api_version`, `disabled`, `name`, `url`; risk:
  external mutation exposing business data to a third-party URL; approval required.
- `update_webhook_endpoint`: POST `/webhook_endpoints/{{ record.id }}` - kind `update`; body type
  `form`; path fields `id`; required record fields `id`; accepted fields `disabled`, `id`, `name`,
  `url`; risk: external mutation exposing business data to a third-party URL; approval required.
- `delete_webhook_endpoint`: POST `/webhook_endpoints/{{ record.id }}/delete` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversible external deletion; approval required.
- `create_comment`: POST `/comments` - kind `create`; body type `form`; required record fields
  `entity_type`, `entity_id`, `notes`; accepted fields `added_by`, `entity_id`, `entity_type`,
  `notes`; risk: external mutation; approval required.
- `delete_comment`: POST `/comments/{{ record.id }}/delete` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `add_promotional_credit`: POST `/promotional_credits/add` - kind `update`; body type `form`;
  required record fields `customer_id`, `description`; accepted fields `amount`, `credit_type`,
  `currency_code`, `customer_id`, `description`; risk: external mutation with a direct
  billing-credit financial effect; approval required.
- `deduct_promotional_credit`: POST `/promotional_credits/deduct` - kind `update`; body type `form`;
  required record fields `customer_id`, `description`; accepted fields `amount`, `credit_type`,
  `currency_code`, `customer_id`, `description`; risk: external mutation with a direct
  billing-credit financial effect; approval required.
- `create_virtual_bank_account`: POST `/virtual_bank_accounts` - kind `create`; body type `form`;
  required record fields `customer_id`; accepted fields `customer_id`, `email`,
  `gateway_account_id`; risk: external mutation; approval required.
- `delete_virtual_bank_account`: POST `/virtual_bank_accounts/{{ record.id }}/delete` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: irreversible external deletion;
  approval required.
- `create_card_payment_source`: POST `/payment_sources/create_card` - kind `create`; body type
  `form`; required record fields `customer_id`; accepted fields `card[cvv]`, `card[expiry_month]`,
  `card[expiry_year]`, `card[first_name]`, `card[last_name]`, `card[number]`, `customer_id`; risk:
  external mutation carrying raw payment-card data; approval required.
- `delete_payment_source`: POST `/payment_sources/{{ record.id }}/delete` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversible external deletion; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 32 stream-backed endpoint group(s), 36 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=29, deprecated=1, destructive_admin=5, duplicate_of=16, non_data_endpoint=16,
  out_of_scope=247, requires_elevated_scope=46.
