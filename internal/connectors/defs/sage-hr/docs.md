# Overview

Reads Sage HR employees, teams, time off, recruitment, and onboarding/offboarding data, and writes
employee/leave/task lifecycle mutations, through the Sage HR API.

Readable streams: `employees`, `teams`, `timeoff_requests`, `terminated_employees`, `positions`,
`termination_reasons`, `leave_policies`, `out_of_office_today`, `individual_allowances`,
`recruitment_positions`, `recruitment_applicants`, `onboarding_categories`,
`offboarding_categories`, `document_categories`.

Write actions: `create_employee`, `update_employee`, `update_employee_custom_field`,
`terminate_employee`, `create_timeoff_request`, `create_kit_day`, `update_kit_day_status`,
`update_leave_policy_kit_days`, `create_onboarding_task`, `create_offboarding_task`.

Service API documentation: https://developers.sage.hr/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Sage HR API key, sent as the X-Auth-Token request header.
  Never logged.
- `base_url` (optional, string); default `https://api.sage.hr/v1`; format `uri`; Sage HR API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.sage.hr/v1`.

Authentication behavior:

- API key authentication in `X-Auth-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/employees`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `employees`, `teams`, `timeoff_requests`, `leave_policies`,
`out_of_office_today`, `onboarding_categories`, `offboarding_categories`, `document_categories`;
page_number: `terminated_employees`, `positions`, `termination_reasons`, `individual_allowances`,
`recruitment_positions`, `recruitment_applicants`.

- `employees`: GET `/employees` - records path `data`; emits passthrough records.
- `teams`: GET `/teams` - records path `data`; emits passthrough records.
- `timeoff_requests`: GET `/leave-management/requests` - records path `data`; emits passthrough
  records.
- `terminated_employees`: GET `/terminated-employees` - records path `data`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 2; emits passthrough
  records.
- `positions`: GET `/positions` - records path `data`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 2; emits passthrough records.
- `termination_reasons`: GET `/termination-reasons` - records path `data`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 2; emits passthrough
  records.
- `leave_policies`: GET `/leave-management/policies` - records path `data`; emits passthrough
  records.
- `out_of_office_today`: GET `/leave-management/out-of-office-today` - records path `data`; emits
  passthrough records.
- `individual_allowances`: GET `/leave-management/reports/individual-allowances` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 2; emits passthrough records.
- `recruitment_positions`: GET `/recruitment/positions` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 2; emits
  passthrough records.
- `recruitment_applicants`: GET `/recruitment/positions/{{ fanout.id }}/applicants` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 2; fan-out; ids from request `/recruitment/positions`; id-list records path `data`; id
  field `id`; id inserted into the request path; stamps `position_id`; emits passthrough records.
- `onboarding_categories`: GET `/onboarding/categories` - records path `data`; emits passthrough
  records.
- `offboarding_categories`: GET `/offboarding/categories` - records path `data`; emits passthrough
  records.
- `document_categories`: GET `/documents/categories` - records path `data`; emits passthrough
  records.

## Write actions & risks

Overall write risk: external Sage HR mutations: employee create/update/termination, custom-field
update, time off/KIT-day requests and approvals, leave policy KIT-day configuration,
onboarding/offboarding task creation.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_employee`: POST `/employees` - kind `create`; body type `form`; required record fields
  `email`, `first_name`, `last_name`; accepted fields `email`, `first_name`, `last_name`,
  `send_email`, `work_start_date`; risk: creates a new employee record and may email the new hire
  (send_email); external mutation, approval required.
- `update_employee`: PUT `/employees/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `approver_ids`, `employee_number`,
  `first_name`, `id`, `last_name`, `leader_id`, `location_id`, `position_id`,
  `selected_leave_types`, `team_id`, `work_start_date`; risk: external mutation updating an employee
  record (org placement, leave types, reporting line); approval required.
- `update_employee_custom_field`: PUT `/employees/{{ record.employee_id }}/custom-fields/{{
  record.custom_field_id }}` - kind `update`; body type `form`; path fields `employee_id`,
  `custom_field_id`; required record fields `employee_id`, `custom_field_id`, `value`; accepted
  fields `custom_field_id`, `employee_id`, `value`; risk: external mutation of an employee custom
  field; approval required.
- `terminate_employee`: POST `/employees/{{ record.employee_id }}/terminations` - kind `create`;
  body type `form`; path fields `employee_id`; required record fields `employee_id`, `date`,
  `termination_reason_id`; accepted fields `comments`, `date`, `employee_id`,
  `termination_reason_id`; risk: destructive/irreversible: terminates an employee's record in Sage
  HR; external mutation, approval required.
- `create_timeoff_request`: POST `/leave-management/requests` - kind `create`; body type `form`;
  required record fields `employee_id`, `time_off_policy_id`, `type`, `part_of_day`; accepted fields
  `date`, `date_from`, `date_to`, `details`, `employee_id`, `hours`, `part_of_day`, `time_from`,
  `time_off_policy_id`, `time_to`, `type`; risk: creates a new time off request against an
  employee's leave balance; external mutation, approval required.
- `create_kit_day`: POST `/leave-management/kit-days` - kind `create`; body type `form`; required
  record fields `employee_id`, `policy_id`; accepted fields `date`, `date_from`, `date_to`,
  `employee_id`, `policy_id`; risk: creates a Keeping-In-Touch day entry against an employee's leave
  policy; external mutation, approval required.
- `update_kit_day_status`: PATCH `/leave-management/kit-days/{{ record.id }}` - kind `update`; body
  type `form`; path fields `id`; required record fields `id`, `status`; accepted fields `id`,
  `status`; risk: approves, declines, or cancels a KIT day request; external mutation, approval
  required.
- `update_leave_policy_kit_days`: PATCH `/leave-management/policies/{{ record.id }}` - kind
  `update`; body type `form`; path fields `id`; required record fields `id`, `kit_days_enabled`,
  `kit_days_quantity`; accepted fields `id`, `kit_days_enabled`, `kit_days_quantity`; risk: changes
  a company-wide leave policy's KIT-day configuration; external mutation, approval required.
- `create_onboarding_task`: POST `/onboarding/tasks` - kind `create`; body type `form`; required
  record fields `title`, `boarding_task_template_category_id`, `due_in`; accepted fields
  `add_after`, `assignee_id`, `boarding_task_template_category_id`, `default_assignee_type`,
  `description`, `due_in`, `require_attachment`, `title`; risk: creates a new onboarding task
  template; external mutation, approval required.
- `create_offboarding_task`: POST `/offboarding/tasks` - kind `create`; body type `form`; required
  record fields `title`, `boarding_task_template_category_id`, `due_in`; accepted fields
  `assignee_id`, `boarding_task_template_category_id`, `default_assignee_type`, `description`,
  `due_in`, `require_attachment`, `title`; risk: creates a new offboarding task template; external
  mutation, approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 14 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, duplicate_of=4, non_data_endpoint=4, out_of_scope=16, requires_elevated_scope=2.
