---
name: pm-outlook
description: Outlook connector knowledge and safe action guide.
---

# pm-outlook

## Purpose

Reads Outlook messages, mail folders, and calendar events through Microsoft Graph using an OAuth 2.0 refresh-token grant.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- scope
- tenant_id
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)
- refresh_token (secret) (required)

## ETL Streams

- messages:
  - primary key: id
  - cursor: received_date_time
  - fields: id(string), last_modified_date_time(string), received_date_time(string), subject(string), web_link(string)
- mail_folders:
  - primary key: id
  - fields: display_name(string), id(string), total_item_count(integer), unread_item_count(integer)
- events:
  - primary key: id
  - cursor: last_modified_date_time
  - fields: created_date_time(string), id(string), last_modified_date_time(string), subject(string), web_link(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Microsoft Graph API read of the authenticated mailbox's messages, mail folders, and calendar events
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect outlook
```

### Inspect as structured JSON

```bash
pm connectors inspect outlook --json
```

## Agent Rules

- Run pm connectors inspect outlook before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
