# Overview

Reads and writes the documented Keka HRMS REST API surface for Core HR, documents, leave,
attendance, payroll, PSA, PMS, hire, expense, assets, requisitions, skills, and BGV resources.

Readable streams: `employees`, `attendance`, `leave_types`, `leave_requests`, `clients`, `projects`,
`employee`, `employee_update_fields`, `groups`, `group_types`, `departments`, `locations`,
`job_titles`, `currencies`, `notice_periods`, `exit_reasons`, `document_types`,
`employee_documents`, `employee_document_attachment_download_urls`, `leave_balances`, `leave_plans`,
`capture_schemes`, `shift_policies`, `holiday_calendars`, `tracking_policies`,
`weekly_off_policies`, `salary_components`, `pay_groups`, `pay_cycles`, `pay_register`,
`pay_batches`, `batch_payments`, `pay_grades`, `pay_bands`, `employee_salaries`,
`employee_fnf_details`, `client`, `billing_roles`, `project_phases`, `project`,
`project_allocations`, `project_time_entries`, `project_tasks`, `project_task_time_entries`,
`project_task_assignees`, `timesheet_entries`, `pms_timeframes`, `goals`, `badges`, `praise`,
`review_groups`, `review_cycles`, `reviews`, `hire_jobs`, `job_application_fields`, `candidates`,
`candidate_interviews`, `candidate_scorecards`, `preboarding_candidates`, `expense_categories`,
`expense_claims`, `expense_policies`, `assets`, `asset_types`, `asset_categories`,
`asset_conditions`, `requisition_requests`, `employee_skills`, `bgv_requests`.

Write actions: `create_employee`, `update_employee_personal_details`, `update_employee_job_details`,
`create_exit_request`, `update_exit_request`, `create_leave_request`, `update_payment_status`,
`create_client`, `update_client`, `create_project_phase`, `create_project`,
`update_project_details`, `add_project_allocation`, `create_project_task`, `update_project_task`,
`update_goal_progress`, `create_praise`, `update_candidate`, `add_candidate_notes`,
`create_candidate`, `create_preboarding_candidate`, `update_preboarding_candidate`,
`update_asset_assignment`, `add_bgv_request_report`.

Service API documentation: https://apidocs.keka.com/.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Keka API key, sent as the api_key form field in the token
  exchange alongside client_id/client_secret. Marked x-secret for its credential-shaped nature even
  though it is a token-exchange input, not a Bearer credential itself.
- `attachment_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference attachment_id.
- `base_url` (required, string); format `uri`; Company-specific Keka API base URL (e.g.
  https://<company>.keka.com/api/v1). No default: Keka has no global base URL.
- `bgv_id` (optional, string); Optional Keka resource identifier used by path-scoped streams that
  reference bgv_id.
- `candidate_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference candidate_id.
- `client_id` (required, string); Keka OAuth2 client ID.
- `client_secret` (required, secret, string); Keka OAuth2 client secret. Used only in the token
  exchange body; never logged.
- `document_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference document_id.
- `document_type_id` (optional, string); Optional Keka resource identifier used by path-scoped
  streams that reference document_type_id.
- `employee_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference employee_id.
- `grant_type` (optional, string); default `kekaapi`; Keka's custom OAuth2 grant_type value for the
  token exchange (not the standard client_credentials grant).
- `job_id` (optional, string); Optional Keka resource identifier used by path-scoped streams that
  reference job_id.
- `mode` (optional, string).
- `pay_batch_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference pay_batch_id.
- `pay_cycle_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference pay_cycle_id.
- `pay_group_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference pay_group_id.
- `project_id` (optional, string); Optional Keka resource identifier used by path-scoped streams
  that reference project_id.
- `scope` (optional, string); default `kekaapi`; OAuth2 scope requested in the token exchange.
- `task_id` (optional, string); Optional Keka resource identifier used by path-scoped streams that
  reference task_id.
- `token_url` (optional, string); default `https://login.keka.com/connect/token`; format `uri`; Keka
  OAuth2 token endpoint override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`, `client_secret`.

Default configuration values: `grant_type=kekaapi`, `scope=kekaapi`,
`token_url=https://login.keka.com/connect/token`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/hris/employees` with query `pageNumber`=`1`; `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `pageNumber`; size parameter `pageSize`;
starts at 1; page size 100.

