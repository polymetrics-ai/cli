---
name: pm-datascope
description: DataScope connector knowledge and safe action guide.
---

# pm-datascope

## Purpose

Reads DataScope locations, form answers, lists, notifications, task assignments, tickets (findings), and generated files, and writes location/list/task-assignment/form-answer mutations, through the DataScope external REST API (full-refresh).

## Icon

- asset: icons/datascope.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://app.mydatascope.com/api/external/docs/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- locations:
  - primary key: id
  - fields: address(), city(), code(), company_code(), company_name(), country(), description(), id(), latitude(), longitude(), name(), phone(), region()
- answers:
  - primary key: form_answer_id
  - cursor: created_at
  - fields: code(), created_at(), form_answer_id(), form_id(), form_name(), form_state(), latitude(), longitude(), user_identifier(), user_name()
- lists:
  - primary key: id
  - cursor: updated_at
  - fields: account_id(), attribute1(), attribute2(), code(), created_at(), description(), id(), list_id(), name(), updated_at()
- notifications:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), form_code(), form_name(), id(), type(), url(), user()
- task_assigns:
  - primary key: id
  - cursor: created_at
  - fields: assign_id(), completed(), completed_datetime(), confirmation_status(), created_at(), created_by(), description(), form_name(), gap(), id(), location_address(), location_code(), location_email(), location_latitude(), location_longitude(), location_name(), location_phone(), mandatory(), on_time(), priority(), response_code(), response_end(), response_start(), start_time(), status(), time_to_perform_minutes(), user_email()
- findings:
  - primary key: id
  - fields: closure_date(), closure_message(), code(), creation_date(), creator_email(), creator_id(), creator_name(), description(), expiration_date(), form_answer_code(), form_answer_id(), id(), last_updated_by(), location_code(), location_id(), location_name(), name(), priority(), status(), task_form_question(), task_form_title(), type()
- files:
  - primary key: id
  - fields: form_code(), form_name(), id(), type(), url(), user()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_location:
  - endpoint: POST /locations
  - risk: creates a new field-data-collection location record; low-risk external mutation, no approval required
- update_location:
  - endpoint: POST /locations/{{ record.id }}
  - required fields: id
  - risk: mutates an existing location's address/contact metadata; external mutation, approval required
- assign_task:
  - endpoint: POST /assign_task
  - risk: assigns a new field task/inspection to a user for a scheduled date; low-risk external mutation, no approval required
- create_metadata_object:
  - endpoint: POST /metadata_object
  - risk: creates a new list (metadata object) element; low-risk external mutation, no approval required
- update_metadata_object:
  - endpoint: POST /metadata_object/{{ record.id }}
  - required fields: id
  - risk: mutates an existing list element's fields; external mutation, approval required
- bulk_update_metadata_objects:
  - endpoint: POST /metadata_objects/bulk_update
  - risk: replaces/updates many list elements of one metadata_type in a single call; higher blast radius than a single-object update, approval required
- create_metadata_type:
  - endpoint: POST /metadata_types
  - risk: creates a new empty list (metadata type/category); low-risk external mutation, no approval required
- update_metadata_type:
  - endpoint: POST /metadata_types/{{ record.id }}
  - required fields: id
  - risk: renames/reconfigures an existing list definition; every list element under it is affected, external mutation, approval required
- change_form_answer:
  - endpoint: POST /change_form_answer
  - risk: overwrites a previously-submitted form answer's value in place, rewriting collected field data after the fact; external mutation, approval required

## Security

- read risk: external DataScope API read of field-data-collection form submissions, location data, task assignments, tickets, and generated files
- write risk: external mutation of DataScope locations, lists (metadata objects/types), task assignments, and previously-submitted form answers; change_form_answer rewrites collected field data after the fact and bulk_update_metadata_objects affects many list elements in one call, so every write ships an explicit per-action risk string
- approval: required for update_location/update_metadata_object/update_metadata_type/bulk_update_metadata_objects/change_form_answer; create_location/assign_task/create_metadata_object/create_metadata_type are low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect datascope
```

### Inspect as structured JSON

```bash
pm connectors inspect datascope --json
```

## Agent Rules

- Run pm connectors inspect datascope before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
