---
name: pm-docuseal
description: DocuSeal connector knowledge and safe action guide.
---

# pm-docuseal

## Purpose

Reads DocuSeal templates, submissions, and submitters, and writes submission/submitter/template mutations through the DocuSeal REST API.

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
- page_size
- template_id
- api_key (secret) (required)

## ETL Streams

- templates:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(string), author_id(integer), created_at(string), external_id(string), folder_name(string), id(integer), name(string), slug(string), updated_at(string)
- submissions:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(string), audit_log_url(string), combined_document_url(string), completed_at(string), created_at(string), expire_at(string), id(integer), name(string), slug(string), source(string), status(string), template_id(integer), template_name(string), updated_at(string)
- submitters:
  - primary key: id
  - cursor: updated_at
  - fields: completed_at(string), created_at(string), email(string), external_id(string), id(integer), name(string), opened_at(string), phone(string), role(string), sent_at(string), slug(string), status(string), submission_id(integer), updated_at(string), uuid(string)
- template_detail:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(string), author(object), author_id(integer), created_at(string), documents(array), external_id(string), fields(array), folder_id(integer), folder_name(string), id(integer), name(string), preferences(object), schema(array), slug(string), source(string), submitters(array), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_submission:
  - endpoint: POST /submissions
  - required fields: template_id, submitters
  - risk: external mutation; dispatches a live signature-request email/SMS to every listed submitter unless send_email/send_sms are explicitly set false; approval required
- archive_submission:
  - endpoint: DELETE /submissions/{{ record.id }}
  - required fields: id
  - risk: external mutation; archives a live DocuSeal submission (soft-delete, still recoverable via the DocuSeal UI); approval required
- update_submitter:
  - endpoint: PUT /submitters/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live DocuSeal submitter's pre-filled values/contact info, can re-send signature request notifications, and can force-mark the submitter completed/auto-signed; approval required
- update_template:
  - endpoint: PUT /templates/{{ record.id }}
  - required fields: id
  - risk: external mutation; renames/moves/relabels a live DocuSeal template and can unarchive it; approval required
- archive_template:
  - endpoint: DELETE /templates/{{ record.id }}
  - required fields: id
  - risk: external mutation; archives a live DocuSeal template (soft-delete, recoverable by unarchiving via update_template); approval required
- clone_template:
  - endpoint: POST /templates/{{ record.id }}/clone
  - required fields: id
  - risk: external mutation; creates a new live DocuSeal template by cloning an existing one; approval required

## Security

- read risk: external DocuSeal API read of document template, submission, and submitter data
- write risk: external mutation; sends live signature requests, archives submissions/templates, and edits submitter/template records in DocuSeal
- approval: required for every write action; create_submission dispatches real signature-request emails/SMS to submitters unless send_email/send_sms are explicitly disabled
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run DocuSeal's declared streams and reverse-ETL actions.
- Usage: pm docuseal <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Read streams
- Reverse ETL writes
- Other Commands
  - archive submission apply - Plan and execute the archive submission reverse-ETL action [intent=reverse_etl availability=implemented write=archive_submission]; approval: requires plan, preview, approval, and execute; risk: external mutation; archives a live DocuSeal submission (soft-delete, still recoverable via the DocuSeal UI); approval required; flags: --id (required) (integer): Required id record field.: maps_to=record.id
  - archive template apply - Plan and execute the archive template reverse-ETL action [intent=reverse_etl availability=implemented write=archive_template]; approval: requires plan, preview, approval, and execute; risk: external mutation; archives a live DocuSeal template (soft-delete, recoverable by unarchiving via update_template); approval required; flags: --id (required) (integer): Required id record field.: maps_to=record.id
  - clone template apply - Plan and execute the clone template reverse-ETL action [intent=reverse_etl availability=implemented write=clone_template]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new live DocuSeal template by cloning an existing one; approval required; flags: --id (required) (integer): Required id record field.: maps_to=record.id
  - submissions list - Run the submissions ETL stream [intent=etl availability=implemented stream=submissions]
  - submitters list - Run the submitters ETL stream [intent=etl availability=implemented stream=submitters]
  - template detail list - Run the template detail ETL stream [intent=etl availability=implemented stream=template_detail]
  - templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
  - update submitter apply - Plan and execute the update submitter reverse-ETL action [intent=reverse_etl availability=implemented write=update_submitter]; approval: requires plan, preview, approval, and execute; risk: external mutation; overwrites a live DocuSeal submitter's pre-filled values/contact info, can re-send signature request notifications, and can force-mark the submitter completed/auto-signed; approval required; flags: --id (required) (integer): Required id record field.: maps_to=record.id
  - update template apply - Plan and execute the update template reverse-ETL action [intent=reverse_etl availability=implemented write=update_template]; approval: requires plan, preview, approval, and execute; risk: external mutation; renames/moves/relabels a live DocuSeal template and can unarchive it; approval required; flags: --id (required) (integer): Required id record field.: maps_to=record.id

## Commands

### Inspect as a manual

```bash
pm connectors inspect docuseal
```

### Inspect as structured JSON

```bash
pm connectors inspect docuseal --json
```

## Agent Rules

- Run pm connectors inspect docuseal before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
