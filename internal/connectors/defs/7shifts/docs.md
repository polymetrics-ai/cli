# Overview

Reads the documented 7shifts v2 REST API surface and executes single-request reverse-ETL
actions for supported mutations.

Readable streams: `companies`, `locations`, `departments`, `roles`, `users`, `shifts`,
`time_punches`, `fetch_tip_pool_manual_entry`, `find_by_id`, `get_assignments`,
`get_availability_by_id`, `get_daily_sales_and_labor`, `get_engage_overview_by_location_id`,
`get_event`, `get_events`, `get_hours_and_wages`, `get_location_by_id`, `get_receipt`,
`get_role_assignments`, `get_task_list`, `get_task_list_daily_summary`, `get_task_list_template`,
`get_task_list_templates`, `get_task_lists`, `get_task_management_settings`,
`get_time_clocking_payroll_period`, `get_time_clocking_payroll_periods`, `get_time_off_list`,
`get_time_off_settings`, `get_time_punch_by_id`, `get_tip_pool_settings`, `get_total_hours`,
`get_user`, `get_user_wages`, `list_availabilities`, `list_availability_reasons`,
`list_company_webhooks`, `list_department_assignments`, `list_employment_record`,
`list_external_user_mappings`, `list_inactive_reasons`, `list_location_assignments`,
`list_log_book_categories`, `list_log_book_comments`, `list_log_book_posts`, `list_sales_receipts`,
`list_scheduled_shifts`, `list_shift_feedback`, `list_user_contacts`,
`list_users_authorized_locations`, `retrieve_company_labor_settings`, `retrieve_daily_stats`,
`retrieve_day_part_settings`, `retrieve_department`, `retrieve_external_user_mapping`,
`retrieve_log_book_comment`, `retrieve_log_book_post`, `retrieve_receipts_summary`, `retrieve_role`,
`retrieve_shift`, `retrieve_tip_pool_detailed_report`, `retrieve_tip_pool_summary_report`,
`retrieve_user_contact`, `view_company`, `who_am_i`.

Write actions: `approve_time_off`, `clear_task`, `complete_task`, `create_availability`,
`create_availability_reason`, `create_bulk_forecast_overrides`, `create_complete_receipt`,
`create_department`, `create_department_assignment`, `create_employment_record`, `create_event`,
`create_external_user_mappings`, `create_forecast_override`, `create_location`,
`create_location_assignment`, `create_log_book_category`, `create_log_book_comment`,
`create_log_book_post`, `create_projected_sales_interval_override`, `create_role`,
`create_role_assignment`, `create_task_list_template`, `create_task_tags`, `create_time_off`,
`create_user_mappings_bulk`, `create_user_wages`, `create_webhook`, `deactivate_user`,
`decline_time_off`, `delete_availability`, `delete_availability_reason`, `delete_company_webhook`,
`delete_department`, `delete_department_assignment`, `delete_employment_record`, `delete_event`,
`delete_external_user_mappings`, `delete_forecast_override`, `delete_location`,
`delete_location_assignment`, `delete_log_book_category`, `delete_log_book_comment`,
`delete_log_book_post`, `delete_role`, `delete_role_assignment`, `delete_shift`,
`delete_task_list_template`, `delete_task_tags`, `delete_time_off`, `delete_time_punch_by_id`,
`edit_availability`, `edit_availability_reason`, `edit_company_webhook`, `edit_event`,
`edit_task_list_template`, `edit_time_off`, `post_shift`, `post_time_punch`, `post_user`,
`put_time_punch`, `put_user`, `save_time_off_settings`, `save_tip_pool_manual_entry`,
`sync_overridden_projected_sales_interval`, `update_availability_status`, `update_company`,
`update_complete_receipt`, `update_department`, `update_employment_record`,
`update_external_user_mappings`, `update_location`, `update_log_book_category`, `update_role`,
`update_role_assignment`, `update_shift`, `upsert_bulk_employment_records`.

Service API documentation: https://developers.7shifts.com/reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); 7shifts API access token used only for Bearer auth.
- `availability_id` (optional, string); Required path or query value for streams that reference
  availability_id.
- `base_url` (optional, string); default `https://api.7shifts.com`; format `uri`; 7shifts API base
  URL override for tests or proxies.
- `company_id` (required, string); 7shifts company id used by company-scoped streams and write
  actions.
- `date` (optional, string); Required date or date-time filter for streams that reference date.
- `date_end` (optional, string); Required date or date-time filter for streams that reference
  date_end.
- `date_start` (optional, string); Required date or date-time filter for streams that reference
  date_start.
- `department_id` (optional, string); Required path or query value for streams that reference
  department_id.
- `employee_id` (optional, string); Required path or query value for streams that reference
  employee_id.
- `end_date` (optional, string); Required date or date-time filter for streams that reference
  end_date.
- `event_id` (optional, string); Required path or query value for streams that reference event_id.
- `from` (optional, string); Required date or date-time filter for streams that reference from.
- `id` (optional, string); Required path or query value for streams that reference id.
- `identifier` (optional, string); Required path or query value for streams that reference
  identifier.
- `list_id` (optional, string); Required path or query value for streams that reference list_id.
- `location_id` (optional, string); Required path or query value for streams that reference
  location_id.
- `log_book_ids` (optional, string); Required path or query value for streams that reference
  log_book_ids.
- `payroll_period_id` (optional, string); Required path or query value for streams that reference
  payroll_period_id.
- `punches` (optional, string); Required path or query value for streams that reference punches.
- `receipt_id` (optional, string); Required path or query value for streams that reference
  receipt_id.
- `role_id` (optional, string); Required path or query value for streams that reference role_id.
- `shift_id` (optional, string); Required path or query value for streams that reference shift_id.
- `start_date` (optional, string); format `date-time`.
- `time_off_id` (optional, string); Required path or query value for streams that reference
  time_off_id.
- `time_punch_id` (optional, string); Required path or query value for streams that reference
  time_punch_id.
- `tip_pool_settings_uuid` (optional, string); Required path or query value for streams that
  reference tip_pool_settings_uuid.
