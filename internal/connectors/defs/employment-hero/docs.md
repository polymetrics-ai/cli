# Overview

Reads Employment Hero organisations, employees, HR reference data, forms, goals, rosters, employee
subresources, and exposes documented JSON mutations through the Employment Hero REST API.

Readable streams: `organisations`, `organisation`, `employees`, `employee`, `teams`,
`team_employees`, `leave_requests`, `leave_request`, `certifications`, `certification`,
`cost_centres`, `custom_fields`, `employing_entities`, `forms`, `form`, `form_responses`,
`form_response`, `form_assignments`, `member_form_responses`, `form_categories`, `form_templates`,
`form_template`, `goals`, `goal`, `goal_comments`, `goal_key_results`, `goal_key_result`,
`kiosk_members`, `leave_categories`, `pay_categories`, `policies`, `rostered_shifts`,
`rostered_shift`, `roles`, `unavailabilities`, `unavailability`, `work_locations`, `work_sites`,
`work_types`, `bank_accounts_v1`, `bank_accounts_v2`, `contractor_job_histories`, `documents`,
`emergency_contacts`, `employee_certification_details`, `employee_custom_fields`,
`employment_histories`, `leave_balances`, `pay_details`, `payslips`, `payslip`, `timesheet_entries`,
`superannuation_detail_v1`, `superannuation_detail_v2`, `tax_declaration_v1`, `tax_declaration_v2`,
`work_eligibility`.

Write actions: `create_certification`, `update_certification`, `archive_certification`,
`delete_certification`, `create_department`, `update_department`, `quick_add_employee`,
`quick_add_contractor`, `onboard_employee_async`, `update_employee_personal_details`,
`update_employee_employment_details`, `update_employee_contractor_details`, `delete_employee`,
`update_employee_certification`, `delete_form`, `create_form_category`, `update_form_category`,
`delete_form_category`, `create_form_template`, `update_form_template`, `delete_form_template`,
`update_goal_archive_status`, `update_goal_health_status`, `bulk_grant_kiosk_access`,
`bulk_revoke_kiosk_access`, `update_leave_balance`, `create_leave_request`, `create_position`,
`update_position`, `bulk_create_rostered_shifts`, `create_timesheet_entries`, `create_work_site`,
`update_work_site`.

Service API documentation: https://developer.employmenthero.com/api-references.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Employment Hero access token, sent as Authorization: Bearer
  <api_key>. Never logged.
- `base_url` (optional, string); default `https://api.employmenthero.com/api`; format `uri`;
  Employment Hero API root. Keep /api as the root; stream paths include /v1 or /v2.
- `certification_id` (optional, string).
- `employee_id` (optional, string); Employee UUID used for single employee-scoped streams and write
  actions.
- `employee_ids` (optional, string); Comma-separated employee UUIDs used by single-object employee
  detail streams that cannot fan out through a separately paginated ID request.
- `form_id` (optional, string).
- `goal_id` (optional, string).
- `key_result_id` (optional, string).
- `leave_request_id` (optional, string).
- `member_ids` (optional, string); Comma-separated member UUIDs used by member_form_responses
  fan-out.
- `organization_id` (optional, string); Employment Hero organisation UUID used for
  organisation-scoped endpoints. The official docs call this organisation_id.
