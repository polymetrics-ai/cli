---
name: pm-ringcentral
description: RingCentral connector knowledge and safe action guide.
---

# pm-ringcentral

## Purpose

Reads RingCentral extensions, call logs, messages, contacts, and devices through the REST API.

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
- dateFrom
- dateTo
- direction
- messageType
- type
- access_token (secret) (required)

## ETL Streams

- extensions:
  - primary key: id
  - fields: extension_number(string), id(string), name(string), status(string), stream(string), type(string)
- call_log:
  - primary key: id
  - cursor: start_time
  - fields: direction(string), id(string), result(string), start_time(string), stream(string), type(string)
- messages:
  - primary key: id
  - cursor: creation_time
  - fields: creation_time(string), direction(string), id(string), stream(string), subject(string), type(string)
- contacts:
  - primary key: id
  - fields: company(string), email(string), first_name(string), id(string), last_name(string), stream(string)
- devices:
  - primary key: id
  - fields: id(string), name(string), status(string), stream(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external RingCentral API read of account extension, call-log, message, contact, and device data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ringcentral
```

### Inspect as structured JSON

```bash
pm connectors inspect ringcentral --json
```

## Agent Rules

- Run pm connectors inspect ringcentral before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
