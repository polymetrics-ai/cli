---
name: pm-tvmaze-schedule
description: TVmaze Schedule connector knowledge and safe action guide.
---

# pm-tvmaze-schedule

## Purpose

Reads public TVmaze broadcast and web schedules without credentials.

## Icon

- asset: icons/tvmazeschedule.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.tvmaze.com/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- country
- date

## ETL Streams

- schedule:
  - primary key: id
  - cursor: airdate
  - fields: airdate(), airtime(), id(), name(), show_id(), show_name()
- web_schedule:
  - primary key: id
  - cursor: airdate
  - fields: airdate(), airtime(), id(), name(), show_id(), show_name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external TVmaze public API read of broadcast/web schedule data
- approval: none; read-only public schedule API, no credentials
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tvmaze-schedule
```

### Inspect as structured JSON

```bash
pm connectors inspect tvmaze-schedule --json
```

## Agent Rules

- Run pm connectors inspect tvmaze-schedule before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
