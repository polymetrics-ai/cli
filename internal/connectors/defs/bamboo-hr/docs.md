# Overview

Reads and writes BambooHR employee, metadata, reporting, time off, applicant tracking, benefits,
goals, training, time tracking, scheduling, and webhook resources that are available through the
documented Basic-auth API surface.

Readable streams: `employees`, `meta_fields`, `meta_lists`, `time_off_types`, `applications`,
`application_details`, `hiring_leads`, `job_summaries`, `company_locations`, `statuses`,
`company_benefits`, `company_benefit_types`, `company_benefit`, `employee_benefits`,
`member_benefit_events`, `benefit_coverages`, `member_benefits`, `benefit_deduction_types`,
`company_profile_integrations`, `company_eins`, `company_information`, `reports`, `report_by_id`,
`datasets_v1`, `fields_from_dataset_v1`, `employee_dependents`, `employee_dependent`,
`employee_roster`, `changed_employee_ids`, `changed_employee_table_data`, `time_off_balance`,
`employee_time_off_policies`, `employee_table_data`, `all_currency_types`, `states_by_country_id`,
`tabular_fields`, `time_off_policies`, `users`, `goals`, `goals_aggregate_v1`,
`alignable_goal_options`, `goal_creation_permission`, `goals_filters_v1`, `goal_share_options`,
`goal_aggregate`, `goal_comments`, `company_report`, `scheduling_list_schedules`,
`scheduling_get_schedule`, `scheduling_list_shift_assessments`, `scheduling_list_shifts`,
`scheduling_get_shift`, `scheduling_list_timezones`, `break_assessments`, `break_policies`,
`break_policy`, `break_policy_breaks`, `break_policy_employees`, `break`,
`employee_break_availabilities`, `employee_break_policies`, `projects`, `project`, `project_tasks`,
`shift_differentials`, `shift_differential`, `task`, `time_off_requests`, `whos_out`,
`timesheet_entries`, `time_tracking_record`, `training_categories`, `training_types`, `webhooks`,
`monitor_fields`, `post_fields`, `webhook`, `employee_time_off_policies_v1_1`,
`goals_aggregate_v1_1`, `goals_filters_v1_1`, `datasets_v1_2`, `fields_from_dataset_v1_2`,
`goals_aggregate_v1_2`, `goals_filters_v1_2`.

Write actions: `create_application_comment`, `update_applicant_status`, `add_new_company_benefit`,
`delete_company_benefit`, `update_company_benefit`, `create_employee_benefit`,
`add_benefit_group_employee`, `clear_employee_deposit`, `add_employee_deposit`,
`add_employee_paystub`, `clear_employee_paystub`, `add_employee_unpaid_paystubs`,
`clear_employee_unpaid_paystubs`, `clear_employee_withholding`, `add_employee_withholding`,
`create_employee_dependent`, `update_employee_dependent`, `adjust_time_off_balance`,
`create_time_off_history`, `assign_time_off_policies`, `create_time_off_request`,
`create_table_row`, `delete_employee_table_row`, `update_table_row`, `update_list_field_values`,
`create_goal`, `delete_goal`, `update_goal_v1`, `close_goal`, `create_goal_comment`,
`delete_goal_comment`, `update_goal_comment`, `update_goal_milestone_progress`,
`update_goal_progress`, `reopen_goal`, `update_goal_sharing`, `create_scheduling_create_schedule`,
`delete_scheduling_delete_schedule`, `update_scheduling_update_schedule`,
`create_scheduling_create_shift`, `create_scheduling_publish_shifts`,
`delete_scheduling_delete_shift`, `update_scheduling_update_shift`, `create_break_policy`,
`delete_break_policy`, `update_break_policy`, `assign_employees_to_break_policy`,
`set_break_policy_employees`, `create_break`, `replace_breaks_for_break_policy`,
`sync_break_policy`, `create_unassign_employees_from_break_policy`, `delete_break`, `update_break`,
`create_project`, `delete_project`, `update_project`, `create_project_task`,
`create_shift_differential`, `delete_shift_differential`, `update_shift_differential`,
`delete_task`, `update_task`, `update_time_off_request_status`, `delete_clock_entries`,
`store_clock_entries`, `delete_timesheet_clock_entries_via_post`,
`create_or_update_timesheet_clock_entries`, `clock_in`, `clock_out`, `store_daily_entries`,
`clock_in_data`, `clock_out_employee_at_specific_time`, `create_timesheet_clock_in_entry`,
`create_timesheet_clock_out_entry`, `delete_timesheet_hour_entries_via_post`,
`create_or_update_timesheet_hour_entries`, `create_time_tracking_project`,
`approve_employee_timesheets`, `clock_out_and_approve_employee_timesheets`, and 21 more.

Service API documentation: https://documentation.bamboohr.com/reference/getting-started.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); BambooHR API key used as the HTTP Basic username with
  literal password x.
- `application_id` (optional, string).
- `break_id` (optional, string).
- `break_policy_breaks_id` (optional, string).
- `break_policy_employees_id` (optional, string).
- `break_policy_id` (optional, string).
- `changed_employee_ids_since` (optional, string).
- `changed_employee_table_data_since` (optional, string).
- `company_benefit_id` (optional, string).
- `company_report_id` (optional, string).
- `country_id` (optional, string).
- `dataset_name` (optional, string).
- `employee_break_availabilities_id` (optional, string).
- `employee_break_policies_id` (optional, string).
- `employee_dependent_id` (optional, string).
- `employee_id` (optional, string).
- `employee_table_data_id` (optional, string).
- `goal_id` (optional, string).
- `goal_share_options_search` (optional, string).
- `member_benefits_calendar_year` (optional, string).
- `project_id` (optional, string).
- `report_id` (optional, string).
- `scheduling_get_schedule_id` (optional, string).
- `scheduling_get_shift_id` (optional, string).
- `shift_differential_id` (optional, string).
- `subdomain` (required, string); Your BambooHR account subdomain, the companySubDomain in
  https://companySubDomain.bamboohr.com.
- `table` (optional, string).
- `task_id` (optional, string).
- `time_off_requests_end` (optional, string).
- `time_off_requests_start` (optional, string).
- `time_tracking_record_id` (optional, string).
- `timesheet_entries_end` (optional, string).
- `timesheet_entries_start` (optional, string).
- `webhook_id` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use base URL `https://{{ config.subdomain }}.bamboohr.com` after applying configuration
defaults.

