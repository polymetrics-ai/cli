---
name: pm-fulcrum
description: Fulcrum connector knowledge and safe action guide.
---

# pm-fulcrum

## Purpose

Reads Fulcrum forms, records, projects, choice lists, and classification sets through the Fulcrum REST API v2.

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
- page_size
- api_key (secret) (required)

## ETL Streams

- forms:
  - primary key: id
  - cursor: updated_at
  - fields: auto_assign(boolean), created_at(string), description(string), id(string), name(string), record_count(integer), status(string), updated_at(string)
- records:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), created_by(string), form_id(string), id(string), latitude(number), longitude(number), project_id(string), status(string), updated_at(string), updated_by(string)
- projects:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), id(string), name(string), updated_at(string)
- choice_lists:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), id(string), name(string), updated_at(string)
- classification_sets:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), id(string), name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Fulcrum API read of form, record, and project data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect fulcrum
```

### Inspect as structured JSON

```bash
pm connectors inspect fulcrum --json
```

## Agent Rules

- Run pm connectors inspect fulcrum before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