Pagination by stream: none: `employee`, `employee_update_fields`, `exit_reasons`,
`employee_document_attachment_download_urls`, `client`, `project`; page_number: `employees`,
`attendance`, `leave_types`, `leave_requests`, `clients`, `projects`, `groups`, `group_types`,
`departments`, `locations`, `job_titles`, `currencies`, `notice_periods`, `document_types`,
`employee_documents`, `leave_balances`, `leave_plans`, `capture_schemes`, `shift_policies`,
`holiday_calendars`, `tracking_policies`, `weekly_off_policies`, `salary_components`, `pay_groups`,
`pay_cycles`, `pay_register`, `pay_batches`, `batch_payments`, `pay_grades`, `pay_bands`,
`employee_salaries`, `employee_fnf_details`, `billing_roles`, `project_phases`,
`project_allocations`, `project_time_entries`, `project_tasks`, `project_task_time_entries`,
`project_task_assignees`, `timesheet_entries`, `pms_timeframes`, `goals`, `badges`, `praise`,
`review_groups`, `review_cycles`, `reviews`, `hire_jobs`, `job_application_fields`, `candidates`,
`candidate_interviews`, `candidate_scorecards`, `preboarding_candidates`, `expense_categories`,
`expense_claims`, `expense_policies`, `assets`, `asset_types`, `asset_categories`,
`asset_conditions`, and 3 more.

- `employees`: GET `/hris/employees` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `attendance`: GET `/time/attendance` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `leave_types`: GET `/time/leavetypes` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `leave_requests`: GET `/time/leaverequests` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `clients`: GET `/psa/clients` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `projects`: GET `/psa/projects` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100.
- `employee`: GET `/hris/employees/{{ config.employee_id }}` - records path `data`; emits
  passthrough records.
- `employee_update_fields`: GET `/hris/employees/updatefields` - records path `data`; emits
  passthrough records.
- `groups`: GET `/hris/groups` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `group_types`: GET `/hris/grouptypes` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `departments`: GET `/hris/departments` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `locations`: GET `/hris/locations` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `job_titles`: GET `/hris/jobtitles` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `currencies`: GET `/hris/currencies` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `notice_periods`: GET `/hris/noticeperiods` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `exit_reasons`: GET `/hris/exitreasons` - records path `.`; emits passthrough records.
- `document_types`: GET `/hris/documents/types` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `employee_documents`: GET `/hris/employees/{{ config.employee_id }}/documenttypes/{{
  config.document_type_id }}/documents` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `employee_document_attachment_download_urls`: GET `/hris/employees/{{ config.employee_id
  }}/documents/{{ config.document_id }}/attachment/{{ config.attachment_id }}` - records path
  `data`; emits passthrough records.
- `leave_balances`: GET `/time/leavebalance` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `leave_plans`: GET `/time/leaveplans` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `capture_schemes`: GET `/time/capturescheme` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `shift_policies`: GET `/time/shiftpolicies` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `holiday_calendars`: GET `/time/holidayscalendar` - records path `data`; page-number pagination;
  page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits
  passthrough records.
- `tracking_policies`: GET `/time/penalisationpolicies` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.
- `weekly_off_policies`: GET `/time/weeklyoffpolicies` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.
- `salary_components`: GET `/payroll/salarycomponents` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.
- `pay_groups`: GET `/payroll/paygroups` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `pay_cycles`: GET `/payroll/paygroups/{{ config.pay_group_id }}/paycycles` - records path `data`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100; emits passthrough records.
- `pay_register`: GET `/payroll/paygroups/{{ config.pay_group_id }}/paycycles/{{ config.pay_cycle_id
  }}/payregister` - records path `data`; page-number pagination; page parameter `pageNumber`; size
  parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `pay_batches`: GET `/payroll/paygroups/{{ config.pay_group_id }}/paycycles/{{ config.pay_cycle_id
  }}/paybatches` - records path `data`; page-number pagination; page parameter `pageNumber`; size
  parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `batch_payments`: GET `/payroll/paygroups/{{ config.pay_group_id }}/paycycles/{{
  config.pay_cycle_id }}/paybatches/{{ config.pay_batch_id }}/payments` - records path `data`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100; emits passthrough records.
- `pay_grades`: GET `/time/payroll/paygrades` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `pay_bands`: GET `/payroll/payband` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `employee_salaries`: GET `/payroll/salaries` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `employee_fnf_details`: GET `/payroll/employees/fnf` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.
- `client`: GET `/psa/clients/{{ config.client_id }}` - records path `data`; emits passthrough
  records.
- `billing_roles`: GET `/psa/clients/{{ config.client_id }}/billingroles` - records path `data`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100; emits passthrough records.
- `project_phases`: GET `/psa/projects/{{ config.project_id }}/phases` - records path `data`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100; emits passthrough records.
- `project`: GET `/psa/projects/{{ config.project_id }}` - records path `data`; emits passthrough
  records.