- `to` (optional, string); Required date or date-time filter for streams that reference to.
- `user_id` (optional, string); Required path or query value for streams that reference user_id.
- `uuid` (optional, string); Required path or query value for streams that reference uuid.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.7shifts.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/companies` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`meta.cursor.next`.

Pagination by stream: cursor: `companies`, `locations`, `departments`, `roles`, `users`, `shifts`,
`time_punches`, `get_task_list_templates`, `get_time_clocking_payroll_periods`, `get_time_off_list`,
`list_availabilities`, `list_availability_reasons`, `list_company_webhooks`,
`list_external_user_mappings`, `list_log_book_posts`, `list_sales_receipts`, `list_shift_feedback`,
`list_user_contacts`; none: `fetch_tip_pool_manual_entry`, `find_by_id`, `get_assignments`,
`get_availability_by_id`, `get_daily_sales_and_labor`, `get_engage_overview_by_location_id`,
`get_event`, `get_events`, `get_hours_and_wages`, `get_location_by_id`, `get_receipt`,
`get_role_assignments`, `get_task_list`, `get_task_list_daily_summary`, `get_task_list_template`,
`get_task_lists`, `get_task_management_settings`, `get_time_clocking_payroll_period`,
`get_time_off_settings`, `get_time_punch_by_id`, `get_tip_pool_settings`, `get_total_hours`,
`get_user`, `get_user_wages`, `list_department_assignments`, `list_employment_record`,
`list_inactive_reasons`, `list_location_assignments`, `list_log_book_categories`,
`list_log_book_comments`, `list_scheduled_shifts`, `list_users_authorized_locations`,
`retrieve_company_labor_settings`, `retrieve_daily_stats`, `retrieve_day_part_settings`,
`retrieve_department`, `retrieve_external_user_mapping`, `retrieve_log_book_comment`,
`retrieve_log_book_post`, `retrieve_receipts_summary`, `retrieve_role`, `retrieve_shift`,
`retrieve_tip_pool_detailed_report`, `retrieve_tip_pool_summary_report`, `retrieve_user_contact`,
`view_company`, `who_am_i`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `companies`: GET `/v2/companies` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `meta.cursor.next`; incremental cursor `modified`; sent
  as `modified_since`; formatted as YYYY-MM-DD date; initial lower bound from `start_date`.
- `locations`: GET `/v2/company/{{ config.company_id }}/locations` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `meta.cursor.next`;
  incremental cursor `modified`; sent as `modified_since`; formatted as YYYY-MM-DD date; initial
  lower bound from `start_date`.
- `departments`: GET `/v2/company/{{ config.company_id }}/departments` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `meta.cursor.next`;
  incremental cursor `modified`; sent as `modified_since`; formatted as YYYY-MM-DD date; initial
  lower bound from `start_date`.
- `roles`: GET `/v2/company/{{ config.company_id }}/roles` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `meta.cursor.next`;
  incremental cursor `modified`; sent as `modified_since`; formatted as YYYY-MM-DD date; initial
  lower bound from `start_date`.
- `users`: GET `/v2/company/{{ config.company_id }}/users` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `meta.cursor.next`;
  incremental cursor `modified`; sent as `modified_since`; formatted as YYYY-MM-DD date; initial
  lower bound from `start_date`.
- `shifts`: GET `/v2/company/{{ config.company_id }}/shifts` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `meta.cursor.next`;
  incremental cursor `modified`; sent as `modified_since`; formatted as YYYY-MM-DD date; initial
  lower bound from `start_date`.
- `time_punches`: GET `/v2/company/{{ config.company_id }}/time_punches` - records path `data`;
  query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`; incremental cursor `modified`; sent as `modified_since`; formatted as
  YYYY-MM-DD date; initial lower bound from `start_date`.
- `fetch_tip_pool_manual_entry`: GET `/v2/company/{{ config.company_id }}/tip_pool/{{
  config.tip_pool_settings_uuid }}/manual_entry` - records path `data`; query `end_date`=`{{
  config.end_date }}`; `start_date`=`{{ config.start_date }}`.
- `find_by_id`: GET `/v2/time_off/{{ config.time_off_id }}` - single-object response; records path
  `.`.
- `get_assignments`: GET `/v2/company/{{ config.company_id }}/users/{{ config.user_id
  }}/assignments` - single-object response; records path `data`; computed output fields
  `stream_key`.
- `get_availability_by_id`: GET `/v2/company/{{ config.company_id }}/availabilities/{{
  config.availability_id }}` - single-object response; records path `data`.
- `get_daily_sales_and_labor`: GET `/v2/reports/daily_sales_and_labor` - records path `data`; query
  `end_date`=`{{ config.end_date }}`; `location_id`=`{{ config.location_id }}`; `start_date`=`{{
  config.start_date }}`.
- `get_engage_overview_by_location_id`: GET `/v2/company/{{ config.company_id }}/locations/{{
  config.location_id }}/engage_overview` - single-object response; records path `data`; query
  `date`=`{{ config.date }}`; computed output fields `stream_key`.
- `get_event`: GET `/v2/company/{{ config.company_id }}/events/{{ config.event_id }}` -
  single-object response; records path `data`.
- `get_events`: GET `/v2/company/{{ config.company_id }}/events` - records path `data`; query
  `end_date`=`{{ config.end_date }}`; `limit`=`100`; `start_date`=`{{ config.start_date }}`.
- `get_hours_and_wages`: GET `/v2/reports/hours_and_wages` - single-object response; records path
  `.`; query `company_id`=`{{ config.company_id }}`; `from`=`{{ config.from }}`; `punches`=`{{
  config.punches }}`; `to`=`{{ config.to }}`; computed output fields `stream_key`.
- `get_location_by_id`: GET `/v2/company/{{ config.company_id }}/locations/{{ config.location_id }}`
  - single-object response; records path `data`.
- `get_receipt`: GET `/v2/company/{{ config.company_id }}/receipts/{{ config.receipt_id }}` -
  single-object response; records path `data`.
