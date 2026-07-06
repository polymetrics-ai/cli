# Overview

Reads and writes Dolibarr ERP/CRM third parties, contacts, products, customer invoices, and orders
through the Dolibarr REST API.

Readable streams: `thirdparties`, `contacts`, `products`, `invoices`, `orders`, `thirdparty_detail`,
`contact_detail`, `product_detail`, `invoice_detail`, `order_detail`.

Write actions: `create_thirdparty`, `update_thirdparty`, `delete_thirdparty`, `create_contact`,
`update_contact`, `delete_contact`, `create_product`, `update_product`, `delete_product`,
`create_invoice`, `update_invoice`, `delete_invoice`, `validate_invoice`, `create_order`,
`update_order`, `delete_order`, `validate_order`.

Service API documentation: https://www.dolibarr.org/webservices.html.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Dolibarr API key, sent as the DOLAPIKEY header. Never
  logged.
- `base_url` (optional, string); default `https://demo.dolibarr.org/api/index.php`; format `uri`;
  Dolibarr REST API base URL, e.g. https://your-dolibarr-host/api/index.php.
- `contact_id` (optional, string); Contact (rowid) the 'contact_detail' stream reads a single record
  for. Required only when reading the 'contact_detail' stream.
- `invoice_id` (optional, string); Invoice (rowid) the 'invoice_detail' stream reads a single record
  for. Required only when reading the 'invoice_detail' stream.
- `mode` (optional, string).
- `order_id` (optional, string); Order (rowid) the 'order_detail' stream reads a single record for.
  Required only when reading the 'order_detail' stream.
- `page_size` (optional, string); default `100`; Records per page (1-1000).
- `product_id` (optional, string); Product (rowid) the 'product_detail' stream reads a single record
  for. Required only when reading the 'product_detail' stream.
- `thirdparty_id` (optional, string); Third party (rowid) the 'thirdparty_detail' stream reads a
  single record for. Required only when reading the 'thirdparty_detail' stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://demo.dolibarr.org/api/index.php`, `page_size=100`.

Authentication behavior:

- API key authentication in `DOLAPIKEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/thirdparties` with query `limit`=`1`; `page`=`0`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
0; page size 100.

Pagination by stream: none: `thirdparty_detail`, `contact_detail`, `product_detail`,
`invoice_detail`, `order_detail`; page_number: `thirdparties`, `contacts`, `products`, `invoices`,
`orders`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `thirdparties`: GET `/thirdparties` - records at response root; query `sortfield`=`t.rowid`;
  `sortorder`=`ASC`; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 0; page size 100; incremental cursor `date_modification`; formatted as `rfc3339`.
- `contacts`: GET `/contacts` - records at response root; query `sortfield`=`t.rowid`;
  `sortorder`=`ASC`; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 0; page size 100; incremental cursor `date_modification`; formatted as `rfc3339`.
- `products`: GET `/products` - records at response root; query `sortfield`=`t.rowid`;
  `sortorder`=`ASC`; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 0; page size 100; incremental cursor `date_modification`; formatted as `rfc3339`.
- `invoices`: GET `/invoices` - records at response root; query `sortfield`=`t.rowid`;
  `sortorder`=`ASC`; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 0; page size 100; incremental cursor `date_modification`; formatted as `rfc3339`.
- `orders`: GET `/orders` - records at response root; query `sortfield`=`t.rowid`;
  `sortorder`=`ASC`; page-number pagination; page parameter `page`; size parameter `limit`; starts
  at 0; page size 100; incremental cursor `date_modification`; formatted as `rfc3339`.
- `thirdparty_detail`: GET `/thirdparties/{{ config.thirdparty_id }}` - single-object response;
  records at response root.
- `contact_detail`: GET `/contacts/{{ config.contact_id }}` - single-object response; records at
  response root.
- `product_detail`: GET `/products/{{ config.product_id }}` - single-object response; records at
  response root.
- `invoice_detail`: GET `/invoices/{{ config.invoice_id }}` - single-object response; records at
  response root.
