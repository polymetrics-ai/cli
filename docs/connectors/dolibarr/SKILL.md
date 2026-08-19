---
name: pm-dolibarr
description: Dolibarr connector knowledge and safe action guide.
---

# pm-dolibarr

## Purpose

Reads and writes Dolibarr ERP/CRM third parties, contacts, products, customer invoices, and orders through the Dolibarr REST API.

## Icon

- id: simple-icons-dolibarr
- asset: icons/simple-icons/dolibarr.svg
- title: Dolibarr
- simple_icon_slug: dolibarr
- simple_icon_hex: 263C5C
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Dolibarr
- match: exact-name-or-slug
- matched_by: dolibarr

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
- api_key (secret) (required)

## ETL Streams

- thirdparties:
  - primary key: id
  - cursor: date_modification
  - fields: client(string), country_code(string), date_creation(integer), date_modification(integer), email(string), fournisseur(string), id(string), name(string), name_alias(string), phone(string), status(string), town(string), zip(string)
- contacts:
  - primary key: id
  - cursor: date_modification
  - fields: country_code(string), date_creation(integer), date_modification(integer), email(string), firstname(string), id(string), lastname(string), phone_mobile(string), phone_pro(string), socid(string), statut(string), town(string), zip(string)
- products:
  - primary key: id
  - cursor: date_modification
  - fields: date_creation(integer), date_modification(integer), id(string), label(string), price(string), price_ttc(string), ref(string), status(string), status_buy(string), stock_reel(string), tva_tx(string), type(string)
- invoices:
  - primary key: id
  - cursor: date_modification
  - fields: date(integer), date_creation(integer), date_modification(integer), id(string), paye(string), ref(string), socid(string), status(string), total_ht(string), total_ttc(string), total_tva(string), type(string)
- orders:
  - primary key: id
  - cursor: date_modification
  - fields: billed(string), date(integer), date_creation(integer), date_modification(integer), id(string), ref(string), socid(string), status(string), total_ht(string), total_ttc(string), total_tva(string)
- thirdparty_detail:
  - primary key: id
  - cursor: date_modification
  - fields: address(string), client(string), code_client(string), code_fournisseur(string), country_code(string), date_creation(integer), date_modification(integer), email(string), fournisseur(string), id(string), name(string), name_alias(string), phone(string), siren(string), siret(string), status(string), town(string), tva_intra(string), zip(string)
- contact_detail:
  - primary key: id
  - cursor: date_modification
  - fields: address(string), country_code(string), date_creation(integer), date_modification(integer), email(string), firstname(string), id(string), lastname(string), phone_mobile(string), phone_pro(string), poste(string), socid(string), statut(string), town(string), zip(string)
- product_detail:
  - primary key: id
  - cursor: date_modification
  - fields: barcode(string), date_creation(integer), date_modification(integer), description(string), id(string), label(string), length(string), price(string), price_ttc(string), ref(string), status(string), status_buy(string), stock_reel(string), tva_tx(string), type(string), weight(string)
- invoice_detail:
  - primary key: id
  - cursor: date_modification
  - fields: date(integer), date_creation(integer), date_lim_reglement(integer), date_modification(integer), id(string), note_private(string), note_public(string), paye(string), ref(string), remise_percent(string), socid(string), status(string), total_ht(string), total_ttc(string), total_tva(string), type(string)
- order_detail:
  - primary key: id
  - cursor: date_modification
  - fields: billed(string), date(integer), date_creation(integer), date_livraison(integer), date_modification(integer), id(string), note_private(string), note_public(string), ref(string), remise_percent(string), socid(string), status(string), total_ht(string), total_ttc(string), total_tva(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_thirdparty:
  - endpoint: POST /thirdparties
  - required fields: name
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
  - required fields: lastname
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
  - required fields: ref, label
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
  - required fields: socid
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
  - required fields: socid
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
