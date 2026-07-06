# Overview

Reads Flexport logistics, network, billing, booking, purchase order, product, document, port, and
webhook-event data through the Flexport REST API; writes supported JSON create/update actions.

Readable streams: `booking_line_items`, `bookings`, `commercial_invoices`, `customs_entries`,
`documents`, `events`, `invoices`, `companies`, `company_entities`, `contacts`, `locations`,
`my_company`, `container_legs`, `containers`, `ports`, `products`, `purchase_order_line_items`,
`purchase_orders`, `shipment_legs`, `shipments`.

Write actions: `create_booking_amendment`, `create_booking_line_item`, `create_booking`,
`create_document`, `create_company`, `update_company`, `create_company_entity`,
`update_company_entity`, `create_contact`, `update_contact`, `create_location`, `update_location`,
`create_product`, `update_product`, `update_shipment`, `create_shipments_shareable`.

Service API documentation: https://developers.flexport.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Flexport API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.flexport.com`; format `uri`; Flexport API base
  URL override for tests or proxies.
- `page_size` (optional, string); default `100`; Records per page, sent as the 'per' query param on
  list requests. Flexport v2 documents a maximum of 100.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.flexport.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products` with query `per`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `data.next`; next URLs
stay on the configured API host.

Pagination by stream: next_url: `booking_line_items`, `bookings`, `commercial_invoices`,
`customs_entries`, `documents`, `events`, `invoices`, `companies`, `company_entities`, `contacts`,
`locations`, `container_legs`, `containers`, `ports`, `products`, `purchase_order_line_items`,
`purchase_orders`, `shipment_legs`, `shipments`; none: `my_company`.

- `booking_line_items`: GET `/booking_line_items` - records path `data.data`; query `per` from
  template `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body;
  URL path `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `bookings`: GET `/bookings` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `commercial_invoices`: GET `/commercial_invoices` - records path `data.data`; query `per` from
  template `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body;
  URL path `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `customs_entries`: GET `/customs_entries` - records path `data.data`; query `per` from template
  `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `documents`: GET `/documents` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `events`: GET `/events` - records path `data.data`; query `per` from template `{{ config.page_size
  }}`, default `100`; follows a next-page URL from the response body; URL path `data.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `invoices`: GET `/invoices` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host.
- `companies`: GET `/companies` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host.
- `company_entities`: GET `/network/company_entities` - records path `data.data`; query `per` from
  template `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body;
  URL path `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `contacts`: GET `/network/contacts` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `locations`: GET `/locations` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host.
- `my_company`: GET `/network/me/companies` - single-object response; records path `data`; emits
  passthrough records.
- `container_legs`: GET `/ocean/shipment_container_legs` - records path `data.data`; query `per`
  from template `{{ config.page_size }}`, default `100`; follows a next-page URL from the response
  body; URL path `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `containers`: GET `/ocean/shipment_containers` - records path `data.data`; query `per` from
  template `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body;
  URL path `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `ports`: GET `/ports` - records path `data.data`; query `per` from template `{{ config.page_size
  }}`, default `100`; follows a next-page URL from the response body; URL path `data.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `products`: GET `/products` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host.
- `purchase_order_line_items`: GET `/purchase_order_line_items` - records path `data.data`; query
  `per` from template `{{ config.page_size }}`, default `100`; follows a next-page URL from the
  response body; URL path `data.next`; next URLs stay on the configured API host; emits passthrough
  records.
- `purchase_orders`: GET `/purchase_orders` - records path `data.data`; query `per` from template
  `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `shipment_legs`: GET `/shipment_legs` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host; emits passthrough records.
- `shipments`: GET `/shipments` - records path `data.data`; query `per` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `data.next`; next URLs stay on the configured API host.

## Write actions & risks

