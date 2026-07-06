# Overview

Reads Cin7 Core (DEAR Inventory) products, customers, suppliers, sales, purchases, inventory
availability, and reference/lookup data, and writes products, customers, suppliers, and
reference-table records, through the Cin7 Core External API v2.

Readable streams: `products`, `customers`, `suppliers`, `sale_list`, `purchase_list`,
`product_families`, `product_availability`, `locations`, `product_categories`, `brands`, `carriers`,
`chart_of_accounts`, `payment_terms`, `tax_rules`, `units_of_measure`, `price_tiers`.

Write actions: `create_product`, `update_product`, `create_customer`, `update_customer`,
`create_supplier`, `update_supplier`, `create_product_category`, `update_product_category`,
`delete_product_category`, `create_brand`, `update_brand`, `create_payment_term`,
`update_payment_term`.

Service API documentation: https://dearinventory.docs.apiary.io/.

## Auth setup

Connection fields:

- `accountid` (required, string); Cin7 Core account ID, sent as the api-auth-accountid header on
  every request.
- `api_key` (required, secret, string); Cin7 Core application key, sent as the
  api-auth-applicationkey header. Never logged.
- `base_url` (optional, string); default `https://inventory.dearsystems.com/externalapi/v2`; format
  `uri`; Cin7 Core API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://inventory.dearsystems.com/externalapi/v2`.

Authentication behavior:

- API key authentication in `api-auth-applicationkey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/product` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `products`: GET `/product` - records path `Products`; query `IncludeDeprecated`=`true`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `brand`, `category`, `cost`, `id`, `last_modified`, `name`, `price_tier1`,
  `sku`, `status`, `type`, `uom`.
- `customers`: GET `/customer` - records path `CustomerList`; query `IncludeDeprecated`=`true`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `currency`, `email`, `id`, `last_modified`, `name`, `payment_term`,
  `phone`, `status`, `tax_rule`.
- `suppliers`: GET `/supplier` - records path `SupplierList`; query `IncludeDeprecated`=`true`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `currency`, `email`, `id`, `last_modified`, `name`, `payment_term`,
  `phone`, `status`.
- `sale_list`: GET `/saleList` - records path `SaleList`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `customer`,
  `customer_id`, `id`, `invoice_amount`, `invoice_status`, `last_modified`, `order_date`,
  `order_number`, `order_status`, `status`.
- `purchase_list`: GET `/purchaseList` - records path `PurchaseList`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`,
  `invoice_amount`, `last_modified`, `order_date`, `order_number`, `order_status`, `status`,
  `supplier`, `supplier_id`.
- `product_families`: GET `/productFamily` - records path `ProductFamilies`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; incremental sent as
  `ModifiedSince`; formatted as `rfc3339`; computed output fields `brand`, `category`, `id`,
  `last_modified`, `name`, `sku`, `uom`.
- `product_availability`: GET `/ref/productavailability` - records path `ProductAvailabilityList`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `allocated`, `available`, `bin`, `id`, `in_transit`, `location`, `name`,
  `on_hand`, `on_order`, `sku`, `stock_on_hand`.
- `locations`: GET `/ref/location` - records path `LocationList`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `product_categories`: GET `/ref/category` - records path `CategoryList`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough
  records.
- `brands`: GET `/ref/brand` - records path `BrandList`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `carriers`: GET `/ref/carrier` - records path `CarrierList`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `chart_of_accounts`: GET `/ref/account` - records path `AccountsList`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough
  records.
- `payment_terms`: GET `/ref/paymentterm` - records path `PaymentTermList`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough
  records.
- `tax_rules`: GET `/ref/tax` - records path `TaxRuleList`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `units_of_measure`: GET `/ref/unit` - records path `UnitList`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `price_tiers`: GET `/ref/priceTier` - records path `PriceTiers`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of live Cin7 Core catalog, customer, supplier, and
reference-table records; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_product`: POST `/product` - kind `create`; body type `json`; required record fields `SKU`,
  `Name`, `Category`, `CostingMethod`, `UOM`, `Status`; accepted fields `Barcode`, `Brand`,
  `Category`, `CostingMethod`, `DefaultLocation`, `Description`, `Name`, `PriceTier1`, `SKU`,
  `Status`, `Type`, `UOM`; risk: external mutation; creates a live Cin7 Core product-catalog entry;
  approval required.
