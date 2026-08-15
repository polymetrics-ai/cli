---
name: pm-salesflare
description: Salesflare connector knowledge and safe action guide.
---

# pm-salesflare

## Purpose

Reads Salesflare accounts, contacts, opportunities, users, tags, tasks, workflows, groups, stages, pipelines, persons, currencies, custom-field types, and email data sources, and writes CRM lifecycle mutations, through the Salesflare REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- accounts:
  - primary key: id
  - fields: city(string), country(string), created_at(string), domain(string), email(string), id(integer), name(string), phone_number(string), updated_at(string)
- contacts:
  - primary key: id
  - fields: account_id(integer), created_at(string), email(string), first_name(string), id(integer), last_name(string), name(string), phone_number(string), updated_at(string)
- opportunities:
  - primary key: id
  - fields: account_id(integer), created_at(string), currency(string), email(string), id(integer), name(string), stage_id(integer), updated_at(string), value(number)
- users:
  - primary key: id
  - fields: email(string), enabled(boolean), id(integer), name(string)
- tags:
  - primary key: id
  - fields: id(integer), name(string)
- tasks:
  - primary key: id
  - fields: account_id(integer), assignee_id(integer), completed(boolean), description(string), due_date(string), id(integer), name(string)
- workflows:
  - primary key: id
  - fields: id(integer), name(string), state(string)
- groups:
  - primary key: id
  - fields: id(integer), name(string)
- stages:
  - primary key: id
  - fields: id(integer), name(string), pipeline_id(integer)
- pipelines:
  - primary key: id
  - fields: id(integer), name(string)
- persons:
  - primary key: id
  - fields: email(string), id(integer), name(string)
- currencies:
  - primary key: code
  - fields: code(string), name(string)
- custom_field_types:
  - primary key: type
  - fields: name(string), type(string)
- email_data_sources:
  - primary key: id
  - fields: email(string), id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_account:
  - endpoint: POST /accounts
  - required fields: name
  - risk: creates a new CRM account; low-risk external mutation, no approval required
- update_account:
  - endpoint: PUT /accounts/{{ record.id }}
  - required fields: id
  - risk: external mutation updating a CRM account; approval required
- delete_account:
  - endpoint: DELETE /accounts/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a CRM account; approval required
- create_contact:
  - endpoint: POST /contacts
  - required fields: name
  - risk: creates a new CRM contact; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation updating a CRM contact; approval required
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a CRM contact; approval required
- create_opportunity:
  - endpoint: POST /opportunities
  - required fields: name
  - risk: creates a new CRM opportunity/deal; low-risk external mutation, no approval required
- update_opportunity:
  - endpoint: PUT /opportunities/{{ record.id }}
  - required fields: id
  - risk: external mutation updating a CRM opportunity/deal (may change stage/close state); approval required
- delete_opportunity:
  - endpoint: DELETE /opportunities/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a CRM opportunity/deal; approval required
- create_tag:
  - endpoint: POST /tags
  - required fields: name
  - risk: creates a new CRM tag; low-risk external mutation, no approval required
- update_tag:
  - endpoint: PUT /tags/{{ record.id }}
  - required fields: id
  - risk: external mutation renaming a CRM tag; approval required
- delete_tag:
  - endpoint: DELETE /tags/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a CRM tag from every record it's applied to; approval required
- create_task:
  - endpoint: POST /tasks
  - required fields: name
  - risk: creates a new CRM task; low-risk external mutation, no approval required
- update_task:
  - endpoint: PUT /tasks/{{ record.id }}
  - required fields: id
  - risk: external mutation updating a CRM task (may mark complete); approval required
- delete_task:
  - endpoint: DELETE /tasks/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a CRM task; approval required
- create_meeting:
  - endpoint: POST /meetings
  - required fields: title, start_date, end_date
  - risk: creates a new CRM meeting/calendar entry; low-risk external mutation, no approval required
- update_meeting:
  - endpoint: PUT /meetings/{{ record.meeting_id }}
  - required fields: meeting_id
  - risk: external mutation updating a CRM meeting/calendar entry; approval required
- delete_meeting:
  - endpoint: DELETE /meetings/{{ record.meeting_id }}
  - required fields: meeting_id
  - risk: destructive/irreversible: permanently deletes a CRM meeting/calendar entry; approval required
- create_call:
  - endpoint: POST /calls
  - required fields: account_id
  - risk: logs a new call activity against a CRM account; low-risk external mutation, no approval required
- create_internal_note:
  - endpoint: POST /messages
  - required fields: content
  - risk: creates a new internal note on a CRM record; low-risk external mutation, no approval required
- update_internal_note:
  - endpoint: PUT /messages/{{ record.message_id }}
  - required fields: message_id
  - risk: external mutation editing a CRM internal note; approval required
- delete_internal_note:
  - endpoint: DELETE /messages/{{ record.message_id }}
  - required fields: message_id
  - risk: destructive/irreversible: permanently deletes a CRM internal note; approval required

## Security

- read risk: external Salesflare API read of CRM account, contact, opportunity, task, workflow, and reference data
- write risk: external Salesflare mutations: account/contact/opportunity/tag/task/meeting/call/internal-note create-update-delete
- approval: required for update/delete actions; create_* actions on accounts/contacts/opportunities/tags/tasks/meetings/calls/internal notes are low-risk and do not require approval; delete_* actions are destructive and irreversible
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect salesflare
```

### Inspect as structured JSON

```bash
pm connectors inspect salesflare --json
```

## Agent Rules

- Run pm connectors inspect salesflare before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
