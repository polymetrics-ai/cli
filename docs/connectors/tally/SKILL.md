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
- api_key (secret) (required)

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
