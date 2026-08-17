---
name: pm-campayn
description: Campayn connector knowledge and safe action guide.
---

# pm-campayn

## Purpose

Reads and writes Campayn subscriber lists, signup forms, contacts, email campaigns, and calendar reports through the Campayn REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- report_from
- report_to
- api_key (secret)

## ETL Streams

- lists:
  - primary key: id
  - fields: contact_count(), id(), list_name(), tags()
- emails:
  - primary key: id
  - fields: id(), name(), percent_responses(), percent_views(), preview_thumb(), preview_url(), send_count(), send_now(), status(), unique_responses(), unique_views()
- reports:
  - primary key: id
  - fields: id(), name(), preview_url(), scheduled_date(), status()
- forms:
  - primary key: id
  - fields: contact_list_id(), form_html(), form_title(), form_type(), id(), list_id(), signup_count()
- contacts:
  - primary key: id
  - fields: confirmed(), email(), first_name(), id(), image_url(), last_name(), list_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_contact:
  - endpoint: POST /lists/{{ record.list_id }}/contacts.json
  - required fields: list_id
  - risk: adds a new contact to a Campayn subscriber list; low-risk external mutation, no approval required
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}.json
  - required fields: id
  - risk: replaces a contact's full field set (the upstream API's own docs warn any field not sent in the body is removed); external mutation, no approval required
- unsubscribe_contact:
  - endpoint: POST /lists/{{ record.list_id }}/unsubscribe.json
  - required fields: list_id
  - risk: unsubscribes a contact from a list by id (single contact) or email (every contact on the list sharing that email address); the docs note neither path shows up in Reporting; low-risk external mutation, no approval required

## Security

- read risk: external Campayn API read of subscriber lists, campaigns, and reports
- write risk: external mutation of Campayn contacts and list-subscription state (add contact, update contact, unsubscribe by id or email); no destructive delete endpoint is documented by the upstream API
- approval: none; low-risk marketing-list mutations, no documented destructive writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect campayn
```

### Inspect as structured JSON

```bash
pm connectors inspect campayn --json
```

## Agent Rules

- Run pm connectors inspect campayn before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
