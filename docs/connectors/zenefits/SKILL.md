---
name: pm-zenefits
description: Zenefits connector knowledge and safe action guide.
---

# pm-zenefits

## Purpose

Reads Zenefits people, companies, departments, locations, employments, custom fields/values, bank accounts, labor groups, and time-off data.

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
- token (secret)

## ETL Streams

- people:
  - primary key: id
  - fields: first_name(), id(), last_name(), status()
- companies:
  - primary key: id
  - fields: id(), name()
- departments:
  - primary key: id
  - fields: id(), name()
- locations:
  - primary key: id
  - fields: city(), company(), country(), id(), name(), people(), phone(), state(), street1(), street2(), zip()
- employments:
  - primary key: id
  - fields: annual_salary(), comp_type(), employment_type(), hire_date(), id(), is_active(), pay_rate(), person(), termination_date(), termination_type(), working_hours_per_week()
- custom_fields:
  - primary key: id
  - fields: can_manager_view_field(), can_person_edit_field(), can_person_view_field(), custom_field_type(), custom_field_values(), help_text(), help_url(), id(), is_field_required(), is_sensitive(), name()
- custom_field_values:
  - primary key: id
  - fields: custom_field(), id(), person(), value()
- company_banks:
  - primary key: id
  - fields: account_number(), account_type(), bank_name(), company(), id(), routing_number()
- employee_banks:
  - primary key: id
  - fields: account_number(), account_type(), amount_per_paycheck(), bank_name(), id(), is_primary_account(), is_salary_account(), percentage_per_paycheck(), person(), priority(), routing_number()
- labor_group_types:
  - primary key: id
  - fields: id(), labor_groups(), name()
- labor_groups:
  - primary key: id
  - fields: assigned_members(), code(), id(), labor_group_type(), name()
- vacation_types:
  - primary key: id
  - fields: company(), counts_as(), id(), name(), status(), vacation_requests()
- vacation_requests:
  - primary key: id
  - fields: approved_date(), created_date(), creator(), deny_reason(), end_date(), hours(), id(), person(), reason(), start_date(), status(), vacation_type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Zenefits account read of people, companies, departments, locations, employments, custom field definitions/values, company and employee bank account details, labor groups, and time-off vacation types/requests
- approval: none; read-only bearer token access. The entire documented Zenefits API is read-only (no write endpoint exists), so there is no write risk to assess
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zenefits
```

### Inspect as structured JSON

```bash
pm connectors inspect zenefits --json
```

## Agent Rules

- Run pm connectors inspect zenefits before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
