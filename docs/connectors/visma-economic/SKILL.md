---
name: pm-visma-economic
description: Visma e-conomic connector knowledge and safe action guide.
---

# pm-visma-economic

## Purpose

Reads customers, suppliers, products, invoices, orders, quotes, departments, payment terms, units, and accounts from the Visma e-conomic REST API, and writes customers, suppliers, products, units, and payment terms.

## Icon

- asset: icons/visma-economic.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://restdocs.e-conomic.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- agreement_grant_token (secret)
- app_secret_token (secret)

## ETL Streams

- customers:
  - primary key: id
  - fields: currency(), id(), name()
- suppliers:
  - primary key: id
  - cursor: lastUpdated
  - fields: address(), balance(), city(), corporateIdentificationNumber(), country(), currency(), ean(), email(), id(), lastUpdated(), name(), telephoneAndFaxNumber(), vatNumber(), zip()
- products:
  - primary key: id
  - cursor: lastUpdated
  - fields: barred(), costPrice(), description(), id(), lastUpdated(), name(), salesPrice()
- invoices_booked:
  - primary key: id
  - cursor: lastUpdated
  - fields: currency(), date(), dueDate(), grossAmount(), id(), lastUpdated(), netAmount(), paymentTerms(), vatAmount()
- invoices_drafts:
  - primary key: id
  - cursor: lastUpdated
  - fields: currency(), date(), dueDate(), grossAmount(), id(), lastUpdated(), netAmount(), notes(), paymentTerms(), references(), vatAmount()
- orders_drafts:
  - primary key: id
  - cursor: lastUpdated
  - fields: currency(), date(), grossAmount(), id(), lastUpdated(), netAmount(), notes(), references(), vatAmount()
- orders_sent:
  - primary key: id
  - cursor: lastUpdated
  - fields: currency(), date(), grossAmount(), id(), lastUpdated(), netAmount(), notes(), references(), vatAmount()
- quotes_drafts:
  - primary key: id
  - cursor: lastUpdated
  - fields: currency(), date(), grossAmount(), id(), lastUpdated(), netAmount(), notes(), references(), vatAmount()
- quotes_sent:
  - primary key: id
  - cursor: lastUpdated
  - fields: currency(), date(), grossAmount(), id(), lastUpdated(), netAmount(), notes(), references(), vatAmount()
- departments:
  - primary key: id
  - fields: id(), name()
- payment_terms:
  - primary key: id
  - fields: duration(), id(), name(), paymentTermsType()
- units:
  - primary key: id
  - fields: id(), name()
- vat_types:
  - primary key: id
  - fields: accountingApplication(), id(), name(), vatPercentage()
- vat_zones:
  - primary key: id
  - fields: enabledForCustomer(), enabledForSupplier(), id(), name()
- accounts:
  - primary key: id
  - fields: accountType(), balance(), blocked(), id(), name()
- customer_groups:
  - primary key: id
  - fields: accountNumber(), id(), name()
- product_groups:
  - primary key: id
  - fields: id(), name(), salesAccount()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation; approval required
- update_customer:
  - endpoint: PUT /customers/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_customer:
  - endpoint: DELETE /customers/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation (deletes a customer permanently); approval required
- create_supplier:
  - endpoint: POST /suppliers
  - risk: external mutation; approval required
- update_supplier:
  - endpoint: PUT /suppliers/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_supplier:
  - endpoint: DELETE /suppliers/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation (deletes a supplier permanently); approval required
- create_product:
  - endpoint: POST /products
  - risk: external mutation; approval required
- update_product:
  - endpoint: PUT /products/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_product:
  - endpoint: DELETE /products/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation (deletes a product permanently); approval required
- create_unit:
  - endpoint: POST /units
  - risk: external mutation; approval required
- update_unit:
  - endpoint: PUT /units/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_unit:
  - endpoint: DELETE /units/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation (deletes a unit permanently); approval required
- create_payment_term:
  - endpoint: POST /payment-terms
  - risk: external mutation; approval required
- update_payment_term:
  - endpoint: PUT /payment-terms/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_payment_term:
  - endpoint: DELETE /payment-terms/{{ record.id }}
  - required fields: id
  - risk: destructive external mutation (deletes a payment term permanently); approval required

## Security

- read risk: external Visma e-conomic API read of customer, supplier, product, and accounting data
- write risk: external mutation of Visma e-conomic customers, suppliers, products, units, and payment terms; approval required
- approval: read: none; write: required for every action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect visma-economic
```

### Inspect as structured JSON

```bash
pm connectors inspect visma-economic --json
```

## Agent Rules

- Run pm connectors inspect visma-economic before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
