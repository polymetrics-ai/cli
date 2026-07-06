---
name: pm-track-pms
description: Track PMS connector knowledge and safe action guide.
---

# pm-track-pms

## Purpose

Reads and writes Track PMS reservations, guests, units, owners, CRM contacts, and unit types through the Track PMS API.

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
- access_token (secret)

## ETL Streams

- reservations:
  - primary key: id
  - cursor: arrival_date
  - fields: arrival_date(), confirmation_number(), id(), status()
- guests:
  - primary key: id
  - fields: id(), name(), status()
- units:
  - primary key: id
  - fields: id(), name(), status()
- owners:
  - primary key: id
  - fields: id(), name(), status()
- contacts:
  - primary key: id
  - fields: cell_phone(), country(), created_at(), first_name(), home_phone(), id(), is_owner_contact(), is_vip(), last_name(), locality(), name(), notes(), postal_code(), primary_email(), region(), secondary_email(), street_address(), updated_at(), work_phone()
- unit_types:
  - primary key: id
  - fields: bedrooms(), created_at(), id(), is_active(), is_bookable(), lodging_type_id(), max_occupancy(), name(), node_id(), short_name(), type_code(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_reservation:
  - endpoint: POST /pms/reservations
  - risk: creates a new guest reservation and blocks the unit's availability for the given date range; external mutation, approval required
- create_unit:
  - endpoint: POST /pms/units
  - risk: creates a new rentable unit/property record; external mutation, approval required
- update_unit:
  - endpoint: PUT /pms/units/{{ record.id }}
  - required fields: id
  - optional fields: name, shortName, shortDescription, longDescription, notes, maxPets, minimumAgeLimit, phone, websiteUrl, nodeId, unitTypeId
  - risk: mutates an existing unit's descriptive/configuration fields; a changed nodeId or unitTypeId affects rate/availability grouping for future reservations
- create_owner:
  - endpoint: POST /pms/owners
  - risk: creates a new property owner record; external mutation, approval required
- update_owner:
  - endpoint: PATCH /pms/owners/{{ record.id }}
  - required fields: id
  - optional fields: name, isActive, streetAddress, locality, region, postal, country, phone, email, notes
  - risk: mutates an existing owner's contact/status fields; setting isActive:false affects that owner's active-unit reporting
- create_contact:
  - endpoint: POST /crm/contacts
  - risk: creates a new CRM contact (guest, lead, or owner-linked person record); external mutation, approval required. Tremendous-adjacent restricted fields (taxId, paymentType, ACH banking fields) are not modeled — see docs.md Known limits
- update_contact:
  - endpoint: PATCH /crm/contacts/{{ record.id }}
  - required fields: id
  - optional fields: firstName, lastName, primaryEmail, secondaryEmail, homePhone, cellPhone, streetAddress, locality, region, postalCode, country, notes, isVip
  - risk: mutates an existing CRM contact's identity/contact-method fields
- delete_contact:
  - endpoint: DELETE /crm/contacts/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a CRM contact; irreversible, and may disassociate the contact from any reservations that reference it

## Security

- read risk: external Track PMS API read of reservation, guest, unit, owner, CRM contact, and unit type data
- write risk: external mutation of Track PMS reservations, units, owners, and CRM contacts; no destructive-admin or elevated-scope actions modeled
- approval: create_reservation, create_unit, and create_owner require approval (each creates a billable/bookable real-world resource record); update_unit, update_owner, create_contact, update_contact, and delete_contact execute without approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect track-pms
```

### Inspect as structured JSON

```bash
pm connectors inspect track-pms --json
```

## Agent Rules

- Run pm connectors inspect track-pms before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
