# Overview

Reads and writes e-conomic customers, products, suppliers, accounts, invoices (booked/draft),
orders, and reference data (currencies, payment terms, VAT zones, customer/product/supplier groups)
through the e-conomic REST API.

Readable streams: `customers`, `products`, `suppliers`, `accounts`, `invoices`, `invoices_drafts`,
`customer_groups`, `product_groups`, `supplier_groups`, `payment_terms`, `vat_zones`, `currencies`,
`orders_drafts`, `orders_archived`.

Write actions: `create_customer`, `update_customer`, `delete_customer`, `create_product`,
`update_product`, `delete_product`, `create_supplier`, `update_supplier`, `delete_supplier`,
`create_draft_invoice`, `update_draft_invoice`, `book_invoice`.

Service API documentation: https://restdocs.e-conomic.com/.

## Auth setup

Connection fields:

- `agreement_grant_token` (required, secret, string); e-conomic per-agreement grant token, sent as
  the X-AgreementGrantToken request header. Never logged.
- `app_secret_token` (required, secret, string); e-conomic app secret token, sent as the
  X-AppSecretToken request header. Never logged.
- `base_url` (optional, string); default `https://restapi.e-conomic.com`; format `uri`; e-conomic
  API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `agreement_grant_token`, `app_secret_token`.

