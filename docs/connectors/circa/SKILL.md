---
name: pm-circa
description: Circa connector knowledge and safe action guide.
---

# pm-circa

## Purpose

Reads and writes Circa events, contacts, companies, teams, custom fields, and event/company sub-resources through the Circa REST API.

## Icon

- asset: icons/circa.svg
- source: official
- review_status: official_verified
- review_url: https://docs.circa.co/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- api_key (secret)

## ETL Streams

- events:
  - primary key: id
  - cursor: updated_at
  - fields: actual_total(), brief_url(), created_at(), id(), name(), paid_total(), planned_total(), status(), time_zone(), updated_at(), website()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: company(), created_at(), email(), first_name(), id(), last_name(), updated_at()
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), created_method(), email_opt_in(), id(), name(), sync_status(), updated_at(), updated_method()
- teams:
  - primary key: id
  - fields: created_at(), created_by(), id(), name()
- fields:
  - primary key: id
  - fields: field_for(), field_name(), field_type(), id(), label(), options(), order(), required(), section()
- event_contacts:
  - primary key: event_id, id
  - fields: address(), check_in_status(), city(), company(), contact_type(), country(), created_at(), created_by(), created_method(), description(), email(), email_opt_in(), event_id(), first_name(), hot_lead(), id(), last_name(), linkedin(), mobile_phone(), office_phone(), owner(), postal_index(), registration_status(), state(), title(), twitter(), updated_at(), updated_by(), updated_method(), website()
- event_staff:
  - primary key: event_id, email
  - fields: based(), custom_fields(), email(), event_id(), first_name(), last_name()
- event_expenses:
  - primary key: event_id, id
  - fields: actual_amount(), budget_category(), event_id(), id(), name(), note(), paid_amount(), team_allocations()
- company_contacts:
  - primary key: company_id, id
  - fields: address(), check_in_status(), city(), company(), company_id(), contact_type(), country(), created_at(), created_by(), created_method(), description(), email(), email_opt_in(), first_name(), hot_lead(), id(), last_name(), linkedin(), mobile_phone(), office_phone(), owner(), postal_index(), registration_status(), state(), title(), twitter(), updated_at(), updated_by(), updated_method(), website()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - risk: external mutation; creates a new CRM contact record
- update_contact:
  - endpoint: PATCH /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing CRM contact record
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a CRM contact; approval required
- create_event:
  - endpoint: POST /events
  - risk: external mutation; creates a new event record
- update_event:
  - endpoint: PATCH /events/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing event record
- delete_event:
  - endpoint: DELETE /events/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of an event; approval required
- create_company:
  - endpoint: POST /companies
  - risk: external mutation; creates a new company record
- update_company:
  - endpoint: PATCH /companies/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing company record
- delete_company:
  - endpoint: DELETE /companies/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a company; approval required
- add_event_contact:
  - endpoint: POST /events/{{ record.event_id }}/contacts
  - required fields: event_id
  - risk: external mutation; registers an existing contact onto an event
- update_event_contact:
  - endpoint: PATCH /events/{{ record.event_id }}/contacts/{{ record.contact_id }}
  - required fields: event_id, contact_id
  - risk: external mutation; updates an event-contact's attendance/registration status
- remove_event_contact:
  - endpoint: DELETE /events/{{ record.event_id }}/contacts/{{ record.contact_id }}
  - required fields: event_id, contact_id
  - risk: irreversible external removal of a contact's event registration; approval required

## Security

- read risk: external Circa API read of event, contact, company, team, custom-field, and event/company sub-resource data
- write risk: external mutation of Circa contacts, events, companies, and event-contact registrations; create/update/delete affect live CRM/event data an operator relies on
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect circa
```

### Inspect as structured JSON

```bash
pm connectors inspect circa --json
```

## Agent Rules

- Run pm connectors inspect circa before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
