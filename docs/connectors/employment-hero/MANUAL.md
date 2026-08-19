# pm connectors inspect employment-hero

```text
NAME
  pm connectors inspect employment-hero - Employment Hero connector manual

SYNOPSIS
  pm connectors inspect employment-hero
  pm connectors inspect employment-hero --json
  pm credentials add <name> --connector employment-hero [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Employment Hero organisations, employees, HR reference data, forms, goals, rosters, employee subresources, and exposes documented JSON mutations through the Employment Hero REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  certification_id
  employee_id
  employee_ids
  form_id
  goal_id
  key_result_id
  leave_request_id
  member_ids
  organization_id
  payslip_id
  response_id
  rostered_shift_id
  template_id
  unavailability_id
  api_key (secret) (required)

ETL STREAMS
  organisations:
    primary key: id
    fields: country(string), id(string), logo_url(string), name(string), phone(string)
  organisation:
    primary key: id
    fields: country(string), id(string), logo_url(string), name(string), phone(string)
  employees:
    primary key: id
    fields: account_email(string), company_email(string), company_mobile(string), country(string), date_of_birth(string), employing_entity(string), first_name(string), gender(string), id(string), job_title(string), known_as(string), last_name(string), location(string), middle_name(string), personal_email(string), personal_mobile_number(string), primary_manager(string), role(string), start_date(string), title(string)
  employee:
    primary key: id
    fields: account_email(string), company_email(string), company_mobile(string), country(string), date_of_birth(string), employing_entity(string), first_name(string), gender(string), id(string), job_title(string), known_as(string), last_name(string), location(string), middle_name(string), personal_email(string), personal_mobile_number(string), primary_manager(string), role(string), start_date(string), title(string)
  teams:
    primary key: id
    fields: id(string), name(string), status(string)
  team_employees:
    primary key: id
    fields: company_email(string), first_name(string), id(string), last_name(string), role(string), team_id(string)
  leave_requests:
    primary key: id
    fields: comment(string), employee_id(string), end_date(string), id(string), leave_balance_amount(string), leave_category_name(string), start_date(string), status(string), total_hours(string)
  leave_request:
    primary key: id
    fields: comment(string), employee_id(string), end_date(string), id(string), leave_balance_amount(string), leave_category_name(string), start_date(string), status(string), total_hours(string)
  certifications:
    primary key: id
    fields: archived(string), description(string), id(string), mandatory(string), name(string), state(string), type(string)
  certification:
    primary key: id
    fields: archived(string), description(string), id(string), mandatory(string), name(string), state(string), type(string)
  cost_centres:
    primary key: id
    fields: code(string), id(string), name(string)
  custom_fields:
    primary key: id
    fields: field_type(string), id(string), name(string), required(string)
  employing_entities:
    primary key: id
    fields: country(string), id(string), name(string)
  forms:
    primary key: id
    fields: category_id(string), id(string), name(string), status(string)
  form:
    primary key: id
    fields: category_id(string), id(string), name(string), status(string)
  form_responses:
    primary key: id
    fields: form_id(string), id(string), member_id(string), status(string), submitted_at(string)
  form_response:
    primary key: id
    fields: form_id(string), id(string), member_id(string), status(string), submitted_at(string)
  form_assignments:
    primary key: id
    fields: form_id(string), id(string), member_id(string), status(string)
  member_form_responses:
    primary key: id
    fields: form_id(string), id(string), member_id(string), status(string), submitted_at(string)
  form_categories:
    primary key: id
    fields: id(string), name(string)
  form_templates:
    primary key: id
    fields: category_id(string), description(string), id(string), name(string)
  form_template:
    primary key: id
    fields: category_id(string), description(string), id(string), name(string)
  goals:
    primary key: id
    fields: health_status(string), id(string), owner_id(string), status(string), title(string)
  goal:
    primary key: id
    fields: health_status(string), id(string), owner_id(string), status(string), title(string)
  goal_comments:
    primary key: id
    fields: author_id(string), body(string), created_at(string), goal_id(string), id(string)
  goal_key_results:
    primary key: id
    fields: goal_id(string), health_status(string), id(string), progress(string), status(string), title(string)
  goal_key_result:
    primary key: id
    fields: goal_id(string), health_status(string), id(string), progress(string), status(string), title(string)
  kiosk_members:
    primary key: id
    fields: id(string), kiosk_access_status(string), member_id(string), name(string)
  leave_categories:
    primary key: id
    fields: code(string), id(string), name(string)
  pay_categories:
    primary key: id
    fields: code(string), id(string), name(string)
  policies:
    primary key: id
    fields: id(string), name(string), policy_type(string)
  rostered_shifts:
    primary key: id
    fields: end_date_time(string), id(string), member_id(string), published(string), start_date_time(string), status(string)
  rostered_shift:
    primary key: id
    fields: end_date_time(string), id(string), member_id(string), published(string), start_date_time(string), status(string)
  roles:
    primary key: id
    fields: id(string), name(string)
  unavailabilities:
    primary key: id
    fields: from_date(string), id(string), member_id(string), status(string), to_date(string)
  unavailability:
    primary key: id
    fields: from_date(string), id(string), member_id(string), status(string), to_date(string)
  work_locations:
    primary key: id
    fields: country(string), id(string), name(string)
  work_sites:
    primary key: id
    fields: city(string), country(string), id(string), name(string), state(string)
  work_types:
    primary key: id
    fields: id(string), name(string), payroll_info_id(string)
  bank_accounts_v1:
    primary key: id
    fields: account_name(string), account_number(string), bsb(string), employee_id(string), id(string)
  bank_accounts_v2:
    primary key: id
    fields: account_name(string), account_number(string), bsb(string), employee_id(string), id(string)
  contractor_job_histories:
    primary key: id
    fields: employee_id(string), end_date(string), id(string), job_title(string), start_date(string)
  documents:
    primary key: id
    fields: created_at(string), document_type(string), employee_id(string), id(string), name(string)
  emergency_contacts:
    primary key: id
    fields: employee_id(string), id(string), name(string), phone(string), relationship(string)
  employee_certification_details:
    primary key: id
    fields: employee_id(string), expiry_date(string), id(string), name(string), status(string)
  employee_custom_fields:
    primary key: id
    fields: employee_id(string), id(string), name(string), value(string)
  employment_histories:
    primary key: id
    fields: employee_id(string), end_date(string), id(string), job_title(string), start_date(string)
  leave_balances:
    primary key: id
    fields: balance(string), employee_id(string), id(string), leave_category_name(string)
  pay_details:
    primary key: id
    fields: employee_id(string), id(string), pay_category_id(string), rate(string)
  payslips:
    primary key: id
    fields: employee_id(string), gross_pay(string), id(string), period_end(string), period_start(string)
  payslip:
    primary key: id
    fields: employee_id(string), gross_pay(string), id(string), period_end(string), period_start(string)
  timesheet_entries:
    primary key: id
    fields: date(string), employee_id(string), end_time(string), id(string), start_time(string), status(string), units(string)
  superannuation_detail_v1:
    primary key: employee_id
    fields: employee_id(string), fund_name(string), id(string), member_number(string)
  superannuation_detail_v2:
    primary key: employee_id
    fields: employee_id(string), fund_name(string), id(string), member_number(string)
  tax_declaration_v1:
    primary key: employee_id
    fields: employee_id(string), id(string), residency_status(string), tax_file_number_status(string)
  tax_declaration_v2:
    primary key: employee_id
    fields: employee_id(string), id(string), residency_status(string), tax_file_number_status(string)
  work_eligibility:
    primary key: employee_id
    fields: employee_id(string), expiry_date(string), id(string), status(string), visa_type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_certification:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/certifications
    required fields: name
    risk: external Employment Hero mutation; approval required before execution
  update_certification:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/certifications/{{ record.certification_id }}
    required fields: certification_id
    risk: external Employment Hero mutation; approval required before execution
  archive_certification:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/certifications/{{ record.certification_id }}/archive_status
    required fields: certification_id, status
    risk: archives or restores an Employment Hero certification configuration
  delete_certification:
    endpoint: DELETE /v1/organisations/{{ config.organization_id }}/certifications/{{ record.certification_id }}
    required fields: certification_id
    risk: deletes an Employment Hero certification configuration
  create_department:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/departments
    required fields: name
    risk: external Employment Hero mutation; approval required before execution
  update_department:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/departments/{{ record.department_id }}
    required fields: department_id
    risk: external Employment Hero mutation; approval required before execution
  quick_add_employee:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/quick_add_employee
    required fields: first_name, last_name, email
    risk: external Employment Hero mutation; approval required before execution
  quick_add_contractor:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/quick_add_contractor
    required fields: first_name, last_name, email
    risk: external Employment Hero mutation; approval required before execution
  onboard_employee_async:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/polling_onboard_employee
    required fields: first_name, last_name, user_attributes
    risk: starts an asynchronous employee onboarding job; approval required before execution
  update_employee_personal_details:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/personal_details
    required fields: employee_id
    risk: external Employment Hero mutation; approval required before execution
  update_employee_employment_details:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/employment_details
    required fields: employee_id
    risk: external Employment Hero mutation; approval required before execution
  update_employee_contractor_details:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/contractor_details
    required fields: employee_id
    risk: external Employment Hero mutation; approval required before execution
  delete_employee:
    endpoint: DELETE /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}
    required fields: employee_id
    risk: deletes or removes an Employment Hero employee record
  update_employee_certification:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/certifications/{{ record.id }}
    required fields: employee_id, id
    risk: external Employment Hero mutation; approval required before execution
  delete_form:
    endpoint: DELETE /v1/organisations/{{ config.organization_id }}/forms/{{ record.form_id }}
    required fields: form_id
    risk: deletes an Employment Hero form
  create_form_category:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/form_categories
    required fields: name
    risk: external Employment Hero mutation; approval required before execution
  update_form_category:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/form_categories/{{ record.form_category_id }}
    required fields: form_category_id
    risk: external Employment Hero mutation; approval required before execution
  delete_form_category:
    endpoint: DELETE /v1/organisations/{{ config.organization_id }}/form_categories/{{ record.form_category_id }}
    required fields: form_category_id
    risk: external Employment Hero mutation; approval required before execution
  create_form_template:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/form_templates
    required fields: name
    risk: external Employment Hero mutation; approval required before execution
  update_form_template:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/form_templates/{{ record.template_id }}
    required fields: template_id
    risk: external Employment Hero mutation; approval required before execution
  delete_form_template:
    endpoint: DELETE /v1/organisations/{{ config.organization_id }}/form_templates/{{ record.template_id }}
    required fields: template_id
    risk: external Employment Hero mutation; approval required before execution
  update_goal_archive_status:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/goals/{{ record.goal_id }}/archive_status
    required fields: goal_id, status
    risk: archives or restores an Employment Hero goal
  update_goal_health_status:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/goals/{{ record.goal_id }}/update_status
    required fields: goal_id, health_status
    risk: changes an Employment Hero goal health status
  bulk_grant_kiosk_access:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/kiosk_members/bulk_grant_access
    required fields: member_ids
    risk: grants kiosk access to multiple members
  bulk_revoke_kiosk_access:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/kiosk_members/bulk_revoke_access
    required fields: member_ids
    risk: revokes kiosk access from multiple members
  update_leave_balance:
    endpoint: PUT /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/leave_balances/{{ record.id }}
    required fields: employee_id, id
    risk: adjusts an employee leave balance
  create_leave_request:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/employees/{{ record.employee_id }}/leave_requests
    required fields: employee_id, leave_category_id, start_date, end_date
    risk: creates an employee leave request
  create_position:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/positions
    required fields: name
    risk: external Employment Hero mutation; approval required before execution
  update_position:
    endpoint: PATCH /v1/organisations/{{ config.organization_id }}/positions/{{ record.position_id }}
    required fields: position_id
    risk: external Employment Hero mutation; approval required before execution
  bulk_create_rostered_shifts:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/rostered_shifts/bulk_create
    required fields: start_date_time, end_date_time, number_of_shifts
    risk: creates rostered shifts in bulk and may publish them
  create_timesheet_entries:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/timesheet_entries
    required fields: timesheets
    risk: creates employee timesheet entries
  create_work_site:
    endpoint: POST /v1/organisations/{{ config.organization_id }}/work_sites
    required fields: name
    risk: external Employment Hero mutation; approval required before execution
  update_work_site:
    endpoint: PUT /v1/organisations/{{ config.organization_id }}/work_sites/{{ record.work_site_id }}
    required fields: work_site_id
    risk: external Employment Hero mutation; approval required before execution

SECURITY
  read risk: external Employment Hero API reads of organisation, employee, leave, form, goal, roster, payroll-reference, document metadata, and employee subresource data
  write risk: creates, updates, archives, or deletes Employment Hero HR objects such as employees, certifications, form assets, leave requests, positions, rostered shifts, timesheets, kiosk access, and work sites
  approval: reverse ETL writes require plan preview and approval token; destructive deletes and status-changing actions are marked high risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect employment-hero

  # Inspect as structured JSON
  pm connectors inspect employment-hero --json

AGENT WORKFLOW
  - Run pm connectors inspect employment-hero before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
