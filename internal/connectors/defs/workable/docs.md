# Overview

Reads Workable recruiting, account, employee, time tracking, time off, review, subscription,
requisition, and offer data; writes Workable candidate, employee, department, member, subscription,
time tracking, time off, offer, and requisition mutations.

Readable streams: `jobs`, `candidates`, `members`, `accounts`, `account`,
`collaboration_permissions`, `departments`, `disqualification_reasons`, `legal_entities`,
`permission_sets`, `recruiters`, `stages`, `subscriptions`, `employee_fields`, `employees_orgchart`,
`employees`, `employee`, `employee_documents`, `review_templates`, `review_template`,
`time_entries`, `timeoff_balances`, `timeoff_categories`, `timeoff_requests`, `work_schedules`,
`candidate_activities`, `candidate_files`, `candidate_offer`, `custom_attributes`, `events`,
`event`, `job_activities`, `job_custom_attributes`, `job_members`, `job_questions`, `job_stages`,
`job`, `job_application_form`, `job_recruiters`, `requisitions`, `requisition`, `offer`.

Write actions: `create_department`, `update_department`, `merge_department`, `delete_department`,
`invite_member`, `update_member`, `deactivate_member`, `enable_member`, `create_subscription`,
`delete_subscription`, `create_employee`, `update_employee`, `create_review_template`,
`bulk_create_time_entries`, `create_time_entry`, `update_time_entry`, `archive_time_entry`,
`decide_timeoff_approval`, `create_timeoff_request`, `update_candidate_custom_attribute`,
`comment_on_candidate`, `copy_candidate`, `disqualify_candidate`, `create_job_candidate`,
`move_candidate`, `relocate_candidate`, `revert_candidate_disqualification`,
`update_candidate_tags`, `rate_candidate`, `update_candidate_rating`, `update_candidate`,
`approve_offer`, `reject_offer`, `create_requisition`, `update_requisition`, `approve_requisition`,
`reject_requisition`, `create_talent_pool_candidate`.

Service API documentation: https://workable.readme.io/reference.

## Auth setup

Connection fields:

- `account_subdomain` (optional, string); Account subdomain for /accounts/{subdomain}.
- `api_key` (required, secret, string); Workable API access token sent as a Bearer token.
- `base_url` (required, string); format `uri`; Full Workable SPI v3 base URL, for example
  https://example.workable.com/spi/v3.
- `candidate_id` (optional, string); Candidate id for candidate detail and sub-resource streams.
- `employee_id` (optional, string); Employee id for employee detail, documents, and employee-scoped
  filters.
- `event_id` (optional, string); Event id for event detail stream.
- `job_shortcode` (optional, string); Job shortcode for job detail and sub-resource streams.
- `offer_id` (optional, string); Offer id for offer detail stream.
- `requisition_code` (optional, string); Requisition code for requisition detail stream.
- `review_template_id` (optional, string); Performance review template id.
- `start_date` (optional, string).
- `timeoff_from_date` (optional, string); default `2020-01-01`; Lower date bound for
  timeoff_requests, whose API requires from_date.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `timeoff_from_date=2020-01-01`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/jobs` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `paging.next`; next
URLs stay on the configured API host.

Pagination by stream: next_url: `jobs`, `candidates`, `members`, `recruiters`,
`candidate_activities`, `events`, `job_activities`, `requisitions`; none: `accounts`, `account`,
`collaboration_permissions`, `departments`, `disqualification_reasons`, `legal_entities`,
`permission_sets`, `stages`, `subscriptions`, `employee_fields`, `employees_orgchart`, `employee`,
`review_template`, `timeoff_balances`, `timeoff_categories`, `work_schedules`, `candidate_files`,
`candidate_offer`, `custom_attributes`, `event`, `job_custom_attributes`, `job_members`,
`job_questions`, `job_stages`, `job`, `job_application_form`, `job_recruiters`, `requisition`,
`offer`; offset_limit: `employees`, `employee_documents`, `review_templates`, `time_entries`,
`timeoff_requests`.

- `jobs`: GET `/jobs` - records path `jobs`; query `created_after` from template `{{
  config.start_date }}`, omitted when absent; `limit`=`100`; follows a next-page URL from the
  response body; URL path `paging.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `candidates`: GET `/candidates` - records path `candidates`; query `created_after` from template
  `{{ config.start_date }}`, omitted when absent; `limit`=`100`; follows a next-page URL from the
  response body; URL path `paging.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `members`: GET `/members` - records path `members`; query `created_after` from template `{{
  config.start_date }}`, omitted when absent; `limit`=`100`; follows a next-page URL from the
  response body; URL path `paging.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `accounts`: GET `/accounts` - records path `accounts`; emits passthrough records.
