---
name: pm-tally
description: Tally connector knowledge and safe action guide.
---

# pm-tally

## Purpose

Reads Tally.so forms, form-scoped submissions, webhooks, and workspaces, and writes form/webhook/workspace mutations through the Tally REST API.

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
- page_size
- start_date
- api_key (secret)

## ETL Streams

- forms:
  - primary key: id
  - fields: createdAt(string), id(string), isClosed(boolean), name(string), numberOfSubmissions(integer), status(string), updatedAt(string), workspaceId(string)
- workspaces:
  - primary key: id
  - fields: createdAt(string), createdByUserId(string), folders(array), id(string), index(integer), invites(array), members(array), name(string), updatedAt(string)
- webhooks:
  - primary key: id
  - fields: createdAt(string), eventTypes(array), externalSubscriber(string), formId(string), httpHeaders(array), id(string), isEnabled(boolean), lastSyncedAt(string), signingSecret(string), updatedAt(string), url(string)
- submissions:
  - primary key: id
  - cursor: submitted_at
  - fields: formId(string), form_id(string), id(string), isCompleted(boolean), pdfUrl(string), previewUrl(string), responses(array), submittedAt(string), submitted_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_webhook:
  - endpoint: POST /webhooks
  - required fields: formId, url, eventTypes
  - risk: registers an external endpoint to receive form submission events
- update_webhook:
  - endpoint: PATCH /webhooks/{{ record.id }}
  - required fields: id, formId, url, eventTypes, isEnabled
  - risk: changes where and whether an existing webhook delivers form submission events
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.id }}
  - required fields: id
  - risk: stops delivery of form submission events to the webhook's registered endpoint; if this is the form's last webhook, the webhooks integration is also marked deleted
- create_form:
  - endpoint: POST /forms
  - required fields: blocks, status
  - risk: creates a new live form in the Tally account
- update_form:
  - endpoint: PATCH /forms/{{ record.id }}
  - required fields: id
  - risk: changes a live form's name, status, blocks, or settings
- delete_form:
  - endpoint: DELETE /forms/{{ record.id }}
  - required fields: id
  - risk: moves a form to the trash, stopping new submissions
- delete_submission:
  - endpoint: DELETE /forms/{{ record.form_id }}/submissions/{{ record.id }}
  - required fields: form_id, id
  - risk: permanently removes a respondent's submission and its answers from Tally
- create_workspace:
  - endpoint: POST /workspaces
  - required fields: name
  - risk: creates a new workspace; requires the account to have a Pro subscription

## Security

