---
name: pm-clazar
description: Clazar connector knowledge and safe action guide.
---

# pm-clazar

## Purpose

Reads Clazar cloud GTM data (buyers, listings, contracts, opportunities, private offers, reseller offers, contacts, and metering records) and writes buyer/opportunity/contract/private-offer/contact/metering mutations, contract activation, and metering-record submission, through the Clazar REST API using OAuth2 client credentials.

## Icon

- id: clazar
- asset: icons/clazar.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://clazar.io/api-docs

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- buyers:
  - primary key: id
  - cursor: last_modified_at
  - fields: cloud(string), cloud_account_id(string), domain(string), id(string), last_modified_at(string), latest_contract_id(string), listing_id(string), name(string), status(string)
- listings:
  - primary key: id
  - cursor: last_modified_at
  - fields: cloud(string), cloud_id(string), cloud_url(string), eula_type(string), id(string), last_modified_at(string), long_description(string), short_description(string), status(string), title(string)
- contracts:
  - primary key: id
  - cursor: last_modified_at
  - fields: accepted_at(string), auto_renew(boolean), buyer_id(string), cloud(string), cloud_id(string), duration(string), end_at(string), id(string), last_modified_at(string), latest_offer_id(string), listing_id(string), offer_type(string), start_at(string), status(string)
- opportunities:
  - primary key: id
  - cursor: last_modified_at
  - fields: accept_by(string), cloud(string), cloud_id(string), created_at(string), customer_company(string), customer_website(string), id(string), last_modified_at(string), stage(string), status(string), target_close_date(string), title(string)
- private_offers:
  - primary key: id
  - cursor: last_modified_at
  - fields: accepted_at(string), archived(string), cloud(string), cloud_id(string), duration(string), eula_type(string), expiration_at(string), id(string), last_modified_at(string), listing_id(string), name(string), offer_type(string), published_at(string), status(string)
- reseller_offers:
  - primary key: id
  - fields: accepted_at(string), archived(boolean), cloud(string), cloud_id(string), cloud_url(string), eula_type(string), expiration_at(string), id(string), listing_id(string), name(string), published_at(string), status(string)
- contacts:
  - primary key: id
  - fields: created_at(string), email(string), full_name(string), id(string), is_editable(boolean), phone_number(string), updated_at(string), uuid(string)
- metering:
  - primary key: id
  - fields: cloud(string), contract_id(string), custom_properties(object), dimension(string), end_time(string), id(string), quantity(string), start_time(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- update_buyer:
  - endpoint: PATCH /buyers/{{ record.id }}/
  - required fields: id
  - risk: external mutation of a buyer's custom properties / external-system associations; approval required
- update_opportunity:
  - endpoint: PATCH /opportunities/{{ record.id }}/
  - required fields: id
  - risk: external mutation of an opportunity's custom properties / external-system associations; approval required
- update_private_offer:
  - endpoint: PATCH /private_offers/{{ record.id }}/
  - required fields: id
  - risk: external mutation of a private offer's custom properties / external-system associations; approval required
- update_contract:
  - endpoint: PATCH /contracts/{{ record.id }}/
  - required fields: id
  - risk: external mutation of a contract's custom properties / external-system associations; approval required
- activate_contract:
  - endpoint: POST /contracts/{{ record.id }}/activate/
  - required fields: id
  - risk: irreversibly activates a pending Clazar contract in the underlying cloud marketplace; approval required (destructive/high-impact state transition)
- create_contact:
  - endpoint: POST /contacts/
  - risk: creates a new Clazar contact record; low-risk (no external marketplace side effects)
- update_contact:
  - endpoint: PATCH /contacts/{{ record.id }}/
  - required fields: id
  - risk: updates a Clazar contact record; low-risk (no external marketplace side effects)
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}/
  - required fields: id
  - risk: permanently deletes a Clazar contact record; approval required (destructive, irreversible)
- update_metering_record:
  - endpoint: PATCH /metering/{{ record.id }}/
  - required fields: id, custom_properties
  - risk: updates only the custom_properties of a submitted metering record; low-risk
- create_metering_records:
  - endpoint: POST /metering/
  - required fields: request
  - risk: submits usage-based billing metering records that drive cloud marketplace invoicing for the buyer's contract; approval required (financial impact, effectively irreversible once billed)

## Security

- read risk: external Clazar API read of cloud marketplace GTM data
- write risk: external mutation of Clazar buyer/opportunity/contract/private-offer/contact/metering-record data; activate_contract irreversibly transitions a contract's state in the underlying cloud marketplace, and create_metering_records submits usage data that drives marketplace billing — every write ships with an explicit per-action risk string
- approval: required for activate_contract, delete_contact, and create_metering_records (financial/state-transition impact); custom_properties/external_object_associations updates and contact create/update are low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect clazar
```

### Inspect as structured JSON

```bash
pm connectors inspect clazar --json
```

## Agent Rules

- Run pm connectors inspect clazar before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
