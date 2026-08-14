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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  agreement_grant_token (secret) (required)
  app_secret_token (secret) (required)

ETL STREAMS
  customers:
    primary key: customer_number
    fields: address(string), balance(number), barred(boolean), city(string), country(string), credit_limit(number), currency(string), customer_group_number(integer), customer_number(integer), email(string), name(string), self(string), vat_zone_number(integer), zip(string)
  products:
    primary key: product_number
    fields: barred(boolean), cost_price(number), description(string), name(string), product_group_number(integer), product_number(string), recommended_price(number), sales_price(number), self(string), unit_number(integer)
  suppliers:
    primary key: supplier_number
    fields: address(string), barred(boolean), city(string), country(string), currency(string), email(string), name(string), self(string), supplier_group_number(integer), supplier_number(integer), vat_zone_number(integer), zip(string)
  accounts:
    primary key: account_number
    fields: account_number(integer), account_type(string), balance(number), block_direct_entries(boolean), debit_credit(string), name(string), self(string), vat_code(string)
  invoices:
    primary key: booked_invoice_number
    fields: booked_invoice_number(integer), currency(string), customer_number(integer), date(string), due_date(string), gross_amount(number), net_amount(number), payment_terms_number(integer), remainder(number), self(string), vat_amount(number)
  invoices_drafts:
    primary key: draft_invoice_number
    fields: currency(string), customer_number(integer), date(string), draft_invoice_number(integer), due_date(string), gross_amount(number), net_amount(number), payment_terms_number(integer), self(string), vat_amount(number)
  customer_groups:
    primary key: customer_group_number
    fields: customer_group_number(integer), name(string), self(string)
  product_groups:
    primary key: product_group_number
    fields: name(string), product_group_number(integer), self(string)
  supplier_groups:
    primary key: supplier_group_number
    fields: name(string), self(string), supplier_group_number(integer)
  payment_terms:
    primary key: payment_terms_number
    fields: days_of_credit(integer), name(string), payment_terms_number(integer), self(string)
  vat_zones:
    primary key: vat_zone_number
    fields: name(string), self(string), vat_zone_number(integer)
  currencies:
    primary key: code
    fields: code(string), name(string), self(string)
  orders_drafts:
    primary key: draft_order_number
    fields: currency(string), customer_number(integer), date(string), draft_order_number(integer), gross_amount(number), net_amount(number), self(string)
  orders_archived:
    primary key: order_number
    fields: currency(string), customer_number(integer), date(string), gross_amount(number), net_amount(number), order_number(integer), self(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    required fields: customerNumber, name, currency, customerGroup, vatZone, paymentTerms
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
    required fields: productNumber, name, productGroup
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
    required fields: supplierNumber, name, currency, supplierGroup, vatZone
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
    required fields: date, currency, customer, paymentTerms, layout, lines
    risk: creates a new draft (work-in-progress, not yet legally binding) invoice; not yet booked, so reversible by deleting the draft — low-risk
  update_draft_invoice:
    endpoint: PUT /invoices/drafts/{{ record.draftInvoiceNumber }}
    required fields: draftInvoiceNumber
    risk: overwrites an existing draft invoice's stored details; only draft (unbooked) invoices are mutable — a booked invoice number here is rejected by e-conomic
  book_invoice:
    endpoint: POST /invoices/booked
    required fields: draftInvoice
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
