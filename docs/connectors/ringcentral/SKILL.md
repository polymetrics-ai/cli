---
name: pm-ringcentral
description: RingCentral connector knowledge and safe action guide.
---

# pm-ringcentral

## Purpose

Reads RingCentral extensions, call logs, messages, contacts, and devices through the REST API.

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
- dateFrom
- dateTo
- direction
- messageType
- type
- access_token (secret)

## ETL Streams

- extensions:
  - primary key: id
  - fields: extension_number(), id(), name(), status(), stream(), type()
- call_log:
  - primary key: id
  - cursor: start_time
  - fields: direction(), id(), result(), start_time(), stream(), type()
- messages:
  - primary key: id
  - cursor: creation_time
  - fields: creation_time(), direction(), id(), stream(), subject(), type()
- contacts:
  - primary key: id
  - fields: company(), email(), first_name(), id(), last_name(), stream()
- devices:
  - primary key: id
  - fields: id(), name(), status(), stream(), type()

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