Connection checks call GET `/api/v1/meta/fields`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `reports`, `report_by_id`, `fields_from_dataset_v1`,
`employee_roster`, `fields_from_dataset_v1_2`; none: `meta_fields`, `meta_lists`, `time_off_types`,
`applications`, `application_details`, `hiring_leads`, `job_summaries`, `company_locations`,
`statuses`, `company_benefits`, `company_benefit_types`, `company_benefit`, `employee_benefits`,
`member_benefit_events`, `benefit_coverages`, `member_benefits`, `benefit_deduction_types`,
`company_profile_integrations`, `company_eins`, `company_information`, `datasets_v1`,
`employee_dependents`, `employee_dependent`, `changed_employee_ids`, `changed_employee_table_data`,
`time_off_balance`, `employee_time_off_policies`, `employee_table_data`, `all_currency_types`,
`states_by_country_id`, `tabular_fields`, `time_off_policies`, `users`, `goals`,
`goals_aggregate_v1`, `alignable_goal_options`, `goal_creation_permission`, `goals_filters_v1`,
`goal_share_options`, `goal_aggregate`, `goal_comments`, `company_report`,
`scheduling_list_schedules`, `scheduling_get_schedule`, `scheduling_list_shift_assessments`,
`scheduling_list_shifts`, `scheduling_get_shift`, `scheduling_list_timezones`, `break_assessments`,
`break_policies`, `break_policy`, `break_policy_breaks`, `break_policy_employees`, `break`,
`employee_break_availabilities`, `employee_break_policies`, `projects`, `project`, `project_tasks`,
`shift_differentials`, and 18 more; page_number: `employees`.

- `employees`: GET `/api/v1/employees/directory` - records path `employees`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `display_name`, `first_name`, `id`, `job_title`, `last_name`, `mobile_phone`, `photo_url`,
  `preferred_name`, `work_email`, `work_phone`.
- `meta_fields`: GET `/api/v1/meta/fields` - records at response root; computed output fields `id`.
- `meta_lists`: GET `/api/v1/meta/lists` - records at response root; computed output fields
  `field_id`.
- `time_off_types`: GET `/api/v1/meta/time_off/types` - records path `timeOffTypes`; computed output
  fields `id`.
- `applications`: GET `/api/v1/applicant_tracking/applications` - records path `applications`.
- `application_details`: GET `/api/v1/applicant_tracking/applications/{{ config.application_id }}` -
  records path `questionsAndAnswers`.
- `hiring_leads`: GET `/api/v1/applicant_tracking/hiring_leads` - records at response root.
- `job_summaries`: GET `/api/v1/applicant_tracking/jobs` - records at response root.
- `company_locations`: GET `/api/v1/applicant_tracking/locations` - records at response root.
- `statuses`: GET `/api/v1/applicant_tracking/statuses` - records at response root.
- `company_benefits`: GET `/api/v1/benefit/company_benefit` - records path `companyBenefits`.
- `company_benefit_types`: GET `/api/v1/benefit/company_benefit/type` - records at response root.
- `company_benefit`: GET `/api/v1/benefit/company_benefit/{{ config.company_benefit_id }}` -
  single-object response; records path `.`.
- `employee_benefits`: GET `/api/v1/benefit/employee_benefit` - records path `employeeBenefits`.
- `member_benefit_events`: GET `/api/v1/benefit/member_benefit` - records path `members`.
- `benefit_coverages`: GET `/api/v1/benefitcoverages` - records path `Benefit Coverages`.
- `member_benefits`: GET `/api/v1/benefits/member-benefits` - records path `data`; query
  `calendarYear`=`{{ config.member_benefits_calendar_year }}`.
- `benefit_deduction_types`: GET `/api/v1/benefits/settings/deduction_types/all` - records at
  response root.
- `company_profile_integrations`: GET `/api/v1/company-profile-integrations` - single-object
  response; records path `.`.
- `company_eins`: GET `/api/v1/company_eins` - single-object response; records path `.`.
- `company_information`: GET `/api/v1/company_information` - single-object response; records path
  `.`.
- `reports`: GET `/api/v1/custom-reports` - records path `reports`; query `page_size`=`100`; follows
  a next-page URL from the response body; URL path `pagination.next_page`; next URLs stay on the
  configured API host.
- `report_by_id`: GET `/api/v1/custom-reports/{{ config.report_id }}` - records path `data`; query
  `page_size`=`100`; follows a next-page URL from the response body; URL path
  `pagination.next_page`; next URLs stay on the configured API host.
- `datasets_v1`: GET `/api/v1/datasets` - records path `datasets`.
- `fields_from_dataset_v1`: GET `/api/v1/datasets/{{ config.dataset_name }}/fields` - records path
  `fields`; query `page_size`=`100`; follows a next-page URL from the response body; URL path
  `pagination.next_page`; next URLs stay on the configured API host.
- `employee_dependents`: GET `/api/v1/employeedependents` - records path `Employee Dependents`.
- `employee_dependent`: GET `/api/v1/employeedependents/{{ config.employee_dependent_id }}` -
  records path `Employee Dependents`.
- `employee_roster`: GET `/api/v1/employees` - records path `data`; query `page[limit]`=`250`;
  follows a next-page URL from the response body; URL path `_links.next.href`; next URLs stay on the
  configured API host.
- `changed_employee_ids`: GET `/api/v1/employees/changed` - single-object response; records path
  `.`; query `since`=`{{ config.changed_employee_ids_since }}`.
- `changed_employee_table_data`: GET `/api/v1/employees/changed/tables/{{ config.table }}` -
  single-object response; records path `.`; query `since`=`{{
  config.changed_employee_table_data_since }}`.
- `time_off_balance`: GET `/api/v1/employees/{{ config.employee_id }}/time_off/calculator` - records
  at response root.
- `employee_time_off_policies`: GET `/api/v1/employees/{{ config.employee_id }}/time_off/policies` -
  records at response root.
- `employee_table_data`: GET `/api/v1/employees/{{ config.employee_table_data_id }}/tables/{{
  config.table }}` - records at response root.
- `all_currency_types`: GET `/api/v1/meta/currency/types` - records at response root.
- `states_by_country_id`: GET `/api/v1/meta/provinces/{{ config.country_id }}` - records path
  `options`.
