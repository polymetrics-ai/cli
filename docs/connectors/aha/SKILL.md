---
name: pm-aha
description: Aha! connector knowledge and safe action guide.
---

# pm-aha

## Purpose

Reads Aha! features, products, ideas, releases, initiatives, goals, epics, and users through the Aha! REST API (read-only).

## Icon

- id: aha
- asset: icons/aha.svg
- source: official
- review_status: official_verified
- review_url: https://www.aha.io/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- features:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), due_date(string), id(string), name(string), reference_num(string), resource(string), start_date(string), updated_at(string), url(string), workflow_status(object)
- products:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), product_line(boolean), reference_prefix(string), resource(string), updated_at(string), url(string)
- ideas:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), reference_num(string), resource(string), score(number), updated_at(string), url(string), votes(integer), workflow_status(object)
- releases:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), reference_num(string), release_date(string), released(boolean), resource(string), start_date(string), updated_at(string), url(string)
- initiatives:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), reference_num(string), resource(string), updated_at(string), url(string), workflow_status(object)
- goals:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), reference_num(string), resource(string), updated_at(string), url(string), workflow_status(object)
- epics:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), due_date(string), id(string), name(string), reference_num(string), resource(string), start_date(string), updated_at(string), url(string), workflow_status(object)
- users:
  - primary key: id
  - fields: administrator(boolean), email(string), enabled(boolean), id(string), name(string), resource(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Aha! API read of planning and roadmap data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect aha
```

### Inspect as structured JSON

```bash
pm connectors inspect aha --json
```

## Agent Rules

- Run pm connectors inspect aha before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
