# pm connectors inspect sage-hr

```text
NAME
  pm connectors inspect sage-hr - Sage HR connector manual

SYNOPSIS
  pm connectors inspect sage-hr
  pm connectors inspect sage-hr --json
  pm credentials add <name> --connector sage-hr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Sage HR employees, teams, time off, recruitment, and onboarding/offboarding data, and writes employee/leave/task lifecycle mutations, through the Sage HR API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  employees:
    primary key: id
    fields: first_name(), id(), last_name()
  teams:
    primary key: id
    fields: id(), name()
  timeoff_requests:
    primary key: id
    fields: id()
  terminated_employees:
    primary key: id
    fields: email(), employee_number(), employment_start_date(), first_name(), id(), last_name(), position(), termination_date()
  positions:
    primary key: id
    fields: code(), description(), id(), title()
  termination_reasons:
    primary key: id
    fields: code(), id(), name(), type()
  leave_policies:
    primary key: id
    fields: accrue_type(), color(), default_allowance(), do_not_accrue(), id(), max_carryover(), name(), unit()
  out_of_office_today:
    primary key: id
    fields: details(), employee_id(), end_date(), hours(), id(), policy_id(), start_date()
  individual_allowances:
    primary key: id
    fields: eligibilities(), full_name(), id()
  recruitment_positions:
    primary key: id
    fields: applicants_count(), applicants_required(), created_at(), employment_type(), group(), group_id(), id(), location(), location_id(), status(), team(), title(), visibility()
  recruitment_applicants:
    primary key: id
    fields: created_at(), disqualified_date(), email(), first_name(), full_name(), hired_date(), id(), last_name(), position_id(), source(), stage()
  onboarding_categories:
    primary key: id
    fields: id(), title()
  offboarding_categories:
    primary key: id
    fields: id(), title()
  document_categories:
    primary key: id
    fields: documents_count(), id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_employee:
    endpoint: POST /employees
    risk: creates a new employee record and may email the new hire (send_email); external mutation, approval required
  update_employee:
    endpoint: PUT /employees/{{ record.id }}
    required fields: id
    risk: external mutation updating an employee record (org placement, leave types, reporting line); approval required
  update_employee_custom_field:
    endpoint: PUT /employees/{{ record.employee_id }}/custom-fields/{{ record.custom_field_id }}
    required fields: employee_id, custom_field_id
    risk: external mutation of an employee custom field; approval required
  terminate_employee:
    endpoint: POST /employees/{{ record.employee_id }}/terminations
    required fields: employee_id
    risk: destructive/irreversible: terminates an employee's record in Sage HR; external mutation, approval required
  create_timeoff_request:
    endpoint: POST /leave-management/requests
    risk: creates a new time off request against an employee's leave balance; external mutation, approval required
  create_kit_day:
    endpoint: POST /leave-management/kit-days
    risk: creates a Keeping-In-Touch day entry against an employee's leave policy; external mutation, approval required
  update_kit_day_status:
    endpoint: PATCH /leave-management/kit-days/{{ record.id }}
    required fields: id
    risk: approves, declines, or cancels a KIT day request; external mutation, approval required
  update_leave_policy_kit_days:
    endpoint: PATCH /leave-management/policies/{{ record.id }}
    required fields: id
    risk: changes a company-wide leave policy's KIT-day configuration; external mutation, approval required
  create_onboarding_task:
    endpoint: POST /onboarding/tasks
    risk: creates a new onboarding task template; external mutation, approval required
  create_offboarding_task:
    endpoint: POST /offboarding/tasks
    risk: creates a new offboarding task template; external mutation, approval required

SECURITY
  read risk: external Sage HR API read of employee, team, time off, recruitment, and onboarding/offboarding data
  write risk: external Sage HR mutations: employee create/update/termination, custom-field update, time off/KIT-day requests and approvals, leave policy KIT-day configuration, onboarding/offboarding task creation
  approval: required for all write actions; terminate_employee is destructive/irreversible
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sage-hr

  # Inspect as structured JSON
  pm connectors inspect sage-hr --json

AGENT WORKFLOW
  - Run pm connectors inspect sage-hr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
