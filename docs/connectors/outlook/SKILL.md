---
name: pm-outlook
description: Outlook connector knowledge and safe action guide.
---

# pm-outlook

## Purpose

Reads Outlook messages, mail folders, and calendar events through Microsoft Graph using an OAuth 2.0 refresh-token grant.

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
- max_pages
- mode
- page_size
- scope
- tenant_id
- token_url
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- messages:
  - primary key: id
  - cursor: received_date_time
  - fields: id(), last_modified_date_time(), received_date_time(), subject(), web_link()
- mail_folders:
  - primary key: id
  - fields: display_name(), id(), total_item_count(), unread_item_count()
- events:
  - primary key: id
  - cursor: last_modified_date_time
  - fields: created_date_time(), id(), last_modified_date_time(), subject(), web_link()

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
