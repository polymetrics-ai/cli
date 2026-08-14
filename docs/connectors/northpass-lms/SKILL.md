---
name: pm-northpass-lms
description: Northpass LMS connector knowledge and safe action guide.
---

# pm-northpass-lms

## Purpose

Reads Northpass LMS people, courses, course enrollments, and groups through the Northpass REST API. Read-only.

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

- people:
  - primary key: id
  - fields: created_at(string), email(string), first_name(string), id(string), last_name(string), status(string), type(string), updated_at(string)
- courses:
  - primary key: id
  - fields: created_at(string), id(string), name(string), slug(string), status(string), type(string), updated_at(string)
- course_enrollments:
  - primary key: id
  - fields: completed_at(string), course_id(string), created_at(string), id(string), learner_id(string), percentage(integer), status(string), type(string), updated_at(string)
- groups:
  - primary key: id
  - fields: created_at(string), id(string), name(string), slug(string), type(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Northpass LMS API read of learner and course data
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect northpass-lms
```

### Inspect as structured JSON

```bash
pm connectors inspect northpass-lms --json
```

## Agent Rules

- Run pm connectors inspect northpass-lms before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