- `tabular_fields`: GET `/api/v1/meta/tables` - records at response root.
- `time_off_policies`: GET `/api/v1/meta/time_off/policies` - records at response root.
- `users`: GET `/api/v1/meta/users` - single-object response; records path `.`.
- `goals`: GET `/api/v1/performance/employees/{{ config.employee_id }}/goals` - records path
  `goals`.
- `goals_aggregate_v1`: GET `/api/v1/performance/employees/{{ config.employee_id }}/goals/aggregate`
  - records path `goals`.
- `alignable_goal_options`: GET `/api/v1/performance/employees/{{ config.employee_id
  }}/goals/alignmentOptions` - records path `alignsWithOptions`.
- `goal_creation_permission`: GET `/api/v1/performance/employees/{{ config.employee_id
  }}/goals/canCreateGoals` - single-object response; records path `.`.
- `goals_filters_v1`: GET `/api/v1/performance/employees/{{ config.employee_id }}/goals/filters` -
  records path `filters`.
- `goal_share_options`: GET `/api/v1/performance/employees/{{ config.employee_id
  }}/goals/shareOptions` - records path `persons`; query `search`=`{{
  config.goal_share_options_search }}`.
- `goal_aggregate`: GET `/api/v1/performance/employees/{{ config.employee_id }}/goals/{{
  config.goal_id }}/aggregate` - records path `comments`.
- `goal_comments`: GET `/api/v1/performance/employees/{{ config.employee_id }}/goals/{{
  config.goal_id }}/comments` - records path `comments`.
- `company_report`: GET `/api/v1/reports/{{ config.company_report_id }}` - records path `employees`.
- `scheduling_list_schedules`: GET `/api/v1/scheduling/schedules` - records path `data`.
- `scheduling_get_schedule`: GET `/api/v1/scheduling/schedules/{{ config.scheduling_get_schedule_id
  }}` - single-object response; records path `.`.
- `scheduling_list_shift_assessments`: GET `/api/v1/scheduling/shift-assessments` - records path
  `data`.
- `scheduling_list_shifts`: GET `/api/v1/scheduling/shifts` - records path `data`.
- `scheduling_get_shift`: GET `/api/v1/scheduling/shifts/{{ config.scheduling_get_shift_id }}` -
  single-object response; records path `.`.
- `scheduling_list_timezones`: GET `/api/v1/scheduling/timezones` - records path `data`.
- `break_assessments`: GET `/api/v1/time-tracking/break-assessments` - single-object response;
  records path `.`.
- `break_policies`: GET `/api/v1/time-tracking/break-policies` - single-object response; records
  path `.`.
- `break_policy`: GET `/api/v1/time-tracking/break-policies/{{ config.break_policy_id }}` -
  single-object response; records path `.`.
- `break_policy_breaks`: GET `/api/v1/time-tracking/break-policies/{{ config.break_policy_breaks_id
  }}/breaks` - single-object response; records path `.`.
- `break_policy_employees`: GET `/api/v1/time-tracking/break-policies/{{
  config.break_policy_employees_id }}/employees` - single-object response; records path `.`.
- `break`: GET `/api/v1/time-tracking/breaks/{{ config.break_id }}` - single-object response;
  records path `.`.
- `employee_break_availabilities`: GET `/api/v1/time-tracking/employees/{{
  config.employee_break_availabilities_id }}/break-availabilities` - records at response root.
- `employee_break_policies`: GET `/api/v1/time-tracking/employees/{{
  config.employee_break_policies_id }}/break-policies` - single-object response; records path `.`.
- `projects`: GET `/api/v1/time-tracking/projects` - single-object response; records path `.`.
- `project`: GET `/api/v1/time-tracking/projects/{{ config.project_id }}` - single-object response;
  records path `.`.
- `project_tasks`: GET `/api/v1/time-tracking/projects/{{ config.project_id }}/tasks` - records path
  `data`.
- `shift_differentials`: GET `/api/v1/time-tracking/shift-differentials` - single-object response;
  records path `.`.
- `shift_differential`: GET `/api/v1/time-tracking/shift-differentials/{{
  config.shift_differential_id }}` - records path `times`.
- `task`: GET `/api/v1/time-tracking/tasks/{{ config.task_id }}` - single-object response; records
  path `.`.
- `time_off_requests`: GET `/api/v1/time_off/requests` - records at response root; query `end`=`{{
  config.time_off_requests_end }}`; `start`=`{{ config.time_off_requests_start }}`.
- `whos_out`: GET `/api/v1/time_off/whos_out` - records at response root.
- `timesheet_entries`: GET `/api/v1/time_tracking/timesheet_entries` - records at response root;
  query `end`=`{{ config.timesheet_entries_end }}`; `start`=`{{ config.timesheet_entries_start }}`.
- `time_tracking_record`: GET `/api/v1/timetracking/record/{{ config.time_tracking_record_id }}` -
  single-object response; records path `.`.
- `training_categories`: GET `/api/v1/training/category` - single-object response; records path `.`.
- `training_types`: GET `/api/v1/training/type` - single-object response; records path `.`.
- `webhooks`: GET `/api/v1/webhooks` - records path `webhooks`.
- `monitor_fields`: GET `/api/v1/webhooks/monitor_fields` - records path `fields`.
- `post_fields`: GET `/api/v1/webhooks/post-fields` - records path `fields`.
- `webhook`: GET `/api/v1/webhooks/{{ config.webhook_id }}` - records path `errors`.
- `employee_time_off_policies_v1_1`: GET `/api/v1_1/employees/{{ config.employee_id
  }}/time_off/policies` - records at response root.
- `goals_aggregate_v1_1`: GET `/api/v1_1/performance/employees/{{ config.employee_id
  }}/goals/aggregate` - records path `goals`.
- `goals_filters_v1_1`: GET `/api/v1_1/performance/employees/{{ config.employee_id }}/goals/filters`
  - records path `filters`.
- `datasets_v1_2`: GET `/api/v1_2/datasets` - records path `datasets`.
- `fields_from_dataset_v1_2`: GET `/api/v1_2/datasets/{{ config.dataset_name }}/fields` - records
  path `fields`; query `page_size`=`100`; follows a next-page URL from the response body; URL path
  `pagination.next_page`; next URLs stay on the configured API host.
- `goals_aggregate_v1_2`: GET `/api/v1_2/performance/employees/{{ config.employee_id
  }}/goals/aggregate` - records path `goals`.