- `project_allocations`: GET `/psa/projects/{{ config.project_id }}/allocations` - records path
  `data`; page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at
  1; page size 100; emits passthrough records.
- `project_time_entries`: GET `/psa/projects/{{ config.project_id }}/timeentries` - records path
  `data`; page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at
  1; page size 100; emits passthrough records.
- `project_tasks`: GET `/psa/projects/{{ config.project_id }}/tasks` - records path `data`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100; emits passthrough records.
- `project_task_time_entries`: GET `/psa/projects/{{ config.project_id }}/tasks/{{ config.task_id
  }}/timeentries` - records path `data`; page-number pagination; page parameter `pageNumber`; size
  parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `project_task_assignees`: GET `/psa/projects/{{ config.project_id }}/tasks/{{ config.task_id }}` -
  records path `data`; page-number pagination; page parameter `pageNumber`; size parameter
  `pageSize`; starts at 1; page size 100; emits passthrough records.
- `timesheet_entries`: GET `/psa/timeentries` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `pms_timeframes`: GET `/pms/timeframes` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `goals`: GET `/pms/goals` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `badges`: GET `/pms/badges` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `praise`: GET `/pms/praise` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `review_groups`: GET `/pms/reviewgroups` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `review_cycles`: GET `/pms/reviewcycles` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `reviews`: GET `/pms/reviews` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `hire_jobs`: GET `/hire/jobs` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `job_application_fields`: GET `/hire/jobs/{{ config.job_id }}/applicationfields` - records path
  `data`; page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at
  1; page size 100; emits passthrough records.
- `candidates`: GET `/hire/jobs/{{ config.job_id }}/candidates` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.
- `candidate_interviews`: GET `/hire/jobs/{{ config.job_id }}/candidate/{{ config.candidate_id
  }}/interviews` - records path `data`; page-number pagination; page parameter `pageNumber`; size
  parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `candidate_scorecards`: GET `/hire/jobs/{{ config.job_id }}/candidate/{{ config.candidate_id
  }}/scorecards` - records path `data`; page-number pagination; page parameter `pageNumber`; size
  parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `preboarding_candidates`: GET `/hire/preboarding/candiates` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.
- `expense_categories`: GET `/expense/categories` - records path `data`; page-number pagination;
  page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits
  passthrough records.
- `expense_claims`: GET `/expense/claims` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `expense_policies`: GET `/expensepolicies` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `assets`: GET `/assets` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `asset_types`: GET `/assets/types` - records path `data`; page-number pagination; page parameter
  `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough records.
- `asset_categories`: GET `/assets/categories` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `asset_conditions`: GET `/assets/conditions` - records path `data`; page-number pagination; page
  parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits passthrough
  records.
- `requisition_requests`: GET `/requisition/requests` - records path `data`; page-number pagination;
  page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100; emits
  passthrough records.
- `employee_skills`: GET `/hris/employees/{{ config.employee_id }}/skills` - records path `data`;
  page-number pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page
  size 100; emits passthrough records.
- `bgv_requests`: GET `/hris/bgv/{{ config.bgv_id }}/requests` - records path `data`; page-number
  pagination; page parameter `pageNumber`; size parameter `pageSize`; starts at 1; page size 100;
  emits passthrough records.

## Write actions & risks

Overall write risk: live Keka API mutations can create or update employee, leave, payroll payment,
client, project, performance, hiring, asset, skill, and BGV records.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_employee`: POST `/hris/employees` - kind `create`; body type `json`; risk: Create Employee
  through the Keka API.
- `update_employee_personal_details`: PUT `/hris/employees/{{ record.employee_id }}/personaldetails`
  - kind `update`; body type `json`; path fields `employee_id`; required record fields
  `employee_id`; accepted fields `employee_id`; risk: Update Employee Personal Details through the
  Keka API.
- `update_employee_job_details`: PUT `/hris/employees/{{ record.employee_id }}/jobdetails` - kind
  `update`; body type `json`; path fields `employee_id`; required record fields `employee_id`;
  accepted fields `employee_id`; risk: Update Employee Job Details through the Keka API.
- `create_exit_request`: POST `/hris/employees/{{ record.employee_id }}/exitrequest` - kind
  `create`; body type `json`; path fields `employee_id`; required record fields `employee_id`;
  accepted fields `employee_id`; risk: Create Exit Request through the Keka API.
- `update_exit_request`: PUT `/hris/employees/{{ record.employee_id }}/exitrequest` - kind `update`;
  body type `json`; path fields `employee_id`; required record fields `employee_id`; accepted fields
  `employee_id`; risk: Update Exit Request through the Keka API.
