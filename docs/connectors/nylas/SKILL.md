---
name: pm-nylas
description: Nylas connector knowledge and safe action guide.
---

# pm-nylas

## Purpose

Reads Nylas calendars, contacts, messages, and events for a connected grant through the Nylas v3 REST API.

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
- calendar_id
- grant_id
- max_pages
- mode
- page_size
- api_key (secret)

## ETL Streams

- calendars:
  - primary key: id
  - fields: description(), grant_id(), hex_color(), id(), is_primary(), name(), object(), read_only(), timezone()
- contacts:
  - primary key: id
  - fields: company_name(), emails(), given_name(), grant_id(), id(), job_title(), object(), phone_numbers(), source(), surname()
- messages:
  - primary key: id
  - cursor: date
  - fields: date(), folders(), from(), grant_id(), id(), object(), snippet(), starred(), subject(), thread_id(), to(), unread()
- events:
  - primary key: id
  - cursor: updated_at
  - fields: busy(), calendar_id(), description(), grant_id(), id(), location(), object(), read_only(), status(), title(), updated_at(), when()

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