Default configuration values: `base_url=https://restapi.e-conomic.com`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers` with query `pagesize`=`1`; `skippages`=`0`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `pagination.nextPage`;
next URLs stay on the configured API host.

- `customers`: GET `/customers` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `address`, `balance`, `barred`,
  `city`, `country`, `credit_limit`, `currency`, `customer_group_number`, `customer_number`,
  `email`, `name`, `self`, `vat_zone_number`, `zip`.
- `products`: GET `/products` - records path `collection`; query `pagesize`=`100`; `skippages`=`0`;
  follows a next-page URL from the response body; URL path `pagination.nextPage`; next URLs stay on
  the configured API host; computed output fields `barred`, `cost_price`, `description`, `name`,
  `product_group_number`, `product_number`, `recommended_price`, `sales_price`, `self`,
  `unit_number`.
- `suppliers`: GET `/suppliers` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `address`, `barred`, `city`,
  `country`, `currency`, `email`, `name`, `self`, `supplier_group_number`, `supplier_number`,
  `vat_zone_number`, `zip`.
- `accounts`: GET `/accounts` - records path `collection`; query `pagesize`=`100`; `skippages`=`0`;
  follows a next-page URL from the response body; URL path `pagination.nextPage`; next URLs stay on
  the configured API host; computed output fields `account_number`, `account_type`, `balance`,
  `block_direct_entries`, `debit_credit`, `name`, `self`, `vat_code`.
- `invoices`: GET `/invoices/booked` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `booked_invoice_number`,
  `currency`, `customer_number`, `date`, `due_date`, `gross_amount`, `net_amount`,
  `payment_terms_number`, `remainder`, `self`, `vat_amount`.
- `invoices_drafts`: GET `/invoices/drafts` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `currency`, `customer_number`,
  `date`, `draft_invoice_number`, `due_date`, `gross_amount`, `net_amount`, `payment_terms_number`,
  `self`, `vat_amount`.
- `customer_groups`: GET `/customer-groups` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `customer_group_number`, `name`,
  `self`.
- `product_groups`: GET `/product-groups` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `name`, `product_group_number`,
  `self`.
- `supplier_groups`: GET `/supplier-groups` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `name`, `self`,
  `supplier_group_number`.
- `payment_terms`: GET `/payment-terms` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `days_of_credit`, `name`,
  `payment_terms_number`, `self`.
- `vat_zones`: GET `/vat-zones` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `name`, `self`,
  `vat_zone_number`.
- `currencies`: GET `/currencies` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `code`, `name`, `self`.
- `orders_drafts`: GET `/orders/drafts` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `currency`, `customer_number`,
  `date`, `draft_order_number`, `gross_amount`, `net_amount`, `self`.
- `orders_archived`: GET `/orders/archived` - records path `collection`; query `pagesize`=`100`;
  `skippages`=`0`; follows a next-page URL from the response body; URL path `pagination.nextPage`;
  next URLs stay on the configured API host; computed output fields `currency`, `customer_number`,
  `date`, `gross_amount`, `net_amount`, `order_number`, `self`.

## Write actions & risks

Overall write risk: creates/updates/deletes customer, product, and supplier master-data records;
creates/updates draft invoices; books a draft invoice into a legally-binding, thereafter-immutable
booked invoice.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `customerNumber`, `name`, `currency`, `customerGroup`, `vatZone`, `paymentTerms`; accepted fields
  `address`, `city`, `country`, `currency`, `customerGroup`, `customerNumber`, `email`, `name`,
  `paymentTerms`, `vatZone`, `zip`; risk: creates a new customer record in the live e-conomic
  bookkeeping ledger; low-risk additive mutation, no approval required.
- `update_customer`: PUT `/customers/{{ record.customerNumber }}` - kind `update`; body type `json`;
  path fields `customerNumber`; required record fields `customerNumber`; accepted fields `address`,
  `barred`, `city`, `country`, `customerNumber`, `email`, `name`, `zip`; risk: overwrites an
  existing customer's stored details; e-conomic's PUT is a full replace of the resource, so omitted
  optional fields may be cleared.
- `delete_customer`: DELETE `/customers/{{ record.customerNumber }}` - kind `delete`; body type
  `none`; path fields `customerNumber`; required record fields `customerNumber`; accepted fields
  `customerNumber`; missing records treated as success for status `404`; risk: permanently removes a
  customer record; e-conomic rejects the delete (409) if the customer has any booked entries, but a
  customer with no bookkeeping history is removed irreversibly.
- `create_product`: POST `/products` - kind `create`; body type `json`; required record fields
  `productNumber`, `name`, `productGroup`; accepted fields `costPrice`, `description`, `name`,
  `productGroup`, `productNumber`, `salesPrice`, `unit`; risk: creates a new sellable/purchasable
  product in the live e-conomic catalog; low-risk additive mutation, no approval required.
- `update_product`: PUT `/products/{{ record.productNumber }}` - kind `update`; body type `json`;
  path fields `productNumber`; required record fields `productNumber`; accepted fields `barred`,
  `costPrice`, `description`, `name`, `productNumber`, `salesPrice`; risk: overwrites an existing
  product's stored details, including its sales/cost price used on future invoices; e-conomic's PUT
  is a full replace, so omitted optional fields may be cleared.
- `delete_product`: DELETE `/products/{{ record.productNumber }}` - kind `delete`; body type `none`;
  path fields `productNumber`; required record fields `productNumber`; accepted fields
  `productNumber`; missing records treated as success for status `404`; risk: permanently removes a
  product from the catalog; e-conomic rejects the delete (409) if the product is referenced by any
  booked invoice line.
- `create_supplier`: POST `/suppliers` - kind `create`; body type `json`; required record fields
  `supplierNumber`, `name`, `currency`, `supplierGroup`, `vatZone`; accepted fields `address`,
  `city`, `country`, `currency`, `email`, `name`, `supplierGroup`, `supplierNumber`, `vatZone`,
  `zip`; risk: creates a new supplier record in the live e-conomic bookkeeping ledger; low-risk
  additive mutation, no approval required.
- `update_supplier`: PUT `/suppliers/{{ record.supplierNumber }}` - kind `update`; body type `json`;
  path fields `supplierNumber`; required record fields `supplierNumber`; accepted fields `address`,
  `barred`, `city`, `country`, `email`, `name`, `supplierNumber`, `zip`; risk: overwrites an
  existing supplier's stored details; e-conomic's PUT is a full replace of the resource, so omitted
  optional fields may be cleared.
- `delete_supplier`: DELETE `/suppliers/{{ record.supplierNumber }}` - kind `delete`; body type
  `none`; path fields `supplierNumber`; required record fields `supplierNumber`; accepted fields
  `supplierNumber`; missing records treated as success for status `404`; risk: permanently removes a
  supplier record; e-conomic rejects the delete (409) if the supplier has any booked entries.
- `create_draft_invoice`: POST `/invoices/drafts` - kind `create`; body type `json`; required record
  fields `date`, `currency`, `customer`, `paymentTerms`, `layout`, `lines`; accepted fields
  `currency`, `customer`, `date`, `layout`, `lines`, `paymentTerms`, `recipient`; risk: creates a
  new draft (work-in-progress, not yet legally binding) invoice; not yet booked, so reversible by
  deleting the draft - low-risk.
- `update_draft_invoice`: PUT `/invoices/drafts/{{ record.draftInvoiceNumber }}` - kind `update`;
  body type `json`; path fields `draftInvoiceNumber`; required record fields `draftInvoiceNumber`;
  accepted fields `currency`, `customer`, `date`, `draftInvoiceNumber`, `lines`, `paymentTerms`;
  risk: overwrites an existing draft invoice's stored details; only draft (unbooked) invoices are
  mutable - a booked invoice number here is rejected by e-conomic.
- `book_invoice`: POST `/invoices/booked` - kind `create`; body type `json`; required record fields
  `draftInvoice`; accepted fields `bookWithNumber`, `draftInvoice`, `sendBy`; risk: irreversibly
  transitions a draft invoice to a legally-binding booked invoice; e-conomic core invoice fields
  become immutable after booking (a correction requires issuing a credit note against it, not an
  update/delete).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 14 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, deprecated=1, destructive_admin=4, duplicate_of=18, non_data_endpoint=3,
  out_of_scope=30.
