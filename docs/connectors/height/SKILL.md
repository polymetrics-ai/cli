---
name: pm-height
description: Height connector knowledge and safe action guide.
---

# pm-height

## Purpose

Reads Height tasks, lists, field templates, users, and workspace through the Height REST API.

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
- api_key (secret)

## ETL Streams

- tasks:
  - primary key: id
  - cursor: createdAt
  - fields: assigneesIds(), completed(), completedAt(), createdAt(), createdUserId(), deleted(), description(), id(), index(), lastActivityAt(), listIds(), model(), name(), parentTaskId(), status(), url()
- lists:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(), defaultList(), description(), id(), key(), model(), name(), type(), updatedAt(), url(), userId(), visualization()
- field_templates:
  - primary key: id
  - fields: archived(), hidden(), id(), labels(), model(), name(), required(), standardType(), type()
- users:
  - primary key: id
  - cursor: createdAt
  - fields: admin(), createdAt(), deleted(), email(), firstname(), id(), key(), lastname(), model(), signedUpAt(), state(), username()
- workspace:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(), createdUserId(), frozen(), id(), key(), model(), name(), url(), urlType()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Height API read of task, list, field-template, user, and workspace data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect height
```

### Inspect as structured JSON

```bash
pm connectors inspect height --json
```

## Agent Rules

- Run pm connectors inspect height before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