Overall write risk: creates or updates Flexport bookings, booking amendments, booking line items,
network companies/entities/contacts/locations, products, shipment metadata, shipment share links,
and base64 JSON document uploads.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_booking_amendment`: POST `/booking_amendments` - kind `create`; body type `json`; body
  fields `booking_id`, `new_name`, `amendment_note`, `new_container_counts`,
  `new_wants_pickup_service`, `new_wants_import_customs_service`, `new_wants_flexport_freight`,
  `new_wants_bco`, `new_wants_214_filing`, `new_wants_ftz_entry`, and 12 more; required record
  fields `booking_id`; accepted fields `amendment_note`, `booking_id`, `new_cargo`,
  `new_cargo_ready_date`, `new_container_counts`, `new_container_references`, `new_delivery_date`,
  `new_destination_address_ref`, `new_destination_port_loc_code`,
  `new_destination_port_us_cbp_port_code`, `new_metadata`, `new_name`, `new_origin_address_ref`,
  `new_origin_port_loc_code`, `new_origin_port_us_cbp_port_code`, `new_product_descriptions`,
  `new_wants_214_filing`, `new_wants_bco`, and 4 more; risk: requests a booking amendment; Flexport
  may apply the change immediately or queue it for approval depending on shipment state.
- `create_booking_line_item`: POST `/booking_line_items` - kind `create`; body type `json`; body
  fields `purchase_order_line_item_id`, `booking_id`, `units`; required record fields
  `purchase_order_line_item_id`, `booking_id`, `units`; accepted fields `booking_id`,
  `purchase_order_line_item_id`, `units`; risk: adds units from a purchase-order line item to a
  booking.
- `create_booking`: POST `/bookings` - kind `create`; body type `json`; body fields `name`,
  `shipper_entity_ref`, `consignee_entity_ref`, `notify_party`, `ocean_booking`, `air_booking`,
  `trucking_booking`, `origin_address_ref`, `destination_address_ref`, `cargo_ready_date`, and 12
  more; required record fields `name`, `shipper_entity_ref`, `consignee_entity_ref`,
  `origin_address_ref`, `destination_address_ref`, `cargo_ready_date`,
  `wants_export_customs_service`, `cargo`; accepted fields `air_booking`, `cargo`,
  `cargo_ready_date`, `consignee_entity_ref`, `declared_as_strategy`, `delivery_date`,
  `destination_address_ref`, `eccn_codes`, `flow_direct`, `metadata`, `name`, `notify_party`,
  `ocean_booking`, `origin_address_ref`, `shipper_entity_ref`, `special_instructions`,
  `trucking_booking`, `user_email`, and 4 more; risk: creates a real Flexport booking request and
  can initiate operational freight workflows.
- `create_document`: POST `/documents` - kind `create`; body type `json`; body fields `file_name`,
  `mime_type`, `document_type`, `memo`, `document`, `user_email`, `shipment_id`; required record
  fields `file_name`, `mime_type`, `document_type`, `document`, `shipment_id`; accepted fields
  `document`, `document_type`, `file_name`, `memo`, `mime_type`, `shipment_id`, `user_email`; risk:
  uploads a base64-encoded document to a shipment; incorrect documents can affect operational
  shipment records.
- `create_company`: POST `/network/companies` - kind `create`; body type `json`; body fields `name`,
  `ref`; required record fields `name`; accepted fields `name`, `ref`; risk: creates a company in
  the Flexport network.
- `update_company`: PATCH `/network/companies/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `name`, `ref`; required record fields `id`; accepted fields `id`,
  `name`, `ref`; risk: updates company name or external reference in the Flexport network.
- `create_company_entity`: POST `/network/company_entities` - kind `create`; body type `json`; body
  fields `name`, `company_id`, `company_ref`, `mailing_address`, `ref`, `vat_numbers`; required
  record fields `name`, `mailing_address`; accepted fields `company_id`, `company_ref`,
  `mailing_address`, `name`, `ref`, `vat_numbers`; risk: creates a legal company entity under a
  Flexport network company.
- `update_company_entity`: PATCH `/network/company_entities/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; body fields `name`, `mailing_address`, `ref`, `vat_numbers`;
  required record fields `id`; accepted fields `id`, `mailing_address`, `name`, `ref`,
  `vat_numbers`; risk: updates company entity legal name, mailing address, reference, or VAT
  numbers.
- `create_contact`: POST `/network/contacts` - kind `create`; body type `json`; body fields `name`,
  `email`, `phone_number`, `company_id`; required record fields `name`, `email`, `phone_number`;
  accepted fields `company_id`, `email`, `name`, `phone_number`; risk: creates a contact in the
  Flexport network.
- `update_contact`: PATCH `/network/contacts/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `name`, `email`, `phone_number`; required record fields `id`;
  accepted fields `email`, `id`, `name`, `phone_number`; risk: updates contact details in the
  Flexport network.
- `create_location`: POST `/network/locations` - kind `create`; body type `json`; body fields
  `name`, `company_id`, `address`, `contact_ids`, `ref`, `metadata`; required record fields `name`,
  `company_id`, `address`; accepted fields `address`, `company_id`, `contact_ids`, `metadata`,
  `name`, `ref`; risk: creates a network location and address in Flexport.
- `update_location`: PATCH `/network/locations/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `name`, `address`, `contact_ids`, `ref`, `metadata`; required record
  fields `id`; accepted fields `address`, `contact_ids`, `id`, `metadata`, `name`, `ref`; risk:
  updates location identity, address, contacts, reference, or metadata.
- `create_product`: POST `/products` - kind `create`; body type `json`; body fields `name`, `sku`,
  `description`, `product_category`, `country_of_origin`, `client_verified`, `product_properties`,
  `classifications`, `suppliers`; required record fields `name`, `sku`; accepted fields
  `classifications`, `client_verified`, `country_of_origin`, `description`, `name`,
  `product_category`, `product_properties`, `sku`, `suppliers`; risk: creates a product in the
  Flexport Product Library.
- `update_product`: PATCH `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `name`, `sku`, `description`, `product_category`, `country_of_origin`,
  `client_verified`, `product_properties`, `classifications`, `suppliers`; required record fields
  `id`; accepted fields `classifications`, `client_verified`, `country_of_origin`, `description`,
  `id`, `name`, `product_category`, `product_properties`, `sku`, `suppliers`; risk: updates product
  library fields; arrays replace the existing values when provided.
- `update_shipment`: PATCH `/shipments/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `metadata`; required record fields `id`, `metadata`; accepted fields
  `id`, `metadata`; risk: replaces shipment metadata tags; incorrect metadata can affect downstream
  shipment workflows.
- `create_shipments_shareable`: POST `/shipments_shareable` - kind `create`; body type `json`; body
  fields `shipment_ids`; required record fields `shipment_ids`; accepted fields `shipment_ids`;
  risk: creates shareable shipment URLs for the listed shipment ids.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 20 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=16, out_of_scope=5.
