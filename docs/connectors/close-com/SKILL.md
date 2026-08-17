---
name: pm-close-com
description: Close.com connector knowledge and safe action guide.
---

# pm-close-com

## Purpose

Reads Close CRM leads, contacts, opportunities, activities, users, tasks, lead/opportunity statuses, pipelines, roles, groups, and custom field definitions, and writes leads/contacts/opportunities/tasks through the Close REST API.

## Icon

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
- api_key (secret)

## ETL Streams

- leads:
  - primary key: id
  - cursor: date_updated
  - fields: created_by(), date_created(), date_updated(), description(), display_name(), id(), name(), organization_id(), status_id(), status_label(), url()
- contacts:
  - primary key: id
  - cursor: date_updated
  - fields: created_by(), date_created(), date_updated(), id(), lead_id(), name(), organization_id(), title()
- opportunities:
  - primary key: id
  - cursor: date_updated
  - fields: confidence(), date_created(), date_updated(), date_won(), id(), lead_id(), lead_name(), organization_id(), pipeline_id(), status_id(), status_label(), status_type(), user_id(), value(), value_currency(), value_formatted()
- activities:
  - primary key: id
  - cursor: date_updated
  - fields: _type(), contact_id(), created_by(), date_created(), date_updated(), direction(), id(), lead_id(), organization_id(), status(), user_id(), user_name()
- users:
  - primary key: id
  - cursor: date_updated
  - fields: date_created(), date_updated(), email(), first_name(), id(), image(), last_name()
- tasks:
  - primary key: id
  - fields: _type(), assigned_to(), assigned_to_name(), contact_id(), created_by(), created_by_name(), date(), date_created(), date_updated(), due_date(), id(), is_complete(), is_dateless(), lead_id(), organization_id(), text(), view()
- lead_statuses:
  - primary key: id
  - fields: id(), label(), organization_id()
- opportunity_statuses:
  - primary key: id
  - fields: id(), label(), organization_id(), pipeline_id(), type()
- pipelines:
  - primary key: id
  - fields: id(), name(), organization_id(), statuses()
- roles:
  - primary key: id
  - fields: id(), name(), organization_id()
- groups:
  - primary key: id
  - fields: id(), members(), name(), organization_id()
- custom_fields_lead:
  - primary key: id
  - fields: accepts_multiple_values(), choices(), date_created(), date_updated(), editable_with_roles(), id(), name(), organization_id(), required(), type()
- custom_fields_contact:
  - primary key: id
  - fields: accepts_multiple_values(), choices(), date_created(), date_updated(), editable_with_roles(), id(), name(), organization_id(), required(), type()
- custom_fields_opportunity:
  - primary key: id
  - fields: accepts_multiple_values(), choices(), date_created(), date_updated(), editable_with_roles(), id(), name(), organization_id(), required(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_lead:
  - endpoint: POST /lead/
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
