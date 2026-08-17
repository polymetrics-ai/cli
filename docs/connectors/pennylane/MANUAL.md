# pm connectors inspect pennylane

```text
NAME
  pm connectors inspect pennylane - Pennylane connector manual

SYNOPSIS
  pm connectors inspect pennylane
  pm connectors inspect pennylane --json
  pm credentials add <name> --connector pennylane [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pennylane v2 customers, customer invoices, suppliers, supplier invoices, products, categories, transactions, and bank accounts, and writes customer/supplier/product/category mutations through the REST API.

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
  filter
  mode
  page_size
  sort
  api_key (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), updated_at()
  customer_invoices:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), updated_at()
  suppliers:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), updated_at()
  products:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), updated_at()
  categories:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), updated_at()
  supplier_invoices:
    primary key: id
    fields: created_at(), date(), id(), invoice_number(), supplier_id(), updated_at()
  transactions:
    primary key: id
    fields: attachment_required(), date(), id(), label(), outstanding_balance()
  bank_accounts:
    primary key: id
    fields: created_at(), currency(), id(), name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_company_customer:
    endpoint: POST /company_customers
    risk: external mutation; creates a company customer record in Pennylane's accounting ledger; approval required
  update_company_customer:
    endpoint: PUT /company_customers/{{ record.id }}
    required fields: id
    risk: external mutation; updates a company customer record in Pennylane's accounting ledger; approval required
  create_individual_customer:
    endpoint: POST /individual_customers
    risk: external mutation; creates an individual customer record in Pennylane's accounting ledger; approval required
  update_individual_customer:
    endpoint: PUT /individual_customers/{{ record.id }}
    required fields: id
    risk: external mutation; updates an individual customer record in Pennylane's accounting ledger; approval required
  create_supplier:
    endpoint: POST /suppliers
    risk: external mutation; creates a supplier record in Pennylane's accounting ledger; approval required
  update_supplier:
    endpoint: PUT /suppliers/{{ record.id }}
    required fields: id
    risk: external mutation; updates a supplier record in Pennylane's accounting ledger; approval required
  create_product:
    endpoint: POST /products
    risk: external mutation; creates a sellable product in Pennylane's accounting ledger; approval required
  update_product:
    endpoint: PUT /products/{{ record.id }}
    required fields: id
    risk: external mutation; updates a product's pricing/VAT metadata in Pennylane; approval required
  create_category:
    endpoint: POST /categories
    risk: external mutation; creates an analytical category in Pennylane's chart of accounts; approval required
  update_category:
    endpoint: PUT /categories/{{ record.id }}
    required fields: id
    risk: external mutation; updates an analytical category in Pennylane's chart of accounts; approval required

SECURITY
  read risk: external Pennylane API read of accounting data (customers, invoices, suppliers, products, categories, transactions, bank accounts)
  write risk: external mutation; creates/updates company and individual customers, suppliers, products, and analytical categories in Pennylane's accounting ledger
  approval: approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pennylane

  # Inspect as structured JSON
  pm connectors inspect pennylane --json

AGENT WORKFLOW
  - Run pm connectors inspect pennylane before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
