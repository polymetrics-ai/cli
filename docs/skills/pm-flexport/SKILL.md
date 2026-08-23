---
name: pm-flexport
description: Flexport connector knowledge and safe action guide.
---

# pm-flexport

## Purpose

Reads Flexport logistics, network, billing, booking, purchase order, product, document, port, and webhook-event data through the Flexport REST API; writes supported JSON create/update actions.

## Icon

- id: flexport
- asset: icons/flexport.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.flexport.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- api_key (secret) (required)

## ETL Streams

- booking_line_items:
  - primary key: id
  - fields: _object(string), id(string), units(integer)
- bookings:
  - primary key: id
  - fields: _object(string), created_at(string), id(string), name(string), status(string), updated_at(string)
- commercial_invoices:
  - primary key: id
  - fields: _object(string), digitization_status(string), id(string), invoice_number(string), updated_at(string)
- customs_entries:
  - primary key: id
  - fields: _object(string), entry_number(string), id(string), status(string)
- documents:
  - primary key: id
  - fields: _object(string), archived_at(string), document_type(string), file_link(string), file_name(string), id(string), memo(string)
- events:
  - primary key: id
  - fields: _object(string), created_at(string), id(string), occurred_at(string), type(string), version(string)
- invoices:
  - primary key: id
  - cursor: updated_at
  - fields: _object(string), created_at(string), currency(string), due_date(string), id(string), invoice_number(string), issued_date(string), status(string), total(string), updated_at(string)
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: _object(string), created_at(string), dba_name(string), id(string), legal_name(string), name(string), updated_at(string)
- company_entities:
  - primary key: id
  - fields: _object(string), id(string), name(string), ref(string)
- contacts:
  - primary key: id
  - fields: _object(string), email(string), id(string), name(string), phone_number(string)
- locations:
  - primary key: id
  - cursor: updated_at
  - fields: _object(string), city(string), country_code(string), created_at(string), id(string), name(string), state(string), street_address(string), updated_at(string), zip(string)
- my_company:
  - primary key: id
  - fields: _object(string), editable(boolean), id(string), name(string), ref(string)
- container_legs:
  - primary key: id
  - fields: _object(string), id(string)
- containers:
  - primary key: id
  - fields: _object(string), container_number(string), container_size(string), container_type(string), id(string)
- ports:
  - primary key: id
  - fields: _object(string), country_code(string), id(string), name(string), port_name(string), port_type(string)
- products:
  - primary key: id
  - cursor: updated_at
  - fields: _object(string), country_of_origin(string), created_at(string), description(string), hts_code(string), id(string), name(string), sku(string), updated_at(string)
- purchase_order_line_items:
  - primary key: id
  - fields: _object(string), id(string), item_key(string), line_item_number(integer), units(integer)
- purchase_orders:
  - primary key: id
  - fields: _object(string), created_at(string), id(string), name(string), status(string), updated_at(string)
- shipment_legs:
  - primary key: id
  - fields: _object(string), id(string), transportation_mode(string)