- `account`: GET `/accounts/{{ config.account_subdomain }}` - single-object response; records at
  response root; emits passthrough records.
- `collaboration_permissions`: GET `/collaboration_permissions` - single-object response; records at
  response root; computed output fields `_pm_id`; emits passthrough records.
- `departments`: GET `/departments` - records at response root; emits passthrough records.
- `disqualification_reasons`: GET `/disqualification_reasons` - single-object response; records at
  response root; computed output fields `_pm_id`; emits passthrough records.
- `legal_entities`: GET `/legal_entities` - records at response root; emits passthrough records.
- `permission_sets`: GET `/permission_sets` - records at response root; emits passthrough records.
- `recruiters`: GET `/recruiters` - records path `recruiters`; query `shortcode` from template `{{
  config.job_shortcode }}`, omitted when absent; follows a next-page URL from the response body; URL
  path `paging.next`; next URLs stay on the configured API host; emits passthrough records.
- `stages`: GET `/stages` - records path `stages`; emits passthrough records.
- `subscriptions`: GET `/subscriptions` - records path `subscriptions`; emits passthrough records.
- `employee_fields`: GET `/employee_fields` - records at response root; emits passthrough records.
- `employees_orgchart`: GET `/employees/orgchart` - records at response root; emits passthrough
  records.
- `employees`: GET `/employees` - records path `employees`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 10; emits passthrough records.
- `employee`: GET `/employees/{{ config.employee_id }}` - single-object response; records at
  response root; emits passthrough records.
- `employee_documents`: GET `/employees/{{ config.employee_id }}/documents` - records path
  `documents`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page
  size 10; emits passthrough records.
- `review_templates`: GET `/review-cycles/templates` - records path `templates`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 10; emits passthrough
  records.
- `review_template`: GET `/review-cycles/templates/{{ config.review_template_id }}` - single-object
  response; records at response root; emits passthrough records.
- `time_entries`: GET `/time-tracking/time-entries` - records path `time_entries`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 10; emits passthrough
  records.
- `timeoff_balances`: GET `/timeoff/balances` - records path `balances`; query `employee_id` from
  template `{{ config.employee_id }}`, omitted when absent; emits passthrough records.
- `timeoff_categories`: GET `/timeoff/categories` - records path `categories`; emits passthrough
  records.
- `timeoff_requests`: GET `/timeoff/requests` - records path `requests`; query `from_date` from
  template `{{ config.timeoff_from_date }}`, default `2020-01-01`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 10; emits passthrough records.
- `work_schedules`: GET `/work_schedules` - records at response root; emits passthrough records.
- `candidate_activities`: GET `/candidates/{{ config.candidate_id }}/activities` - records path
  `activities`; query `limit`=`100`; follows a next-page URL from the response body; URL path
  `paging.next`; next URLs stay on the configured API host; emits passthrough records.
- `candidate_files`: GET `/candidates/{{ config.candidate_id }}/files` - records path `files`; emits
  passthrough records.
- `candidate_offer`: GET `/candidates/{{ config.candidate_id }}/offer` - single-object response;
  records at response root; emits passthrough records.
- `custom_attributes`: GET `/custom_attributes` - records path `custom_attributes`; emits
  passthrough records.
- `events`: GET `/events` - records path `events`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `paging.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `event`: GET `/events/{{ config.event_id }}` - single-object response; records at response root;
  emits passthrough records.
- `job_activities`: GET `/jobs/{{ config.job_shortcode }}/activities` - records path `activities`;
  query `limit`=`100`; follows a next-page URL from the response body; URL path `paging.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `job_custom_attributes`: GET `/jobs/{{ config.job_shortcode }}/custom_attributes` - records path
  `custom_attributes`; emits passthrough records.
- `job_members`: GET `/jobs/{{ config.job_shortcode }}/members` - records path `members`; emits
  passthrough records.