- `payslip_id` (optional, string).
- `response_id` (optional, string).
- `rostered_shift_id` (optional, string).
- `template_id` (optional, string).
- `unavailability_id` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.employmenthero.com/api`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/organisations` with query `items_per_page`=`1`; `page_index`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page_index`; size parameter
`items_per_page`; starts at 1; page size 100.

Pagination by stream: none: `organisation`, `employee`, `leave_request`, `certification`, `form`,
`form_response`, `form_template`, `goal`, `goal_key_result`, `rostered_shift`, `unavailability`,
`payslip`, `superannuation_detail_v1`, `superannuation_detail_v2`, `tax_declaration_v1`,
`tax_declaration_v2`, `work_eligibility`; page_number: `organisations`, `employees`, `teams`,
`team_employees`, `leave_requests`, `certifications`, `cost_centres`, `custom_fields`,
`employing_entities`, `forms`, `form_responses`, `form_assignments`, `member_form_responses`,
`form_categories`, `form_templates`, `goals`, `goal_comments`, `goal_key_results`, `kiosk_members`,
`leave_categories`, `pay_categories`, `policies`, `rostered_shifts`, `roles`, `unavailabilities`,
`work_locations`, `work_sites`, `work_types`, `bank_accounts_v1`, `bank_accounts_v2`,
`contractor_job_histories`, `documents`, `emergency_contacts`, `employee_certification_details`,
`employee_custom_fields`, `employment_histories`, `leave_balances`, `pay_details`, `payslips`,
`timesheet_entries`.

- `organisations`: GET `/v1/organisations` - records path `data.items`; page-number pagination; page
  parameter `page_index`; size parameter `items_per_page`; starts at 1; page size 100.
- `organisation`: GET `/v1/organisations/{{ config.organization_id }}` - single-object response;
  records path `data`.
- `employees`: GET `/v1/organisations/{{ config.organization_id }}/employees` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `employee`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ config.employee_id
  }}` - single-object response; records path `data`.
- `teams`: GET `/v1/organisations/{{ config.organization_id }}/teams` - records path `data.items`;
  page-number pagination; page parameter `page_index`; size parameter `items_per_page`; starts at 1;
  page size 100.
- `team_employees`: GET `/v1/organisations/{{ config.organization_id }}/teams/{{ fanout.id
  }}/employees` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/teams`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `team_id`.
- `leave_requests`: GET `/v1/organisations/{{ config.organization_id }}/leave_requests` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `leave_request`: GET `/v1/organisations/{{ config.organization_id }}/leave_requests/{{
  config.leave_request_id }}` - single-object response; records path `data`.
- `certifications`: GET `/v1/organisations/{{ config.organization_id }}/certifications` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `certification`: GET `/v1/organisations/{{ config.organization_id }}/certifications/{{
  config.certification_id }}` - single-object response; records path `data`.
- `cost_centres`: GET `/v1/organisations/{{ config.organization_id }}/cost_centres` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `custom_fields`: GET `/v1/organisations/{{ config.organization_id }}/custom_fields` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `employing_entities`: GET `/v1/organisations/{{ config.organization_id }}/employing_entities` -
  records path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `forms`: GET `/v1/organisations/{{ config.organization_id }}/forms` - records path `data.items`;
  page-number pagination; page parameter `page_index`; size parameter `items_per_page`; starts at 1;
  page size 100.
- `form`: GET `/v1/organisations/{{ config.organization_id }}/forms/{{ config.form_id }}` -
  single-object response; records path `data`.
- `form_responses`: GET `/v1/organisations/{{ config.organization_id }}/forms/{{ fanout.id
  }}/responses` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/forms`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `form_id`.
- `form_response`: GET `/v1/organisations/{{ config.organization_id }}/forms/{{ config.form_id
  }}/responses/{{ config.response_id }}` - single-object response; records path `data`.
- `form_assignments`: GET `/v1/organisations/{{ config.organization_id }}/forms/{{ fanout.id
  }}/assignments` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/forms`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `form_id`.
- `member_form_responses`: GET `/v1/organisations/{{ config.organization_id }}/members/{{ fanout.id
  }}/form_responses` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  config field `member_ids`; id inserted into the request path; stamps `member_id`.
- `form_categories`: GET `/v1/organisations/{{ config.organization_id }}/form_categories` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `form_templates`: GET `/v1/organisations/{{ config.organization_id }}/form_templates` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `form_template`: GET `/v1/organisations/{{ config.organization_id }}/form_templates/{{
  config.template_id }}` - single-object response; records path `data`.
- `goals`: GET `/v1/organisations/{{ config.organization_id }}/goals` - records path `data.items`;
  page-number pagination; page parameter `page_index`; size parameter `items_per_page`; starts at 1;
  page size 100.
- `goal`: GET `/v1/organisations/{{ config.organization_id }}/goals/{{ config.goal_id }}` -
  single-object response; records path `data`.
- `goal_comments`: GET `/v1/organisations/{{ config.organization_id }}/goals/{{ fanout.id
  }}/comments` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/goals`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `goal_id`.
