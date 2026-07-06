---
name: pm-onfleet
description: Onfleet connector knowledge and safe action guide.
---

# pm-onfleet

## Purpose

Reads Onfleet tasks, workers, teams, hubs, and administrators through the Onfleet REST API.

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
- max_pages
- mode
- api_key (secret)

## ETL Streams

- tasks:
  - primary key: id
  - cursor: timeLastModified
  - fields: completed(), creator(), executor(), id(), merchant(), shortId(), state(), timeCreated(), timeLastModified(), trackingURL(), worker()
- workers:
  - primary key: id
  - cursor: timeLastModified
  - fields: activeTask(), id(), name(), onDuty(), phone(), timeCreated(), timeLastModified(), timeLastSeen()
- teams:
  - primary key: id
  - cursor: timeLastModified
  - fields: hub(), id(), name(), timeCreated(), timeLastModified()
- hubs:
  - primary key: id
  - fields: address(), id(), name()
- administrators:
  - primary key: id
  - cursor: timeLastModified
  - fields: email(), id(), isActive(), name(), timeCreated(), timeLastModified(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Onfleet API read of delivery task and workforce data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect onfleet
```

### Inspect as structured JSON

```bash
pm connectors inspect onfleet --json
```

## Agent Rules

- Run pm connectors inspect onfleet before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