- `job_questions`: GET `/jobs/{{ config.job_shortcode }}/questions` - records path `questions`;
  emits passthrough records.
- `job_stages`: GET `/jobs/{{ config.job_shortcode }}/stages` - records path `stages`; emits
  passthrough records.
- `job`: GET `/jobs/{{ config.job_shortcode }}` - single-object response; records at response root;
  emits passthrough records.
- `job_application_form`: GET `/jobs/{{ config.job_shortcode }}/application_form` - single-object
  response; records at response root; computed output fields `_pm_id`; emits passthrough records.
- `job_recruiters`: GET `/jobs/{{ config.job_shortcode }}/recruiters` - records path `recruiters`;
  emits passthrough records.
- `requisitions`: GET `/requisitions` - records path `requisitions`; query `created_after` from
  template `{{ config.start_date }}`, omitted when absent; `limit`=`100`; follows a next-page URL
  from the response body; URL path `paging.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `requisition`: GET `/requisitions/{{ config.requisition_code }}` - single-object response; records
  at response root; emits passthrough records.
- `offer`: GET `/offers/{{ config.offer_id }}` - single-object response; records at response root;
  emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, approves, rejects, archives, deactivates, or deletes Workable
recruiting/HR resources according to the selected action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_department`: POST `/departments` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: POST /departments mutates Workable data; approval required.
- `update_department`: PUT `/departments` - kind `update`; body type `json`; required record fields
  `id`; accepted fields `id`, `name`; risk: PUT /departments mutates Workable data; approval
  required.
- `merge_department`: POST `/departments/{{ record.department_id }}/merge` - kind `custom`; body
  type `json`; path fields `department_id`; required record fields `department_id`,
  `target_department_id`; accepted fields `department_id`, `target_department_id`; risk: POST
  /departments/{{ record.department_id }}/merge mutates Workable data; approval required.
- `delete_department`: DELETE `/departments/{{ record.department_id }}?force={{ record.force }}` -
  kind `delete`; body type `none`; path fields `department_id`, `force`; required record fields
  `department_id`, `force`; accepted fields `department_id`, `force`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: DELETE /departments/{{
  record.department_id }}?force={{ record.force }} mutates Workable data; approval required.
- `invite_member`: POST `/members/invite` - kind `create`; body type `json`; required record fields
  `email`; accepted fields `email`, `name`; risk: POST /members/invite mutates Workable data;
  approval required.
- `update_member`: PUT `/members` - kind `update`; body type `json`; required record fields `id`;
  accepted fields `id`, `name`; risk: PUT /members mutates Workable data; approval required.
- `deactivate_member`: DELETE `/members/{{ record.member_id }}` - kind `delete`; body type `none`;
  path fields `member_id`; required record fields `member_id`; accepted fields `member_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE /members/{{
  record.member_id }} mutates Workable data; approval required.
- `enable_member`: POST `/members/{{ record.member_id }}/enable` - kind `custom`; body type `none`;
  path fields `member_id`; required record fields `member_id`; accepted fields `member_id`; risk:
  POST /members/{{ record.member_id }}/enable mutates Workable data; approval required.
- `create_subscription`: POST `/subscriptions` - kind `create`; body type `json`; required record
  fields `target`, `event`; accepted fields `args`, `event`, `target`; risk: POST /subscriptions
  mutates Workable data; approval required.
- `delete_subscription`: DELETE `/subscriptions/{{ record.subscription_id }}` - kind `delete`; body
  type `none`; path fields `subscription_id`; required record fields `subscription_id`; accepted
  fields `subscription_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /subscriptions/{{ record.subscription_id }} mutates Workable data;
  approval required.
- `create_employee`: POST `/employees` - kind `create`; body type `json`; accepted fields
  `firstname`, `lastname`, `personal_email`; risk: POST /employees mutates Workable data; approval
  required.
- `update_employee`: PATCH `/employees/{{ record.employee_id }}` - kind `update`; body type `json`;
  path fields `employee_id`; required record fields `employee_id`; accepted fields `employee_id`,
  `firstname`; risk: PATCH /employees/{{ record.employee_id }} mutates Workable data; approval
  required.
- `create_review_template`: POST `/review-cycles/templates` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: POST /review-cycles/templates mutates
  Workable data; approval required.
