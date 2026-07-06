# Overview

Reads Pennylane v2 customers, customer invoices, suppliers, supplier invoices, products, categories,
transactions, and bank accounts, and writes customer/supplier/product/category mutations through the
REST API.

Readable streams: `customers`, `customer_invoices`, `suppliers`, `products`, `categories`,
`supplier_invoices`, `transactions`, `bank_accounts`.

Write actions: `create_company_customer`, `update_company_customer`, `create_individual_customer`,
`update_individual_customer`, `create_supplier`, `update_supplier`, `create_product`,
`update_product`, `create_category`, `update_category`.

Service API documentation: https://pennylane.readme.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Pennylane API key. Used only for Bearer auth; never logged.
- `base_url` (optional, string); default `https://app.pennylane.com/api/external/v2`; format `uri`;
  Pennylane API base URL override for tests or proxies.
- `filter` (optional, string); Optional Pennylane filter query expression, sent verbatim as the
  filter query param.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-100).
- `sort` (optional, string); Optional Pennylane sort expression, sent verbatim as the sort query
  param.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.pennylane.com/api/external/v2`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `next_cursor`.

- `customers`: GET `/customers` - records path `items`; query `filter` from template `{{
  config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`, default
  `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`; emits passthrough records.
- `customer_invoices`: GET `/customer_invoices` - records path `items`; query `filter` from template
  `{{ config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`; emits passthrough records.
- `suppliers`: GET `/suppliers` - records path `items`; query `filter` from template `{{
  config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`, default
  `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`; emits passthrough records.
- `products`: GET `/products` - records path `items`; query `filter` from template `{{ config.filter
  }}`, omitted when absent; `limit` from template `{{ config.page_size }}`, default `50`; `sort`
  from template `{{ config.sort }}`, omitted when absent; cursor pagination; cursor parameter
  `cursor`; next token from `next_cursor`; emits passthrough records.
- `categories`: GET `/categories` - records path `items`; query `filter` from template `{{
  config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`, default
  `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`; emits passthrough records.
- `supplier_invoices`: GET `/supplier_invoices` - records path `items`; query `filter` from template
  `{{ config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`; emits passthrough records.
- `transactions`: GET `/transactions` - records path `items`; query `filter` from template `{{
  config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`, default
  `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`; emits passthrough records.
- `bank_accounts`: GET `/bank_accounts` - records path `items`; query `filter` from template `{{
  config.filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`, default
  `50`; `sort` from template `{{ config.sort }}`, omitted when absent; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation; creates/updates company and individual customers, suppliers,
products, and analytical categories in Pennylane's accounting ledger.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_company_customer`: POST `/company_customers` - kind `create`; body type `json`; required
  record fields `name`, `billing_address`; accepted fields `billing_address`, `billing_iban`,
  `delivery_address`, `emails`, `external_reference`, `name`, `notes`, `payment_conditions`,
  `phone`, `recipient`, `reference`, `reg_no`, `vat_number`; risk: external mutation; creates a
  company customer record in Pennylane's accounting ledger; approval required.
- `update_company_customer`: PUT `/company_customers/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `billing_address`,
  `billing_iban`, `delivery_address`, `emails`, `external_reference`, `id`, `name`, `notes`,
  `payment_conditions`, `phone`, `recipient`, `reference`, `reg_no`, `vat_number`; risk: external
  mutation; updates a company customer record in Pennylane's accounting ledger; approval required.
- `create_individual_customer`: POST `/individual_customers` - kind `create`; body type `json`;
  required record fields `first_name`, `last_name`, `billing_address`; accepted fields
  `billing_address`, `billing_iban`, `delivery_address`, `emails`, `external_reference`,
  `first_name`, `last_name`, `notes`, `payment_conditions`, `phone`, `recipient`, `reference`; risk:
  external mutation; creates an individual customer record in Pennylane's accounting ledger;
  approval required.
- `update_individual_customer`: PUT `/individual_customers/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `billing_address`,
  `billing_iban`, `delivery_address`, `emails`, `external_reference`, `first_name`, `id`,
  `last_name`, `notes`, `payment_conditions`, `phone`, `recipient`, `reference`; risk: external
  mutation; updates an individual customer record in Pennylane's accounting ledger; approval
  required.
- `create_supplier`: POST `/suppliers` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `emails`, `establishment_no`, `external_reference`, `iban`, `name`,
  `postal_address`, `reg_no`, `supplier_due_date_delay`, `supplier_due_date_rule`,
  `supplier_payment_method`, `vat_number`; risk: external mutation; creates a supplier record in
  Pennylane's accounting ledger; approval required.
- `update_supplier`: PUT `/suppliers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `emails`, `establishment_no`,
  `external_reference`, `iban`, `id`, `name`, `postal_address`, `reg_no`, `supplier_due_date_delay`,
  `supplier_due_date_rule`, `supplier_payment_method`, `vat_number`; risk: external mutation;
  updates a supplier record in Pennylane's accounting ledger; approval required.
- `create_product`: POST `/products` - kind `create`; body type `json`; accepted fields `currency`,
  `description`, `external_reference`, `label`, `ledger_account_id`, `price_before_tax`,
  `reference`, `unit`, `vat_rate`; risk: external mutation; creates a sellable product in
  Pennylane's accounting ledger; approval required.
- `update_product`: PUT `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `currency`, `description`,
  `external_reference`, `id`, `label`, `ledger_account_id`, `price_before_tax`, `reference`, `unit`,
  `vat_rate`; risk: external mutation; updates a product's pricing/VAT metadata in Pennylane;
  approval required.
- `create_category`: POST `/categories` - kind `create`; body type `json`; required record fields
  `label`, `category_group_id`; accepted fields `analytical_code`, `category_group_id`, `direction`,
  `label`; risk: external mutation; creates an analytical category in Pennylane's chart of accounts;
  approval required.
- `update_category`: PUT `/categories/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `analytical_code`, `direction`, `id`,
  `label`; risk: external mutation; updates an analytical category in Pennylane's chart of accounts;
  approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 8 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=8, deprecated=1, duplicate_of=68, non_data_endpoint=2, out_of_scope=4,
  requires_elevated_scope=57.