- `get_role_assignments`: GET `/v2/company/{{ config.company_id }}/users/{{ config.user_id
  }}/role_assignments` - records path `data`.
- `get_task_list`: GET `/v2/company/{{ config.company_id }}/task_lists/{{ config.list_id }}` -
  single-object response; records path `data`.
- `get_task_list_daily_summary`: GET `/v2/company/{{ config.company_id }}/task_list_daily_summary` -
  single-object response; records path `data`; query `date`=`{{ config.date }}`; `location_id`=`{{
  config.location_id }}`.
- `get_task_list_template`: GET `/v2/company/{{ config.company_id }}/task_list_templates/{{
  config.uuid }}` - single-object response; records path `data`.
- `get_task_list_templates`: GET `/v2/company/{{ config.company_id }}/task_list_templates` - records
  path `data`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`.
- `get_task_lists`: GET `/v2/company/{{ config.company_id }}/task_lists` - records path `data`.
- `get_task_management_settings`: GET `/v2/company/{{ config.company_id }}/task_management_settings`
  - single-object response; records path `data`.
- `get_time_clocking_payroll_period`: GET `/v2/company/{{ config.company_id }}/payroll_periods/{{
  config.payroll_period_id }}` - single-object response; records path `data`.
- `get_time_clocking_payroll_periods`: GET `/v2/time_clocking/payroll_periods` - records path
  `data`; query `company_id`=`{{ config.company_id }}`; `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `meta.cursor.next`.
- `get_time_off_list`: GET `/v2/time_off` - records path `data`; query `company_id`=`{{
  config.company_id }}`; `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token
  from `meta.cursor.next`.
- `get_time_off_settings`: GET `/v2/time_off_settings/{{ config.company_id }}` - single-object
  response; records path `.`.
- `get_time_punch_by_id`: GET `/v2/company/{{ config.company_id }}/time_punches/{{
  config.time_punch_id }}` - single-object response; records path `data`.
- `get_tip_pool_settings`: GET `/v2/company/{{ config.company_id }}/tip_pool_settings` - records
  path `data`.
- `get_total_hours`: GET `/v2/time_off/total_hours` - records path `data`; query `company_id`=`{{
  config.company_id }}`; `date_end`=`{{ config.date_end }}`; `date_start`=`{{ config.date_start }}`;
  `employee_id`=`{{ config.employee_id }}`.
- `get_user`: GET `/v2/company/{{ config.company_id }}/users/{{ config.identifier }}` -
  single-object response; records path `data`.
- `get_user_wages`: GET `/v2/company/{{ config.company_id }}/users/{{ config.user_id }}/wages` -
  single-object response; records path `data`; computed output fields `stream_key`.
- `list_availabilities`: GET `/v2/company/{{ config.company_id }}/availabilities` - records path
  `data`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`.
- `list_availability_reasons`: GET `/v2/company/{{ config.company_id }}/availability_reasons` -
  records path `data`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token
  from `meta.cursor.next`.
- `list_company_webhooks`: GET `/v2/company/{{ config.company_id }}/webhooks` - records path `data`;
  query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`.
- `list_department_assignments`: GET `/v2/company/{{ config.company_id }}/users/{{ config.user_id
  }}/department_assignments` - records path `data`.
- `list_employment_record`: GET `/v2/company/{{ config.company_id }}/employment_records` - records
  path `data`.
- `list_external_user_mappings`: GET `/v2/company/{{ config.company_id }}/external_user_mappings` -
  records path `data`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token
  from `meta.cursor.next`.
- `list_inactive_reasons`: GET `/v2/company/{{ config.company_id }}/inactive_reasons` - records path
  `data`; computed output fields `stream_key`.
- `list_location_assignments`: GET `/v2/company/{{ config.company_id }}/users/{{ config.user_id
  }}/location_assignments` - records path `data`.
- `list_log_book_categories`: GET `/v2/company/{{ config.company_id }}/log_book_categories` -
  records path `data`.
- `list_log_book_comments`: GET `/v2/company/{{ config.company_id }}/log_book_comments` - records
  path `data`; query `log_book_ids`=`{{ config.log_book_ids }}`.
- `list_log_book_posts`: GET `/v2/company/{{ config.company_id }}/log_book_posts` - records path
  `data`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`.
- `list_sales_receipts`: GET `/v2/company/{{ config.company_id }}/receipts` - records path `data`;
  query `limit`=`100`; `location_id`=`{{ config.location_id }}`; cursor pagination; cursor parameter
  `cursor`; next token from `meta.cursor.next`.
- `list_scheduled_shifts`: GET `/v2/company/{{ config.company_id }}/shifts_scheduled/{{ config.id
  }}` - records path `data`; query `location_id`=`{{ config.location_id }}`.
- `list_shift_feedback`: GET `/v2/company/{{ config.company_id }}/shift_feedback` - records path
  `data`; query `end_date`=`{{ config.end_date }}`; `limit`=`100`; `start_date`=`{{
  config.start_date }}`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`.
- `list_user_contacts`: GET `/v2/company/{{ config.company_id }}/contacts` - records path `data`;
  query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.cursor.next`.
- `list_users_authorized_locations`: GET `/v2/company/{{ config.company_id }}/users/{{
  config.user_id }}/authorized_locations` - records path `data`.
- `retrieve_company_labor_settings`: GET `/v2/company/{{ config.company_id }}/labor_settings` -
  single-object response; records path `data`.
- `retrieve_daily_stats`: GET `/v2/company/{{ config.company_id }}/location/{{ config.location_id
  }}/daily_stats` - single-object response; records path `data`; query `date`=`{{ config.date }}`;
  computed output fields `stream_key`.
- `retrieve_day_part_settings`: GET `/v2/company/{{ config.company_id }}/day_part/settings` -
  records path `data`.
- `retrieve_department`: GET `/v2/company/{{ config.company_id }}/departments/{{
  config.department_id }}` - single-object response; records path `data`.
- `retrieve_external_user_mapping`: GET `/v2/company/{{ config.company_id
  }}/external_user_mappings/{{ config.identifier }}` - single-object response; records path `data`.
