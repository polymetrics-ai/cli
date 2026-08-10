---
name: pm-fillout
description: Fillout connector knowledge and safe action guide.
---

# pm-fillout

## Purpose

Reads Fillout forms and manages webhooks/submission deletion through the Fillout REST API. Question definitions and submissions LIST remain on the legacy connector pending an engine fan_out fallback-mode gap (see docs.md Known limits).

## Icon

- id: simple-icons-fillout
- asset: icons/simple-icons/fillout.svg
- title: Fillout
- simple_icon_slug: fillout
- simple_icon_hex: FFC738
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Fillout
- match: exact-name-or-slug
- matched_by: fillout

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

- forms:
  - primary key: id
  - fields: id(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_webhook:
  - endpoint: POST /webhook/create
  - required fields: formId, url
  - risk: registers a new outbound webhook subscription that will POST live form-submission data to an external URL; external mutation, approval required
- remove_webhook:
  - endpoint: POST /webhook/delete
  - required fields: webhookId
  - risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately
- delete_submission_by_id:
  - endpoint: DELETE /forms/{{ record.form_id }}/submissions/{{ record.submission_id }}
  - required fields: form_id, submission_id
  - risk: permanently deletes a single form response; irreversible, approval required

## Security

- read risk: external Fillout API read of form metadata
- write risk: creates/removes outbound webhook subscriptions and deletes individual form submissions; external mutation, approval required
- approval: required for write actions; none for read
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Fillout's declared streams and reverse-ETL actions.
- Usage: pm fillout <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api get forms formid submissions submissionid - Documented GET /forms/{formId}/submissions/{submissionId} (not implemented) [intent=direct_read availability=not_implemented operation=fillout.get.forms-formid-submissions-submissionid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms id - Documented GET /forms/{id} (not implemented) [intent=direct_read availability=not_implemented operation=fillout.get.forms-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get forms id submissions - Documented GET /forms/{id}/submissions (not implemented) [intent=direct_read availability=not_implemented operation=fillout.get.forms-id-submissions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post forms formid submissions - Documented POST /forms/{formId}/submissions (not implemented) [intent=direct_write availability=not_implemented operation=fillout.post.forms-formid-submissions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: registers a new outbound webhook subscription that will POST live form-submission data to an external URL; external mutation, approval required; flags: --formId (required), --url (required)
  - delete submission by id apply - Plan and execute the delete submission by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_submission_by_id]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a single form response; irreversible, approval required; flags: --form_id (required), --submission_id (required)
  - forms list - Run the forms ETL stream [intent=etl availability=implemented stream=forms]
  - remove webhook apply - Plan and execute the remove webhook reverse-ETL action [intent=reverse_etl availability=implemented write=remove_webhook]; approval: requires plan, preview, approval, and execute; risk: permanently removes a webhook subscription; event delivery to its target URL stops immediately; flags: --webhookId (required)

## Commands

### Inspect as a manual

```bash
pm connectors inspect fillout
```

### Inspect as structured JSON

```bash
pm connectors inspect fillout --json
```

## Agent Rules

- Run pm connectors inspect fillout before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
