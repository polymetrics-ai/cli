---
name: pm-cin7
description: Cin7 connector knowledge and safe action guide.
---

# pm-cin7

## Purpose

Reads Cin7 Core (DEAR Inventory) products, customers, suppliers, sales, purchases, inventory availability, and reference/lookup data, and writes products, customers, suppliers, and reference-table records, through the Cin7 Core External API v2.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- accountid (required)
- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- products:
  - primary key: id
  - fields: brand(string), category(string), cost(number), id(string), last_modified(string), name(string), price_tier1(number), sku(string), status(string), type(string), uom(string)
- customers:
  - primary key: id
  - fields: currency(string), email(string), id(string), last_modified(string), name(string), payment_term(string), phone(string), status(string), tax_rule(string)
- suppliers:
  - primary key: id
  - fields: currency(string), email(string), id(string), last_modified(string), name(string), payment_term(string), phone(string), status(string)
- sale_list:
  - primary key: id
  - fields: customer(string), customer_id(string), id(string), invoice_amount(number), invoice_status(string), last_modified(string), order_date(string), order_number(string), order_status(string), status(string)
- purchase_list:
  - primary key: id
  - fields: id(string), invoice_amount(number), last_modified(string), order_date(string), order_number(string), order_status(string), status(string), supplier(string), supplier_id(string)
- product_families:
  - primary key: id
  - cursor: last_modified
  - fields: brand(string), category(string), id(string), last_modified(string), name(string), sku(string), uom(string)
- product_availability:
  - primary key: id, location, bin
  - fields: allocated(number), available(number), bin(string), id(string), in_transit(number), location(string), name(string), on_hand(number), on_order(number), sku(string), stock_on_hand(number)
- locations:
  - primary key: ID
  - fields: AddressCitySuburb(string), AddressCountry(string), AddressLine1(string), AddressLine2(string), AddressStateProvince(string), AddressZipPostCode(string), Bins(array), FixedAssetsLocation(boolean), ID(string), IsCoMan(boolean), IsDefault(boolean), IsDeprecated(boolean), IsShopfloor(boolean), IsStaging(boolean), Name(string), ParentID(string), ParentName(string), PickZones(string), ReferenceCount(integer)
- product_categories:
  - primary key: ID
  - fields: ID(string), Name(string)
- brands:
  - primary key: ID
  - fields: ID(string), Name(string)
- carriers:
  - primary key: CarrierID
  - fields: CarrierID(string), Description(string)
- chart_of_accounts:
  - primary key: Code
  - fields: BankAccountId(string), BankAccountNumber(string), Class(string), Code(string), Description(string), DisplayName(string), ForPayments(boolean), Name(string), OldCode(string), Status(string), SystemAccount(string), SystemAccountCode(string), Type(string)
- payment_terms:
  - primary key: ID
  - fields: Duration(integer), ID(string), IsActive(boolean), IsDefault(boolean), Method(string), Name(string)
- tax_rules:
  - primary key: ID
  - fields: Account(string), Components(array), ID(string), IsActive(boolean), IsTaxForPurchase(boolean), IsTaxForSale(boolean), Name(string), TaxInclusive(boolean), TaxPercent(number)
- units_of_measure:
  - primary key: ID
  - fields: ID(string), Name(string)
- price_tiers:
  - primary key: Code
  - fields: Code(integer), Name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_product:
  - endpoint: POST /product
  - required fields: SKU, Name, Category, CostingMethod, UOM, Status
  - risk: external mutation; creates a live Cin7 Core product-catalog entry; approval required
- update_product:
  - endpoint: PUT /product
  - required fields: ID
  - risk: external mutation; overwrites live Cin7 Core product-catalog fields; approval required
- create_customer:
  - endpoint: POST /customer
  - required fields: Name, Currency, PaymentTerm, AccountReceivable, RevenueAccount, TaxRule
  - risk: external mutation; creates a live Cin7 Core customer record used for future sales; approval required
- update_customer:
  - endpoint: PUT /customer
  - required fields: ID
  - risk: external mutation; overwrites live Cin7 Core customer fields (billing terms, tax rule, credit settings); approval required
- create_supplier:
  - endpoint: POST /supplier
  - required fields: Name, Currency, PaymentTerm, AccountPayable, TaxRule
  - risk: external mutation; creates a live Cin7 Core supplier record used for future purchases; approval required
- update_supplier:
  - endpoint: PUT /supplier
  - required fields: ID
  - risk: external mutation; overwrites live Cin7 Core supplier fields (billing terms, tax rule); approval required
- create_product_category:
  - endpoint: POST /ref/category
  - required fields: Name
  - risk: external mutation; creates a live Cin7 Core product category, immediately selectable on any product; approval required
- update_product_category:
  - endpoint: PUT /ref/category
  - required fields: ID, Name
  - risk: external mutation; renames a live Cin7 Core product category referenced by existing products; approval required
- delete_product_category:
  - endpoint: DELETE /ref/category?ID={{ record.ID }}
  - required fields: ID
  - risk: external mutation; irreversibly deletes a live Cin7 Core product category; approval required
- create_brand:
  - endpoint: POST /ref/brand
  - required fields: Name
  - risk: external mutation; creates a live Cin7 Core product brand, immediately selectable on any product; approval required
- update_brand:
  - endpoint: PUT /ref/brand
  - required fields: ID, Name
  - risk: external mutation; renames a live Cin7 Core product brand referenced by existing products; approval required
- create_payment_term:
  - endpoint: POST /ref/paymentterm
  - required fields: Name
  - risk: external mutation; creates a live Cin7 Core payment term, immediately selectable on customers/suppliers; approval required
- update_payment_term:
  - endpoint: PUT /ref/paymentterm
  - required fields: ID, Name
  - risk: external mutation; overwrites a live Cin7 Core payment term's duration/method, affecting due-date calculation on future customer/supplier transactions; approval required

## Security

- read risk: external Cin7 Core API read of inventory, customer, order, and reference/lookup data
- write risk: external mutation of live Cin7 Core catalog, customer, supplier, and reference-table records; approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect cin7
```

### Inspect as structured JSON

```bash
pm connectors inspect cin7 --json
```

## Agent Rules

- Run pm connectors inspect cin7 before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