- `goals_filters_v1_2`: GET `/api/v1_2/performance/employees/{{ config.employee_id }}/goals/filters`
  - records path `filters`.

## Write actions & risks

Overall write risk: creates, updates, assigns, approves, or deletes BambooHR HR records according to
the selected reverse-ETL action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_application_comment`: POST `/api/v1/applicant_tracking/applications/{{
  record.application_id }}/comments` - kind `create`; body type `json`; path fields
  `application_id`; required record fields `application_id`, `comment`; accepted fields
  `application_id`, `comment`, `type`; risk: Create Job Application Comment through the BambooHR
  API.
- `update_applicant_status`: POST `/api/v1/applicant_tracking/applications/{{ record.application_id
  }}/status` - kind `update`; body type `json`; path fields `application_id`; required record fields
  `application_id`, `status`; accepted fields `application_id`, `status`; risk: Update Applicant
  Status through the BambooHR API.
- `add_new_company_benefit`: POST `/api/v1/benefit/company_benefit` - kind `create`; body type
  `json`; risk: Add a new company benefit through the BambooHR API.
- `delete_company_benefit`: DELETE `/api/v1/benefit/company_benefit/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Delete a company benefit.
- `update_company_benefit`: PUT `/api/v1/benefit/company_benefit/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `benefitType`,
  `benefitVendorId`, `companyBenefitName`, `deductionTypeId`, `description`, `endDate`, `id`,
  `meetAcaMin`, `minEssentialCoverage`, `planUrl`, `reimbursementAmount`,
  `reimbursementCurrencyCode`, `reimbursementFrequency`, `safeHarbor`, `ssoLoginUrl`,
  `ssoLoginUrlLinkText`, `startDate`; risk: Update a company benefit through the BambooHR API.
- `create_employee_benefit`: POST `/api/v1/benefit/employee_benefit` - kind `create`; body type
  `json`; accepted fields `benefitPlanCoverageId`, `companyAmount`, `companyAmountType`,
  `companyAnnualMax`, `companyBenefitId`, `companyBenefitName`, `companyCapAmount`,
  `companyCapAmountType`, `companyPercentBasedOn`, `coverageLevel`, `currencyCode`,
  `deductionEndDate`, `deductionStartDate`, `effectiveDate`, `employeeAmount`, `employeeAmountType`,
  `employeeAnnualMax`, `employeeCapAmount`, and 4 more; risk: Add an employee benefit through the
  BambooHR API.
- `add_benefit_group_employee`: POST `/api/v1/benefitgroupemployees` - kind `create`; body type
  `json`; accepted fields `benefitGroupId`, `employeeId`, `endDate`, `startDate`; risk: Add a
  benefit group employee through the BambooHR API.
- `clear_employee_deposit`: DELETE `/api/v1/employee_direct_deposit_accounts/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Clear an employee's direct deposit information.
- `add_employee_deposit`: POST `/api/v1/employee_direct_deposit_accounts/{{ record.id }}` - kind
  `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `accounts`, `id`; risk: Add an employee's direct deposit information through the BambooHR API.
- `add_employee_paystub`: POST `/api/v1/employee_pay_stub` - kind `create`; body type `json`;
  accepted fields `additionalFed`, `additionalLocal`, `additionalState`, `currencyCode`,
  `deductions`, `deductionsAmount`, `dependentsAmount`, `deposits`, `employeeId`,
  `externalRecordId`, `fedWitholding`, `federalType`, `gross`, `localWithholding`, `net`,
  `otherIncome`, `payDate`, `payPeriodFrom`, and 20 more; risk: Add an employee's paystub through
  the BambooHR API.
- `clear_employee_paystub`: DELETE `/api/v1/employee_pay_stub/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes BambooHR data:
  Delete an employee's paystub.
- `add_employee_unpaid_paystubs`: POST `/api/v1/employee_unpaid_pay_stubs` - kind `create`; body
  type `json`; accepted fields `employeeId`, `unpaidPeriods`; risk: Add an employee's unpaid
  paystubs through the BambooHR API.
- `clear_employee_unpaid_paystubs`: DELETE `/api/v1/employee_unpaid_pay_stubs/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Deletes BambooHR data: Clear an employee's unpaid paystubs.
- `clear_employee_withholding`: DELETE `/api/v1/employee_withholding/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Clear an employee's default withholdings.
- `add_employee_withholding`: POST `/api/v1/employee_withholding/{{ record.id }}` - kind `create`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `additionalFed`,
  `additionalLocal`, `additionalState`, `fedWithholding`, `id`, `localWithholding`,
  `stateWithholding`, `taxLocal`, `taxState`; risk: Add an employee's default withholdings through
  the BambooHR API.
- `create_employee_dependent`: POST `/api/v1/employeedependents` - kind `create`; body type `json`;
  required record fields `employeeId`; accepted fields `addressLine1`, `addressLine2`, `city`,
  `country`, `dateOfBirth`, `employeeId`, `firstName`, `gender`, `homePhone`, `isStudent`,
  `isUsCitizen`, `lastName`, `middleName`, `relationship`, `sin`, `ssn`, `state`, `zipCode`; risk:
  Create Employee Dependent through the BambooHR API.
