---
name: pm-zendesk-chat
description: Zendesk Chat connector knowledge and safe action guide.
---

# pm-zendesk-chat

## Purpose

Reads Zendesk Chat agents, chats, departments, shortcuts, and triggers through the Zendesk Chat REST API v2.

## Icon

- id: zendesk-chat
- asset: icons/zendesk-chat.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://support.zendesk.com/hc/en-us/sections/4405298889242-Developer-updates

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- start_date
- access_token (secret)

## ETL Streams

- agents:
  - primary key: id
  - fields: create_date(string), display_name(string), email(string), enabled(boolean), first_name(string), id(integer), last_login(string), last_name(string), role_id(integer)
- chats:
  - primary key: id
  - cursor: timestamp
  - fields: comment(string), department_id(integer), duration(integer), id(string), rating(string), session(object), timestamp(string), type(string), visitor(object)
- departments:
  - primary key: id
  - fields: description(string), enabled(boolean), id(integer), members(array), name(string), settings(object)
- shortcuts:
  - primary key: id
  - fields: id(integer), message(string), name(string), options(string), scope(string), tags(array)
- triggers:
  - primary key: id
  - fields: definition(object), description(string), enabled(boolean), id(integer), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Zendesk Chat API read of agent, chat, and configuration data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zendesk-chat
```

### Inspect as structured JSON

```bash
pm connectors inspect zendesk-chat --json
```

## Agent Rules

- Run pm connectors inspect zendesk-chat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
