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
- form_id (required)
- mode
- page_size
- start_date
- token_url
- client_id (secret) (required)
- client_refresh_token (secret) (required)
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