- `bulk_create_time_entries`: POST `/time-tracking/time-entries` - kind `create`; body type `json`;
  required record fields `time_entries`; accepted fields `time_entries`; risk: POST
  /time-tracking/time-entries mutates Workable data; approval required.
- `create_time_entry`: POST `/time-tracking/employees/{{ record.employee_id }}/time-entries` - kind
  `create`; body type `json`; path fields `employee_id`; required record fields `employee_id`;
  accepted fields `duration`, `employee_id`, `from_date`, `type`; risk: POST
  /time-tracking/employees/{{ record.employee_id }}/time-entries mutates Workable data; approval
  required.
- `update_time_entry`: PATCH `/time-tracking/employees/{{ record.employee_id }}/time-entries/{{
  record.uuid }}` - kind `update`; body type `json`; path fields `employee_id`, `uuid`; required
  record fields `employee_id`, `uuid`; accepted fields `duration`, `employee_id`, `uuid`; risk:
  PATCH /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }} mutates
  Workable data; approval required.
- `archive_time_entry`: DELETE `/time-tracking/employees/{{ record.employee_id }}/time-entries/{{
  record.uuid }}` - kind `delete`; body type `none`; path fields `employee_id`, `uuid`; required
  record fields `employee_id`, `uuid`; accepted fields `employee_id`, `uuid`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }} mutates Workable
  data; approval required.
- `decide_timeoff_approval`: PATCH `/timeoff/approvals/{{ record.approval_key }}` - kind `update`;
  body type `json`; path fields `approval_key`; required record fields `approval_key`, `state`;
  accepted fields `approval_key`, `state`; risk: PATCH /timeoff/approvals/{{ record.approval_key }}
  mutates Workable data; approval required.
- `create_timeoff_request`: POST `/timeoff/requests` - kind `create`; body type `json`; required
  record fields `from_date`; accepted fields `category_id`, `employee_id`, `from_date`, `to_date`;
  risk: POST /timeoff/requests mutates Workable data; approval required.
- `update_candidate_custom_attribute`: PATCH `/candidates/{{ record.candidate_id
  }}/update_custom_attribute_value` - kind `update`; body type `json`; path fields `candidate_id`;
  required record fields `candidate_id`, `custom_attribute_id`, `value`; accepted fields
  `candidate_id`, `custom_attribute_id`, `value`; risk: PATCH /candidates/{{ record.candidate_id
  }}/update_custom_attribute_value mutates Workable data; approval required.
- `comment_on_candidate`: POST `/candidates/{{ record.candidate_id }}/comments` - kind `create`;
  body type `json`; path fields `candidate_id`; required record fields `candidate_id`, `comment`;
  accepted fields `candidate_id`, `comment`; risk: POST /candidates/{{ record.candidate_id
  }}/comments mutates Workable data; approval required.
- `copy_candidate`: POST `/candidates/{{ record.candidate_id }}/copy` - kind `create`; body type
  `json`; path fields `candidate_id`; required record fields `candidate_id`, `member_id`,
  `target_job_shortcode`, `target_stage`; accepted fields `candidate_id`, `member_id`,
  `target_job_shortcode`, `target_stage`; risk: POST /candidates/{{ record.candidate_id }}/copy
  mutates Workable data; approval required.
- `disqualify_candidate`: POST `/candidates/{{ record.candidate_id }}/disqualify` - kind `update`;
  body type `json`; path fields `candidate_id`; required record fields `candidate_id`; accepted
  fields `candidate_id`, `disqualification_reason`; risk: POST /candidates/{{ record.candidate_id
  }}/disqualify mutates Workable data; approval required.
- `create_job_candidate`: POST `/jobs/{{ record.job_shortcode }}/candidates` - kind `create`; body
  type `json`; path fields `job_shortcode`; required record fields `job_shortcode`, `name`; accepted
  fields `email`, `job_shortcode`, `name`; risk: POST /jobs/{{ record.job_shortcode }}/candidates
  mutates Workable data; approval required.
- `move_candidate`: POST `/candidates/{{ record.candidate_id }}/move` - kind `update`; body type
  `json`; path fields `candidate_id`; required record fields `candidate_id`, `target_stage`;
  accepted fields `candidate_id`, `target_stage`; risk: POST /candidates/{{ record.candidate_id
  }}/move mutates Workable data; approval required.
