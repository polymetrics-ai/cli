---
name: pm-employment-hero
description: Employment Hero connector knowledge and safe action guide.
---

# pm-employment-hero

## Purpose

Reads Employment Hero organisations, employees, HR reference data, forms, goals, rosters, employee subresources, and exposes documented JSON mutations through the Employment Hero REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- certification_id
- employee_id
- employee_ids
- form_id
- goal_id
- key_result_id
- leave_request_id
- member_ids
- organization_id
- payslip_id
- response_id
- rostered_shift_id
- template_id
- unavailability_id
- api_key (secret)

## ETL Streams

- organisations:
  - primary key: id
  - fields: country(), id(), logo_url(), name(), phone()
- organisation:
  - primary key: id
  - fields: country(), id(), logo_url(), name(), phone()
- employees:
  - primary key: id
  - fields: account_email(), company_email(), company_mobile(), country(), date_of_birth(), employing_entity(), first_name(), gender(), id(), job_title(), known_as(), last_name(), location(), middle_name(), personal_email(), personal_mobile_number(), primary_manager(), role(), start_date(), title()
- employee:
  - primary key: id
  - fields: account_email(), company_email(), company_mobile(), country(), date_of_birth(), employing_entity(), first_name(), gender(), id(), job_title(), known_as(), last_name(), location(), middle_name(), personal_email(), personal_mobile_number(), primary_manager(), role(), start_date(), title()
- teams:
  - primary key: id
  - fields: id(), name(), status()
- team_employees:
  - primary key: id
  - fields: company_email(), first_name(), id(), last_name(), role(), team_id()
- leave_requests:
  - primary key: id
  - fields: comment(), employee_id(), end_date(), id(), leave_balance_amount(), leave_category_name(), start_date(), status(), total_hours()
- leave_request:
  - primary key: id
  - fields: comment(), employee_id(), end_date(), id(), leave_balance_amount(), leave_category_name(), start_date(), status(), total_hours()
- certifications:
  - primary key: id
  - fields: archived(), description(), id(), mandatory(), name(), state(), type()
- certification:
  - primary key: id
  - fields: archived(), description(), id(), mandatory(), name(), state(), type()
- cost_centres:
  - primary key: id
  - fields: code(), id(), name()
- custom_fields:
  - primary key: id
  - fields: field_type(), id(), name(), required()
- employing_entities:
  - primary key: id
  - fields: country(), id(), name()
- forms:
  - primary key: id
  - fields: category_id(), id(), name(), status()
- form:
  - primary key: id
  - fields: category_id(), id(), name(), status()
- form_responses:
  - primary key: id
  - fields: form_id(), id(), member_id(), status(), submitted_at()
- form_response:
  - primary key: id
  - fields: form_id(), id(), member_id(), status(), submitted_at()
- form_assignments:
  - primary key: id
  - fields: form_id(), id(), member_id(), status()
- member_form_responses:
  - primary key: id
  - fields: form_id(), id(), member_id(), status(), submitted_at()
- form_categories:
  - primary key: id
  - fields: id(), name()
- form_templates:
  - primary key: id
  - fields: category_id(), description(), id(), name()
- form_template:
  - primary key: id
  - fields: category_id(), description(), id(), name()
- goals:
  - primary key: id
  - fields: health_status(), id(), owner_id(), status(), title()
- goal:
  - primary key: id
  - fields: health_status(), id(), owner_id(), status(), title()
- goal_comments:
  - primary key: id
  - fields: author_id(), body(), created_at(), goal_id(), id()
- goal_key_results:
  - primary key: id
  - fields: goal_id(), health_status(), id(), progress(), status(), title()
- goal_key_result:
  - primary key: id
  - fields: goal_id(), health_status(), id(), progress(), status(), title()
- kiosk_members:
  - primary key: id
  - fields: id(), kiosk_access_status(), member_id(), name()
- leave_categories:
  - primary key: id
  - fields: code(), id(), name()
- pay_categories:
  - primary key: id
  - fields: code(), id(), name()
- policies:
  - primary key: id
  - fields: id(), name(), policy_type()
