# Overview

Reads and writes Chargify (Maxio Advanced Billing) customers, subscriptions, products, product
families, coupons, transactions, invoices, payment profiles, events, and statements through the
Chargify REST API.

Readable streams: `customers`, `subscriptions`, `products`, `coupons`, `transactions`,
`product_families`, `invoices`, `payment_profiles`, `events`, `statements`.

Write actions: `create_customer`, `update_customer`, `create_subscription`, `update_subscription`,
`cancel_subscription`, `create_product_family`, `create_product`, `update_product`, `create_coupon`,
`update_coupon`.

Service API documentation: https://developers.maxio.com/docs/api-docs/YXBpOjE0MTA4MjYx-chargify-api.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Chargify API key. Sent as the HTTP Basic username with the
  literal password "x" (Chargify's documented API-key auth convention). Used only when
  username/password are not both supplied.
- `base_url` (required, string); format `uri`; Chargify API base URL, e.g.
  https://yoursubdomain.chargify.com.
- `domain` (optional, string); Chargify domain (e.g. yoursubdomain.chargify.com). Used to derive
  base_url when base_url is not set.
- `password` (optional, secret, string); Explicit HTTP Basic password; overrides the api_key/"x"
  default when set together with username.
- `subdomain` (optional, string); Chargify subdomain; combined with ".chargify.com" to derive
  base_url when both domain and base_url are unset.
- `username` (optional, string); Explicit HTTP Basic username; overrides api_key when set together
  with the password secret.

Secret fields are redacted in logs and write previews: `api_key`, `password`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password` when `{{ config.username
  }}`.
- HTTP Basic authentication using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers.json` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `country`, `created_at`, `email`, `first_name`,
  `id`, `last_name`, `organization`, `phone`, `reference`, `updated_at`.
- `subscriptions`: GET `/subscriptions.json` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `balance_in_cents`, `created_at`,
  `current_period_ends_at`, `current_period_started_at`, `customer_id`, `id`, `product_id`, `state`,
  `total_revenue_in_cents`, `updated_at`.
- `products`: GET `/products.json` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `created_at`, `description`, `handle`, `id`,
  `interval`, `interval_unit`, `name`, `price_in_cents`, `product_family_id`, `updated_at`.
- `coupons`: GET `/coupons.json` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`; formatted
  as `rfc3339`; computed output fields `amount_in_cents`, `code`, `created_at`, `description`, `id`,
  `name`, `percentage`, `product_family_id`, `updated_at`.
- `transactions`: GET `/transactions.json` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `created_at`; formatted as `rfc3339`; computed output fields `amount_in_cents`, `created_at`,
  `customer_id`, `id`, `kind`, `product_id`, `subscription_id`, `success`, `transaction_type`.
- `product_families`: GET `/product_families.json` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `accounting_code`, `created_at`,
  `description`, `handle`, `id`, `name`, `updated_at`.
- `invoices`: GET `/invoices.json` - records path `invoices`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `created_at`, `currency`, `customer_id`,
  `due_amount`, `due_date`, `id`, `issue_date`, `number`, `paid_amount`, `state`, `subscription_id`,
  `total_amount`, `updated_at`.
- `payment_profiles`: GET `/payment_profiles.json` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `card_type`, `created_at`, `current_vault`, `customer_id`, `expiration_month`, `expiration_year`,
  `id`, `last_four`, `payment_type`, `updated_at`.
- `events`: GET `/events.json` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; incremental cursor `created_at`; formatted
  as `rfc3339`; computed output fields `created_at`, `customer_id`, `id`, `key`, `message`,
  `subscription_id`.
- `statements`: GET `/statements.json` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `closing_balance_in_cents`, `created_at`, `customer_id`, `id`, `settlement_date`,
  `subscription_id`, `uid`.

## Write actions & risks

Overall write risk: external mutation of Chargify billing data (customers, subscriptions, product
catalog, coupons); subscription create/update/cancel actions have direct billing side effects and
require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers.json` - kind `create`; body type `json`; required record
  fields `customer`; accepted fields `customer`; risk: external mutation; approval required.
- `update_customer`: PUT `/customers/{{ record.id }}.json` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `customer`; accepted fields `customer`, `id`; risk:
  external mutation; approval required.
- `create_subscription`: POST `/subscriptions.json` - kind `create`; body type `json`; required
  record fields `subscription`; accepted fields `subscription`; risk: external mutation with billing
  side effects; approval required.
- `update_subscription`: PUT `/subscriptions/{{ record.id }}.json` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `subscription`; accepted fields `id`,
  `subscription`; risk: external mutation with billing side effects; approval required.
- `cancel_subscription`: POST `/subscriptions/{{ record.id }}/cancel.json` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `subscription`;
  risk: external mutation with billing side effects; approval required.
- `create_product_family`: POST `/product_families.json` - kind `create`; body type `json`; required
  record fields `product_family`; accepted fields `product_family`; risk: external mutation;
  approval required.
- `create_product`: POST `/product_families/{{ record.product_family_id }}/products.json` - kind
  `create`; body type `json`; path fields `product_family_id`; required record fields
  `product_family_id`, `product`; accepted fields `product`, `product_family_id`; risk: external
  mutation; approval required.
- `update_product`: PUT `/products/{{ record.id }}.json` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `product`; accepted fields `id`, `product`; risk:
  external mutation; approval required.
- `create_coupon`: POST `/product_families/{{ record.product_family_id }}/coupons.json` - kind
  `create`; body type `json`; path fields `product_family_id`; required record fields
  `product_family_id`, `coupon`; accepted fields `coupon`, `product_family_id`; risk: external
  mutation; approval required.
- `update_coupon`: PUT `/coupons/{{ record.id }}.json` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `coupon`; accepted fields `coupon`, `id`; risk: external
  mutation; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=4, duplicate_of=10, non_data_endpoint=2, out_of_scope=16,
  requires_elevated_scope=3.