- `retrieve_log_book_comment`: GET `/v2/company/{{ config.company_id }}/log_book_comments/{{
  config.id }}` - single-object response; records path `data`.
- `retrieve_log_book_post`: GET `/v2/company/{{ config.company_id }}/log_book_posts/{{ config.id }}`
  - single-object response; records path `data`.
- `retrieve_receipts_summary`: GET `/v2/company/{{ config.company_id }}/receipts_summary` - records
  path `data`; query `location_id`=`{{ config.location_id }}`.
- `retrieve_role`: GET `/v2/company/{{ config.company_id }}/roles/{{ config.role_id }}` -
  single-object response; records path `data`.
- `retrieve_shift`: GET `/v2/company/{{ config.company_id }}/shifts/{{ config.shift_id }}` -
  single-object response; records path `data`.
- `retrieve_tip_pool_detailed_report`: GET `/v2/company/{{ config.company_id }}/locations/{{
  config.location_id }}/tip_pool_detailed_report` - single-object response; records path `data`;
  query `end_date`=`{{ config.end_date }}`; `start_date`=`{{ config.start_date }}`.
- `retrieve_tip_pool_summary_report`: GET `/v2/company/{{ config.company_id }}/locations/{{
  config.location_id }}/tip_pool_summary_report` - records path `data`; query `end_date`=`{{
  config.end_date }}`; `start_date`=`{{ config.start_date }}`.
- `retrieve_user_contact`: GET `/v2/company/{{ config.company_id }}/contacts/{{ config.identifier
  }}` - single-object response; records path `data`.
- `view_company`: GET `/v2/companies/{{ config.id }}` - single-object response; records path `.`.
- `who_am_i`: GET `/v2/whoami` - single-object response; records path `data`.

## Write actions & risks

Overall write risk: creates, updates, deletes, approves, declines, or otherwise mutates configured
7shifts account resources through single-request REST actions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `approve_time_off`: POST `/v2/time_off/{{ record.time_off_id }}/approve` - kind `custom`; body
  type `json`; path fields `time_off_id`; required record fields `time_off_id`; accepted fields
  `status_action_message`, `time_off_id`; risk: Approve Time Off Request in the configured 7shifts
  account.
- `clear_task`: POST `/v2/company/{{ config.company_id }}/task_lists/{{ record.list_id }}/tasks/{{
  record.task_id }}/clear` - kind `custom`; body type `json`; path fields `list_id`, `task_id`;
  required record fields `task_id`, `list_id`, `user_id`; accepted fields `list_id`, `task_id`,
  `user_id`; risk: Clear Task in the configured 7shifts account.
- `complete_task`: POST `/v2/company/{{ config.company_id }}/task_lists/{{ record.list_id
  }}/tasks/{{ record.task_id }}/complete` - kind `custom`; body type `json`; path fields `list_id`,
  `task_id`; required record fields `task_id`, `list_id`, `user_id`; accepted fields
  `completion_value`, `list_id`, `task_id`, `user_id`; risk: Complete Task in the configured 7shifts
  account.
- `create_availability`: POST `/v2/company/{{ config.company_id }}/availabilities` - kind `create`;
  body type `json`; required record fields `user_id`, `repeat`, `mon`, `mon_from`, `mon_to`,
  `mon_comments`, `mon_reason`, `tue`, `tue_from`, `tue_to`, `tue_comments`, `tue_reason`, and 25
  more; accepted fields `fri`, `fri_comments`, `fri_from`, `fri_reason`, `fri_to`, `mon`,
  `mon_comments`, `mon_from`, `mon_reason`, `mon_to`, `repeat`, `sat`, `sat_comments`, `sat_from`,
  `sat_reason`, `sat_to`, `sun`, `sun_comments`, and 21 more; risk: Create Availability in the
  configured 7shifts account.
- `create_availability_reason`: POST `/v2/company/{{ config.company_id }}/availability_reasons` -
  kind `create`; body type `json`; required record fields `reason`; accepted fields
  `comments_required`, `reason`; risk: Create Availability Reason in the configured 7shifts account.
- `create_bulk_forecast_overrides`: POST `/v2/company/{{ config.company_id }}/location/{{
  record.location_id }}/forecast_overrides` - kind `create`; body type `json`; path fields
  `location_id`; required record fields `location_id`, `data`; accepted fields `data`,
  `location_id`; risk: Create Daily Projected Forecast Overrides in the configured 7shifts account.
- `create_complete_receipt`: POST `/v2/company/{{ config.company_id }}/receipts` - kind `custom`;
  body type `json`; required record fields `location_id`, `receipt_date`, `receipt_lines`,
  `tip_details`, `net_total`, `status`; accepted fields `dining_option`, `external_user_id`,
  `gross_total`, `location_id`, `net_total`, `order_type`, `receipt_date`, `receipt_lines`,
  `revenue_center`, `status`, `tip_details`, `tips`, `total_receipt_discounts`; risk: Create Receipt
  in the configured 7shifts account.
- `create_department`: POST `/v2/company/{{ config.company_id }}/departments` - kind `create`; body
  type `json`; required record fields `location_id`, `name`, `default`; accepted fields `default`,
  `location_id`, `name`; risk: Create Department in the configured 7shifts account.
- `create_department_assignment`: POST `/v2/company/{{ config.company_id }}/users/{{ record.user_id
  }}/department_assignments` - kind `create`; body type `json`; path fields `user_id`; required
  record fields `user_id`, `department_id`; accepted fields `appear_on_schedule`, `department_id`,
  `user_id`; risk: Create Department Assignment in the configured 7shifts account.
- `create_employment_record`: POST `/v2/company/{{ config.company_id }}/employment_records` - kind
  `create`; body type `json`; required record fields `user_id`; accepted fields
  `business_entity_uuid`, `classification`, `hire_date`, `termination_date`, `user_id`; risk: Create
  Employment Record in the configured 7shifts account.
