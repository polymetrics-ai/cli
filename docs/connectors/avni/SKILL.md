---
name: pm-avni
description: Avni connector knowledge and safe action guide.
---

# pm-avni

## Purpose

Reads Avni subjects and encounters through a read-only HTTP API using HTTP Basic authentication.

## Icon

- asset: icons/avni.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- start_date
- username
- password (secret)

## ETL Streams

- subjects:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), updated_at()
- encounters:
  - primary key: id
  - cursor: updated_at
  - fields: encounter_type(), id(), subject_id(), updated_at()
- program_enrolments:
  - primary key: id
  - cursor: updated_at
  - fields: enrolment_date_time(), exit_date_time(), id(), program(), subject_id(), updated_at()
- program_encounters:
  - primary key: id
  - cursor: updated_at
  - fields: encounter_date_time(), encounter_type(), enrolment_id(), id(), program(), subject_id(), updated_at()
- group_subjects:
  - primary key: id
  - cursor: updated_at
  - fields: group_subject_id(), id(), member_subject_id(), membership_end_date(), membership_start_date(), updated_at()
- locations:
  - primary key: id
  - cursor: updated_at
  - fields: id(), level(), parent_id(), title(), type(), updated_at()
- approval_statuses:
  - primary key: entity_id, entity_type
  - cursor: status_date_time
  - fields: approval_status(), approval_status_comment(), entity_id(), entity_type(), entity_type_id(), status_date_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Avni API read of subjects and encounters
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect avni
```

### Inspect as structured JSON

```bash
pm connectors inspect avni --json
```

## Agent Rules

- Run pm connectors inspect avni before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
