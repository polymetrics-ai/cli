---
name: pm-circa
description: Circa connector knowledge and safe action guide.
---

# pm-circa

## Purpose

Reads and writes Circa events, contacts, companies, teams, custom fields, and event/company sub-resources through the Circa REST API.

## Icon

- id: circa
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
- api_key (secret) (required)

## ETL Streams

- events:
  - primary key: id
  - cursor: updated_at
  - fields: actual_total(number), brief_url(string), created_at(string), id(string), name(string), paid_total(number), planned_total(number), status(string), time_zone(string), updated_at(string), website(string)
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: company(object), created_at(string), email(string), first_name(string), id(string), last_name(string), updated_at(string)
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), created_method(string), email_opt_in(boolean), id(string), name(string), sync_status(object), updated_at(string), updated_method(string)
- teams:
  - primary key: id
  - fields: created_at(string), created_by(object), id(string), name(string)
- fields:
  - primary key: id
  - fields: field_for(string), field_name(string), field_type(string), id(string), label(string), options(array), order(integer), required(boolean), section(string)
- event_contacts:
  - primary key: event_id, id
  - fields: address(string), check_in_status(string), city(string), company(object), contact_type(string), country(string), created_at(string), created_by(object), created_method(string), description(string), email(string), email_opt_in(boolean), event_id(string), first_name(string), hot_lead(boolean), id(string), last_name(string), linkedin(string), mobile_phone(string), office_phone(string), owner(object), postal_index(string), registration_status(string), state(string), title(string), twitter(string), updated_at(string), updated_by(object), updated_method(string), website(string)
- event_staff:
  - primary key: event_id, email
  - fields: based(string), custom_fields(array), email(string), event_id(string), first_name(string), last_name(string)
- event_expenses:
  - primary key: event_id, id
  - fields: actual_amount(number), budget_category(string), event_id(string), id(string), name(string), note(string), paid_amount(number), team_allocations(array)
- company_contacts:
  - primary key: company_id, id
  - fields: address(string), check_in_status(string), city(string), company(object), company_id(string), contact_type(string), country(string), created_at(string), created_by(object), created_method(string), description(string), email(string), email_opt_in(boolean), first_name(string), hot_lead(boolean), id(string), last_name(string), linkedin(string), mobile_phone(string), office_phone(string), owner(object), postal_index(string), registration_status(string), state(string), title(string), twitter(string), updated_at(string), updated_by(object), updated_method(string), website(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - required fields: first_name, last_name
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
  - required fields: name
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
  - required fields: name
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
  - required fields: event_id, contact_id
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
