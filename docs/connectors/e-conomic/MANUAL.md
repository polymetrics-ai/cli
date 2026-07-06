# pm connectors inspect e-conomic

```text
NAME
  pm connectors inspect e-conomic - e-conomic connector manual

SYNOPSIS
  pm connectors inspect e-conomic
  pm connectors inspect e-conomic --json
  pm credentials add <name> --connector e-conomic [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes e-conomic customers, products, suppliers, accounts, invoices (booked/draft), orders, and reference data (currencies, payment terms, VAT zones, customer/product/supplier groups) through the e-conomic REST API.

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
  base_url
  mode
  agreement_grant_token (secret)
  app_secret_token (secret)

ETL STREAMS
  customers:
    primary key: customer_number
    fields: address(), balance(), barred(), city(), country(), credit_limit(), currency(), customer_group_number(), customer_number(), email(), name(), self(), vat_zone_number(), zip()
  products:
    primary key: product_number
    fields: barred(), cost_price(), description(), name(), product_group_number(), product_number(), recommended_price(), sales_price(), self(), unit_number()
  suppliers:
    primary key: supplier_number
    fields: address(), barred(), city(), country(), currency(), email(), name(), self(), supplier_group_number(), supplier_number(), vat_zone_number(), zip()
  accounts:
    primary key: account_number
    fields: account_number(), account_type(), balance(), block_direct_entries(), debit_credit(), name(), self(), vat_code()
  invoices:
    primary key: booked_invoice_number
    fields: booked_invoice_number(), currency(), customer_number(), date(), due_date(), gross_amount(), net_amount(), payment_terms_number(), remainder(), self(), vat_amount()
  invoices_drafts:
    primary key: draft_invoice_number
    fields: currency(), customer_number(), date(), draft_invoice_number(), due_date(), gross_amount(), net_amount(), payment_terms_number(), self(), vat_amount()
  customer_groups:
    primary key: customer_group_number
    fields: customer_group_number(), name(), self()
  product_groups:
    primary key: product_group_number
    fields: name(), product_group_number(), self()
  supplier_groups:
    primary key: supplier_group_number
    fields: name(), self(), supplier_group_number()
  payment_terms:
    primary key: payment_terms_number
    fields: days_of_credit(), name(), payment_terms_number(), self()
  vat_zones:
    primary key: vat_zone_number
    fields: name(), self(), vat_zone_number()
  currencies:
    primary key: code
    fields: code(), name(), self()
  orders_drafts:
    primary key: draft_order_number
    fields: currency(), customer_number(), date(), draft_order_number(), gross_amount(), net_amount(), self()
  orders_archived:
    primary key: order_number
    fields: currency(), customer_number(), date(), gross_amount(), net_amount(), order_number(), self()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: creates a new customer record in the live e-conomic bookkeeping ledger; low-risk additive mutation, no approval required
  update_customer:
    endpoint: PUT /customers/{{ record.customerNumber }}
    required fields: customerNumber
    risk: overwrites an existing customer's stored details; e-conomic's PUT is a full replace of the resource, so omitted optional fields may be cleared
  delete_customer:
    endpoint: DELETE /customers/{{ record.customerNumber }}
    required fields: customerNumber
    risk: permanently removes a customer record; e-conomic rejects the delete (409) if the customer has any booked entries, but a customer with no bookkeeping history is removed irreversibly
  create_product:
    endpoint: POST /products
    risk: creates a new sellable/purchasable product in the live e-conomic catalog; low-risk additive mutation, no approval required
  update_product:
    endpoint: PUT /products/{{ record.productNumber }}
    required fields: productNumber
    risk: overwrites an existing product's stored details, including its sales/cost price used on future invoices; e-conomic's PUT is a full replace, so omitted optional fields may be cleared
  delete_product:
    endpoint: DELETE /products/{{ record.productNumber }}
    required fields: productNumber
    risk: permanently removes a product from the catalog; e-conomic rejects the delete (409) if the product is referenced by any booked invoice line
  create_supplier:
    endpoint: POST /suppliers
    risk: creates a new supplier record in the live e-conomic bookkeeping ledger; low-risk additive mutation, no approval required
  update_supplier:
    endpoint: PUT /suppliers/{{ record.supplierNumber }}
    required fields: supplierNumber
    risk: overwrites an existing supplier's stored details; e-conomic's PUT is a full replace of the resource, so omitted optional fields may be cleared
  delete_supplier:
    endpoint: DELETE /suppliers/{{ record.supplierNumber }}
    required fields: supplierNumber
    risk: permanently removes a supplier record; e-conomic rejects the delete (409) if the supplier has any booked entries
  create_draft_invoice:
    endpoint: POST /invoices/drafts
    risk: creates a new draft (work-in-progress, not yet legally binding) invoice; not yet booked, so reversible by deleting the draft — low-risk
  update_draft_invoice:
    endpoint: PUT /invoices/drafts/{{ record.draftInvoiceNumber }}
    required fields: draftInvoiceNumber
    risk: overwrites an existing draft invoice's stored details; only draft (unbooked) invoices are mutable — a booked invoice number here is rejected by e-conomic
  book_invoice:
    endpoint: POST /invoices/booked
    risk: irreversibly transitions a draft invoice to a legally-binding booked invoice; e-conomic core invoice fields become immutable after booking (a correction requires issuing a credit note against it, not an update/delete)

SECURITY
  read risk: external e-conomic API read of customer, product, supplier, account, invoice, order, and reference-data (currencies/payment-terms/vat-zones/groups) records
  write risk: creates/updates/deletes customer, product, and supplier master-data records; creates/updates draft invoices; books a draft invoice into a legally-binding, thereafter-immutable booked invoice
  approval: none for master-data CRUD and draft-invoice authoring; book_invoice is irreversible (correction requires a credit note) and should be gated by the caller's own review step before use
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect e-conomic

  # Inspect as structured JSON
  pm connectors inspect e-conomic --json

AGENT WORKFLOW
  - Run pm connectors inspect e-conomic before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
