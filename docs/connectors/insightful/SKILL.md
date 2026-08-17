---
name: pm-insightful
description: Insightful connector knowledge and safe action guide.
---

# pm-insightful

## Purpose

Reads Insightful workforce-analytics employees, teams, projects, and directory entries through the Insightful REST API.

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
- mode
- api_token (secret)

## ETL Streams

- employee:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), email(), id(), modelName(), name(), updatedAt()
- team:
  - primary key: id
  - fields: default(), description(), employees(), id(), modelName(), name(), projects()
- projects:
  - primary key: id
  - cursor: updatedAt
  - fields: archived(), billable(), createdAt(), creatorId(), employees(), id(), modelName(), name(), organizationId(), updatedAt()
- directory:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), modelName(), name(), organizationId(), updatedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Insightful API read of workforce-analytics employees, teams, projects, and directory entries
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect insightful
```

### Inspect as structured JSON

```bash
pm connectors inspect insightful --json
```

## Agent Rules

- Run pm connectors inspect insightful before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
