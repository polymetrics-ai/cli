# pm connectors inspect track-pms

```text
NAME
  pm connectors inspect track-pms - Track PMS connector manual

SYNOPSIS
  pm connectors inspect track-pms
  pm connectors inspect track-pms --json
  pm credentials add <name> --connector track-pms [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Track PMS reservations, guests, units, owners, CRM contacts, and unit types through the Track PMS API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_token (secret) (required)

ETL STREAMS
  reservations:
    primary key: id
    cursor: arrival_date
    fields: arrival_date(string), confirmation_number(string), id(string), status(string)
  guests:
    primary key: id
    fields: id(string), name(string), status(string)
  units:
    primary key: id
    fields: id(string), name(string), status(string)
  owners:
    primary key: id
    fields: id(string), name(string), status(string)
  contacts:
    primary key: id
    fields: cell_phone(string), country(string), created_at(string), first_name(string), home_phone(string), id(integer), is_owner_contact(boolean), is_vip(boolean), last_name(string), locality(string), name(string), notes(string), postal_code(string), primary_email(string), region(string), secondary_email(string), street_address(string), updated_at(string), work_phone(string)
  unit_types:
    primary key: id
    fields: bedrooms(integer), created_at(string), id(integer), is_active(boolean), is_bookable(boolean), lodging_type_id(integer), max_occupancy(integer), name(string), node_id(integer), short_name(string), type_code(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_reservation:
    endpoint: POST /pms/reservations
    required fields: unitId, arrivalDate, departureDate
    risk: creates a new guest reservation and blocks the unit's availability for the given date range; external mutation, approval required
  create_unit:
    endpoint: POST /pms/units
    required fields: name
    risk: creates a new rentable unit/property record; external mutation, approval required
  update_unit:
    endpoint: PUT /pms/units/{{ record.id }}
    required fields: id
    optional fields: name, shortName, shortDescription, longDescription, notes, maxPets, minimumAgeLimit, phone, websiteUrl, nodeId, unitTypeId
    risk: mutates an existing unit's descriptive/configuration fields; a changed nodeId or unitTypeId affects rate/availability grouping for future reservations
  create_owner:
    endpoint: POST /pms/owners
    required fields: name
    risk: creates a new property owner record; external mutation, approval required
  update_owner:
    endpoint: PATCH /pms/owners/{{ record.id }}
    required fields: id
    optional fields: name, isActive, streetAddress, locality, region, postal, country, phone, email, notes
    risk: mutates an existing owner's contact/status fields; setting isActive:false affects that owner's active-unit reporting
  create_contact:
    endpoint: POST /crm/contacts
    required fields: firstName, lastName
    risk: creates a new CRM contact (guest, lead, or owner-linked person record); external mutation, approval required. Tremendous-adjacent restricted fields (taxId, paymentType, ACH banking fields) are not modeled — see docs.md Known limits
  update_contact:
    endpoint: PATCH /crm/contacts/{{ record.id }}
    required fields: id
    optional fields: firstName, lastName, primaryEmail, secondaryEmail, homePhone, cellPhone, streetAddress, locality, region, postalCode, country, notes, isVip
    risk: mutates an existing CRM contact's identity/contact-method fields
  delete_contact:
    endpoint: DELETE /crm/contacts/{{ record.id }}
    required fields: id
    risk: permanently deletes a CRM contact; irreversible, and may disassociate the contact from any reservations that reference it

SECURITY
  read risk: external Track PMS API read of reservation, guest, unit, owner, CRM contact, and unit type data
  write risk: external mutation of Track PMS reservations, units, owners, and CRM contacts; no destructive-admin or elevated-scope actions modeled
  approval: create_reservation, create_unit, and create_owner require approval (each creates a billable/bookable real-world resource record); update_unit, update_owner, create_contact, update_contact, and delete_contact execute without approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect track-pms

  # Inspect as structured JSON
  pm connectors inspect track-pms --json

AGENT WORKFLOW
  - Run pm connectors inspect track-pms before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