- `create_event`: POST `/v2/company/{{ config.company_id }}/events` - kind `create`; body type
  `json`; required record fields `location_ids`, `start_date`, `start_time`, `end_time`, `end_date`,
  `title`, `is_multi_day`; accepted fields `color`, `description`, `end_date`, `end_time`,
  `is_multi_day`, `location_ids`, `recurrence`, `start_date`, `start_time`, `title`; risk: Create
  Event in the configured 7shifts account.
- `create_external_user_mappings`: POST `/v2/company/{{ config.company_id }}/external_user_mappings`
  - kind `create`; body type `json`; required record fields `user_id`, `external_user_id`; accepted
  fields `external_user_id`, `user_id`; risk: Create External User Mapping in the configured 7shifts
  account.
- `create_forecast_override`: POST `/v2/company/{{ config.company_id }}/location/{{
  record.location_id }}/forecast_override` - kind `create`; body type `json`; path fields
  `location_id`; required record fields `location_id`, `date`, `value`, `report_type`; accepted
  fields `date`, `department_id`, `location_id`, `report_type`, `value`; risk: Create Daily
  Projected Forecast Override in the configured 7shifts account.
- `create_location`: POST `/v2/company/{{ config.company_id }}/locations` - kind `create`; body type
  `json`; required record fields `name`, `country`; accepted fields `city`, `copy_from_id`,
  `country`, `coupon`, `enable_shift_feedback`, `formatted_address`, `fri_hours_close`,
  `fri_hours_open`, `fri_is_closed`, `holiday_pay`, `latitude`, `longitude`, `mon_hours_close`,
  `mon_hours_open`, `mon_is_closed`, `name`, `place_id`, `sat_hours_close`, and 18 more; risk:
  Create Location in the configured 7shifts account.
- `create_location_assignment`: POST `/v2/company/{{ config.company_id }}/users/{{ record.user_id
  }}/location_assignments` - kind `create`; body type `json`; path fields `user_id`; required record
  fields `user_id`, `location_id`; accepted fields `location_id`, `user_id`; risk: Create Location
  Assignments in the configured 7shifts account.
- `create_log_book_category`: POST `/v2/company/{{ config.company_id }}/log_book_categories` - kind
  `create`; body type `json`; required record fields `name`; accepted fields `col`, `field_type`,
  `name`, `notify`, `required`, `sort`; risk: Create Log Book Category in the configured 7shifts
  account.
- `create_log_book_comment`: POST `/v2/company/{{ config.company_id }}/log_book_comments` - kind
  `create`; body type `json`; required record fields `log_book_id`, `message`; accepted fields
  `log_book_id`, `message`; risk: Create Log Book Comment in the configured 7shifts account.
- `create_log_book_post`: POST `/v2/company/{{ config.company_id }}/log_book_posts` - kind `create`;
  body type `json`; required record fields `location_id`, `log_book_category_id`, `date`, `message`;
  accepted fields `attachment_uuid`, `attachments`, `date`, `location_id`, `log_book_category_id`,
  `message`; risk: Create Log Book Post in the configured 7shifts account.
- `create_projected_sales_interval_override`: POST `/v2/company/{{ config.company_id }}/locations/{{
  record.location_id }}/forecast_override_interval` - kind `create`; body type `json`; path fields
  `location_id`; required record fields `location_id`, `start`, `end`, `value`; accepted fields
  `department_id`, `end`, `location_id`, `report_type`, `start`, `value`; risk: Create Sales
  Forecast Override Interval in the configured 7shifts account.
- `create_role`: POST `/v2/company/{{ config.company_id }}/roles` - kind `create`; body type `json`;
  required record fields `name`, `color`, `location_id`, `department_id`; accepted fields `color`,
  `department_id`, `is_tipped_role`, `job_code`, `location_id`, `name`, `sort`, `stations`; risk:
  Create Role in the configured 7shifts account.
- `create_role_assignment`: POST `/v2/company/{{ config.company_id }}/users/{{ record.user_id
  }}/role_assignments` - kind `create`; body type `json`; path fields `user_id`; required record
  fields `user_id`, `role_id`; accepted fields `primary`, `role_id`, `user_id`; risk: Create Role
  Assignment in the configured 7shifts account.
- `create_task_list_template`: POST `/v2/company/{{ config.company_id }}/task_list_templates` - kind
  `create`; body type `json`; required record fields `title`, `recurrence`, `assignments`; accepted
  fields `assignments`, `description`, `due`, `recurrence`, `status`, `task_templates`,
  `time_frame`, `title`; risk: Create Task List Template in the configured 7shifts account.
- `create_task_tags`: POST `/v2/company/{{ config.company_id }}/task_tags` - kind `create`; body
  type `json`; required record fields `company_id`, `tags`; accepted fields `company_id`, `tags`;
  risk: Create Task Tags in the configured 7shifts account.
- `create_time_off`: POST `/v2/time_off` - kind `create`; body type `json`; required record fields
  `user_id`, `company_id`, `from_date`, `to_date`, `partial`, `status`, `category`; accepted fields
  `category`, `comments`, `company_id`, `from_date`, `hours`, `partial`, `partial_from`,
  `partial_to`, `status`, `status_action_message`, `status_action_user_id`, `to_date`, `user_id`;
  risk: Create Time Off in the configured 7shifts account.
- `create_user_mappings_bulk`: POST `/v2/company/{{ config.company_id
  }}/external_user_mappings_bulk` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: Create User External Mappings in the configured 7shifts account.
- `create_user_wages`: POST `/v2/company/{{ config.company_id }}/users/{{ record.user_id }}/wages` -
  kind `create`; body type `json`; path fields `user_id`; required record fields `user_id`,
  `effective_date`, `wage_type`, `wage_cents`; accepted fields `allocations`, `effective_date`,
  `enabled_days`, `role_id`, `user_id`, `wage_cents`, `wage_type`; risk: Create User Wage in the
  configured 7shifts account.
- `create_webhook`: POST `/v2/company/{{ config.company_id }}/webhooks` - kind `create`; body type
  `json`; required record fields `url`, `method`, `event`; accepted fields `event`, `method`, `url`;
  risk: Create Webhook in the configured 7shifts account.
