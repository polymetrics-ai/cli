---
name: pm-front
description: Front connector knowledge and safe action guide.
---

# pm-front

## Purpose

Reads Front contacts, conversations, inboxes, tags, teammates, and channels through the Front Core REST API.

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
- page_limit
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), is_private(), is_spammer(), name(), updated_at()
- conversations:
  - primary key: id
  - cursor: last_message_at
  - fields: created_at(), id(), is_private(), last_message_at(), status(), subject(), waiting_since()
- inboxes:
  - primary key: id
  - fields: custom_fields(), id(), is_private(), is_public(), name()
- tags:
  - primary key: id
  - fields: created_at(), highlight(), id(), is_private(), is_visible_in_conversation_lists(), name(), updated_at()
- teammates:
  - primary key: id
  - fields: email(), first_name(), id(), is_admin(), is_available(), is_blocked(), last_name(), username()
- channels:
  - primary key: id
  - fields: address(), id(), is_private(), is_valid(), name(), send_as(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Front API read of contact, conversation, inbox, tag, teammate, and channel data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect front
```

### Inspect as structured JSON

```bash
pm connectors inspect front --json
```

## Agent Rules

- Run pm connectors inspect front before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
