---
name: pm-jotform
description: Jotform connector knowledge and safe action guide.
---

# pm-jotform

## Purpose

Reads Jotform forms, submissions, reports, folders, and the account profile through the Jotform REST API.

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

- forms:
  - primary key: id
  - cursor: created_at
  - fields: count(), created_at(), id(), last_submission(), new(), status(), title(), type(), updated_at(), url(), username()
- submissions:
  - primary key: id
  - cursor: created_at
  - fields: answers(), created_at(), flag(), form_id(), id(), ip(), new(), notes(), status(), updated_at()
- reports:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), fields(), form_id(), id(), status(), title(), type(), updated_at(), url()
- folders:
  - primary key: id
  - fields: color(), forms(), id(), name(), owner(), parent(), subfolders()
- user:
  - primary key: username
  - fields: account_type(), created_at(), email(), name(), status(), time_zone(), updated_at(), usage(), username()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Jotform API read of form, submission, report, and folder data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect jotform
```

### Inspect as structured JSON

```bash
pm connectors inspect jotform --json
```

## Agent Rules

- Run pm connectors inspect jotform before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
