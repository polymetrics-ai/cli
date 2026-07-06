# Overview

Reads customers, suppliers, products, invoices, orders, quotes, departments, payment terms, units,
and accounts from the Visma e-conomic REST API, and writes customers, suppliers, products, units,
and payment terms.

Readable streams: `customers`, `suppliers`, `products`, `invoices_booked`, `invoices_drafts`,
`orders_drafts`, `orders_sent`, `quotes_drafts`, `quotes_sent`, `departments`, `payment_terms`,
`units`, `vat_types`, `vat_zones`, `accounts`, `customer_groups`, `product_groups`.

Write actions: `create_customer`, `update_customer`, `delete_customer`, `create_supplier`,
`update_supplier`, `delete_supplier`, `create_product`, `update_product`, `delete_product`,
`create_unit`, `update_unit`, `delete_unit`, `create_payment_term`, `update_payment_term`,
`delete_payment_term`.

Service API documentation: https://restdocs.e-conomic.com/.

## Auth setup

Connection fields:

- `agreement_grant_token` (required, secret, string); e-conomic agreement grant token, sent as the
  X-AgreementGrantToken header on every request.
- `app_secret_token` (required, secret, string); e-conomic app secret token, sent as the
  X-AppSecretToken header on every request.
- `base_url` (optional, string); default `https://restapi.e-conomic.com`; format `uri`; Visma
  e-conomic REST API base URL. Defaults to the public e-conomic API host.
- `start_date` (optional, string); format `date-time`; Optional RFC3339 lower-bound value used to
  seed the initial incremental filter (lastUpdated$gte:<date>) on streams that support it, when no
  persisted cursor state exists yet.

Secret fields are redacted in logs and write previews: `agreement_grant_token`, `app_secret_token`.

Default configuration values: `base_url=https://restapi.e-conomic.com`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: page-number pagination; page parameter `skippages`; size parameter `pagesize`;
starts at 0; page size 100.