- `deactivate_user`: DELETE `/v2/company/{{ config.company_id }}/users/{{ record.identifier }}` -
  kind `delete`; body type `json`; path fields `identifier`; required record fields `identifier`,
  `inactive_reason`; accepted fields `identifier`, `inactive_comments`, `inactive_reason`,
  `would_rehire`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deactivate User in the configured 7shifts account.
- `decline_time_off`: POST `/v2/time_off/{{ record.time_off_id }}/decline` - kind `custom`; body
  type `json`; path fields `time_off_id`; required record fields `time_off_id`; accepted fields
  `status_action_message`, `time_off_id`; risk: Decline Time Off Request in the configured 7shifts
  account.
- `delete_availability`: DELETE `/v2/company/{{ config.company_id }}/availabilities/{{
  record.availability_id }}` - kind `delete`; body type `none`; path fields `availability_id`;
  required record fields `availability_id`; accepted fields `availability_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete Availability in the
  configured 7shifts account.
- `delete_availability_reason`: DELETE `/v2/company/{{ config.company_id }}/availability_reasons/{{
  record.availability_reason_id }}` - kind `delete`; body type `none`; path fields
  `availability_reason_id`; required record fields `availability_reason_id`; accepted fields
  `availability_reason_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Availability Reason in the configured 7shifts account.
- `delete_company_webhook`: DELETE `/v2/company/{{ config.company_id }}/webhooks/{{
  record.webhook_id }}` - kind `delete`; body type `none`; path fields `webhook_id`; required record
  fields `webhook_id`; accepted fields `webhook_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete Webhook in the configured 7shifts account.
- `delete_department`: DELETE `/v2/company/{{ config.company_id }}/departments/{{
  record.department_id }}` - kind `delete`; body type `none`; path fields `department_id`; required
  record fields `department_id`; accepted fields `department_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete Department in the configured 7shifts
  account.
- `delete_department_assignment`: DELETE `/v2/company/{{ config.company_id }}/users/{{
  record.user_id }}/department_assignments/{{ record.department_id }}` - kind `delete`; body type
  `none`; path fields `user_id`, `department_id`; required record fields `department_id`, `user_id`;
  accepted fields `department_id`, `user_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete Department Assignment in the configured 7shifts account.
- `delete_employment_record`: DELETE `/v2/company/{{ config.company_id }}/employment_record/{{
  record.uuid }}` - kind `delete`; body type `none`; path fields `uuid`; required record fields
  `uuid`; accepted fields `uuid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Employment Record in the configured 7shifts account.
- `delete_event`: DELETE `/v2/company/{{ config.company_id }}/events/{{ record.event_id }}` - kind
  `delete`; body type `none`; path fields `event_id`; required record fields `event_id`; accepted
  fields `event_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Event in the configured 7shifts account.
- `delete_external_user_mappings`: DELETE `/v2/company/{{ config.company_id
  }}/external_user_mappings/{{ record.identifier }}` - kind `delete`; body type `none`; path fields
  `identifier`; required record fields `identifier`; accepted fields `identifier`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete External User
  Mapping in the configured 7shifts account.
- `delete_forecast_override`: DELETE `/v2/company/{{ config.company_id }}/location/{{
  record.location_id }}/forecast_override` - kind `delete`; body type `json`; path fields
  `location_id`; required record fields `location_id`, `start_date`, `report_type`; accepted fields
  `end_date`, `location_id`, `report_type`, `start_date`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Sync Daily Projected Forecast Override in the
  configured 7shifts account.
- `delete_location`: DELETE `/v2/company/{{ config.company_id }}/locations/{{ record.location_id }}`
  - kind `delete`; body type `none`; path fields `location_id`; required record fields
  `location_id`; accepted fields `location_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete Location in the configured 7shifts account.
- `delete_location_assignment`: DELETE `/v2/company/{{ config.company_id }}/users/{{ record.user_id
  }}/location_assignments/{{ record.location_id }}` - kind `delete`; body type `none`; path fields
  `user_id`, `location_id`; required record fields `location_id`, `user_id`; accepted fields
  `location_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Location Assignment in the configured 7shifts account.
- `delete_log_book_category`: DELETE `/v2/company/{{ config.company_id }}/log_book_categories/{{
  record.id }}` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Log Book Category in the configured 7shifts account.
- `delete_log_book_comment`: DELETE `/v2/company/{{ config.company_id }}/log_book_comments/{{
  record.id }}` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Log Book Comment in the configured 7shifts account.
- `delete_log_book_post`: DELETE `/v2/company/{{ config.company_id }}/log_book_posts/{{ record.id
  }}` - kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted
  fields `id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete Log Book Post in the configured 7shifts account.
- `delete_role`: DELETE `/v2/company/{{ config.company_id }}/roles/{{ record.role_id }}` - kind
  `delete`; body type `none`; path fields `role_id`; required record fields `role_id`; accepted
  fields `role_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete Role in the configured 7shifts account.
- `delete_role_assignment`: DELETE `/v2/company/{{ config.company_id }}/users/{{ record.user_id
  }}/role_assignments/{{ record.role_id }}` - kind `delete`; body type `none`; path fields
  `user_id`, `role_id`; required record fields `role_id`, `user_id`; accepted fields `role_id`,
  `user_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Delete Role Assignment in the configured 7shifts account.
- `delete_shift`: DELETE `/v2/company/{{ config.company_id }}/shifts/{{ record.shift_id }}` - kind
  `delete`; body type `none`; path fields `shift_id`; required record fields `shift_id`; accepted
  fields `shift_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Shift in the configured 7shifts account.
- `delete_task_list_template`: DELETE `/v2/company/{{ config.company_id }}/task_list_templates/{{
  record.uuid }}` - kind `delete`; body type `none`; path fields `uuid`; required record fields
  `uuid`; accepted fields `uuid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete Task List Template in the configured 7shifts account.