- `goal_key_results`: GET `/v1/organisations/{{ config.organization_id }}/goals/{{ fanout.id
  }}/key_results` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/goals`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `goal_id`.
- `goal_key_result`: GET `/v1/organisations/{{ config.organization_id }}/goals/{{ config.goal_id
  }}/key_results/{{ config.key_result_id }}` - single-object response; records path `data`.
- `kiosk_members`: GET `/v1/organisations/{{ config.organization_id }}/kiosk_members` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `leave_categories`: GET `/v1/organisations/{{ config.organization_id }}/leave_categories` -
  records path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `pay_categories`: GET `/v1/organisations/{{ config.organization_id }}/pay_categories` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `policies`: GET `/v1/organisations/{{ config.organization_id }}/policies` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `rostered_shifts`: GET `/v1/organisations/{{ config.organization_id }}/rostered_shifts` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `rostered_shift`: GET `/v1/organisations/{{ config.organization_id }}/rostered_shifts/{{
  config.rostered_shift_id }}` - single-object response; records path `data`.
- `roles`: GET `/v1/organisations/{{ config.organization_id }}/roles` - records path `data.items`;
  page-number pagination; page parameter `page_index`; size parameter `items_per_page`; starts at 1;
  page size 100.
- `unavailabilities`: GET `/v1/organisations/{{ config.organization_id }}/unavailabilities` -
  records path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `unavailability`: GET `/v1/organisations/{{ config.organization_id }}/unavailabilities/{{
  config.unavailability_id }}` - single-object response; records path `data`.
- `work_locations`: GET `/v1/organisations/{{ config.organization_id }}/work_locations` - records
  path `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `work_sites`: GET `/v1/organisations/{{ config.organization_id }}/work_sites` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `work_types`: GET `/v1/organisations/{{ config.organization_id }}/work_types` - records path
  `data.items`; page-number pagination; page parameter `page_index`; size parameter
  `items_per_page`; starts at 1; page size 100.
- `bank_accounts_v1`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/bank_accounts` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `bank_accounts_v2`: GET `/v2/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/bank_accounts` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `contractor_job_histories`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{
  fanout.id }}/job_histories` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `documents`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/documents` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `employee_id`.
- `emergency_contacts`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/emergency_contacts` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `employee_certification_details`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{
  fanout.id }}/certifications` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `employee_custom_fields`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{
  fanout.id }}/custom_fields` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `employment_histories`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/employment_histories` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `leave_balances`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/leave_balances` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `pay_details`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/pay_details` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `employee_id`.
- `payslips`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/payslips` - records path `data.items`; page-number pagination; page parameter `page_index`;
  size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from request
  `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path `data.items`; id
  field `id`; id inserted into the request path; stamps `employee_id`.
- `payslip`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ config.employee_id
  }}/payslips/{{ config.payslip_id }}` - single-object response; records path `data`.
- `timesheet_entries`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/timesheet_entries` - records path `data.items`; page-number pagination; page parameter
  `page_index`; size parameter `items_per_page`; starts at 1; page size 100; fan-out; ids from
  request `/v1/organisations/{{ config.organization_id }}/employees`; id-list records path
  `data.items`; id field `id`; id inserted into the request path; stamps `employee_id`.
- `superannuation_detail_v1`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{
  fanout.id }}/superannuation_detail` - single-object response; records path `data`; fan-out; ids
  from config field `employee_ids`; id inserted into the request path; stamps `employee_id`.
- `superannuation_detail_v2`: GET `/v2/organisations/{{ config.organization_id }}/employees/{{
  fanout.id }}/superannuation_detail` - single-object response; records path `data`; fan-out; ids
  from config field `employee_ids`; id inserted into the request path; stamps `employee_id`.
