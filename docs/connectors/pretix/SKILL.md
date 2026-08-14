---
name: pm-pretix
description: Pretix connector knowledge and safe action guide.
---

# pm-pretix

## Purpose

Reads pretix organizers, events, items, and orders through the pretix REST API.

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
- event
- organizer
- api_token (secret) (required)

## ETL Streams

- organizers:
  - primary key: id
  - fields: id(string), name(object), slug(string)
- events:
  - primary key: id
  - fields: id(string), name(object), slug(string), updated_at(string)
- items:
  - primary key: id
  - fields: code(string), id(integer), name(object), slug(string)
- orders:
  - primary key: id
  - fields: code(string), id(string), name(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external pretix API read of organizer, event, item, and order data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pretix
```

### Inspect as structured JSON

```bash
pm connectors inspect pretix --json
```

## Agent Rules

- Run pm connectors inspect pretix before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