- rostered_shifts:
  - primary key: id
  - fields: end_date_time(), id(), member_id(), published(), start_date_time(), status()
- rostered_shift:
  - primary key: id
  - fields: end_date_time(), id(), member_id(), published(), start_date_time(), status()
- roles:
  - primary key: id
  - fields: id(), name()
- unavailabilities:
  - primary key: id
  - fields: from_date(), id(), member_id(), status(), to_date()
- unavailability:
  - primary key: id
  - fields: from_date(), id(), member_id(), status(), to_date()
- work_locations:
  - primary key: id
  - fields: country(), id(), name()
- work_sites:
  - primary key: id
  - fields: city(), country(), id(), name(), state()
- work_types:
  - primary key: id
  - fields: id(), name(), payroll_info_id()
- bank_accounts_v1:
  - primary key: id
  - fields: account_name(), account_number(), bsb(), employee_id(), id()
- bank_accounts_v2:
  - primary key: id
  - fields: account_name(), account_number(), bsb(), employee_id(), id()
- contractor_job_histories:
  - primary key: id
  - fields: employee_id(), end_date(), id(), job_title(), start_date()
- documents:
  - primary key: id
  - fields: created_at(), document_type(), employee_id(), id(), name()
- emergency_contacts:
  - primary key: id
  - fields: employee_id(), id(), name(), phone(), relationship()
- employee_certification_details:
  - primary key: id
  - fields: employee_id(), expiry_date(), id(), name(), status()
- employee_custom_fields:
  - primary key: id
  - fields: employee_id(), id(), name(), value()
- employment_histories:
  - primary key: id
  - fields: employee_id(), end_date(), id(), job_title(), start_date()
- leave_balances:
  - primary key: id
  - fields: balance(), employee_id(), id(), leave_category_name()
- pay_details:
  - primary key: id
  - fields: employee_id(), id(), pay_category_id(), rate()
- payslips:
  - primary key: id
  - fields: employee_id(), gross_pay(), id(), period_end(), period_start()
- payslip:
  - primary key: id
  - fields: employee_id(), gross_pay(), id(), period_end(), period_start()
- timesheet_entries:
  - primary key: id
  - fields: date(), employee_id(), end_time(), id(), start_time(), status(), units()
- superannuation_detail_v1:
  - primary key: employee_id
  - fields: employee_id(), fund_name(), id(), member_number()
- superannuation_detail_v2:
  - primary key: employee_id
  - fields: employee_id(), fund_name(), id(), member_number()
- tax_declaration_v1:
  - primary key: employee_id
  - fields: employee_id(), id(), residency_status(), tax_file_number_status()
- tax_declaration_v2:
  - primary key: employee_id
  - fields: employee_id(), id(), residency_status(), tax_file_number_status()
- work_eligibility:
  - primary key: employee_id
  - fields: employee_id(), expiry_date(), id(), status(), visa_type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_certification:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/certifications
  - risk: external Employment Hero mutation; approval required before execution
- update_certification:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/certifications/{{ record.certification_id }}
  - required fields: certification_id
  - risk: external Employment Hero mutation; approval required before execution
- archive_certification:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/certifications/{{ record.certification_id }}/archive_status
  - required fields: certification_id
  - optional fields: status
  - risk: archives or restores an Employment Hero certification configuration
- delete_certification:
  - endpoint: DELETE /v1/organisations/{{ config.organization_id }}/certifications/{{ record.certification_id }}
  - required fields: certification_id
  - risk: deletes an Employment Hero certification configuration
- create_department:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/departments
  - risk: external Employment Hero mutation; approval required before execution
- update_department:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/departments/{{ record.department_id }}
  - required fields: department_id
  - risk: external Employment Hero mutation; approval required before execution
- quick_add_employee:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/quick_add_employee
  - risk: external Employment Hero mutation; approval required before execution
- quick_add_contractor:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/quick_add_contractor
  - risk: external Employment Hero mutation; approval required before execution
- onboard_employee_async:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/polling_onboard_employee
  - risk: starts an asynchronous employee onboarding job; approval required before execution
- update_employee_personal_details:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/personal_details
  - required fields: employee_id
  - risk: external Employment Hero mutation; approval required before execution
