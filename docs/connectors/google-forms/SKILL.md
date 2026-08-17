---
name: pm-google-forms
description: Google Forms connector knowledge and safe action guide.
---

# pm-google-forms

## Purpose

Reads Google Forms metadata, form items, and submitted responses through the Google Forms REST API using an OAuth 2.0 refresh-token grant.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
  - fields: description(), document_title(), form_id(), item_count(), responder_uri(), revision_id(), title()
- form_items:
  - primary key: form_id, item_id
  - fields: description(), form_id(), item_id(), question_id(), title()
- responses:
  - primary key: response_id
  - cursor: last_submitted_time
  - fields: answers(), create_time(), form_id(), last_submitted_time(), respondent_email(), response_id(), total_score()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Google Forms API read of form metadata, form items, and submitted responses
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