- read risk: external Tally API read of form definitions, submission responses, webhook configuration, and workspace membership
- write risk: external Tally API mutation (form/webhook/workspace create-update-delete, submission delete)
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Tally's declared streams and reverse-ETL actions.
- Usage: pm tally <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete organizations organizationid invites inviteid - Documented DELETE /organizations/{organizationId}/invites/{inviteId} (not implemented) [intent=direct_write availability=not_implemented operation=tally.delete.organizations-organizationid-invites-inviteid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete organizations organizationid users userid - Documented DELETE /organizations/{organizationId}/users/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=tally.delete.organizations-organizationid-users-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete workspaces workspaceid - Documented DELETE /workspaces/{workspaceId} (not implemented) [intent=direct_write availability=not_implemented operation=tally.delete.workspaces-workspaceid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete workspaces workspaceid folders id - Documented DELETE /workspaces/{workspaceId}/folders/{id} (not implemented) [intent=direct_write availability=not_implemented operation=tally.delete.workspaces-workspaceid-folders-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get forms formid - Documented GET /forms/{formId} (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid analytics dimensions - Documented GET /forms/{formId}/analytics/dimensions (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-analytics-dimensions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid analytics drop-off - Documented GET /forms/{formId}/analytics/drop-off (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-analytics-drop-off]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid analytics metrics - Documented GET /forms/{formId}/analytics/metrics (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-analytics-metrics]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid analytics submissions - Documented GET /forms/{formId}/analytics/submissions (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-analytics-submissions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid analytics visits - Documented GET /forms/{formId}/analytics/visits (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-analytics-visits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid blocks - Documented GET /forms/{formId}/blocks (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-blocks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid questions - Documented GET /forms/{formId}/questions (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-questions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid submissions submissionid - Documented GET /forms/{formId}/submissions/{submissionId} (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.forms-formid-submissions-submissionid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get organizations organizationid invites - Documented GET /organizations/{organizationId}/invites (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.organizations-organizationid-invites]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get organizations organizationid users - Documented GET /organizations/{organizationId}/users (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.organizations-organizationid-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users me - Documented GET /users/me (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.users-me]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get webhooks webhookid events - Documented GET /webhooks/{webhookId}/events (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.webhooks-webhookid-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get workspaces workspaceid - Documented GET /workspaces/{workspaceId} (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.workspaces-workspaceid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get workspaces workspaceid folders - Documented GET /workspaces/{workspaceId}/folders (not implemented) [intent=direct_read availability=not_implemented operation=tally.get.workspaces-workspaceid-folders]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch forms formid blocks - Documented PATCH /forms/{formId}/blocks (not implemented) [intent=direct_write availability=not_implemented operation=tally.patch.forms-formid-blocks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch forms formid questions questionid - Documented PATCH /forms/{formId}/questions/{questionId} (not implemented) [intent=direct_write availability=not_implemented operation=tally.patch.forms-formid-questions-questionid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch workspaces workspaceid - Documented PATCH /workspaces/{workspaceId} (not implemented) [intent=direct_write availability=not_implemented operation=tally.patch.workspaces-workspaceid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch workspaces workspaceid folders id - Documented PATCH /workspaces/{workspaceId}/folders/{id} (not implemented) [intent=direct_write availability=not_implemented operation=tally.patch.workspaces-workspaceid-folders-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post organizations organizationid invites - Documented POST /organizations/{organizationId}/invites (not implemented) [intent=direct_write availability=not_implemented operation=tally.post.organizations-organizationid-invites]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post webhooks webhookid events eventid - Documented POST /webhooks/{webhookId}/events/{eventId} (not implemented) [intent=direct_write availability=not_implemented operation=tally.post.webhooks-webhookid-events-eventid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post workspaces workspaceid folders - Documented POST /workspaces/{workspaceId}/folders (not implemented) [intent=direct_write availability=not_implemented operation=tally.post.workspaces-workspaceid-folders]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - create form apply - Plan and execute the create form reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_form]; approval: requires plan, preview, approval, and execute; risk: creates a new live form in the Tally account; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: registers an external endpoint to receive form submission events; flags: --eventTypes (required), --formId (required), --url (required)
  - create workspace apply - Plan and execute the create workspace reverse-ETL action [intent=reverse_etl availability=implemented write=create_workspace]; approval: requires plan, preview, approval, and execute; risk: creates a new workspace; requires the account to have a Pro subscription; flags: --name (required)
  - delete form apply - Plan and execute the delete form reverse-ETL action [intent=reverse_etl availability=implemented write=delete_form]; approval: requires plan, preview, approval, and execute; risk: moves a form to the trash, stopping new submissions; flags: --id (required)
  - delete submission apply - Plan and execute the delete submission reverse-ETL action [intent=reverse_etl availability=implemented write=delete_submission]; approval: requires plan, preview, approval, and execute; risk: permanently removes a respondent's submission and its answers from Tally; flags: --form_id (required), --id (required)
  - delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: stops delivery of form submission events to the webhook's registered endpoint; if this is the form's last webhook, the webhooks integration is also marked deleted; flags: --id (required)
  - forms list - Run the forms ETL stream [intent=etl availability=implemented stream=forms]
  - submissions list - Run the submissions ETL stream [intent=etl availability=implemented stream=submissions]
  - update form apply - Plan and execute the update form reverse-ETL action [intent=reverse_etl availability=implemented write=update_form]; approval: requires plan, preview, approval, and execute; risk: changes a live form's name, status, blocks, or settings; flags: --id (required)
  - update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: changes where and whether an existing webhook delivers form submission events; flags: --eventTypes (required), --formId (required), --id (required), --isEnabled (required), --url (required)
  - webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]
  - workspaces list - Run the workspaces ETL stream [intent=etl availability=implemented stream=workspaces]

## Commands

### Inspect as a manual

```bash
pm connectors inspect tally
```

### Inspect as structured JSON

```bash
pm connectors inspect tally --json
```

## Agent Rules

- Run pm connectors inspect tally before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