- `tax_declaration_v1`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/tax_declaration` - single-object response; records path `data`; fan-out; ids from config field
  `employee_ids`; id inserted into the request path; stamps `employee_id`.
- `tax_declaration_v2`: GET `/v2/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/tax_declaration` - single-object response; records path `data`; fan-out; ids from config field
  `employee_ids`; id inserted into the request path; stamps `employee_id`.
- `work_eligibility`: GET `/v1/organisations/{{ config.organization_id }}/employees/{{ fanout.id
  }}/work_eligibility` - single-object response; records path `data`; fan-out; ids from config field
  `employee_ids`; id inserted into the request path; stamps `employee_id`.

## Write actions & risks

Overall write risk: creates, updates, archives, or deletes Employment Hero HR objects such as
employees, certifications, form assets, leave requests, positions, rostered shifts, timesheets,
kiosk access, and work sites.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_certification`: POST `/v1/organisations/{{ config.organization_id }}/certifications` -
  kind `create`; body type `json`; required record fields `name`; accepted fields `name`, `type`;
  risk: external Employment Hero mutation; approval required before execution.
- `update_certification`: PATCH `/v1/organisations/{{ config.organization_id }}/certifications/{{
  record.certification_id }}` - kind `update`; body type `json`; path fields `certification_id`;
  required record fields `certification_id`; accepted fields `certification_id`, `name`; risk:
  external Employment Hero mutation; approval required before execution.
- `archive_certification`: PATCH `/v1/organisations/{{ config.organization_id }}/certifications/{{
  record.certification_id }}/archive_status` - kind `update`; body type `json`; path fields
  `certification_id`; body fields `status`; required record fields `certification_id`, `status`;
  accepted fields `certification_id`, `status`; risk: archives or restores an Employment Hero
  certification configuration.
- `delete_certification`: DELETE `/v1/organisations/{{ config.organization_id }}/certifications/{{
  record.certification_id }}` - kind `delete`; body type `none`; path fields `certification_id`;
  required record fields `certification_id`; accepted fields `certification_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes an Employment Hero
  certification configuration.
- `create_department`: POST `/v1/organisations/{{ config.organization_id }}/departments` - kind
  `create`; body type `json`; required record fields `name`; accepted fields `member_ids`, `name`,
  `work_site_id`; risk: external Employment Hero mutation; approval required before execution.
- `update_department`: PATCH `/v1/organisations/{{ config.organization_id }}/departments/{{
  record.department_id }}` - kind `update`; body type `json`; path fields `department_id`; required
  record fields `department_id`; accepted fields `department_id`, `name`; risk: external Employment
  Hero mutation; approval required before execution.
- `quick_add_employee`: POST `/v1/organisations/{{ config.organization_id
  }}/employees/quick_add_employee` - kind `create`; body type `json`; required record fields
  `first_name`, `last_name`, `email`; accepted fields `email`, `employing_entity_id`, `first_name`,
  `last_name`, `work_location`; risk: external Employment Hero mutation; approval required before
  execution.
- `quick_add_contractor`: POST `/v1/organisations/{{ config.organization_id
  }}/employees/quick_add_contractor` - kind `create`; body type `json`; required record fields
  `first_name`, `last_name`, `email`; accepted fields `business_detail`, `email`, `first_name`,
  `last_name`, `trading_name`; risk: external Employment Hero mutation; approval required before
  execution.
- `onboard_employee_async`: POST `/v1/organisations/{{ config.organization_id
  }}/employees/polling_onboard_employee` - kind `create`; body type `json`; required record fields
  `first_name`, `last_name`, `user_attributes`; accepted fields `first_name`, `last_name`,
  `user_attributes`; risk: starts an asynchronous employee onboarding job; approval required before
  execution.
- `update_employee_personal_details`: PATCH `/v1/organisations/{{ config.organization_id
  }}/employees/{{ record.employee_id }}/personal_details` - kind `update`; body type `json`; path
  fields `employee_id`; required record fields `employee_id`; accepted fields `employee_id`,
  `first_name`, `last_name`, `personal_email`; risk: external Employment Hero mutation; approval
  required before execution.
- `update_employee_employment_details`: PATCH `/v1/organisations/{{ config.organization_id
  }}/employees/{{ record.employee_id }}/employment_details` - kind `update`; body type `json`; path
  fields `employee_id`; required record fields `employee_id`; accepted fields `company_email`,
  `employee_id`, `job_title`, `start_date`; risk: external Employment Hero mutation; approval
  required before execution.
- `update_employee_contractor_details`: PATCH `/v1/organisations/{{ config.organization_id
  }}/employees/{{ record.employee_id }}/contractor_details` - kind `update`; body type `json`; path
  fields `employee_id`; required record fields `employee_id`; accepted fields `business_detail`,
  `employee_id`, `trading_name`; risk: external Employment Hero mutation; approval required before
  execution.
- `delete_employee`: DELETE `/v1/organisations/{{ config.organization_id }}/employees/{{
  record.employee_id }}` - kind `delete`; body type `none`; path fields `employee_id`; required
  record fields `employee_id`; accepted fields `employee_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes or removes an Employment Hero employee
  record.