- shipments:
  - primary key: id
  - cursor: updated_at
  - fields: _object(string), created_at(string), destination_port(string), estimated_arrival_date(string), estimated_departure_date(string), freight_type(string), id(string), origin_port(string), status(string), transportation_mode(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_booking_amendment:
  - endpoint: POST /booking_amendments
  - required fields: booking_id
  - optional fields: new_name, amendment_note, new_container_counts, new_wants_pickup_service, new_wants_import_customs_service, new_wants_flexport_freight, new_wants_bco, new_wants_214_filing, new_wants_ftz_entry, new_origin_address_ref, new_origin_port_us_cbp_port_code, new_origin_port_loc_code, new_destination_address_ref, new_destination_port_us_cbp_port_code, new_destination_port_loc_code, new_cargo_ready_date, new_delivery_date, new_product_descriptions, new_cargo, new_metadata, new_container_references
  - risk: requests a booking amendment; Flexport may apply the change immediately or queue it for approval depending on shipment state
- create_booking_line_item:
  - endpoint: POST /booking_line_items
  - required fields: purchase_order_line_item_id, booking_id, units
  - risk: adds units from a purchase-order line item to a booking
- create_booking:
  - endpoint: POST /bookings
  - required fields: name, shipper_entity_ref, consignee_entity_ref, origin_address_ref, destination_address_ref, cargo_ready_date, wants_export_customs_service, cargo
  - optional fields: notify_party, ocean_booking, air_booking, trucking_booking, delivery_date, wants_flexport_freight, wants_import_customs_service, wants_bco, special_instructions, metadata, declared_as_strategy, eccn_codes, flow_direct, user_email
  - risk: creates a real Flexport booking request and can initiate operational freight workflows
- create_document:
  - endpoint: POST /documents
  - required fields: file_name, mime_type, document_type, document, shipment_id
  - optional fields: memo, user_email
  - risk: uploads a base64-encoded document to a shipment; incorrect documents can affect operational shipment records
- create_company:
  - endpoint: POST /network/companies
  - required fields: name
  - optional fields: ref
  - risk: creates a company in the Flexport network
- update_company:
  - endpoint: PATCH /network/companies/{{ record.id }}
  - required fields: id
  - optional fields: name, ref
  - risk: updates company name or external reference in the Flexport network
- create_company_entity:
  - endpoint: POST /network/company_entities
  - required fields: name, mailing_address
  - optional fields: company_id, company_ref, ref, vat_numbers
  - risk: creates a legal company entity under a Flexport network company
- update_company_entity:
  - endpoint: PATCH /network/company_entities/{{ record.id }}
  - required fields: id
  - optional fields: name, mailing_address, ref, vat_numbers
  - risk: updates company entity legal name, mailing address, reference, or VAT numbers
- create_contact:
  - endpoint: POST /network/contacts
  - required fields: name, email, phone_number
  - optional fields: company_id
  - risk: creates a contact in the Flexport network
- update_contact:
  - endpoint: PATCH /network/contacts/{{ record.id }}
  - required fields: id
  - optional fields: name, email, phone_number
  - risk: updates contact details in the Flexport network
- create_location:
  - endpoint: POST /network/locations
  - required fields: name, company_id, address
  - optional fields: contact_ids, ref, metadata
  - risk: creates a network location and address in Flexport
- update_location:
  - endpoint: PATCH /network/locations/{{ record.id }}
  - required fields: id
  - optional fields: name, address, contact_ids, ref, metadata
  - risk: updates location identity, address, contacts, reference, or metadata
- create_product:
  - endpoint: POST /products
  - required fields: name, sku
  - optional fields: description, product_category, country_of_origin, client_verified, product_properties, classifications, suppliers
  - risk: creates a product in the Flexport Product Library
- update_product:
  - endpoint: PATCH /products/{{ record.id }}
  - required fields: id
  - optional fields: name, sku, description, product_category, country_of_origin, client_verified, product_properties, classifications, suppliers
  - risk: updates product library fields; arrays replace the existing values when provided
- update_shipment:
  - endpoint: PATCH /shipments/{{ record.id }}
  - required fields: id, metadata
  - risk: replaces shipment metadata tags; incorrect metadata can affect downstream shipment workflows
- create_shipments_shareable:
  - endpoint: POST /shipments_shareable
  - required fields: shipment_ids
  - risk: creates shareable shipment URLs for the listed shipment ids

## Security

- read risk: external Flexport API read of freight/logistics, network, purchase order, booking, document, webhook event, and billing data
- write risk: creates or updates Flexport bookings, booking amendments, booking line items, network companies/entities/contacts/locations, products, shipment metadata, shipment share links, and base64 JSON document uploads
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect flexport
```

### Inspect as structured JSON

```bash
pm connectors inspect flexport --json
```

## Agent Rules

- Run pm connectors inspect flexport before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