- `order_detail`: GET `/orders/{{ config.order_id }}` - single-object response; records at response
  root.

## Write actions & risks

Overall write risk: external mutation; creates/updates/deletes live Dolibarr third parties,
contacts, products, invoices, and orders, and validates draft invoices/orders.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_thirdparty`: POST `/thirdparties` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `address`, `client`, `code_client`, `code_fournisseur`,
  `country_code`, `email`, `fournisseur`, `name`, `phone`, `town`, `zip`; risk: external mutation;
  creates a live Dolibarr third party (customer/supplier); approval required.
- `update_thirdparty`: PUT `/thirdparties/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address`, `country_code`, `email`,
  `id`, `name`, `phone`, `town`, `zip`; risk: external mutation; overwrites a live Dolibarr third
  party's record fields; approval required.
- `delete_thirdparty`: DELETE `/thirdparties/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Dolibarr third party; approval required.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `lastname`; accepted fields `address`, `country_code`, `email`, `firstname`, `lastname`,
  `phone_mobile`, `phone_pro`, `socid`, `town`, `zip`; risk: external mutation; creates a live
  Dolibarr contact; approval required.
- `update_contact`: PUT `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `email`, `firstname`, `id`, `lastname`,
  `phone_mobile`, `phone_pro`; risk: external mutation; overwrites a live Dolibarr contact's record
  fields; approval required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Dolibarr contact; approval required.
- `create_product`: POST `/products` - kind `create`; body type `json`; required record fields
  `ref`, `label`; accepted fields `description`, `label`, `price`, `ref`, `status`, `status_buy`,
  `tva_tx`, `type`; risk: external mutation; creates a live Dolibarr product/service; approval
  required.
- `update_product`: PUT `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `description`, `id`, `label`, `price`,
  `status`, `status_buy`, `tva_tx`; risk: external mutation; overwrites a live Dolibarr
  product/service record fields; approval required.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Dolibarr product/service; approval required.
- `create_invoice`: POST `/invoices` - kind `create`; body type `json`; required record fields
  `socid`; accepted fields `date`, `note_private`, `note_public`, `socid`, `type`; risk: external
  mutation; creates a live Dolibarr customer invoice (draft status); approval required.
- `update_invoice`: PUT `/invoices/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `note_private`, `note_public`; risk:
  external mutation; overwrites a live Dolibarr invoice's record fields (only permitted while the
  invoice is in draft status); approval required.
- `delete_invoice`: DELETE `/invoices/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Dolibarr invoice (only permitted while in draft status); approval
  required.
- `validate_invoice`: POST `/invoices/{{ record.id }}/validate` - kind `update`; body type `json`;
  path fields `id`; body fields `idwarehouse`, `notrigger`; required record fields `id`; accepted
  fields `id`, `idwarehouse`, `notrigger`; risk: external mutation; validates a live Dolibarr
  invoice, transitioning it out of draft status irreversibly and assigning its final reference
  number; approval required.
- `create_order`: POST `/orders` - kind `create`; body type `json`; required record fields `socid`;
  accepted fields `date`, `note_private`, `note_public`, `socid`; risk: external mutation; creates a
  live Dolibarr sales order (draft status); approval required.
- `update_order`: PUT `/orders/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `id`, `note_private`, `note_public`; risk: external
  mutation; overwrites a live Dolibarr order's record fields (only permitted while the order is in
  draft status); approval required.
- `delete_order`: DELETE `/orders/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live Dolibarr order (only permitted while in draft status); approval required.
- `validate_order`: POST `/orders/{{ record.id }}/validate` - kind `update`; body type `json`; path
  fields `id`; body fields `idwarehouse`, `notrigger`; required record fields `id`; accepted fields
  `id`, `idwarehouse`, `notrigger`; risk: external mutation; validates a live Dolibarr order,
  transitioning it out of draft status irreversibly and assigning its final reference number;
  approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 17 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=1, duplicate_of=19, out_of_scope=96,
  requires_elevated_scope=10.
