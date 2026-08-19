---
name: pm-sage-hr
description: Sage HR connector knowledge and safe action guide.
---

# pm-sage-hr

## Purpose

Reads Sage HR employees, teams, time off, recruitment, and onboarding/offboarding data, and writes employee/leave/task lifecycle mutations, through the Sage HR API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- employees:
  - primary key: id
  - fields: first_name(string), id(integer), last_name(string)
- teams:
  - primary key: id
  - fields: id(integer), name(string)
- timeoff_requests:
  - primary key: id
  - fields: id(integer)
- terminated_employees:
  - primary key: id
  - fields: email(string), employee_number(string), employment_start_date(string), first_name(string), id(integer), last_name(string), position(string), termination_date(string)
- positions:
  - primary key: id
  - fields: code(string), description(string), id(integer), title(string)
- termination_reasons:
  - primary key: id
  - fields: code(string), id(integer), name(string), type(string)
- leave_policies:
  - primary key: id
  - fields: accrue_type(string), color(string), default_allowance(string), do_not_accrue(boolean), id(integer), max_carryover(string), name(string), unit(string)
- out_of_office_today:
  - primary key: id
  - fields: details(string), employee_id(integer), end_date(string), hours(number), id(integer), policy_id(integer), start_date(string)
- individual_allowances:
  - primary key: id
  - fields: eligibilities(array), full_name(string), id(integer)
- recruitment_positions:
  - primary key: id
  - fields: applicants_count(integer), applicants_required(integer), created_at(string), employment_type(string), group(string), group_id(integer), id(integer), location(string), location_id(integer), status(string), team(string), title(string), visibility(string)
- recruitment_applicants:
  - primary key: id
  - fields: created_at(string), disqualified_date(string), email(string), first_name(string), full_name(string), hired_date(string), id(integer), last_name(string), position_id(string), source(string), stage(object)
- onboarding_categories:
  - primary key: id
  - fields: id(integer), title(string)
- offboarding_categories:
  - primary key: id
  - fields: id(integer), title(string)
- document_categories:
  - primary key: id
  - fields: documents_count(integer), id(integer), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_employee:
  - endpoint: POST /employees
  - required fields: email, first_name, last_name
  - risk: creates a new employee record and may email the new hire (send_email); external mutation, approval required
- update_employee:
  - endpoint: PUT /employees/{{ record.id }}
  - required fields: id
  - risk: external mutation updating an employee record (org placement, leave types, reporting line); approval required
- update_employee_custom_field:
  - endpoint: PUT /employees/{{ record.employee_id }}/custom-fields/{{ record.custom_field_id }}
  - required fields: employee_id, custom_field_id, value
  - risk: external mutation of an employee custom field; approval required
- terminate_employee:
  - endpoint: POST /employees/{{ record.employee_id }}/terminations
  - required fields: employee_id, date, termination_reason_id
  - risk: destructive/irreversible: terminates an employee's record in Sage HR; external mutation, approval required
- create_timeoff_request:
  - endpoint: POST /leave-management/requests
  - required fields: employee_id, time_off_policy_id, type, part_of_day
  - risk: creates a new time off request against an employee's leave balance; external mutation, approval required
- create_kit_day:
  - endpoint: POST /leave-management/kit-days
  - required fields: employee_id, policy_id
  - risk: creates a Keeping-In-Touch day entry against an employee's leave policy; external mutation, approval required
- update_kit_day_status:
  - endpoint: PATCH /leave-management/kit-days/{{ record.id }}
  - required fields: id, status
  - risk: approves, declines, or cancels a KIT day request; external mutation, approval required
- update_leave_policy_kit_days:
  - endpoint: PATCH /leave-management/policies/{{ record.id }}
  - required fields: id, kit_days_enabled, kit_days_quantity
  - risk: changes a company-wide leave policy's KIT-day configuration; external mutation, approval required
- create_onboarding_task:
  - endpoint: POST /onboarding/tasks
  - required fields: title, boarding_task_template_category_id, due_in
  - risk: creates a new onboarding task template; external mutation, approval required
- create_offboarding_task:
  - endpoint: POST /offboarding/tasks
  - required fields: title, boarding_task_template_category_id, due_in
  - risk: creates a new offboarding task template; external mutation, approval required

## Security

- read risk: external Sage HR API read of employee, team, time off, recruitment, and onboarding/offboarding data
- write risk: external Sage HR mutations: employee create/update/termination, custom-field update, time off/KIT-day requests and approvals, leave policy KIT-day configuration, onboarding/offboarding task creation
- approval: required for all write actions; terminate_employee is destructive/irreversible
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sage-hr
```

### Inspect as structured JSON

```bash
pm connectors inspect sage-hr --json
```

## Agent Rules

- Run pm connectors inspect sage-hr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