- `relocate_candidate`: POST `/candidates/{{ record.candidate_id }}/relocate` - kind `update`; body
  type `json`; path fields `candidate_id`; required record fields `candidate_id`,
  `target_job_shortcode`; accepted fields `candidate_id`, `target_job_shortcode`; risk: POST
  /candidates/{{ record.candidate_id }}/relocate mutates Workable data; approval required.
- `revert_candidate_disqualification`: POST `/candidates/{{ record.candidate_id }}/revert` - kind
  `update`; body type `none`; path fields `candidate_id`; required record fields `candidate_id`;
  accepted fields `candidate_id`; risk: POST /candidates/{{ record.candidate_id }}/revert mutates
  Workable data; approval required.
- `update_candidate_tags`: PUT `/candidates/{{ record.candidate_id }}/tags` - kind `update`; body
  type `json`; path fields `candidate_id`; required record fields `candidate_id`, `tags`; accepted
  fields `candidate_id`, `tags`; risk: PUT /candidates/{{ record.candidate_id }}/tags mutates
  Workable data; approval required.
- `rate_candidate`: POST `/candidates/{{ record.candidate_id }}/ratings` - kind `create`; body type
  `json`; path fields `candidate_id`; required record fields `candidate_id`, `rating`; accepted
  fields `candidate_id`, `comment`, `rating`; risk: POST /candidates/{{ record.candidate_id
  }}/ratings mutates Workable data; approval required.
- `update_candidate_rating`: PUT `/candidates/{{ record.candidate_id }}/ratings` - kind `upsert`;
  body type `json`; path fields `candidate_id`; required record fields `candidate_id`, `rating`;
  accepted fields `candidate_id`, `comment`, `rating`; risk: PUT /candidates/{{ record.candidate_id
  }}/ratings mutates Workable data; approval required.
- `update_candidate`: PATCH `/candidates/{{ record.candidate_id }}` - kind `update`; body type
  `json`; path fields `candidate_id`; required record fields `candidate_id`; accepted fields
  `candidate_id`, `name`; risk: PATCH /candidates/{{ record.candidate_id }} mutates Workable data;
  approval required.
- `approve_offer`: PATCH `/offers/{{ record.offer_id }}/approve` - kind `update`; body type `none`;
  path fields `offer_id`; required record fields `offer_id`; accepted fields `offer_id`; risk: PATCH
  /offers/{{ record.offer_id }}/approve mutates Workable data; approval required.
- `reject_offer`: PATCH `/offers/{{ record.offer_id }}/reject` - kind `update`; body type `none`;
  path fields `offer_id`; required record fields `offer_id`; accepted fields `offer_id`; risk: PATCH
  /offers/{{ record.offer_id }}/reject mutates Workable data; approval required.
- `create_requisition`: POST `/requisitions` - kind `create`; body type `json`; accepted fields
  `code`, `reason`; risk: POST /requisitions mutates Workable data; approval required.
- `update_requisition`: PATCH `/requisitions/{{ record.requisition_id }}` - kind `update`; body type
  `json`; path fields `requisition_id`; required record fields `requisition_id`; accepted fields
  `reason`, `requisition_id`; risk: PATCH /requisitions/{{ record.requisition_id }} mutates Workable
  data; approval required.
- `approve_requisition`: PATCH `/requisitions/{{ record.requisition_code }}/approve` - kind
  `update`; body type `none`; path fields `requisition_code`; required record fields
  `requisition_code`; accepted fields `requisition_code`; risk: PATCH /requisitions/{{
  record.requisition_code }}/approve mutates Workable data; approval required.
- `reject_requisition`: PATCH `/requisitions/{{ record.requisition_code }}/reject` - kind `update`;
  body type `none`; path fields `requisition_code`; required record fields `requisition_code`;
  accepted fields `requisition_code`; risk: PATCH /requisitions/{{ record.requisition_code }}/reject
  mutates Workable data; approval required.
- `create_talent_pool_candidate`: POST `/talent_pool/{{ record.stage }}/candidates` - kind `create`;
  body type `json`; path fields `stage`; required record fields `stage`, `name`; accepted fields
  `email`, `name`, `stage`; risk: POST /talent_pool/{{ record.stage }}/candidates mutates Workable
  data; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 42 stream-backed endpoint group(s), 38 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=1.