- `update_employee_certification`: PATCH `/v1/organisations/{{ config.organization_id
  }}/employees/{{ record.employee_id }}/certifications/{{ record.id }}` - kind `update`; body type
  `json`; path fields `employee_id`, `id`; required record fields `employee_id`, `id`; accepted
  fields `completion_date`, `employee_id`, `expiry_date`, `id`; risk: external Employment Hero
  mutation; approval required before execution.
- `delete_form`: DELETE `/v1/organisations/{{ config.organization_id }}/forms/{{ record.form_id }}`
  - kind `delete`; body type `none`; path fields `form_id`; required record fields `form_id`;
  accepted fields `form_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes an Employment Hero form.
- `create_form_category`: POST `/v1/organisations/{{ config.organization_id }}/form_categories` -
  kind `create`; body type `json`; required record fields `name`; accepted fields `name`; risk:
  external Employment Hero mutation; approval required before execution.
- `update_form_category`: PATCH `/v1/organisations/{{ config.organization_id }}/form_categories/{{
  record.form_category_id }}` - kind `update`; body type `json`; path fields `form_category_id`;
  required record fields `form_category_id`; accepted fields `form_category_id`, `name`; risk:
  external Employment Hero mutation; approval required before execution.
- `delete_form_category`: DELETE `/v1/organisations/{{ config.organization_id }}/form_categories/{{
  record.form_category_id }}` - kind `delete`; body type `none`; path fields `form_category_id`;
  required record fields `form_category_id`; accepted fields `form_category_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external Employment Hero
  mutation; approval required before execution.
- `create_form_template`: POST `/v1/organisations/{{ config.organization_id }}/form_templates` -
  kind `create`; body type `json`; required record fields `name`; accepted fields `category_id`,
  `description`, `name`, `sections`; risk: external Employment Hero mutation; approval required
  before execution.
- `update_form_template`: PATCH `/v1/organisations/{{ config.organization_id }}/form_templates/{{
  record.template_id }}` - kind `update`; body type `json`; path fields `template_id`; required
  record fields `template_id`; accepted fields `name`, `sections`, `template_id`; risk: external
  Employment Hero mutation; approval required before execution.