- `update_employee_dependent`: PUT `/api/v1/employeedependents/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `employeeId`; accepted fields
  `addressLine1`, `addressLine2`, `city`, `country`, `dateOfBirth`, `employeeId`, `firstName`,
  `gender`, `homePhone`, `id`, `isStudent`, `isUsCitizen`, `lastName`, `middleName`, `relationship`,
  `sin`, `ssn`, `state`, and 1 more; risk: Update Employee Dependent through the BambooHR API.
- `adjust_time_off_balance`: PUT `/api/v1/employees/{{ record.employee_id
  }}/time_off/balance_adjustment` - kind `update`; body type `json`; path fields `employee_id`;
  required record fields `employee_id`, `amount`, `date`, `timeOffTypeId`; accepted fields `amount`,
  `date`, `employee_id`, `note`, `timeOffTypeId`; risk: Adjust Time Off Balance through the BambooHR
  API.
- `create_time_off_history`: PUT `/api/v1/employees/{{ record.employee_id }}/time_off/history` -
  kind `update`; body type `json`; path fields `employee_id`; required record fields `employee_id`,
  `date`; accepted fields `amount`, `date`, `employee_id`, `eventType`, `note`, `timeOffRequestId`,
  `timeOffTypeId`; risk: Create Time Off History Item through the BambooHR API.
- `assign_time_off_policies`: PUT `/api/v1/employees/{{ record.employee_id }}/time_off/policies` -
  kind `update`; body type `json`; path fields `employee_id`; required record fields `employee_id`;
  accepted fields `employee_id`; risk: Assign Time Off Policies through the BambooHR API.
- `create_time_off_request`: PUT `/api/v1/employees/{{ record.employee_id }}/time_off/request` -
  kind `update`; body type `json`; path fields `employee_id`; required record fields `employee_id`,
  `status`, `start`, `end`, `timeOffTypeId`; accepted fields `amount`, `dates`, `employee_id`,
  `end`, `notes`, `previousRequest`, `start`, `status`, `timeOffTypeId`; risk: Create Time Off
  Request through the BambooHR API.
- `create_table_row`: POST `/api/v1/employees/{{ record.id }}/tables/{{ record.table }}` - kind
  `create`; body type `json`; path fields `id`, `table`; required record fields `id`, `table`;
  accepted fields `date`, `department`, `division`, `id`, `jobTitle`, `location`, `reportsTo`,
  `table`, `teams`; risk: Create Table Row through the BambooHR API.
- `delete_employee_table_row`: DELETE `/api/v1/employees/{{ record.id }}/tables/{{ record.table
  }}/{{ record.row_id }}` - kind `delete`; body type `none`; path fields `id`, `table`, `row_id`;
  required record fields `id`, `table`, `row_id`; accepted fields `id`, `row_id`, `table`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes BambooHR
  data: Delete Employee Table Row.
- `update_table_row`: POST `/api/v1/employees/{{ record.id }}/tables/{{ record.table }}/{{
  record.row_id }}` - kind `update`; body type `json`; path fields `id`, `table`, `row_id`; required
  record fields `id`, `table`, `row_id`; accepted fields `date`, `department`, `division`, `id`,
  `jobTitle`, `location`, `reportsTo`, `row_id`, `table`, `teams`; risk: Update Table Row through
  the BambooHR API.
- `update_list_field_values`: PUT `/api/v1/meta/lists/{{ record.list_field_id }}` - kind `update`;
  body type `json`; path fields `list_field_id`; required record fields `list_field_id`; accepted
  fields `list_field_id`, `options`; risk: Update List Field Values through the BambooHR API.
- `create_goal`: POST `/api/v1/performance/employees/{{ record.employee_id }}/goals` - kind
  `create`; body type `json`; path fields `employee_id`; required record fields `employee_id`,
  `title`, `dueDate`, `sharedWithEmployeeIds`; accepted fields `alignsWithOptionId`,
  `completionDate`, `description`, `dueDate`, `employee_id`, `milestones`, `percentComplete`,
  `sharedWithEmployeeIds`, `title`; risk: Create Goal through the BambooHR API.
- `delete_goal`: DELETE `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}` - kind `delete`; body type `none`; path fields `employee_id`, `goal_id`;
  required record fields `employee_id`, `goal_id`; accepted fields `employee_id`, `goal_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes BambooHR
  data: Delete Goal.
- `update_goal_v1`: PUT `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}` - kind `update`; body type `json`; path fields `employee_id`, `goal_id`;
  required record fields `employee_id`, `goal_id`, `dueDate`, `sharedWithEmployeeIds`, `title`;
  accepted fields `alignsWithOptionId`, `completionDate`, `description`, `dueDate`, `employee_id`,
  `goal_id`, `percentComplete`, `sharedWithEmployeeIds`, `title`; risk: Update Goal (v1) through the
  BambooHR API.
- `close_goal`: POST `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{ record.goal_id
  }}/close` - kind `update`; body type `json`; path fields `employee_id`, `goal_id`; required record
  fields `employee_id`, `goal_id`; accepted fields `comment`, `employee_id`, `goal_id`; risk: Close
  Goal through the BambooHR API.
- `create_goal_comment`: POST `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}/comments` - kind `create`; body type `json`; path fields `employee_id`,
  `goal_id`; required record fields `employee_id`, `goal_id`, `text`; accepted fields `employee_id`,
  `goal_id`, `text`; risk: Create Goal Comment through the BambooHR API.
