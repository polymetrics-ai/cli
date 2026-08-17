---
name: pm-northpass-lms
description: Northpass LMS connector knowledge and safe action guide.
---

# pm-northpass-lms

## Purpose

Reads Northpass LMS people, courses, course enrollments, and groups through the Northpass REST API. Read-only.

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

- people:
  - primary key: id
  - fields: created_at(), email(), first_name(), id(), last_name(), status(), type(), updated_at()
- courses:
  - primary key: id
  - fields: created_at(), id(), name(), slug(), status(), type(), updated_at()
- course_enrollments:
  - primary key: id
  - fields: completed_at(), course_id(), created_at(), id(), learner_id(), percentage(), status(), type(), updated_at()
- groups:
  - primary key: id
  - fields: created_at(), id(), name(), slug(), type(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