- update_employee_employment_details:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/employment_details
  - required fields: employee_id
  - risk: external Employment Hero mutation; approval required before execution
- update_employee_contractor_details:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/contractor_details
  - required fields: employee_id
  - risk: external Employment Hero mutation; approval required before execution
- delete_employee:
  - endpoint: DELETE /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}
  - required fields: employee_id
  - risk: deletes or removes an Employment Hero employee record
- update_employee_certification:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/certifications/{{ record.id }}
  - required fields: employee_id, id
  - risk: external Employment Hero mutation; approval required before execution
- delete_form:
  - endpoint: DELETE /v1/organisations/{{ config.organization_id }}/forms/{{ record.form_id }}
  - required fields: form_id
  - risk: deletes an Employment Hero form
- create_form_category:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/form_categories
  - risk: external Employment Hero mutation; approval required before execution
- update_form_category:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/form_categories/{{ record.form_category_id }}
  - required fields: form_category_id
  - risk: external Employment Hero mutation; approval required before execution
- delete_form_category:
  - endpoint: DELETE /v1/organisations/{{ config.organization_id }}/form_categories/{{ record.form_category_id }}
  - required fields: form_category_id
  - risk: external Employment Hero mutation; approval required before execution
- create_form_template:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/form_templates
  - risk: external Employment Hero mutation; approval required before execution
- update_form_template:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/form_templates/{{ record.template_id }}
  - required fields: template_id
  - risk: external Employment Hero mutation; approval required before execution
- delete_form_template:
  - endpoint: DELETE /v1/organisations/{{ config.organization_id }}/form_templates/{{ record.template_id }}
  - required fields: template_id
  - risk: external Employment Hero mutation; approval required before execution
- update_goal_archive_status:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/goals/{{ record.goal_id }}/archive_status
  - required fields: goal_id
  - risk: archives or restores an Employment Hero goal
- update_goal_health_status:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/goals/{{ record.goal_id }}/update_status
  - required fields: goal_id
  - risk: changes an Employment Hero goal health status
- bulk_grant_kiosk_access:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/kiosk_members/bulk_grant_access
  - risk: grants kiosk access to multiple members
- bulk_revoke_kiosk_access:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/kiosk_members/bulk_revoke_access
  - risk: revokes kiosk access from multiple members
- update_leave_balance:
  - endpoint: PUT /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/leave_balances/{{ record.id }}
  - required fields: employee_id, id
  - risk: adjusts an employee leave balance
- create_leave_request:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/leave_requests
  - required fields: employee_id
  - risk: creates an employee leave request
- create_position:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/positions
  - risk: external Employment Hero mutation; approval required before execution
- update_position:
  - endpoint: PATCH /v1/organisations/{{ config.organization_id }}/positions/{{ record.position_id }}
  - required fields: position_id
  - risk: external Employment Hero mutation; approval required before execution
- bulk_create_rostered_shifts:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/rostered_shifts/bulk_create
  - risk: creates rostered shifts in bulk and may publish them
- create_timesheet_entries:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/timesheet_entries
  - risk: creates employee timesheet entries
- create_work_site:
  - endpoint: POST /v1/organisations/{{ config.organization_id }}/work_sites
  - risk: external Employment Hero mutation; approval required before execution
- update_work_site:
  - endpoint: PUT /v1/organisations/{{ config.organization_id }}/work_sites/{{ record.work_site_id }}
  - required fields: work_site_id
  - risk: external Employment Hero mutation; approval required before execution

## Security

- read risk: external Employment Hero API reads of organisation, employee, leave, form, goal, roster, payroll-reference, document metadata, and employee subresource data
- write risk: creates, updates, archives, or deletes Employment Hero HR objects such as employees, certifications, form assets, leave requests, positions, rostered shifts, timesheets, kiosk access, and work sites
- approval: reverse ETL writes require plan preview and approval token; destructive deletes and status-changing actions are marked high risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect employment-hero
```

### Inspect as structured JSON

```bash
pm connectors inspect employment-hero --json
```

## Agent Rules

- Run pm connectors inspect employment-hero before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
