---
name: pm-front
description: Front connector knowledge and safe action guide.
---

# pm-front

## Purpose

Reads Front contacts, conversations, inboxes, tags, teammates, and channels through the Front Core REST API.

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
- page_limit
- api_key (secret) (required)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(number), description(string), id(string), is_private(boolean), is_spammer(boolean), name(string), updated_at(number)
- conversations:
  - primary key: id
  - cursor: last_message_at
  - fields: created_at(number), id(string), is_private(boolean), last_message_at(number), status(string), subject(string), waiting_since(number)
- inboxes:
  - primary key: id
  - fields: custom_fields(object), id(string), is_private(boolean), is_public(boolean), name(string)
- tags:
  - primary key: id
  - fields: created_at(number), highlight(string), id(string), is_private(boolean), is_visible_in_conversation_lists(boolean), name(string), updated_at(number)
- teammates:
  - primary key: id
  - fields: email(string), first_name(string), id(string), is_admin(boolean), is_available(boolean), is_blocked(boolean), last_name(string), username(string)
- channels:
  - primary key: id
  - fields: address(string), id(string), is_private(boolean), is_valid(boolean), name(string), send_as(string), type(string)

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
