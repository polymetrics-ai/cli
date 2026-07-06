# Overview

Reads and writes Track PMS reservations, guests, units, owners, CRM contacts, and unit types through
the Track PMS API.

Readable streams: `reservations`, `guests`, `units`, `owners`, `contacts`, `unit_types`.

Write actions: `create_reservation`, `create_unit`, `update_unit`, `create_owner`, `update_owner`,
`create_contact`, `update_contact`, `delete_contact`.

Service API documentation: https://developer.trackhs.com/reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Track PMS API access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.trackhs.com`; format `uri`; Track PMS API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.trackhs.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/reservations`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100; maximum 1 page(s).

- `reservations`: GET `/reservations` - records path `reservations`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed
  output fields `arrival_date`, `confirmation_number`.
- `guests`: GET `/guests` - records path `guests`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed output fields
  `name`.
- `units`: GET `/units` - records path `units`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed output fields `name`.
- `owners`: GET `/owners` - records path `owners`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; maximum 1 page(s); computed output fields
  `name`.
- `contacts`: GET `/crm/contacts` - records path `_embedded.contacts`; page-number pagination; page
  parameter `page`; size parameter `size`; starts at 1; page size 100; computed output fields
  `cell_phone`, `created_at`, `first_name`, `home_phone`, `is_owner_contact`, `is_vip`, `last_name`,
  `postal_code`, `primary_email`, `secondary_email`, `street_address`, `updated_at`, `work_phone`.
- `unit_types`: GET `/pms/units/types` - records path `_embedded.units`; page-number pagination;
  page parameter `page`; size parameter `size`; starts at 1; page size 100; computed output fields
  `created_at`, `is_active`, `is_bookable`, `lodging_type_id`, `max_occupancy`, `node_id`,
  `short_name`, `type_code`, `updated_at`.

## Write actions & risks

Overall write risk: external mutation of Track PMS reservations, units, owners, and CRM contacts; no
destructive-admin or elevated-scope actions modeled.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_reservation`: POST `/pms/reservations` - kind `create`; body type `json`; required record
  fields `unitId`, `arrivalDate`, `departureDate`; accepted fields `arrivalDate`, `contactId`,
  `departureDate`, `guaranteePolicyId`, `isTaxable`, `notes`, `occupants`, `rateTypeId`,
  `reservationTypeId`, `status`, `unitId`; risk: creates a new guest reservation and blocks the
  unit's availability for the given date range; external mutation, approval required.
- `create_unit`: POST `/pms/units` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `longDescription`, `maxPets`, `minimumAgeLimit`, `name`, `nodeId`, `notes`,
  `phone`, `shortDescription`, `shortName`, `unitTypeId`, `websiteUrl`; risk: creates a new rentable
  unit/property record; external mutation, approval required.
- `update_unit`: PUT `/pms/units/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `name`, `shortName`, `shortDescription`, `longDescription`, `notes`, `maxPets`,
  `minimumAgeLimit`, `phone`, `websiteUrl`, `nodeId`, and 1 more; required record fields `id`;
  accepted fields `id`, `longDescription`, `maxPets`, `minimumAgeLimit`, `name`, `nodeId`, `notes`,
  `phone`, `shortDescription`, `shortName`, `unitTypeId`, `websiteUrl`; risk: mutates an existing
  unit's descriptive/configuration fields; a changed nodeId or unitTypeId affects rate/availability
  grouping for future reservations.
- `create_owner`: POST `/pms/owners` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `country`, `email`, `isActive`, `locality`, `name`, `notes`, `phone`,
  `postal`, `region`, `streetAddress`; risk: creates a new property owner record; external mutation,
  approval required.
- `update_owner`: PATCH `/pms/owners/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `name`, `isActive`, `streetAddress`, `locality`, `region`, `postal`, `country`,
  `phone`, `email`, `notes`; required record fields `id`; accepted fields `country`, `email`, `id`,
  `isActive`, `locality`, `name`, `notes`, `phone`, `postal`, `region`, `streetAddress`; risk:
  mutates an existing owner's contact/status fields; setting isActive:false affects that owner's
  active-unit reporting.
- `create_contact`: POST `/crm/contacts` - kind `create`; body type `json`; required record fields
  `firstName`, `lastName`; accepted fields `cellPhone`, `country`, `firstName`, `homePhone`,
  `isVip`, `lastName`, `locality`, `notes`, `postalCode`, `primaryEmail`, `region`,
  `secondaryEmail`, `streetAddress`; risk: creates a new CRM contact (guest, lead, or owner-linked
  person record); external mutation, approval required. Tremendous-adjacent restricted fields
  (taxId, paymentType, ACH banking fields) are not modeled - see docs.md Known limits.
- `update_contact`: PATCH `/crm/contacts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `firstName`, `lastName`, `primaryEmail`, `secondaryEmail`, `homePhone`,
  `cellPhone`, `streetAddress`, `locality`, `region`, `postalCode`, and 3 more; required record
  fields `id`; accepted fields `cellPhone`, `country`, `firstName`, `homePhone`, `id`, `isVip`,
  `lastName`, `locality`, `notes`, `postalCode`, `primaryEmail`, `region`, `secondaryEmail`,
  `streetAddress`; risk: mutates an existing CRM contact's identity/contact-method fields.
- `delete_contact`: DELETE `/crm/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently deletes a CRM contact; irreversible, and may disassociate the
  contact from any reservations that reference it.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, destructive_admin=1, duplicate_of=6, out_of_scope=206,
  requires_elevated_scope=46.
