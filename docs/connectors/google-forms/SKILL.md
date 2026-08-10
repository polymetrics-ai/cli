---
name: pm-google-forms
description: Google Forms connector knowledge and safe action guide.
---

# pm-google-forms

## Purpose

Reads Google Forms metadata, form items, and submitted responses through the Google Forms REST API using an OAuth 2.0 refresh-token grant.

## Icon

- id: simple-icons-googleforms
- asset: icons/simple-icons/googleforms.svg
- title: Google Forms
- simple_icon_slug: googleforms
- simple_icon_hex: 7248B9
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Google%20Forms
- match: exact-name-or-slug
- matched_by: google-forms

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- form_id
- mode
- page_size
- start_date
- token_url
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- forms:
  - primary key: form_id
  - fields: description(string), document_title(string), form_id(string), item_count(integer), responder_uri(string), revision_id(string), title(string)
- form_items:
  - primary key: form_id, item_id
  - fields: description(string), form_id(string), item_id(string), question_id(string), title(string)
- responses:
  - primary key: response_id
  - cursor: last_submitted_time
  - fields: answers(object), create_time(string), form_id(string), last_submitted_time(string), respondent_email(string), response_id(string), total_score(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Google Forms API read of form metadata, form items, and submitted responses
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Google Forms's declared streams and reverse-ETL actions.
- Usage: pm google-forms <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete forms formid watches watchid - Documented DELETE /forms/{formId}/watches/{watchId} (not implemented) [intent=direct_write availability=not_implemented operation=google-forms.delete.forms-formid-watches-watchid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get forms formid responses responseid - Documented GET /forms/{formId}/responses/{responseId} (not implemented) [intent=direct_read availability=not_implemented operation=google-forms.get.forms-formid-responses-responseid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms formid watches - Documented GET /forms/{formId}/watches (not implemented) [intent=direct_read availability=not_implemented operation=google-forms.get.forms-formid-watches]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post forms - Documented POST /forms (not implemented) [intent=direct_write availability=not_implemented operation=google-forms.post.forms]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post forms formid batchupdate - Documented POST /forms/{formId}:batchUpdate (not implemented) [intent=direct_write availability=not_implemented operation=google-forms.post.forms-formid-batchupdate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post forms formid watches - Documented POST /forms/{formId}/watches (not implemented) [intent=direct_write availability=not_implemented operation=google-forms.post.forms-formid-watches]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 forms formid setpublishsettings - Documented POST /v1/forms/{formId}:setPublishSettings (not implemented) [intent=direct_write availability=not_implemented operation=google-forms.post.v1-forms-formid-setpublishsettings]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 forms formid watches watchid renew - Documented POST /v1/forms/{formId}/watches/{watchId}:renew (not implemented) [intent=direct_write availability=not_implemented operation=google-forms.post.v1-forms-formid-watches-watchid-renew]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - form items list - Run the form items ETL stream [intent=etl availability=implemented stream=form_items]
  - forms list - Run the forms ETL stream [intent=etl availability=implemented stream=forms]
  - responses list - Run the responses ETL stream [intent=etl availability=implemented stream=responses]

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-forms
```

### Inspect as structured JSON

```bash
pm connectors inspect google-forms --json
```

## Agent Rules

- Run pm connectors inspect google-forms before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
