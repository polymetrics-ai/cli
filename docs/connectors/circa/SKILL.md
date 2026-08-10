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
- api_key (secret)

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

## Command Surface

- Run Circa's declared streams and reverse-ETL actions.
- Usage: pm circa <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - add event contact apply - Plan and execute the add event contact reverse-ETL action [intent=reverse_etl availability=implemented write=add_event_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; registers an existing contact onto an event; flags: --contact_id (required), --event_id (required)
  - api delete api v1 companies company-id - Documented DELETE /api/v1/companies/{company_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.delete.api-v1-companies-company-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 contacts contact-id - Documented DELETE /api/v1/contacts/{contact_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.delete.api-v1-contacts-contact-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 events event-id - Documented DELETE /api/v1/events/{event_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.delete.api-v1-events-event-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 events event-id contacts contact-id - Documented DELETE /api/v1/events/{event_id}/contacts/{contact_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.delete.api-v1-events-event-id-contacts-contact-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete api v1 events event-id contacts exports export-id - Documented DELETE /api/v1/events/{event_id}/contacts/exports/{export_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.delete.api-v1-events-event-id-contacts-exports-export-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete events event-id contacts exports export-id - Documented DELETE /events/{event_id}/contacts/exports/{export_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.delete.events-event-id-contacts-exports-export-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get api v1 companies - Documented GET /api/v1/companies (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-companies]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 companies company-id - Documented GET /api/v1/companies/{company_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-companies-company-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 companies company-id contacts - Documented GET /api/v1/companies/{company_id}/contacts (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-companies-company-id-contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 contacts - Documented GET /api/v1/contacts (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 contacts contact-id - Documented GET /api/v1/contacts/{contact_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-contacts-contact-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events - Documented GET /api/v1/events (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events event-id - Documented GET /api/v1/events/{event_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events-event-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events event-id contacts - Documented GET /api/v1/events/{event_id}/contacts (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events-event-id-contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events event-id contacts contact-id - Documented GET /api/v1/events/{event_id}/contacts/{contact_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events-event-id-contacts-contact-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events event-id contacts exports export-id - Documented GET /api/v1/events/{event_id}/contacts/exports/{export_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events-event-id-contacts-exports-export-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events event-id expenses - Documented GET /api/v1/events/{event_id}/expenses (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events-event-id-expenses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 events event-id staff - Documented GET /api/v1/events/{event_id}/staff (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-events-event-id-staff]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 fields - Documented GET /api/v1/fields (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-fields]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 fields field-id - Documented GET /api/v1/fields/{field_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-fields-field-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 teams - Documented GET /api/v1/teams (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-teams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 teams team-id - Documented GET /api/v1/teams/{team_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.api-v1-teams-team-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get companies company-id - Documented GET /companies/{company_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.companies-company-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get contacts contact-id - Documented GET /contacts/{contact_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.contacts-contact-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get events event-id - Documented GET /events/{event_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.events-event-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get events event-id contacts contact-id - Documented GET /events/{event_id}/contacts/{contact_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.events-event-id-contacts-contact-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get events event-id contacts exports export-id - Documented GET /events/{event_id}/contacts/exports/{export_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.events-event-id-contacts-exports-export-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get fields field-id - Documented GET /fields/{field_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.fields-field-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get teams team-id - Documented GET /teams/{team_id} (not implemented) [intent=direct_read availability=not_implemented operation=circa.get.teams-team-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch api v1 companies company-id - Documented PATCH /api/v1/companies/{company_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.patch.api-v1-companies-company-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch api v1 contacts contact-id - Documented PATCH /api/v1/contacts/{contact_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.patch.api-v1-contacts-contact-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch api v1 events event-id - Documented PATCH /api/v1/events/{event_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.patch.api-v1-events-event-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch api v1 events event-id contacts contact-id - Documented PATCH /api/v1/events/{event_id}/contacts/{contact_id} (not implemented) [intent=direct_write availability=not_implemented operation=circa.patch.api-v1-events-event-id-contacts-contact-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 companies - Documented POST /api/v1/companies (not implemented) [intent=direct_write availability=not_implemented operation=circa.post.api-v1-companies]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 contacts - Documented POST /api/v1/contacts (not implemented) [intent=direct_write availability=not_implemented operation=circa.post.api-v1-contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 events - Documented POST /api/v1/events (not implemented) [intent=direct_write availability=not_implemented operation=circa.post.api-v1-events]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 events event-id contacts - Documented POST /api/v1/events/{event_id}/contacts (not implemented) [intent=direct_write availability=not_implemented operation=circa.post.api-v1-events-event-id-contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 events event-id contacts exports - Documented POST /api/v1/events/{event_id}/contacts/exports (not implemented) [intent=direct_write availability=not_implemented operation=circa.post.api-v1-events-event-id-contacts-exports]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post events event-id contacts exports - Documented POST /events/{event_id}/contacts/exports (not implemented) [intent=direct_write availability=not_implemented operation=circa.post.events-event-id-contacts-exports]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - companies list - Run the companies ETL stream [intent=etl availability=implemented stream=companies]; notes: discrepancy=present-in-surface-absent-from-artifact
  - company contacts list - Run the company contacts ETL stream [intent=etl availability=implemented stream=company_contacts]; notes: discrepancy=present-in-surface-absent-from-artifact
  - contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]; notes: discrepancy=present-in-surface-absent-from-artifact
  - create company apply - Plan and execute the create company reverse-ETL action [intent=reverse_etl availability=implemented write=create_company]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new company record; flags: --name (required)
  - create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new CRM contact record; flags: --first_name (required), --last_name (required)
  - create event apply - Plan and execute the create event reverse-ETL action [intent=reverse_etl availability=implemented write=create_event]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new event record; flags: --name (required)
  - delete company apply - Plan and execute the delete company reverse-ETL action [intent=reverse_etl availability=implemented write=delete_company]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion of a company; approval required; flags: --id (required)
  - delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion of a CRM contact; approval required; flags: --id (required)
  - delete event apply - Plan and execute the delete event reverse-ETL action [intent=reverse_etl availability=implemented write=delete_event]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion of an event; approval required; flags: --id (required)
  - event contacts list - Run the event contacts ETL stream [intent=etl availability=implemented stream=event_contacts]; notes: discrepancy=present-in-surface-absent-from-artifact
  - event expenses list - Run the event expenses ETL stream [intent=etl availability=implemented stream=event_expenses]; notes: discrepancy=present-in-surface-absent-from-artifact
  - event staff list - Run the event staff ETL stream [intent=etl availability=implemented stream=event_staff]; notes: discrepancy=present-in-surface-absent-from-artifact
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]; notes: discrepancy=present-in-surface-absent-from-artifact
  - fields list - Run the fields ETL stream [intent=etl availability=implemented stream=fields]; notes: discrepancy=present-in-surface-absent-from-artifact
  - remove event contact apply - Plan and execute the remove event contact reverse-ETL action [intent=reverse_etl availability=implemented write=remove_event_contact]; approval: requires plan, preview, approval, and execute; risk: irreversible external removal of a contact's event registration; approval required; flags: --contact_id (required), --event_id (required)
  - teams list - Run the teams ETL stream [intent=etl availability=implemented stream=teams]; notes: discrepancy=present-in-surface-absent-from-artifact
  - update company apply - Plan and execute the update company reverse-ETL action [intent=reverse_etl availability=implemented write=update_company]; approval: requires plan, preview, approval, and execute; risk: external mutation; updates an existing company record; flags: --id (required)
  - update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; updates an existing CRM contact record; flags: --id (required)
  - update event apply - Plan and execute the update event reverse-ETL action [intent=reverse_etl availability=implemented write=update_event]; approval: requires plan, preview, approval, and execute; risk: external mutation; updates an existing event record; flags: --id (required)
  - update event contact apply - Plan and execute the update event contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_event_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; updates an event-contact's attendance/registration status; flags: --contact_id (required), --event_id (required)

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
