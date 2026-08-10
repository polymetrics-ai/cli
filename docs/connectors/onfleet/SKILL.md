---
name: pm-onfleet
description: Onfleet connector knowledge and safe action guide.
---

# pm-onfleet

## Purpose

Reads Onfleet tasks, workers, teams, hubs, and administrators through the Onfleet REST API.

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
- max_pages
- mode
- api_key (secret) (required)

## ETL Streams

- tasks:
  - primary key: id
  - cursor: timeLastModified
  - fields: completed(boolean), creator(string), executor(string), id(string), merchant(string), shortId(string), state(integer), timeCreated(integer), timeLastModified(integer), trackingURL(string), worker(string)
- workers:
  - primary key: id
  - cursor: timeLastModified
  - fields: activeTask(string), id(string), name(string), onDuty(boolean), phone(string), timeCreated(integer), timeLastModified(integer), timeLastSeen(integer)
- teams:
  - primary key: id
  - cursor: timeLastModified
  - fields: hub(string), id(string), name(string), timeCreated(integer), timeLastModified(integer)
- hubs:
  - primary key: id
  - fields: address(string), id(string), name(string)
- administrators:
  - primary key: id
  - cursor: timeLastModified
  - fields: email(string), id(string), isActive(boolean), name(string), timeCreated(integer), timeLastModified(integer), type(string)

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