- `delete_goal_comment`: DELETE `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}/comments/{{ record.comment_id }}` - kind `delete`; body type `none`; path fields
  `employee_id`, `goal_id`, `comment_id`; required record fields `employee_id`, `goal_id`,
  `comment_id`; accepted fields `comment_id`, `employee_id`, `goal_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes BambooHR data: Delete Goal
  Comment.
- `update_goal_comment`: PUT `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}/comments/{{ record.comment_id }}` - kind `update`; body type `json`; path fields
  `employee_id`, `goal_id`, `comment_id`; required record fields `employee_id`, `goal_id`,
  `comment_id`, `text`; accepted fields `comment_id`, `employee_id`, `goal_id`, `text`; risk: Update
  Goal Comment through the BambooHR API.
- `update_goal_milestone_progress`: PUT `/api/v1/performance/employees/{{ record.employee_id
  }}/goals/{{ record.goal_id }}/milestones/{{ record.milestone_id }}/progress` - kind `update`; body
  type `json`; path fields `employee_id`, `goal_id`, `milestone_id`; required record fields
  `employee_id`, `goal_id`, `milestone_id`, `complete`; accepted fields `complete`, `employee_id`,
  `goal_id`, `milestone_id`; risk: Update Milestone Progress through the BambooHR API.
- `update_goal_progress`: PUT `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}/progress` - kind `update`; body type `json`; path fields `employee_id`,
  `goal_id`; required record fields `employee_id`, `goal_id`, `percentComplete`; accepted fields
  `completionDate`, `employee_id`, `goal_id`, `percentComplete`; risk: Update Goal Progress through
  the BambooHR API.
- `reopen_goal`: POST `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}/reopen` - kind `update`; body type `json`; path fields `employee_id`, `goal_id`;
  required record fields `employee_id`, `goal_id`; accepted fields `employee_id`, `goal_id`; risk:
  Reopen Goal through the BambooHR API.
- `update_goal_sharing`: PUT `/api/v1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}/sharedWith` - kind `update`; body type `json`; path fields `employee_id`,
  `goal_id`; required record fields `employee_id`, `goal_id`; accepted fields `employee_id`,
  `goal_id`, `sharedWithEmployeeIds`; risk: Update Goal Sharing through the BambooHR API.
- `create_scheduling_create_schedule`: POST `/api/v1/scheduling/schedules` - kind `create`; body
  type `json`; required record fields `name`, `locationId`, `startOfWeek`; accepted fields
  `earlyClockInThreshold`, `employeeIds`, `locationId`, `managerUserIds`, `name`, `startOfWeek`,
  `timezone`; risk: Create Schedule through the BambooHR API.
- `delete_scheduling_delete_schedule`: DELETE `/api/v1/scheduling/schedules/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Delete Schedule.
- `update_scheduling_update_schedule`: PATCH `/api/v1/scheduling/schedules/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `earlyClockInThreshold`, `employeeIds`, `id`, `locationId`, `managerUserIds`, `name`,
  `startOfWeek`, `timezone`; risk: Update Schedule through the BambooHR API.
- `create_scheduling_create_shift`: POST `/api/v1/scheduling/shifts` - kind `create`; body type
  `json`; required record fields `scheduleId`, `status`, `color`, `timezone`, `start`, `end`;
  accepted fields `capacity`, `color`, `employeeIds`, `end`, `name`, `recurrenceDtend`,
  `recurrenceDtstart`, `recurrenceRule`, `recurrenceUntil`, `scheduleId`, `start`, `status`,
  `timezone`; risk: Create Shift through the BambooHR API.
- `create_scheduling_publish_shifts`: POST `/api/v1/scheduling/shifts/publish` - kind `create`; body
  type `json`; required record fields `shiftIds`; accepted fields `shiftIds`; risk: Publish Shifts
  through the BambooHR API.
- `delete_scheduling_delete_shift`: DELETE `/api/v1/scheduling/shifts/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Delete Shift.
- `update_scheduling_update_shift`: PATCH `/api/v1/scheduling/shifts/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `capacity`, `color`, `employeeIds`, `end`, `id`, `name`, `recurrenceDtend`, `recurrenceDtstart`,
  `recurrenceEditOption`, `recurrenceRule`, `recurrenceUntil`, `start`, `timezone`,
  `unpublishedChanges`; risk: Update Shift through the BambooHR API.
- `create_break_policy`: POST `/api/v1/time-tracking/break-policies` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `allEmployeesAssigned`, `breaks`,
  `description`, `employeeIds`, `name`; risk: Create Break Policy through the BambooHR API.
- `delete_break_policy`: DELETE `/api/v1/time-tracking/break-policies/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Delete Break Policy.
- `update_break_policy`: PATCH `/api/v1/time-tracking/break-policies/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `allEmployeesAssigned`, `description`, `id`, `name`; risk: Update Break Policy through the
  BambooHR API.
- `assign_employees_to_break_policy`: POST `/api/v1/time-tracking/break-policies/{{ record.id
  }}/assign` - kind `update`; body type `json`; path fields `id`; required record fields `id`,
  `employeeIds`; accepted fields `employeeIds`, `id`; risk: Assign Employees to Break Policy through
  the BambooHR API.
- `set_break_policy_employees`: PUT `/api/v1/time-tracking/break-policies/{{ record.id }}/assign` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`, `employeeIds`;
  accepted fields `employeeIds`, `id`; risk: Set Employees for Break Policy through the BambooHR
  API.
- `create_break`: POST `/api/v1/time-tracking/break-policies/{{ record.id }}/breaks` - kind
  `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `availabilityEndTime`, `availabilityMaxHoursWorked`, `availabilityMinHoursWorked`,
  `availabilityStartTime`, `availabilityType`, `duration`, `id`, `name`, `paid`, `policyId`; risk:
  Create Break through the BambooHR API.
- `replace_breaks_for_break_policy`: PUT `/api/v1/time-tracking/break-policies/{{ record.id
  }}/breaks` - kind `update`; body type `json`; path fields `id`; required record fields `id`;
  accepted fields `id`; risk: Replace Breaks for Break Policy through the BambooHR API.
- `sync_break_policy`: PUT `/api/v1/time-tracking/break-policies/{{ record.id }}/sync` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `allEmployeesAssigned`, `breaks`, `description`, `employeeIds`, `id`, `name`; risk: Sync Break
  Policy through the BambooHR API.
- `create_unassign_employees_from_break_policy`: POST `/api/v1/time-tracking/break-policies/{{
  record.id }}/unassign` - kind `create`; body type `json`; path fields `id`; required record fields
  `id`, `employeeIds`; accepted fields `employeeIds`, `id`; risk: Unassign Employees from Break
  Policy through the BambooHR API.
- `delete_break`: DELETE `/api/v1/time-tracking/breaks/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes BambooHR data:
  Delete Break.
- `update_break`: PATCH `/api/v1/time-tracking/breaks/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `availabilityEndTime`,
  `availabilityMaxHoursWorked`, `availabilityMinHoursWorked`, `availabilityStartTime`,
  `availabilityType`, `duration`, `id`, `name`, `paid`, `policyId`; risk: Update Break through the
  BambooHR API.
- `create_project`: POST `/api/v1/time-tracking/projects` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `allEmployeesAssigned`, `billable`, `employeeIds`,
  `includeInPayroll`, `name`, `tasks`; risk: Create Time Tracking Project through the BambooHR API.
- `delete_project`: DELETE `/api/v1/time-tracking/projects/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes BambooHR data:
  Delete Time Tracking Project.
- `update_project`: PATCH `/api/v1/time-tracking/projects/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields
  `allEmployeesAssigned`, `archived`, `billable`, `employeeIds`, `hasTasks`, `id`,
  `includeInPayroll`, `name`; risk: Update Time Tracking Project through the BambooHR API.
- `create_project_task`: POST `/api/v1/time-tracking/projects/{{ record.project_id }}/tasks` - kind
  `create`; body type `json`; path fields `project_id`; required record fields `project_id`, `name`;
  accepted fields `billable`, `name`, `project_id`; risk: Create Time Tracking Project Task through
  the BambooHR API.
- `create_shift_differential`: POST `/api/v1/time-tracking/shift-differentials` - kind `create`;
  body type `json`; required record fields `name`, `rate`, `rateType`, `times`; accepted fields
  `allowAllEmployees`, `employeeIds`, `name`, `rate`, `rateType`, `times`; risk: Create Time
  Tracking Shift Differential through the BambooHR API.
- `delete_shift_differential`: DELETE `/api/v1/time-tracking/shift-differentials/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Deletes BambooHR data: Delete Time Tracking Shift Differential.
- `update_shift_differential`: PATCH `/api/v1/time-tracking/shift-differentials/{{ record.id }}` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `allowAllEmployees`, `archived`, `employeeIds`, `id`, `name`, `rate`, `rateType`, `times`; risk:
  Update Time Tracking Shift Differential through the BambooHR API.