- `delete_task_tags`: DELETE `/v2/company/{{ config.company_id }}/task_tags` - kind `delete`; body
  type `json`; required record fields `company_id`, `uuids`; accepted fields `company_id`, `uuids`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete Task
  Tags in the configured 7shifts account.
- `delete_time_off`: DELETE `/v2/time_off/{{ record.time_off_id }}` - kind `delete`; body type
  `none`; path fields `time_off_id`; required record fields `time_off_id`; accepted fields
  `time_off_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete Time Off in the configured 7shifts account.
- `delete_time_punch_by_id`: DELETE `/v2/company/{{ config.company_id }}/time_punches/{{
  record.time_punch_id }}` - kind `delete`; body type `none`; path fields `time_punch_id`; required
  record fields `time_punch_id`; accepted fields `time_punch_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete Time Punch in the configured 7shifts
  account.
- `edit_availability`: PUT `/v2/company/{{ config.company_id }}/availabilities/{{
  record.availability_id }}` - kind `update`; body type `json`; path fields `availability_id`;
  required record fields `availability_id`; accepted fields `availability_id`, `fri`,
  `fri_comments`, `fri_from`, `fri_reason`, `fri_to`, `mon`, `mon_comments`, `mon_from`,
  `mon_reason`, `mon_to`, `repeat`, `sat`, `sat_comments`, `sat_from`, `sat_reason`, `sat_to`,
  `sun`, and 21 more; risk: Update Availability in the configured 7shifts account.
- `edit_availability_reason`: PUT `/v2/company/{{ config.company_id }}/availability_reasons/{{
  record.availability_reason_id }}` - kind `update`; body type `json`; path fields
  `availability_reason_id`; required record fields `availability_reason_id`, `reason`; accepted
  fields `availability_reason_id`, `comments_required`, `reason`; risk: Update Availability Reason
  in the configured 7shifts account.
- `edit_company_webhook`: PUT `/v2/company/{{ config.company_id }}/webhooks/{{ record.webhook_id }}`
  - kind `update`; body type `json`; path fields `webhook_id`; required record fields `webhook_id`,
  `url`; accepted fields `url`, `webhook_id`; risk: Update Webhook in the configured 7shifts
  account.
- `edit_event`: PATCH `/v2/company/{{ config.company_id }}/events/{{ record.event_id }}` - kind
  `update`; body type `json`; path fields `event_id`; required record fields `event_id`,
  `location_ids`, `start_date`, `start_time`, `end_time`, `end_date`, `title`, `is_multi_day`;
  accepted fields `color`, `description`, `end_date`, `end_time`, `event_id`, `is_multi_day`,
  `location_ids`, `recurrence`, `recurrence_target`, `start_date`, `start_time`, `title`; risk:
  Update Event in the configured 7shifts account.
- `edit_task_list_template`: PUT `/v2/company/{{ config.company_id }}/task_list_templates/{{
  record.uuid }}` - kind `update`; body type `json`; path fields `uuid`; required record fields
  `uuid`; accepted fields `assignments`, `description`, `due`, `recurrence`, `status`,
  `task_templates`, `time_frame`, `title`, `uuid`; risk: Update Task List Template in the configured
  7shifts account.
- `edit_time_off`: PATCH `/v2/time_off/{{ record.time_off_id }}` - kind `update`; body type `json`;
  path fields `time_off_id`; required record fields `time_off_id`; accepted fields `category`,
  `comments`, `from_date`, `hours`, `partial`, `partial_from`, `partial_to`, `status`,
  `status_action_message`, `time_off_id`, `to_date`, `user_id`; risk: Update Time Off in the
  configured 7shifts account.
- `post_shift`: POST `/v2/company/{{ config.company_id }}/shifts` - kind `create`; body type `json`;
  required record fields `location_id`, `start`, `end`; accepted fields `breaks`,
  `business_decline`, `close`, `department_id`, `draft`, `end`, `job_network_wage`, `late_minutes`,
  `list_in_job_network`, `location_id`, `notes`, `notified`, `open`, `open_offer_type`, `role_id`,
  `start`, `station_id`, `status`, and 3 more; risk: Create Shift in the configured 7shifts account.
- `post_time_punch`: POST `/v2/company/{{ config.company_id }}/time_punches` - kind `create`; body
  type `json`; required record fields `location_id`, `user_id`, `clocked_in`; accepted fields
  `breaks`, `clocked_in`, `clocked_out`, `department_id`, `location_id`, `notes`, `role_id`, `tips`,
  `user_id`; risk: Create Time Punch in the configured 7shifts account.
- `post_user`: POST `/v2/company/{{ config.company_id }}/users` - kind `create`; body type `json`;
  required record fields `first_name`, `last_name`, `location_ids`, `department_ids`, `type`;
  accepted fields `address`, `appear_as_employee`, `birth_date`, `city`, `department_ids`, `email`,
  `employee_id`, `first_name`, `hire_date`, `home_number`, `invite_user`, `language`, `last_name`,
  `location_ids`, `max_weekly_hours`, `mobile_number`, `notes`, `permissions_template_id`, and 12
  more; risk: Create User in the configured 7shifts account.
- `put_time_punch`: PUT `/v2/company/{{ config.company_id }}/time_punches/{{ record.time_punch_id
  }}` - kind `update`; body type `json`; path fields `time_punch_id`; required record fields
  `time_punch_id`; accepted fields `breaks`, `clocked_in`, `clocked_out`, `department_id`, `notes`,
  `role_id`, `time_punch_id`, `tips`; risk: Update Time Punch in the configured 7shifts account.
- `put_user`: PUT `/v2/company/{{ config.company_id }}/users/{{ record.identifier }}` - kind
  `update`; body type `json`; path fields `identifier`; required record fields `identifier`;
  accepted fields `active`, `address`, `appear_as_employee`, `birth_date`, `city`, `email`,
  `employee_id`, `first_name`, `hire_date`, `home_number`, `identifier`, `invite`, `language`,
  `last_name`, `max_weekly_hours`, `mobile_number`, `notes`, `permissions_template_id`, and 13 more;
  risk: Update User in the configured 7shifts account.
