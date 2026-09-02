---
name: pm-prestashop
description: PrestaShop connector knowledge and safe action guide.
---

# pm-prestashop

## Purpose

Reads PrestaShop customers, orders, products, addresses, and carts through the PrestaShop Webservice REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: prestashop
- asset: icons/prestashop.svg
- source: official
- review_status: official_verified
- review_url: https://devdocs.prestashop-project.org/9/webservice/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date (required)
- url (required)
- access_key (secret) (required)

## ETL Streams

- customers:
  - primary key: id
  - cursor: date_upd
  - fields: active(boolean), company(string), date_add(string), date_upd(string), email(string), firstname(string), id(integer), id_default_group(integer), id_lang(integer), lastname(string), newsletter(boolean)
- orders:
  - primary key: id
  - cursor: date_upd
  - fields: current_state(integer), date_add(string), date_upd(string), id(integer), id_address_delivery(integer), id_address_invoice(integer), id_customer(integer), payment(string), reference(string), total_paid(string), total_paid_real(string), valid(boolean)
- products:
  - primary key: id
  - cursor: date_upd
  - fields: active(boolean), date_add(string), date_upd(string), id(integer), id_category_default(integer), id_manufacturer(integer), id_supplier(integer), price(string), quantity(integer), reference(string)
- addresses:
  - primary key: id
  - cursor: date_upd
  - fields: city(string), date_add(string), date_upd(string), firstname(string), id(integer), id_country(integer), id_customer(integer), id_state(integer), lastname(string), phone(string), postcode(string)
- carts:
  - primary key: id
  - cursor: date_upd
  - fields: date_add(string), date_upd(string), id(integer), id_address_delivery(integer), id_address_invoice(integer), id_carrier(integer), id_currency(integer), id_customer(integer), id_lang(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external PrestaShop API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect prestashop
```

### Inspect as structured JSON

```bash
pm connectors inspect prestashop --json
```

## Agent Rules

- Run pm connectors inspect prestashop before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