- `delete_task`: DELETE `/api/v1/time-tracking/tasks/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes BambooHR data:
  Delete Time Tracking Task.
- `update_task`: PATCH `/api/v1/time-tracking/tasks/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `billable`, `id`, `name`;
  risk: Update Time Tracking Task through the BambooHR API.
- `update_time_off_request_status`: PUT `/api/v1/time_off/requests/{{ record.request_id }}/status` -
  kind `update`; body type `json`; path fields `request_id`; required record fields `request_id`,
  `status`; accepted fields `note`, `request_id`, `status`; risk: Update Time Off Request Status
  through the BambooHR API.
- `delete_clock_entries`: DELETE `/api/v1/time_tracking/clock_entries` - kind `delete`; body type
  `none`; accepted fields `clockEntryIds`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes BambooHR data: Delete clock entries.
- `store_clock_entries`: POST `/api/v1/time_tracking/clock_entries` - kind `create`; body type
  `json`; accepted fields `entries`; risk: Store clock entries through the BambooHR API.
- `delete_timesheet_clock_entries_via_post`: POST `/api/v1/time_tracking/clock_entries/delete` -
  kind `create`; body type `json`; required record fields `clockEntryIds`; accepted fields
  `clockEntryIds`; risk: Delete Timesheet Clock Entries through the BambooHR API.
- `create_or_update_timesheet_clock_entries`: POST `/api/v1/time_tracking/clock_entries/store` -
  kind `create`; body type `json`; required record fields `entries`; accepted fields `entries`;
  risk: Create or Update Timesheet Clock Entries through the BambooHR API.
- `clock_in`: POST `/api/v1/time_tracking/clock_in/{{ record.employee_id }}` - kind `create`; body
  type `json`; path fields `employee_id`; required record fields `employee_id`; accepted fields
  `clockInLocation`, `employee_id`, `note`, `projectId`, `start`, `taskId`, `timezone`; risk: Clock
  in (employee id optional) through the BambooHR API.
- `clock_out`: POST `/api/v1/time_tracking/clock_out/{{ record.employee_id }}` - kind `create`; body
  type `json`; path fields `employee_id`; required record fields `employee_id`; accepted fields
  `clockOutLocation`, `employee_id`; risk: Clock out (employee id optional) through the BambooHR
  API.
- `store_daily_entries`: POST `/api/v1/time_tracking/daily_entries` - kind `create`; body type
  `json`; accepted fields `entries`; risk: Store daily entries through the BambooHR API.
- `clock_in_data`: POST `/api/v1/time_tracking/employee/{{ record.employee_id }}/clock_in/data` -
  kind `create`; body type `json`; path fields `employee_id`; required record fields `employee_id`;
  accepted fields `clockInLocation`, `clockOutLocation`, `employee_id`, `note`, `projectId`,
  `start`, `taskId`, `timezone`; risk: Edit information on the currently clocked in entry through
  the BambooHR API.
- `clock_out_employee_at_specific_time`: POST `/api/v1/time_tracking/employee/{{ record.employee_id
  }}/clock_out/datetime` - kind `create`; body type `json`; path fields `employee_id`; required
  record fields `employee_id`; accepted fields `datetime`, `employeeId`, `employee_id`, `timezone`;
  risk: Clock out an employee at a specific time through the BambooHR API.
- `create_timesheet_clock_in_entry`: POST `/api/v1/time_tracking/employees/{{ record.employee_id
  }}/clock_in` - kind `create`; body type `json`; path fields `employee_id`; required record fields
  `employee_id`; accepted fields `breakId`, `date`, `employee_id`, `note`, `offline`, `projectId`,
  `start`, `taskId`, `timezone`; risk: Create Timesheet Clock-In Entry through the BambooHR API.
- `create_timesheet_clock_out_entry`: POST `/api/v1/time_tracking/employees/{{ record.employee_id
  }}/clock_out` - kind `create`; body type `json`; path fields `employee_id`; required record fields
  `employee_id`; accepted fields `date`, `employee_id`, `end`, `timezone`; risk: Create Timesheet
  Clock-Out Entry through the BambooHR API.
- `delete_timesheet_hour_entries_via_post`: POST `/api/v1/time_tracking/hour_entries/delete` - kind
  `create`; body type `json`; required record fields `hourEntryIds`; accepted fields `hourEntryIds`;
  risk: Delete Timesheet Hour Entries through the BambooHR API.
- `create_or_update_timesheet_hour_entries`: POST `/api/v1/time_tracking/hour_entries/store` - kind
  `create`; body type `json`; required record fields `hours`; accepted fields `hours`; risk: Create
  or Update Timesheet Hour Entries through the BambooHR API.
- `create_time_tracking_project`: POST `/api/v1/time_tracking/projects` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `allowAllEmployees`, `billable`,
  `employeeIds`, `hasTasks`, `name`, `tasks`; risk: Create Time Tracking Project through the
  BambooHR API.
- `approve_employee_timesheets`: POST `/api/v1/time_tracking/timesheets/approve` - kind `create`;
  body type `json`; required record fields `lastChanged`, `timesheets`; accepted fields
  `lastChanged`, `timesheets`; risk: Approve employee timesheets through the BambooHR API.
- `clock_out_and_approve_employee_timesheets`: POST
  `/api/v1/time_tracking/timesheets/clock_out_and_approve` - kind `create`; body type `json`;
  accepted fields `clockOuts`; risk: Approve timesheets for employees that are currently clocked in
  through the BambooHR API.
- `create_time_tracking_hour_record`: POST `/api/v1/timetracking/add` - kind `create`; body type
  `json`; required record fields `dateHoursWorked`, `employeeId`, `hoursWorked`, `rateType`,
  `timeTrackingId`; accepted fields `dateHoursWorked`, `departmentId`, `divisionId`, `employeeId`,
  `hoursWorked`, `jobCode`, `jobData`, `jobTitleId`, `payCode`, `payRate`, `rateType`,
  `timeTrackingId`; risk: Create Hour Record through the BambooHR API.
