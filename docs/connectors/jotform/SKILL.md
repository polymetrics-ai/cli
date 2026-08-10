---
name: pm-jotform
description: Jotform connector knowledge and safe action guide.
---

# pm-jotform

## Purpose

Reads Jotform forms, submissions, reports, folders, and the account profile through the Jotform REST API.

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
- api_key (secret) (required)

## ETL Streams

- forms:
  - primary key: id
  - cursor: created_at
  - fields: count(string), created_at(string), id(string), last_submission(string), new(string), status(string), title(string), type(string), updated_at(string), url(string), username(string)
- submissions:
  - primary key: id
  - cursor: created_at
  - fields: answers(object), created_at(string), flag(string), form_id(string), id(string), ip(string), new(string), notes(string), status(string), updated_at(string)
- reports:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), fields(string), form_id(string), id(string), status(string), title(string), type(string), updated_at(string), url(string)
- folders:
  - primary key: id
  - fields: color(string), forms(object), id(string), name(string), owner(string), parent(string), subfolders(object)
- user:
  - primary key: username
  - fields: account_type(string), created_at(string), email(string), name(string), status(string), time_zone(string), updated_at(string), usage(string), username(string)

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
