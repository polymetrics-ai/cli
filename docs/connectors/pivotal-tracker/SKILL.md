---
name: pm-pivotal-tracker
description: Pivotal Tracker connector knowledge and safe action guide.
---

# pm-pivotal-tracker

## Purpose

Reads Pivotal Tracker projects, stories, iterations, and epics through API v5.

## Icon

- asset: icons/pivotal-tracker.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- project_id
- api_token (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: id(), name(), state(), updated_at()
- stories:
  - primary key: id
  - fields: id(), name(), state(), updated_at()
- iterations:
  - primary key: id
  - fields: id(), name(), state(), updated_at()
- epics:
  - primary key: id
  - fields: id(), name(), state(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Pivotal Tracker API read of project, story, iteration, and epic data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pivotal-tracker
```

### Inspect as structured JSON

```bash
pm connectors inspect pivotal-tracker --json
```

## Agent Rules

- Run pm connectors inspect pivotal-tracker before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
