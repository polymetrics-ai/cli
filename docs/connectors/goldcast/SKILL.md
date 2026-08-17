---
name: pm-goldcast
description: Goldcast connector knowledge and safe action guide.
---

# pm-goldcast

## Purpose

Reads Goldcast organizations, events, agenda items, discussion groups, and tracks through the Goldcast customapi REST API.

## Icon

- asset: icons/goldcast.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.goldcast.io/api-docs

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- access_key (secret)

## ETL Streams

- organizations:
  - primary key: id
  - fields: created_at(), domain(), id(), name(), slug()
- events:
  - primary key: id
  - fields: created_at(), end_time(), id(), organization(), start_time(), status(), timezone(), title()
- agenda_items:
  - primary key: id
  - fields: description(), end_time(), event(), id(), start_time(), title()
- discussion_groups:
  - primary key: id
  - fields: capacity(), created_at(), event(), id(), name()
- tracks:
  - primary key: id
  - fields: color(), event(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Goldcast API read of organization, event, and event-scoped data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect goldcast
```

### Inspect as structured JSON

```bash
pm connectors inspect goldcast --json
```

## Agent Rules

- Run pm connectors inspect goldcast before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
