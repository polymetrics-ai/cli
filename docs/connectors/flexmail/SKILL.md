---
name: pm-flexmail
description: Flexmail connector knowledge and safe action guide.
---

# pm-flexmail

## Purpose

Reads Flexmail contacts, custom fields, interests, segments, and sources through the Flexmail REST API.

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

- account_id
- base_url
- mode
- page_size
- personal_access_token (secret)

## ETL Streams

- contacts:
  - primary key: id
  - fields: custom_fields(), email(), first_name(), id(), language(), name()
- custom_fields:
  - primary key: id
  - fields: id(), name(), placeholder(), type()
- interests:
  - primary key: id
  - fields: description(), id(), label(), name(), visibility()
- segments:
  - primary key: id
  - fields: id(), name(), number_of_contacts()
- sources:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Flexmail API read of contact and marketing-list data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect flexmail
```

### Inspect as structured JSON

```bash
pm connectors inspect flexmail --json
```

## Agent Rules

- Run pm connectors inspect flexmail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