- `create_leave_request`: POST `/time/leaverequests` - kind `create`; body type `json`; risk: Create
  Leave Request through the Keka API.
- `update_payment_status`: PUT `/payroll/paygroups/{{ record.pay_group_id }}/paycycles/{{
  record.pay_cycle_id }}/paybatches/{{ record.pay_batch_id }}/payments` - kind `update`; body type
  `json`; path fields `pay_group_id`, `pay_cycle_id`, `pay_batch_id`; required record fields
  `pay_group_id`, `pay_cycle_id`, `pay_batch_id`; accepted fields `pay_batch_id`, `pay_cycle_id`,
  `pay_group_id`; risk: Update Payment Status through the Keka API.
- `create_client`: POST `/psa/clients` - kind `create`; body type `json`; risk: Create Client
  through the Keka API.
- `update_client`: PUT `/psa/clients/{{ record.client_id }}` - kind `update`; body type `json`; path
  fields `client_id`; required record fields `client_id`; accepted fields `client_id`; risk: Update
  Client through the Keka API.
- `create_project_phase`: POST `/psa/projects/{{ record.project_id }}/phases` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Create Project Phase through the Keka API.
- `create_project`: POST `/psa/projects` - kind `create`; body type `json`; risk: Create Project
  through the Keka API.
- `update_project_details`: PUT `/psa/projects/{{ record.project_id }}` - kind `update`; body type
  `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Update Project Details through the Keka API.
- `add_project_allocation`: POST `/psa/projects/{{ record.project_id }}/allocations` - kind
  `create`; body type `json`; path fields `project_id`; required record fields `project_id`;
  accepted fields `project_id`; risk: Add Project Allocation through the Keka API.
- `create_project_task`: POST `/psa/projects/{{ record.project_id }}/tasks` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Create Project Task through the Keka API.
- `update_project_task`: PUT `/psa/projects/{{ record.project_id }}/tasks/{{ record.task_id }}` -
  kind `update`; body type `json`; path fields `project_id`, `task_id`; required record fields
  `project_id`, `task_id`; accepted fields `project_id`, `task_id`; risk: Update Project Task
  through the Keka API.
- `update_goal_progress`: PUT `/pms/goals/{{ record.goal_id }}/progress` - kind `update`; body type
  `json`; path fields `goal_id`; required record fields `goal_id`; accepted fields `goal_id`; risk:
  Update Goal Progress through the Keka API.
- `create_praise`: POST `/pms/praise` - kind `create`; body type `json`; risk: Create Praise through
  the Keka API.
- `update_candidate`: PUT `/hire/jobs/{{ record.job_id }}/candidate/{{ record.candidate_id }}` -
  kind `update`; body type `json`; path fields `job_id`, `candidate_id`; required record fields
  `job_id`, `candidate_id`; accepted fields `candidate_id`, `job_id`; risk: Update Candidate through
  the Keka API.
- `add_candidate_notes`: POST `/hire/jobs/{{ record.job_id }}/candidate/{{ record.candidate_id
  }}/notes` - kind `create`; body type `json`; path fields `job_id`, `candidate_id`; required record
  fields `job_id`, `candidate_id`; accepted fields `candidate_id`, `job_id`; risk: Add Candidate
  Notes through the Keka API.
- `create_candidate`: POST `/v1/hire/jobs/{{ record.job_id }}/candidate` - kind `create`; body type
  `json`; path fields `job_id`; required record fields `job_id`; accepted fields `job_id`; risk:
  Create Candidate through the Keka API.
- `create_preboarding_candidate`: POST `/hire/preboarding/candidates` - kind `create`; body type
  `json`; risk: Create Preboarding Candidate through the Keka API.
- `update_preboarding_candidate`: PUT `/hire/preboarding/candidates/{{
  record.preboarding_candidate_id }}` - kind `update`; body type `json`; path fields
  `preboarding_candidate_id`; required record fields `preboarding_candidate_id`; accepted fields
  `preboarding_candidate_id`; risk: Update Preboarding Candidate through the Keka API.
- `update_asset_assignment`: PUT `/assets/{{ record.asset_id }}/allocation` - kind `update`; body
  type `json`; path fields `asset_id`; required record fields `asset_id`; accepted fields
  `asset_id`; risk: Update Asset Assignment through the Keka API.
- `add_bgv_request_report`: PUT `/hris/bgv/{{ record.bgv_id }}/requests/{{ record.request_id }}` -
  kind `update`; body type `json`; path fields `bgv_id`, `request_id`; required record fields
  `bgv_id`, `request_id`; accepted fields `bgv_id`, `request_id`; risk: Add Bgv Request Report
  through the Keka API.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 69 stream-backed endpoint group(s), 24 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
