---
name: pm-workflowmax
description: WorkflowMax connector knowledge and safe action guide.
---

# pm-workflowmax

## Purpose

Reads and writes WorkflowMax jobs, clients, and client contacts through the real WorkflowMax API v2 (api.workflowmax2.com/v2).

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

- account_id (required)
- base_url
- mode
- updated_since
- access_token (secret) (required)

## ETL Streams

- jobs:
  - primary key: uuid
  - cursor: updated_at
  - fields: budget(number), clientContactUUID(string), clientOrderNumber(string), clientUUID(string), completedDate(string), description(string), dueDate(string), jobCategoryUUID(string), jobNumber(string), jobStatusUUID(string), name(string), priority(string), startDate(string), updated_at(string), uuid(string)
- clients:
  - primary key: uuid
  - cursor: updated_at
  - fields: archived(boolean), clientManagerUUID(string), exportCode(string), favorite(boolean), jobManagerUUID(string), name(string), prospect(boolean), referralSource(string), updated_at(string), uuid(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_client:
  - endpoint: POST /v2/clients
  - risk: creates a WorkflowMax client record; approval required
- update_client:
  - endpoint: PUT /v2/clients/{{ record.uuid }}
  - required fields: uuid
  - risk: updates a WorkflowMax client record; approval required
- delete_client:
  - endpoint: DELETE /v2/clients/{{ record.uuid }}
  - required fields: uuid
  - risk: permanently deletes a WorkflowMax client record; approval required
- create_job:
  - endpoint: POST /v2/jobs
  - required fields: clientUUID, jobName, statusUUID, startDate, dueDate, priority
  - risk: creates a WorkflowMax job; approval required
- delete_job:
  - endpoint: DELETE /v2/jobs/{{ record.uuid }}
  - required fields: uuid
  - risk: permanently deletes a WorkflowMax job; approval required
- create_client_contact:
  - endpoint: POST /v2/clients/contacts
  - required fields: firstName
  - risk: creates a WorkflowMax client-contact record (not attached to any client until linked); approval required
- update_client_contact:
  - endpoint: PUT /v2/clients/contacts/{{ record.uuid }}
  - required fields: uuid
  - risk: updates a WorkflowMax client-contact record; approval required
- delete_client_contact:
  - endpoint: DELETE /v2/clients/contacts/{{ record.uuid }}
  - required fields: uuid
  - risk: permanently deletes a WorkflowMax client-contact record; approval required

## Security

- read risk: external WorkflowMax API v2 read of job, client, and client-contact data
- write risk: external mutation of WorkflowMax jobs, clients, and client contacts (create/update/delete); approval required
- approval: writes require approval; reads are none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect workflowmax
```

### Inspect as structured JSON

```bash
pm connectors inspect workflowmax --json
```

## Agent Rules

- Run pm connectors inspect workflowmax before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