- `save_time_off_settings`: POST `/v2/time_off_settings/{{ config.company_id }}` - kind `create`;
  body type `json`; accepted fields `paid_time_off`, `sick_time_off`, `time_off_request_comment`,
  `time_off_request_notice`; risk: Create Time Off Settings in the configured 7shifts account.
- `save_tip_pool_manual_entry`: PUT `/v2/company/{{ config.company_id }}/tip_pool/{{
  record.tip_pool_settings_uuid }}/manual_entry` - kind `update`; body type `json`; path fields
  `tip_pool_settings_uuid`; required record fields `tip_pool_settings_uuid`, `data`; accepted fields
  `data`, `object`, `tip_pool_settings_uuid`; risk: Update Tip Pool Manual Entries in the configured
  7shifts account.
- `sync_overridden_projected_sales_interval`: DELETE `/v2/company/{{ config.company_id
  }}/locations/{{ record.location_id }}/forecast_override_interval` - kind `delete`; body type
  `json`; path fields `location_id`; required record fields `location_id`, `start`, `end`; accepted
  fields `end`, `location_id`, `report_type`, `start`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete Sales Forecast Override Interval in the configured
  7shifts account.
- `update_availability_status`: PUT `/v2/company/{{ config.company_id }}/availabilities/{{
  record.availability_id }}/status` - kind `update`; body type `json`; path fields
  `availability_id`; required record fields `availability_id`, `status`; accepted fields
  `availability_id`, `message`, `status`; risk: Update Availability Status in the configured 7shifts
  account.
- `update_company`: PATCH `/v2/companies/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `country`, `id`, `name`, `photo`, `pos`;
  risk: Update Company in the configured 7shifts account.
- `update_complete_receipt`: PUT `/v2/company/{{ config.company_id }}/receipts/{{ record.receipt_id
  }}` - kind `update`; body type `json`; path fields `receipt_id`; required record fields
  `receipt_id`, `net_total`; accepted fields `dining_option`, `external_user_id`, `gross_total`,
  `net_total`, `order_type`, `receipt_date`, `receipt_id`, `receipt_lines`, `revenue_center`,
  `status`, `tip_details`, `tips`, `total_receipt_discounts`; risk: Update Receipt in the configured
  7shifts account.
- `update_department`: PUT `/v2/company/{{ config.company_id }}/departments/{{ record.department_id
  }}` - kind `update`; body type `json`; path fields `department_id`; required record fields
  `department_id`, `name`, `default`; accepted fields `default`, `department_id`, `name`; risk:
  Update Department in the configured 7shifts account.
- `update_employment_record`: PUT `/v2/company/{{ config.company_id }}/employment_record/{{
  record.uuid }}` - kind `update`; body type `json`; path fields `uuid`; required record fields
  `uuid`; accepted fields `business_entity_uuid`, `classification`, `hire_date`, `termination_date`,
  `uuid`; risk: Update Employment Record in the configured 7shifts account.
- `update_external_user_mappings`: PUT `/v2/company/{{ config.company_id
  }}/external_user_mappings/{{ record.identifier }}` - kind `update`; body type `json`; path fields
  `identifier`; required record fields `identifier`; accepted fields `external_user_id`,
  `identifier`, `user_id`; risk: Update External User Mappings in the configured 7shifts account.
- `update_location`: PUT `/v2/company/{{ config.company_id }}/locations/{{ record.location_id }}` -
  kind `update`; body type `json`; path fields `location_id`; required record fields `location_id`;
  accepted fields `auto_send_log_book_time`, `city`, `country`, `department_based_budget`,
  `formatted_address`, `fri_hours_close`, `fri_hours_open`, `fri_is_closed`, `hash`, `holiday_pay`,
  `lat`, `lng`, `location_id`, `message`, `mon_hours_close`, `mon_hours_open`, `mon_is_closed`,
  `name`, and 19 more; risk: Update Location in the configured 7shifts account.
- `update_log_book_category`: PATCH `/v2/company/{{ config.company_id }}/log_book_categories/{{
  record.id }}` - kind `update`; body type `json`; path fields `id`; required record fields `id`;
  accepted fields `col`, `field_type`, `id`, `name`, `notify`, `required`, `sort`; risk: Update Log
  Book Category in the configured 7shifts account.
- `update_role`: PUT `/v2/company/{{ config.company_id }}/roles/{{ record.role_id }}` - kind
  `update`; body type `json`; path fields `role_id`; required record fields `role_id`; accepted
  fields `color`, `department_id`, `is_tipped_role`, `job_code`, `name`, `role_id`, `sort`,
  `stations`; risk: Update Role in the configured 7shifts account.
- `update_role_assignment`: PUT `/v2/company/{{ config.company_id }}/users/{{ record.user_id
  }}/role_assignments/{{ record.role_id }}` - kind `update`; body type `json`; path fields
  `user_id`, `role_id`; required record fields `role_id`, `user_id`; accepted fields `primary`,
  `role_id`, `skill_level`, `sort`, `user_id`; risk: Update Role Assignment in the configured
  7shifts account.
- `update_shift`: PUT `/v2/company/{{ config.company_id }}/shifts/{{ record.shift_id }}` - kind
  `update`; body type `json`; path fields `shift_id`; required record fields `shift_id`; accepted
  fields `breaks`, `business_decline`, `close`, `custom_flag_id`, `department_id`, `draft`, `end`,
  `late_minutes`, `location_id`, `notes`, `open`, `open_offer_type`, `role_id`, `shift_id`, `start`,
  `station_id`, `status`, `unassigned`, and 2 more; risk: Update Shift in the configured 7shifts
  account.
- `upsert_bulk_employment_records`: POST `/v2/company/{{ config.company_id
  }}/bulk_employment_records` - kind `upsert`; body type `json`; required record fields `records`;
  accepted fields `records`; risk: Create Many Employment Records in the configured 7shifts account.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 65 stream-backed endpoint group(s), 76 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=7, non_data_endpoint=2, out_of_scope=3, requires_elevated_scope=1.