- `delete_form_template`: DELETE `/v1/organisations/{{ config.organization_id }}/form_templates/{{
  record.template_id }}` - kind `delete`; body type `none`; path fields `template_id`; required
  record fields `template_id`; accepted fields `template_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external Employment Hero mutation; approval
  required before execution.
- `update_goal_archive_status`: PATCH `/v1/organisations/{{ config.organization_id }}/goals/{{
  record.goal_id }}/archive_status` - kind `update`; body type `json`; path fields `goal_id`;
  required record fields `goal_id`, `status`; accepted fields `goal_id`, `reason`, `status`; risk:
  archives or restores an Employment Hero goal.
- `update_goal_health_status`: PATCH `/v1/organisations/{{ config.organization_id }}/goals/{{
  record.goal_id }}/update_status` - kind `update`; body type `json`; path fields `goal_id`;
  required record fields `goal_id`, `health_status`; accepted fields `comment`, `goal_id`,
  `health_status`; risk: changes an Employment Hero goal health status.
- `bulk_grant_kiosk_access`: PATCH `/v1/organisations/{{ config.organization_id
  }}/kiosk_members/bulk_grant_access` - kind `update`; body type `json`; required record fields
  `member_ids`; accepted fields `member_ids`; risk: grants kiosk access to multiple members.
- `bulk_revoke_kiosk_access`: PATCH `/v1/organisations/{{ config.organization_id
  }}/kiosk_members/bulk_revoke_access` - kind `update`; body type `json`; required record fields
  `member_ids`; accepted fields `member_ids`, `notify_members`; risk: revokes kiosk access from
  multiple members.
- `update_leave_balance`: PUT `/v1/organisations/{{ config.organization_id }}/employees/{{
  record.employee_id }}/leave_balances/{{ record.id }}` - kind `update`; body type `json`; path
  fields `employee_id`, `id`; required record fields `employee_id`, `id`; accepted fields
  `accrued_amount`, `employee_id`, `id`, `reason`; risk: adjusts an employee leave balance.
- `create_leave_request`: POST `/v1/organisations/{{ config.organization_id }}/employees/{{
  record.employee_id }}/leave_requests` - kind `create`; body type `json`; path fields
  `employee_id`; required record fields `employee_id`, `leave_category_id`, `start_date`,
  `end_date`; accepted fields `comment`, `employee_id`, `end_date`, `hours_per_day`,
  `leave_category_id`, `start_date`; risk: creates an employee leave request.
- `create_position`: POST `/v1/organisations/{{ config.organization_id }}/positions` - kind
  `create`; body type `json`; required record fields `name`; accepted fields `cost_centre_id`,
  `member_ids`, `name`, `team_ids`, `work_site_id`; risk: external Employment Hero mutation;
  approval required before execution.
- `update_position`: PATCH `/v1/organisations/{{ config.organization_id }}/positions/{{
  record.position_id }}` - kind `update`; body type `json`; path fields `position_id`; required
  record fields `position_id`; accepted fields `member_ids`, `name`, `position_id`, `team_ids`;
  risk: external Employment Hero mutation; approval required before execution.
- `bulk_create_rostered_shifts`: POST `/v1/organisations/{{ config.organization_id
  }}/rostered_shifts/bulk_create` - kind `create`; body type `json`; required record fields
  `start_date_time`, `end_date_time`, `number_of_shifts`; accepted fields `breaks`, `end_date_time`,
  `member_ids`, `number_of_shifts`, `published`, `start_date_time`; risk: creates rostered shifts in
  bulk and may publish them.
- `create_timesheet_entries`: POST `/v1/organisations/{{ config.organization_id
  }}/timesheet_entries` - kind `create`; body type `json`; required record fields `timesheets`;
  accepted fields `timesheets`; risk: creates employee timesheet entries.
- `create_work_site`: POST `/v1/organisations/{{ config.organization_id }}/work_sites` - kind
  `create`; body type `json`; required record fields `name`; accepted fields `address`, `name`;
  risk: external Employment Hero mutation; approval required before execution.
- `update_work_site`: PUT `/v1/organisations/{{ config.organization_id }}/work_sites/{{
  record.work_site_id }}` - kind `update`; body type `json`; path fields `work_site_id`; required
  record fields `work_site_id`; accepted fields `address`, `name`, `work_site_id`; risk: external
  Employment Hero mutation; approval required before execution.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 57 stream-backed endpoint group(s), 33 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, non_data_endpoint=7.
