# pm connectors inspect flexport

```text
NAME
  pm connectors inspect flexport - Flexport connector manual

SYNOPSIS
  pm connectors inspect flexport
  pm connectors inspect flexport --json
  pm credentials add <name> --connector flexport [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Flexport logistics, network, billing, booking, purchase order, product, document, port, and webhook-event data through the Flexport REST API; writes supported JSON create/update actions.

ICON
  asset: icons/flexport.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.flexport.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  page_size
  api_key (secret)

ETL STREAMS
  booking_line_items:
    primary key: id
    fields: _object(), id(), units()
  bookings:
    primary key: id
    fields: _object(), created_at(), id(), name(), status(), updated_at()
  commercial_invoices:
    primary key: id
    fields: _object(), digitization_status(), id(), invoice_number(), updated_at()
  customs_entries:
    primary key: id
    fields: _object(), entry_number(), id(), status()
  documents:
    primary key: id
    fields: _object(), archived_at(), document_type(), file_link(), file_name(), id(), memo()
  events:
    primary key: id
    fields: _object(), created_at(), id(), occurred_at(), type(), version()
  invoices:
    primary key: id
    cursor: updated_at
    fields: _object(), created_at(), currency(), due_date(), id(), invoice_number(), issued_date(), status(), total(), updated_at()
  companies:
    primary key: id
    cursor: updated_at
    fields: _object(), created_at(), dba_name(), id(), legal_name(), name(), updated_at()
  company_entities:
    primary key: id
    fields: _object(), id(), name(), ref()
  contacts:
    primary key: id
    fields: _object(), email(), id(), name(), phone_number()
  locations:
    primary key: id
    cursor: updated_at
    fields: _object(), city(), country_code(), created_at(), id(), name(), state(), street_address(), updated_at(), zip()
  my_company:
    primary key: id
    fields: _object(), editable(), id(), name(), ref()
  container_legs:
    primary key: id
    fields: _object(), id()
  containers:
    primary key: id
    fields: _object(), container_number(), container_size(), container_type(), id()
  ports:
    primary key: id
    fields: _object(), country_code(), id(), name(), port_name(), port_type()
  products:
    primary key: id
    cursor: updated_at
    fields: _object(), country_of_origin(), created_at(), description(), hts_code(), id(), name(), sku(), updated_at()
  purchase_order_line_items:
    primary key: id
    fields: _object(), id(), item_key(), line_item_number(), units()
  purchase_orders:
    primary key: id
    fields: _object(), created_at(), id(), name(), status(), updated_at()
  shipment_legs:
    primary key: id
    fields: _object(), id(), transportation_mode()
  shipments:
    primary key: id
    cursor: updated_at
    fields: _object(), created_at(), destination_port(), estimated_arrival_date(), estimated_departure_date(), freight_type(), id(), origin_port(), status(), transportation_mode(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_booking_amendment:
    endpoint: POST /booking_amendments
    optional fields: booking_id, new_name, amendment_note, new_container_counts, new_wants_pickup_service, new_wants_import_customs_service, new_wants_flexport_freight, new_wants_bco, new_wants_214_filing, new_wants_ftz_entry, new_origin_address_ref, new_origin_port_us_cbp_port_code, new_origin_port_loc_code, new_destination_address_ref, new_destination_port_us_cbp_port_code, new_destination_port_loc_code, new_cargo_ready_date, new_delivery_date, new_product_descriptions, new_cargo, new_metadata, new_container_references
    risk: requests a booking amendment; Flexport may apply the change immediately or queue it for approval depending on shipment state
  create_booking_line_item:
    endpoint: POST /booking_line_items
    optional fields: purchase_order_line_item_id, booking_id, units
    risk: adds units from a purchase-order line item to a booking
  create_booking:
    endpoint: POST /bookings
    optional fields: name, shipper_entity_ref, consignee_entity_ref, notify_party, ocean_booking, air_booking, trucking_booking, origin_address_ref, destination_address_ref, cargo_ready_date, delivery_date, wants_export_customs_service, wants_flexport_freight, wants_import_customs_service, wants_bco, cargo, special_instructions, metadata, declared_as_strategy, eccn_codes, flow_direct, user_email
    risk: creates a real Flexport booking request and can initiate operational freight workflows
  create_document:
    endpoint: POST /documents
    optional fields: file_name, mime_type, document_type, memo, document, user_email, shipment_id
    risk: uploads a base64-encoded document to a shipment; incorrect documents can affect operational shipment records
  create_company:
    endpoint: POST /network/companies
    optional fields: name, ref
    risk: creates a company in the Flexport network
  update_company:
    endpoint: PATCH /network/companies/{{ record.id }}
    required fields: id
    optional fields: name, ref
    risk: updates company name or external reference in the Flexport network
  create_company_entity:
    endpoint: POST /network/company_entities
    optional fields: name, company_id, company_ref, mailing_address, ref, vat_numbers
    risk: creates a legal company entity under a Flexport network company
  update_company_entity:
    endpoint: PATCH /network/company_entities/{{ record.id }}
    required fields: id
    optional fields: name, mailing_address, ref, vat_numbers
    risk: updates company entity legal name, mailing address, reference, or VAT numbers
  create_contact:
    endpoint: POST /network/contacts
    optional fields: name, email, phone_number, company_id
    risk: creates a contact in the Flexport network
  update_contact:
    endpoint: PATCH /network/contacts/{{ record.id }}
    required fields: id
    optional fields: name, email, phone_number
    risk: updates contact details in the Flexport network
  create_location:
    endpoint: POST /network/locations
    optional fields: name, company_id, address, contact_ids, ref, metadata
    risk: creates a network location and address in Flexport
  update_location:
    endpoint: PATCH /network/locations/{{ record.id }}
    required fields: id
    optional fields: name, address, contact_ids, ref, metadata
    risk: updates location identity, address, contacts, reference, or metadata
  create_product:
    endpoint: POST /products
    optional fields: name, sku, description, product_category, country_of_origin, client_verified, product_properties, classifications, suppliers
    risk: creates a product in the Flexport Product Library
  update_product:
    endpoint: PATCH /products/{{ record.id }}
    required fields: id
    optional fields: name, sku, description, product_category, country_of_origin, client_verified, product_properties, classifications, suppliers
    risk: updates product library fields; arrays replace the existing values when provided
  update_shipment:
    endpoint: PATCH /shipments/{{ record.id }}
    required fields: id
    optional fields: metadata
    risk: replaces shipment metadata tags; incorrect metadata can affect downstream shipment workflows
  create_shipments_shareable:
    endpoint: POST /shipments_shareable
    optional fields: shipment_ids
    risk: creates shareable shipment URLs for the listed shipment ids

SECURITY
  read risk: external Flexport API read of freight/logistics, network, purchase order, booking, document, webhook event, and billing data
  write risk: creates or updates Flexport bookings, booking amendments, booking line items, network companies/entities/contacts/locations, products, shipment metadata, shipment share links, and base64 JSON document uploads
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect flexport

  # Inspect as structured JSON
  pm connectors inspect flexport --json

AGENT WORKFLOW
  - Run pm connectors inspect flexport before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