- `update_product`: PUT `/product` - kind `update`; body type `json`; required record fields `ID`;
  accepted fields `Barcode`, `Brand`, `Category`, `Description`, `ID`, `Name`, `PriceTier1`, `SKU`,
  `Status`, `UOM`; risk: external mutation; overwrites live Cin7 Core product-catalog fields;
  approval required.
- `create_customer`: POST `/customer` - kind `create`; body type `json`; required record fields
  `Name`, `Currency`, `PaymentTerm`, `AccountReceivable`, `RevenueAccount`, `TaxRule`; accepted
  fields `AccountReceivable`, `Carrier`, `Comments`, `Currency`, `Discount`, `DisplayName`,
  `Location`, `Name`, `PaymentTerm`, `PriceTier`, `RevenueAccount`, `Status`, `TaxRule`; risk:
  external mutation; creates a live Cin7 Core customer record used for future sales; approval
  required.
- `update_customer`: PUT `/customer` - kind `update`; body type `json`; required record fields `ID`;
  accepted fields `Comments`, `Currency`, `Discount`, `ID`, `Name`, `PaymentTerm`, `Status`,
  `TaxRule`; risk: external mutation; overwrites live Cin7 Core customer fields (billing terms, tax
  rule, credit settings); approval required.
- `create_supplier`: POST `/supplier` - kind `create`; body type `json`; required record fields
  `Name`, `Currency`, `PaymentTerm`, `AccountPayable`, `TaxRule`; accepted fields `AccountPayable`,
  `Comments`, `Currency`, `Discount`, `Name`, `PaymentTerm`, `Status`, `TaxRule`; risk: external
  mutation; creates a live Cin7 Core supplier record used for future purchases; approval required.
- `update_supplier`: PUT `/supplier` - kind `update`; body type `json`; required record fields `ID`;
  accepted fields `AccountPayable`, `Comments`, `Currency`, `Discount`, `ID`, `Name`, `PaymentTerm`,
  `Status`, `TaxRule`; risk: external mutation; overwrites live Cin7 Core supplier fields (billing
  terms, tax rule); approval required.
- `create_product_category`: POST `/ref/category` - kind `create`; body type `json`; required record
  fields `Name`; accepted fields `Name`; risk: external mutation; creates a live Cin7 Core product
  category, immediately selectable on any product; approval required.
- `update_product_category`: PUT `/ref/category` - kind `update`; body type `json`; required record
  fields `ID`, `Name`; accepted fields `ID`, `Name`; risk: external mutation; renames a live Cin7
  Core product category referenced by existing products; approval required.
- `delete_product_category`: DELETE `/ref/category?ID={{ record.ID }}` - kind `delete`; body type
  `none`; path fields `ID`; required record fields `ID`; accepted fields `ID`; risk: external
  mutation; irreversibly deletes a live Cin7 Core product category; approval required.
- `create_brand`: POST `/ref/brand` - kind `create`; body type `json`; required record fields
  `Name`; accepted fields `Name`; risk: external mutation; creates a live Cin7 Core product brand,
  immediately selectable on any product; approval required.
- `update_brand`: PUT `/ref/brand` - kind `update`; body type `json`; required record fields `ID`,
  `Name`; accepted fields `ID`, `Name`; risk: external mutation; renames a live Cin7 Core product
  brand referenced by existing products; approval required.
- `create_payment_term`: POST `/ref/paymentterm` - kind `create`; body type `json`; required record
  fields `Name`; accepted fields `Duration`, `IsActive`, `IsDefault`, `Method`, `Name`; risk:
  external mutation; creates a live Cin7 Core payment term, immediately selectable on
  customers/suppliers; approval required.
- `update_payment_term`: PUT `/ref/paymentterm` - kind `update`; body type `json`; required record
  fields `ID`, `Name`; accepted fields `Duration`, `ID`, `IsActive`, `IsDefault`, `Method`, `Name`;
  risk: external mutation; overwrites a live Cin7 Core payment term's duration/method, affecting
  due-date calculation on future customer/supplier transactions; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 16 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, destructive_admin=9, duplicate_of=3, non_data_endpoint=4, out_of_scope=13,
  requires_elevated_scope=4.
