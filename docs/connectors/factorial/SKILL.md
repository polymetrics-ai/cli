---
name: pm-factorial
description: Factorial connector knowledge and safe action guide.
---

# pm-factorial

## Purpose

Reads FactorialHR employees, teams, time-off leaves, leave types, and locations through the Factorial REST API.

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
- api_key (secret)

## ETL Streams

- employees:
  - primary key: id
  - cursor: updated_at
  - fields: active(), birthday_on(), company_id(), created_at(), email(), first_name(), full_name(), gender(), id(), last_name(), legal_entity_id(), location_id(), manager_id(), team_ids(), terminated_on(), updated_at()
- teams:
  - primary key: id
  - fields: avatar(), company_id(), description(), employee_ids(), id(), lead_ids(), name()
- leaves:
  - primary key: id
  - cursor: updated_at
  - fields: approved(), created_at(), description(), employee_id(), finish_on(), half_day(), id(), leave_type_id(), start_on(), updated_at()
- leave_types:
  - primary key: id
  - fields: active(), approval_required(), color(), company_id(), id(), identifier(), name()
- locations:
  - primary key: id
  - fields: address_line_1(), city(), company_id(), country(), id(), main(), name(), postal_code(), state(), timezone()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Factorial API read of employee, team, and time-off data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect factorial
```

### Inspect as structured JSON

```bash
pm connectors inspect factorial --json
```

## Agent Rules

- Run pm connectors inspect factorial before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
