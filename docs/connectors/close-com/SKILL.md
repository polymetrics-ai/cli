---
name: pm-close-com
description: Close.com connector knowledge and safe action guide.
---

# pm-close-com

## Purpose

Reads Close CRM leads, contacts, opportunities, activities, users, tasks, lead/opportunity statuses, pipelines, roles, groups, and custom field definitions, and writes leads/contacts/opportunities/tasks through the Close REST API.

## Icon

- id: close
- asset: icons/close.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.close.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- leads:
  - primary key: id
  - cursor: date_updated
  - fields: created_by(string), date_created(string), date_updated(string), description(string), display_name(string), id(string), name(string), organization_id(string), status_id(string), status_label(string), url(string)
- contacts:
  - primary key: id
  - cursor: date_updated
  - fields: created_by(string), date_created(string), date_updated(string), id(string), lead_id(string), name(string), organization_id(string), title(string)
- opportunities:
  - primary key: id
  - cursor: date_updated
  - fields: confidence(integer), date_created(string), date_updated(string), date_won(string), id(string), lead_id(string), lead_name(string), organization_id(string), pipeline_id(string), status_id(string), status_label(string), status_type(string), user_id(string), value(integer), value_currency(string), value_formatted(string)
- activities:
  - primary key: id
  - cursor: date_updated
  - fields: _type(string), contact_id(string), created_by(string), date_created(string), date_updated(string), direction(string), id(string), lead_id(string), organization_id(string), status(string), user_id(string), user_name(string)
- users:
  - primary key: id
  - cursor: date_updated
  - fields: date_created(string), date_updated(string), email(string), first_name(string), id(string), image(string), last_name(string)
- tasks:
  - primary key: id
  - fields: _type(string), assigned_to(string), assigned_to_name(string), contact_id(string), created_by(string), created_by_name(string), date(string), date_created(string), date_updated(string), due_date(string), id(string), is_complete(boolean), is_dateless(boolean), lead_id(string), organization_id(string), text(string), view(string)
- lead_statuses:
  - primary key: id
  - fields: id(string), label(string), organization_id(string)
- opportunity_statuses:
  - primary key: id
  - fields: id(string), label(string), organization_id(string), pipeline_id(string), type(string)
- pipelines:
  - primary key: id
  - fields: id(string), name(string), organization_id(string), statuses(array)
- roles:
  - primary key: id
  - fields: id(string), name(string), organization_id(string)
- groups:
  - primary key: id
  - fields: id(string), members(array), name(string), organization_id(string)
- custom_fields_lead:
  - primary key: id
  - fields: accepts_multiple_values(boolean), choices(array), date_created(string), date_updated(string), editable_with_roles(array), id(string), name(string), organization_id(string), required(boolean), type(string)
- custom_fields_contact:
  - primary key: id
  - fields: accepts_multiple_values(boolean), choices(array), date_created(string), date_updated(string), editable_with_roles(array), id(string), name(string), organization_id(string), required(boolean), type(string)
- custom_fields_opportunity:
  - primary key: id
  - fields: accepts_multiple_values(boolean), choices(array), date_created(string), date_updated(string), editable_with_roles(array), id(string), name(string), organization_id(string), required(boolean), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_lead:
  - endpoint: POST /lead/
  - required fields: name
  - risk: external mutation; creates a live Close lead; approval required
- update_lead:
  - endpoint: PUT /lead/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Close lead's fields; approval required
- delete_lead:
  - endpoint: DELETE /lead/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Close lead and its contacts/opportunities; approval required
- create_contact:
  - endpoint: POST /contact/
  - required fields: lead_id, name
  - risk: external mutation; creates a live Close contact under a lead; approval required
- update_contact:
  - endpoint: PUT /contact/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Close contact's fields; approval required
- delete_contact:
  - endpoint: DELETE /contact/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Close contact; approval required
- create_opportunity:
  - endpoint: POST /opportunity/
  - required fields: lead_id
  - risk: external mutation; creates a live Close opportunity under a lead; approval required
- update_opportunity:
  - endpoint: PUT /opportunity/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Close opportunity's fields; approval required
- delete_opportunity:
  - endpoint: DELETE /opportunity/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Close opportunity; approval required
- create_task:
  - endpoint: POST /task/
  - required fields: _type, lead_id, text
  - risk: external mutation; creates a live Close task on a lead; approval required
- update_task:
  - endpoint: PUT /task/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Close task's fields; approval required
- delete_task:
  - endpoint: DELETE /task/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Close task; approval required

## Security

- read risk: external Close CRM API read of lead, contact, opportunity, activity, user, task, and account-configuration data
- write risk: external mutation; creates/updates/deletes live Close leads, contacts, opportunities, and tasks
- approval: required for all write actions; reads remain none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect close-com
```

### Inspect as structured JSON

```bash
pm connectors inspect close-com --json
```

## Agent Rules

- Run pm connectors inspect close-com before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
