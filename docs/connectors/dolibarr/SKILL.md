---
name: pm-dolibarr
description: Dolibarr connector knowledge and safe action guide.
---

# pm-dolibarr

## Purpose

Reads and writes Dolibarr ERP/CRM third parties, contacts, products, customer invoices, and orders through the Dolibarr REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- contact_id
- invoice_id
- mode
- order_id
- page_size
- product_id
- thirdparty_id
- api_key (secret)

## ETL Streams

- thirdparties:
  - primary key: id
  - cursor: date_modification
  - fields: client(), country_code(), date_creation(), date_modification(), email(), fournisseur(), id(), name(), name_alias(), phone(), status(), town(), zip()
- contacts:
  - primary key: id
  - cursor: date_modification
  - fields: country_code(), date_creation(), date_modification(), email(), firstname(), id(), lastname(), phone_mobile(), phone_pro(), socid(), statut(), town(), zip()
- products:
  - primary key: id
  - cursor: date_modification
  - fields: date_creation(), date_modification(), id(), label(), price(), price_ttc(), ref(), status(), status_buy(), stock_reel(), tva_tx(), type()
- invoices:
  - primary key: id
  - cursor: date_modification
  - fields: date(), date_creation(), date_modification(), id(), paye(), ref(), socid(), status(), total_ht(), total_ttc(), total_tva(), type()
- orders:
  - primary key: id
  - cursor: date_modification
  - fields: billed(), date(), date_creation(), date_modification(), id(), ref(), socid(), status(), total_ht(), total_ttc(), total_tva()
- thirdparty_detail:
  - primary key: id
  - cursor: date_modification
  - fields: address(), client(), code_client(), code_fournisseur(), country_code(), date_creation(), date_modification(), email(), fournisseur(), id(), name(), name_alias(), phone(), siren(), siret(), status(), town(), tva_intra(), zip()
- contact_detail:
  - primary key: id
  - cursor: date_modification
  - fields: address(), country_code(), date_creation(), date_modification(), email(), firstname(), id(), lastname(), phone_mobile(), phone_pro(), poste(), socid(), statut(), town(), zip()
- product_detail:
  - primary key: id
  - cursor: date_modification
  - fields: barcode(), date_creation(), date_modification(), description(), id(), label(), length(), price(), price_ttc(), ref(), status(), status_buy(), stock_reel(), tva_tx(), type(), weight()
- invoice_detail:
  - primary key: id
  - cursor: date_modification
  - fields: date(), date_creation(), date_lim_reglement(), date_modification(), id(), note_private(), note_public(), paye(), ref(), remise_percent(), socid(), status(), total_ht(), total_ttc(), total_tva(), type()
- order_detail:
  - primary key: id
  - cursor: date_modification
  - fields: billed(), date(), date_creation(), date_livraison(), date_modification(), id(), note_private(), note_public(), ref(), remise_percent(), socid(), status(), total_ht(), total_ttc(), total_tva()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_thirdparty:
  - endpoint: POST /thirdparties
  - risk: external mutation; creates a live Dolibarr third party (customer/supplier); approval required
- update_thirdparty:
  - endpoint: PUT /thirdparties/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Dolibarr third party's record fields; approval required
- delete_thirdparty:
  - endpoint: DELETE /thirdparties/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Dolibarr third party; approval required
- create_contact:
  - endpoint: POST /contacts
  - risk: external mutation; creates a live Dolibarr contact; approval required
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Dolibarr contact's record fields; approval required
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Dolibarr contact; approval required
- create_product:
  - endpoint: POST /products
  - risk: external mutation; creates a live Dolibarr product/service; approval required
- update_product:
  - endpoint: PUT /products/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Dolibarr product/service record fields; approval required
- delete_product:
  - endpoint: DELETE /products/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Dolibarr product/service; approval required
- create_invoice:
  - endpoint: POST /invoices
  - risk: external mutation; creates a live Dolibarr customer invoice (draft status); approval required
- update_invoice:
  - endpoint: PUT /invoices/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Dolibarr invoice's record fields (only permitted while the invoice is in draft status); approval required
- delete_invoice:
  - endpoint: DELETE /invoices/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Dolibarr invoice (only permitted while in draft status); approval required
- validate_invoice:
  - endpoint: POST /invoices/{{ record.id }}/validate
  - required fields: id
  - optional fields: idwarehouse, notrigger
  - risk: external mutation; validates a live Dolibarr invoice, transitioning it out of draft status irreversibly and assigning its final reference number; approval required
- create_order:
  - endpoint: POST /orders
  - risk: external mutation; creates a live Dolibarr sales order (draft status); approval required
- update_order:
  - endpoint: PUT /orders/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Dolibarr order's record fields (only permitted while the order is in draft status); approval required
- delete_order:
  - endpoint: DELETE /orders/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Dolibarr order (only permitted while in draft status); approval required
- validate_order:
  - endpoint: POST /orders/{{ record.id }}/validate
  - required fields: id
  - optional fields: idwarehouse, notrigger
  - risk: external mutation; validates a live Dolibarr order, transitioning it out of draft status irreversibly and assigning its final reference number; approval required

## Security

- read risk: external Dolibarr instance read of ERP/CRM business data
- write risk: external mutation; creates/updates/deletes live Dolibarr third parties, contacts, products, invoices, and orders, and validates draft invoices/orders
- approval: required for every write action; delete_* actions are irreversible in Dolibarr
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dolibarr
```

### Inspect as structured JSON

```bash
pm connectors inspect dolibarr --json
```

## Agent Rules

- Run pm connectors inspect dolibarr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