Pagination by stream: none: `customers`; page_number: `suppliers`, `products`, `invoices_booked`,
`invoices_drafts`, `orders_drafts`, `orders_sent`, `quotes_drafts`, `quotes_sent`, `departments`,
`payment_terms`, `units`, `vat_types`, `vat_zones`, `accounts`, `customer_groups`, `product_groups`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers` - records path `collection`; computed output fields `id`.
- `suppliers`: GET `/suppliers` - records path `collection`; query `filter` from template
  `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; incremental cursor
  `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `id`.
- `products`: GET `/products` - records path `collection`; query `filter` from template
  `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; incremental cursor
  `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `id`.
- `invoices_booked`: GET `/invoices/booked` - records path `collection`; query `filter` from
  template `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number
  pagination; page parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100;
  incremental cursor `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound
  from `start_date`; computed output fields `id`.
- `invoices_drafts`: GET `/invoices/drafts` - records path `collection`; query `filter` from
  template `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number
  pagination; page parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100;
  incremental cursor `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound
  from `start_date`; computed output fields `id`.
- `orders_drafts`: GET `/orders/drafts` - records path `collection`; query `filter` from template
  `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; incremental cursor
  `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `id`.
- `orders_sent`: GET `/orders/sent` - records path `collection`; query `filter` from template
  `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; incremental cursor
  `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `id`.
- `quotes_drafts`: GET `/quotes/drafts` - records path `collection`; query `filter` from template
  `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; incremental cursor
  `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `id`.
- `quotes_sent`: GET `/quotes/sent` - records path `collection`; query `filter` from template
  `lastUpdated$gte:{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; incremental cursor
  `lastUpdated`; sent as `filter`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `id`.
- `departments`: GET `/departments` - records path `collection`; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output
  fields `id`.
- `payment_terms`: GET `/payment-terms` - records path `collection`; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output
  fields `id`.
- `units`: GET `/units` - records path `collection`; page-number pagination; page parameter
  `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output fields `id`.
- `vat_types`: GET `/vat-types` - records path `collection`; page-number pagination; page parameter
  `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output fields `id`.
- `vat_zones`: GET `/vat-zones` - records path `collection`; page-number pagination; page parameter
  `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output fields `id`.
- `accounts`: GET `/accounts` - records path `collection`; page-number pagination; page parameter
  `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output fields `id`.
- `customer_groups`: GET `/customer-groups` - records path `collection`; page-number pagination;
  page parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output
  fields `id`.
- `product_groups`: GET `/product-groups` - records path `collection`; page-number pagination; page
  parameter `skippages`; size parameter `pagesize`; starts at 0; page size 100; computed output
  fields `id`.

## Write actions & risks

Overall write risk: external mutation of Visma e-conomic customers, suppliers, products, units, and
payment terms; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `name`, `currency`, `paymentTerms`, `customerGroup`, `vatZone`; accepted fields `address`, `city`,
  `country`, `currency`, `customerGroup`, `email`, `name`, `paymentTerms`, `vatZone`, `zip`; risk:
  external mutation; approval required.
- `update_customer`: PUT `/customers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `currency`, `paymentTerms`, `customerGroup`, `vatZone`;
  accepted fields `address`, `barred`, `city`, `country`, `currency`, `customerGroup`, `email`,
  `id`, `name`, `paymentTerms`, `vatZone`, `zip`; risk: external mutation; approval required.
- `delete_customer`: DELETE `/customers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: destructive external mutation (deletes a customer permanently); approval
  required.
- `create_supplier`: POST `/suppliers` - kind `create`; body type `json`; required record fields
  `name`, `currency`, `paymentTerms`, `group`, `vatZone`; accepted fields `address`, `city`,
  `country`, `currency`, `email`, `group`, `name`, `paymentTerms`, `vatZone`, `zip`; risk: external
  mutation; approval required.
- `update_supplier`: PUT `/suppliers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `currency`, `paymentTerms`, `group`, `vatZone`;
  accepted fields `address`, `city`, `country`, `currency`, `email`, `group`, `id`, `name`,
  `paymentTerms`, `vatZone`, `zip`; risk: external mutation; approval required.
- `delete_supplier`: DELETE `/suppliers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: destructive external mutation (deletes a supplier permanently); approval
  required.
- `create_product`: POST `/products` - kind `create`; body type `json`; required record fields
  `productNumber`, `name`, `salesPrice`, `productGroup`; accepted fields `barred`, `costPrice`,
  `description`, `name`, `productGroup`, `productNumber`, `salesPrice`, `unit`; risk: external
  mutation; approval required.
- `update_product`: PUT `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `salesPrice`, `productGroup`; accepted fields `barred`,
  `costPrice`, `description`, `id`, `name`, `productGroup`, `salesPrice`, `unit`; risk: external
  mutation; approval required.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: destructive external mutation (deletes a product permanently); approval
  required.
- `create_unit`: POST `/units` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: external mutation; approval required.
- `update_unit`: PUT `/units/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `name`; accepted fields `id`, `name`; risk: external mutation;
  approval required.
- `delete_unit`: DELETE `/units/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: destructive external mutation (deletes a unit permanently); approval required.
- `create_payment_term`: POST `/payment-terms` - kind `create`; body type `json`; required record
  fields `name`, `paymentTermsType`; accepted fields `duration`, `name`, `paymentTermsType`; risk:
  external mutation; approval required.
- `update_payment_term`: PUT `/payment-terms/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `name`, `paymentTermsType`; accepted fields
  `duration`, `id`, `name`, `paymentTermsType`; risk: external mutation; approval required.
- `delete_payment_term`: DELETE `/payment-terms/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: destructive external mutation (deletes a payment term
  permanently); approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 17 stream-backed endpoint group(s), 15 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, duplicate_of=21, non_data_endpoint=4, out_of_scope=23,
  requires_elevated_scope=10.
