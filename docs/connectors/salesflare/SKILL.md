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
- api_key (secret)

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

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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

## Command Surface

- Run Salesflare's declared streams and reverse-ETL actions.
- Usage: pm salesflare <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]
  - api delete customfields itemclass id - Documented DELETE /customfields/{itemClass}/{id} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.delete.customfields-itemclass-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get accounts account-id - Documented GET /accounts/{account_id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.accounts-account-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get accounts account-id feed - Documented GET /accounts/{account_id}/feed (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.accounts-account-id-feed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get accounts account-id messages - Documented GET /accounts/{account_id}/messages (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.accounts-account-id-messages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get campaigns mergefields - Documented GET /campaigns/mergefields (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.campaigns-mergefields]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get conferences conference-id - Documented GET /conferences/{conference_id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.conferences-conference-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get contacts contact-id - Documented GET /contacts/{contact_id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.contacts-contact-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get customfields itemclass - Documented GET /customfields/{itemClass} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.customfields-itemclass]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get customfields itemclass customfieldapifield options - Documented GET /customfields/{itemClass}/{customFieldApiField}/options (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.customfields-itemclass-customfieldapifield-options]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get customfields itemclass id - Documented GET /customfields/{itemClass}/{id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.customfields-itemclass-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get filterfields entity - Documented GET /filterfields/{entity} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.filterfields-entity]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get groups id - Documented GET /groups/{id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.groups-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get me - Documented GET /me (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.me]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get me contacts - Documented GET /me/contacts (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.me-contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get meetings meeting-id - Documented GET /meetings/{meeting_id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.meetings-meeting-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get opportunities id - Documented GET /opportunities/{id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.opportunities-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get settings ai - Documented GET /settings/ai (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.settings-ai]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get stages stage-id - Documented GET /stages/{stage_id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.stages-stage-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get tags tag-id - Documented GET /tags/{tag_id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.tags-tag-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get tags tag-id usage - Documented GET /tags/{tag_id}/usage (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.tags-tag-id-usage]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users id - Documented GET /users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get workflows id - Documented GET /workflows/{id} (not implemented) [intent=direct_read availability=not_implemented operation=salesflare.get.workflows-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post accounts account-id contacts - Documented POST /accounts/{account_id}/contacts (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.post.accounts-account-id-contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post accounts account-id users - Documented POST /accounts/{account_id}/users (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.post.accounts-account-id-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post customfields itemclass - Documented POST /customfields/{itemClass} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.post.customfields-itemclass]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post message message-id feedback - Documented POST /message/{message_id}/feedback (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.post.message-message-id-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post messages message-id feedback - Documented POST /messages/{message_id}/feedback (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.post.messages-message-id-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post workflows - Documented POST /workflows (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.post.workflows]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put accounts account-id contacts - Documented PUT /accounts/{account_id}/contacts (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.accounts-account-id-contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put accounts account-id users - Documented PUT /accounts/{account_id}/users (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.accounts-account-id-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put calls meeting-id - Documented PUT /calls/{meeting_id} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.calls-meeting-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put customfields itemclass id - Documented PUT /customfields/{itemClass}/{id} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.customfields-itemclass-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put datasources email id - Documented PUT /datasources/email/{id} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.datasources-email-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put settings ai - Documented PUT /settings/ai (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.settings-ai]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put workflows id - Documented PUT /workflows/{id} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.workflows-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put workflows id audience record-id - Documented PUT /workflows/{id}/audience/{record_id} (not implemented) [intent=direct_write availability=not_implemented operation=salesflare.put.workflows-id-audience-record-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
  - create account apply - Plan and execute the create account reverse-ETL action [intent=reverse_etl availability=implemented write=create_account]; approval: requires plan, preview, approval, and execute; risk: creates a new CRM account; low-risk external mutation, no approval required; flags: --name (required)
  - create call apply - Plan and execute the create call reverse-ETL action [intent=reverse_etl availability=implemented write=create_call]; approval: requires plan, preview, approval, and execute; risk: logs a new call activity against a CRM account; low-risk external mutation, no approval required; flags: --account_id (required)
  - create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: creates a new CRM contact; low-risk external mutation, no approval required; flags: --name (required)
  - create internal note apply - Plan and execute the create internal note reverse-ETL action [intent=reverse_etl availability=implemented write=create_internal_note]; approval: requires plan, preview, approval, and execute; risk: creates a new internal note on a CRM record; low-risk external mutation, no approval required; flags: --content (required)
  - create meeting apply - Plan and execute the create meeting reverse-ETL action [intent=reverse_etl availability=implemented write=create_meeting]; approval: requires plan, preview, approval, and execute; risk: creates a new CRM meeting/calendar entry; low-risk external mutation, no approval required; flags: --end_date (required), --start_date (required), --title (required)
  - create opportunity apply - Plan and execute the create opportunity reverse-ETL action [intent=reverse_etl availability=implemented write=create_opportunity]; approval: requires plan, preview, approval, and execute; risk: creates a new CRM opportunity/deal; low-risk external mutation, no approval required; flags: --name (required)
  - create tag apply - Plan and execute the create tag reverse-ETL action [intent=reverse_etl availability=implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: creates a new CRM tag; low-risk external mutation, no approval required; flags: --name (required)
  - create task apply - Plan and execute the create task reverse-ETL action [intent=reverse_etl availability=implemented write=create_task]; approval: requires plan, preview, approval, and execute; risk: creates a new CRM task; low-risk external mutation, no approval required; flags: --name (required)
  - currencies list - Run the currencies ETL stream [intent=etl availability=implemented stream=currencies]
  - custom field types list - Run the custom field types ETL stream [intent=etl availability=implemented stream=custom_field_types]
  - delete account apply - Plan and execute the delete account reverse-ETL action [intent=reverse_etl availability=implemented write=delete_account]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM account; approval required; flags: --id (required)
  - delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM contact; approval required; flags: --id (required)
  - delete internal note apply - Plan and execute the delete internal note reverse-ETL action [intent=reverse_etl availability=implemented write=delete_internal_note]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM internal note; approval required; flags: --message_id (required)
  - delete meeting apply - Plan and execute the delete meeting reverse-ETL action [intent=reverse_etl availability=implemented write=delete_meeting]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM meeting/calendar entry; approval required; flags: --meeting_id (required)
  - delete opportunity apply - Plan and execute the delete opportunity reverse-ETL action [intent=reverse_etl availability=implemented write=delete_opportunity]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM opportunity/deal; approval required; flags: --id (required)
  - delete tag apply - Plan and execute the delete tag reverse-ETL action [intent=reverse_etl availability=implemented write=delete_tag]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM tag from every record it's applied to; approval required; flags: --id (required)
  - delete task apply - Plan and execute the delete task reverse-ETL action [intent=reverse_etl availability=implemented write=delete_task]; approval: requires plan, preview, approval, and execute; risk: destructive/irreversible: permanently deletes a CRM task; approval required; flags: --id (required)
  - email data sources list - Run the email data sources ETL stream [intent=etl availability=implemented stream=email_data_sources]
  - groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]
  - opportunities list - Run the opportunities ETL stream [intent=etl availability=implemented stream=opportunities]
  - persons list - Run the persons ETL stream [intent=etl availability=implemented stream=persons]
  - pipelines list - Run the pipelines ETL stream [intent=etl availability=implemented stream=pipelines]
  - stages list - Run the stages ETL stream [intent=etl availability=implemented stream=stages]
  - tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
  - tasks list - Run the tasks ETL stream [intent=etl availability=implemented stream=tasks]
  - update account apply - Plan and execute the update account reverse-ETL action [intent=reverse_etl availability=implemented write=update_account]; approval: requires plan, preview, approval, and execute; risk: external mutation updating a CRM account; approval required; flags: --id (required)
  - update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation updating a CRM contact; approval required; flags: --id (required)
  - update internal note apply - Plan and execute the update internal note reverse-ETL action [intent=reverse_etl availability=implemented write=update_internal_note]; approval: requires plan, preview, approval, and execute; risk: external mutation editing a CRM internal note; approval required; flags: --message_id (required)
  - update meeting apply - Plan and execute the update meeting reverse-ETL action [intent=reverse_etl availability=implemented write=update_meeting]; approval: requires plan, preview, approval, and execute; risk: external mutation updating a CRM meeting/calendar entry; approval required; flags: --meeting_id (required)
  - update opportunity apply - Plan and execute the update opportunity reverse-ETL action [intent=reverse_etl availability=implemented write=update_opportunity]; approval: requires plan, preview, approval, and execute; risk: external mutation updating a CRM opportunity/deal (may change stage/close state); approval required; flags: --id (required)
  - update tag apply - Plan and execute the update tag reverse-ETL action [intent=reverse_etl availability=implemented write=update_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation renaming a CRM tag; approval required; flags: --id (required)
  - update task apply - Plan and execute the update task reverse-ETL action [intent=reverse_etl availability=implemented write=update_task]; approval: requires plan, preview, approval, and execute; risk: external mutation updating a CRM task (may mark complete); approval required; flags: --id (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
  - workflows list - Run the workflows ETL stream [intent=etl availability=implemented stream=workflows]

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
