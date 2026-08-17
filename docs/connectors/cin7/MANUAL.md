# pm connectors inspect cin7

```text
NAME
  pm connectors inspect cin7 - Cin7 connector manual

SYNOPSIS
  pm connectors inspect cin7
  pm connectors inspect cin7 --json
  pm credentials add <name> --connector cin7 [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Cin7 Core (DEAR Inventory) products, customers, suppliers, sales, purchases, inventory availability, and reference/lookup data, and writes products, customers, suppliers, and reference-table records, through the Cin7 Core External API v2.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  accountid
  base_url
  mode
  api_key (secret)

ETL STREAMS
  products:
    primary key: id
    fields: brand(), category(), cost(), id(), last_modified(), name(), price_tier1(), sku(), status(), type(), uom()
  customers:
    primary key: id
    fields: currency(), email(), id(), last_modified(), name(), payment_term(), phone(), status(), tax_rule()
  suppliers:
    primary key: id
    fields: currency(), email(), id(), last_modified(), name(), payment_term(), phone(), status()
  sale_list:
    primary key: id
    fields: customer(), customer_id(), id(), invoice_amount(), invoice_status(), last_modified(), order_date(), order_number(), order_status(), status()
  purchase_list:
    primary key: id
    fields: id(), invoice_amount(), last_modified(), order_date(), order_number(), order_status(), status(), supplier(), supplier_id()
  product_families:
    primary key: id
    cursor: last_modified
    fields: brand(), category(), id(), last_modified(), name(), sku(), uom()
  product_availability:
    primary key: id, location, bin
    fields: allocated(), available(), bin(), id(), in_transit(), location(), name(), on_hand(), on_order(), sku(), stock_on_hand()
  locations:
    primary key: ID
    fields: AddressCitySuburb(), AddressCountry(), AddressLine1(), AddressLine2(), AddressStateProvince(), AddressZipPostCode(), Bins(), FixedAssetsLocation(), ID(), IsCoMan(), IsDefault(), IsDeprecated(), IsShopfloor(), IsStaging(), Name(), ParentID(), ParentName(), PickZones(), ReferenceCount()
  product_categories:
    primary key: ID
    fields: ID(), Name()
  brands:
    primary key: ID
    fields: ID(), Name()
  carriers:
    primary key: CarrierID
    fields: CarrierID(), Description()
  chart_of_accounts:
    primary key: Code
    fields: BankAccountId(), BankAccountNumber(), Class(), Code(), Description(), DisplayName(), ForPayments(), Name(), OldCode(), Status(), SystemAccount(), SystemAccountCode(), Type()
  payment_terms:
    primary key: ID
    fields: Duration(), ID(), IsActive(), IsDefault(), Method(), Name()
  tax_rules:
    primary key: ID
    fields: Account(), Components(), ID(), IsActive(), IsTaxForPurchase(), IsTaxForSale(), Name(), TaxInclusive(), TaxPercent()
  units_of_measure:
    primary key: ID
    fields: ID(), Name()
  price_tiers:
    primary key: Code
    fields: Code(), Name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_product:
    endpoint: POST /product
    risk: external mutation; creates a live Cin7 Core product-catalog entry; approval required
  update_product:
    endpoint: PUT /product
    risk: external mutation; overwrites live Cin7 Core product-catalog fields; approval required
  create_customer:
    endpoint: POST /customer
    risk: external mutation; creates a live Cin7 Core customer record used for future sales; approval required
  update_customer:
    endpoint: PUT /customer
    risk: external mutation; overwrites live Cin7 Core customer fields (billing terms, tax rule, credit settings); approval required
  create_supplier:
    endpoint: POST /supplier
    risk: external mutation; creates a live Cin7 Core supplier record used for future purchases; approval required
  update_supplier:
    endpoint: PUT /supplier
    risk: external mutation; overwrites live Cin7 Core supplier fields (billing terms, tax rule); approval required
  create_product_category:
    endpoint: POST /ref/category
    risk: external mutation; creates a live Cin7 Core product category, immediately selectable on any product; approval required
  update_product_category:
    endpoint: PUT /ref/category
    risk: external mutation; renames a live Cin7 Core product category referenced by existing products; approval required
  delete_product_category:
    endpoint: DELETE /ref/category?ID={{ record.ID }}
    required fields: ID
    risk: external mutation; irreversibly deletes a live Cin7 Core product category; approval required
  create_brand:
    endpoint: POST /ref/brand
    risk: external mutation; creates a live Cin7 Core product brand, immediately selectable on any product; approval required
  update_brand:
    endpoint: PUT /ref/brand
    risk: external mutation; renames a live Cin7 Core product brand referenced by existing products; approval required
  create_payment_term:
    endpoint: POST /ref/paymentterm
    risk: external mutation; creates a live Cin7 Core payment term, immediately selectable on customers/suppliers; approval required
  update_payment_term:
    endpoint: PUT /ref/paymentterm
    risk: external mutation; overwrites a live Cin7 Core payment term's duration/method, affecting due-date calculation on future customer/supplier transactions; approval required

SECURITY
  read risk: external Cin7 Core API read of inventory, customer, order, and reference/lookup data
  write risk: external mutation of live Cin7 Core catalog, customer, supplier, and reference-table records; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect cin7

  # Inspect as structured JSON
  pm connectors inspect cin7 --json

AGENT WORKFLOW
  - Run pm connectors inspect cin7 before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
