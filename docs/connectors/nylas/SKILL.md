---
name: pm-nylas
description: Nylas connector knowledge and safe action guide.
---

# pm-nylas

## Purpose

Reads Nylas calendars, contacts, messages, and events for a connected grant through the Nylas v3 REST API.

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
- calendar_id
- grant_id
- max_pages
- mode
- page_size
- api_key (secret) (required)

## ETL Streams

- calendars:
  - primary key: id
  - fields: description(string), grant_id(string), hex_color(string), id(string), is_primary(boolean), name(string), object(string), read_only(boolean), timezone(string)
- contacts:
  - primary key: id
  - fields: company_name(string), emails(array), given_name(string), grant_id(string), id(string), job_title(string), object(string), phone_numbers(array), source(string), surname(string)
- messages:
  - primary key: id
  - cursor: date
  - fields: date(integer), folders(array), from(array), grant_id(string), id(string), object(string), snippet(string), starred(boolean), subject(string), thread_id(string), to(array), unread(boolean)
- events:
  - primary key: id
  - cursor: updated_at
  - fields: busy(boolean), calendar_id(string), description(string), grant_id(string), id(string), location(string), object(string), read_only(boolean), status(string), title(string), updated_at(integer), when(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Nylas API read of a connected grant's calendar, contact, and message data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect nylas
```

### Inspect as structured JSON

```bash
pm connectors inspect nylas --json
```

## Agent Rules

- Run pm connectors inspect nylas before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
