---
name: pm-pennylane
description: Pennylane connector knowledge and safe action guide.
---

# pm-pennylane

## Purpose

Reads Pennylane v2 customers, customer invoices, suppliers, supplier invoices, products, categories, transactions, and bank accounts, and writes customer/supplier/product/category mutations through the REST API.

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

- base_url
- filter
- mode
- page_size
- sort
- api_key (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- customer_invoices:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- suppliers:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- products:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), name(string), updated_at(string)
- supplier_invoices:
  - primary key: id
  - fields: created_at(string), date(string), id(integer), invoice_number(string), supplier_id(integer), updated_at(string)
- transactions:
  - primary key: id
  - fields: attachment_required(boolean), date(string), id(integer), label(string), outstanding_balance(string)
- bank_accounts:
  - primary key: id
  - fields: created_at(string), currency(string), id(integer), name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_company_customer:
  - endpoint: POST /company_customers
  - required fields: name, billing_address
  - risk: external mutation; creates a company customer record in Pennylane's accounting ledger; approval required
- update_company_customer:
  - endpoint: PUT /company_customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a company customer record in Pennylane's accounting ledger; approval required
- create_individual_customer:
  - endpoint: POST /individual_customers
  - required fields: first_name, last_name, billing_address
  - risk: external mutation; creates an individual customer record in Pennylane's accounting ledger; approval required
- update_individual_customer:
  - endpoint: PUT /individual_customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an individual customer record in Pennylane's accounting ledger; approval required
- create_supplier:
  - endpoint: POST /suppliers
  - required fields: name
  - risk: external mutation; creates a supplier record in Pennylane's accounting ledger; approval required
- update_supplier:
  - endpoint: PUT /suppliers/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a supplier record in Pennylane's accounting ledger; approval required
- create_product:
  - endpoint: POST /products
  - risk: external mutation; creates a sellable product in Pennylane's accounting ledger; approval required
- update_product:
  - endpoint: PUT /products/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a product's pricing/VAT metadata in Pennylane; approval required
- create_category:
  - endpoint: POST /categories
  - required fields: label, category_group_id
  - risk: external mutation; creates an analytical category in Pennylane's chart of accounts; approval required
- update_category:
  - endpoint: PUT /categories/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an analytical category in Pennylane's chart of accounts; approval required

## Security

- read risk: external Pennylane API read of accounting data (customers, invoices, suppliers, products, categories, transactions, bank accounts)
- write risk: external mutation; creates/updates company and individual customers, suppliers, products, and analytical categories in Pennylane's accounting ledger
- approval: approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pennylane
```

### Inspect as structured JSON

```bash
pm connectors inspect pennylane --json
```

## Agent Rules

- Run pm connectors inspect pennylane before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