- `update_time_tracking_record`: PUT `/api/v1/timetracking/adjust` - kind `update`; body type
  `json`; required record fields `timeTrackingId`, `hoursWorked`; accepted fields `holidayId`,
  `hoursWorked`, `projectId`, `shiftDifferentialId`, `taskId`, `timeTrackingId`; risk: Update Hour
  Record through the BambooHR API.
- `delete_time_tracking_hour_record`: DELETE `/api/v1/timetracking/delete/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  BambooHR data: Delete Hour Record.
- `create_or_update_time_tracking_hour_records`: POST `/api/v1/timetracking/record` - kind `create`;
  body type `json`; risk: Create or Update Hour Records through the BambooHR API.
- `create_training_category`: POST `/api/v1/training/category` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: Create Training Category through the
  BambooHR API.
- `delete_training_category`: DELETE `/api/v1/training/category/{{ record.training_category_id }}` -
  kind `delete`; body type `none`; path fields `training_category_id`; required record fields
  `training_category_id`; accepted fields `training_category_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Deletes BambooHR data: Delete Training
  Category.
- `update_training_category`: PUT `/api/v1/training/category/{{ record.training_category_id }}` -
  kind `update`; body type `json`; path fields `training_category_id`; required record fields
  `training_category_id`, `name`; accepted fields `name`, `training_category_id`; risk: Update
  Training Category through the BambooHR API.
- `create_employee_training_record`: POST `/api/v1/training/record/employee/{{ record.employee_id
  }}` - kind `create`; body type `json`; path fields `employee_id`; required record fields
  `employee_id`, `completed`, `type`; accepted fields `completed`, `cost`, `credits`, `employee_id`,
  `hours`, `instructor`, `notes`, `type`; risk: Create Employee Training Record through the BambooHR
  API.
- `delete_employee_training_record`: DELETE `/api/v1/training/record/{{
  record.employee_training_record_id }}` - kind `delete`; body type `none`; path fields
  `employee_training_record_id`; required record fields `employee_training_record_id`; accepted
  fields `employee_training_record_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes BambooHR data: Delete Employee Training Record.
- `update_employee_training_record`: PUT `/api/v1/training/record/{{
  record.employee_training_record_id }}` - kind `update`; body type `json`; path fields
  `employee_training_record_id`; required record fields `employee_training_record_id`, `completed`;
  accepted fields `completed`, `cost`, `credits`, `employee_training_record_id`, `hours`,
  `instructor`, `notes`; risk: Update Employee Training Record through the BambooHR API.
- `create_training_type`: POST `/api/v1/training/type` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `allowEmployeesToMarkComplete`, `category`, `description`,
  `dueFromHireDate`, `frequency`, `linkUrl`, `name`, `renewable`, `required`; risk: Create Training
  Type through the BambooHR API.
- `delete_training_type`: DELETE `/api/v1/training/type/{{ record.training_type_id }}` - kind
  `delete`; body type `none`; path fields `training_type_id`; required record fields
  `training_type_id`; accepted fields `training_type_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes BambooHR data: Delete Training Type.
- `update_training_type`: PUT `/api/v1/training/type/{{ record.training_type_id }}` - kind `update`;
  body type `json`; path fields `training_type_id`; required record fields `training_type_id`;
  accepted fields `allowEmployeesToMarkComplete`, `category`, `description`, `dueFromHireDate`,
  `frequency`, `linkUrl`, `name`, `renewable`, `required`, `training_type_id`; risk: Update Training
  Type through the BambooHR API.
- `create_webhook`: POST `/api/v1/webhooks` - kind `create`; body type `json`; required record
  fields `name`, `url`, `format`; accepted fields `events`, `format`, `includeCompanyDomain`,
  `monitorFields`, `name`, `postFields`, `url`; risk: Create Webhook through the BambooHR API.
- `delete_webhook`: DELETE `/api/v1/webhooks/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes BambooHR data: Delete Webhook.
- `update_webhook`: PUT `/api/v1/webhooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `name`, `url`, `format`; accepted fields `events`,
  `format`, `id`, `includeCompanyDomain`, `monitorFields`, `name`, `postFields`, `url`; risk: Update
  Webhook through the BambooHR API.
- `assign_time_off_policies_v1_1`: PUT `/api/v1_1/employees/{{ record.employee_id
  }}/time_off/policies` - kind `update`; body type `json`; path fields `employee_id`; required
  record fields `employee_id`; accepted fields `employee_id`; risk: Assign Time Off Policies v1.1
  through the BambooHR API.
- `create_table_row_v1_1`: POST `/api/v1_1/employees/{{ record.id }}/tables/{{ record.table }}` -
  kind `create`; body type `json`; path fields `id`, `table`; required record fields `id`, `table`;
  accepted fields `date`, `department`, `division`, `id`, `jobTitle`, `location`, `reportsTo`,
  `table`, `teams`; risk: Create Table Row v1.1 through the BambooHR API.
- `update_table_row_v1_1`: POST `/api/v1_1/employees/{{ record.id }}/tables/{{ record.table }}/{{
  record.row_id }}` - kind `update`; body type `json`; path fields `id`, `table`, `row_id`; required
  record fields `id`, `table`, `row_id`; accepted fields `date`, `department`, `division`, `id`,
  `jobTitle`, `location`, `reportsTo`, `row_id`, `table`, `teams`; risk: Update Table Row v1.1
  through the BambooHR API.
- `update_goal_v1_1`: PUT `/api/v1_1/performance/employees/{{ record.employee_id }}/goals/{{
  record.goal_id }}` - kind `update`; body type `json`; path fields `employee_id`, `goal_id`;
  required record fields `employee_id`, `goal_id`, `title`, `dueDate`, `sharedWithEmployeeIds`;
  accepted fields `alignsWithOptionId`, `completionDate`, `deletedMilestoneIds`, `description`,
  `dueDate`, `employee_id`, `goal_id`, `milestones`, `milestonesEnabled`, `percentComplete`,
  `sharedWithEmployeeIds`, `title`; risk: Update Goal (v1.1) through the BambooHR API.
- `update_company_benefit_properties`: POST `/api/v1_2/benefit/company_benefit/{{ record.id }}` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`, `properties`; risk: Update a company benefit through the BambooHR API.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 84 stream-backed endpoint group(s), 101 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=19, deprecated=3, out_of_scope=27, requires_elevated_scope=106.
