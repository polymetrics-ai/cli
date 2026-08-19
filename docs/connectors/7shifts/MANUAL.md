# pm connectors inspect 7shifts

```text
NAME
  pm connectors inspect 7shifts - 7shifts connector manual

SYNOPSIS
  pm connectors inspect 7shifts
  pm connectors inspect 7shifts --json
  pm credentials add <name> --connector 7shifts [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads the documented 7shifts v2 REST API surface and executes declarative single-request reverse-ETL actions for supported mutations.

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
  availability_id
  base_url
  company_id (required)
  date
  date_end
  date_start
  department_id
  employee_id
  end_date
  event_id
  from
  id
  identifier
  list_id
  location_id
  log_book_ids
  payroll_period_id
  punches
  receipt_id
  role_id
  shift_id
  start_date
  time_off_id
  time_punch_id
  tip_pool_settings_uuid
  to
  user_id
  uuid
  access_token (secret) (required)

ETL STREAMS
  companies:
    primary key: id
    cursor: modified
    fields: active(boolean), country_code(string), created(string), currency(string), id(integer), modified(string), name(string), timezone(string)
  locations:
    primary key: id
    cursor: modified
    fields: address(string), city(string), company_id(integer), country(string), created(string), id(integer), modified(string), name(string), state(string), timezone(string)
  departments:
    primary key: id
    cursor: modified
    fields: company_id(integer), created(string), id(integer), location_id(integer), modified(string), name(string)
  roles:
    primary key: id
    cursor: modified
    fields: color(string), company_id(integer), created(string), department_id(integer), id(integer), location_id(integer), modified(string), name(string)
  users:
    primary key: id
    cursor: modified
    fields: active(boolean), company_id(integer), created(string), email(string), first_name(string), hire_date(string), id(integer), last_name(string), modified(string), type(string)
  shifts:
    primary key: id
    cursor: modified
    fields: company_id(integer), created(string), deleted(boolean), department_id(integer), end(string), id(integer), location_id(integer), modified(string), open(boolean), role_id(integer), start(string), user_id(integer)
  time_punches:
    primary key: id
    cursor: modified
    fields: approved(boolean), clocked_in(string), clocked_out(string), company_id(integer), created(string), department_id(integer), id(integer), location_id(integer), modified(string), role_id(integer), user_id(integer)
  fetch_tip_pool_manual_entry:
    primary key: uuid
    fields: created(string), day_part_uuid(string), entry_date(string), modified(string), tip_amount(number), tip_pool_settings_uuid(string), uuid(string)
  find_by_id:
    primary key: id
    fields: amount_of_hours(number), category(string), comments(string), company_id(integer), created(string), from_date(string), hours(array), id(integer), partial(boolean), partial_from(string), partial_to(string), status(integer), status_action_message(string), status_action_user_id(integer), to_date(string), user_id(integer)
  get_assignments:
    primary key: stream_key
    fields: departments(array), locations(array), roles(array), stream_key(string)
  get_availability_by_id:
    primary key: id
    fields: company_id(integer), created(string), fri(integer), fri_comments(string), fri_from(string), fri_reason(string), fri_to(string), id(integer), modified(string), mon(integer), mon_comments(string), mon_from(string), mon_reason(string), mon_to(string), old_approved_data(object), repeat(boolean), sat(integer), sat_comments(string), sat_from(string), sat_reason(string), sat_to(string), status(integer), status_action_message(string), status_action_user_id(integer), sun(integer), sun_comments(string), sun_from(string), sun_reason(string), sun_to(string), thu(integer), thu_comments(string), thu_from(string), thu_reason(string), thu_to(string), tue(integer), tue_comments(string), tue_from(string), tue_reason(string), tue_to(string), user_id(integer), wed(integer), wed_comments(string), wed_from(string), wed_reason(string), wed_to(string), week(string), week_to(string)
  get_daily_sales_and_labor:
    primary key: id
    fields: actual_items(integer), actual_labor_cost(integer), actual_labor_minutes(integer), actual_ot_minutes(integer), actual_sales(integer), date(string), department_id(integer), id(integer), items_per_labor_hour(number), labor_percent(number), labor_target(number), location_id(integer), projected_items(integer), projected_items_override(integer), projected_items_per_labor_hour(number), projected_labor_cost(integer), projected_labor_minutes(integer), projected_sales(integer), projected_sales_override(number), projected_sales_per_labor_hour(number), sales_per_labor_hour(number)
  get_engage_overview_by_location_id:
    primary key: stream_key
    fields: employee_stats(object), engagement_scores(object), location_stats(object), shift_feedback(object), stream_key(string), tenure(object)
  get_event:
    primary key: id
    fields: all_day(boolean), color(string), description(string), end(array), end_date(string), end_time(string), event_type(string), id(integer), is_multi_day(boolean), location_ids(array), recurrence(string), start(array), start_date(string), start_time(string), title(string)
  get_events:
    primary key: id
    fields: all_day(boolean), color(string), description(string), end(array), end_date(string), end_time(string), event_type(string), id(integer), is_multi_day(boolean), location_ids(array), recurrence(string), start(array), start_date(string), start_time(string), title(string)
  get_hours_and_wages:
    primary key: stream_key
    fields: end(string), show_exception_costs(boolean), show_tips(boolean), start(string), stream_key(string), tip_tracking_enabled(boolean), total(object), users(array)
  get_location_by_id:
    primary key: id
    fields: auto_send_log_book_time(string), city(string), company_id(integer), country(string), created(string), deleted(boolean), department_based_budget(boolean), formatted_address(string), fri_hours_close(string), fri_hours_open(string), fri_is_closed(boolean), hash(string), holiday_pay(boolean), id(integer), lat(number), lng(number), mapping_id(string), message(string), modified(string), mon_hours_close(string), mon_hours_open(string), mon_is_closed(boolean), name(string), place_id(string), sat_hours_close(string), sat_hours_open(string), sat_is_closed(boolean), shift_feedback(boolean), state(string), sun_hours_close(string), sun_hours_open(string), sun_is_closed(boolean), thu_hours_close(string), thu_hours_open(string), thu_is_closed(boolean), timezone(string), timezone_updated(boolean), tue_hours_close(string), tue_hours_open(string), tue_is_closed(boolean), wed_hours_close(string), wed_hours_open(string), wed_is_closed(boolean)
  get_receipt:
    primary key: uuid
    fields: company_id(integer), created_date(string), external_user_id(string), gross_total(integer), location_id(integer), modified_date(string), net_total(integer), pos_id(integer), receipt_close_date(string), receipt_date(string), receipt_id(string), receipt_lines(array), revenue_center(string), status(string), tip_details(array), tips(integer), total_item_discounts(integer), total_receipt_discounts(integer), uuid(string)
  get_role_assignments:
    primary key: location_id
    fields: department_id(integer), is_primary(boolean), location_id(integer), name(string), role_id(integer), skill_level(integer), sort(integer)
  get_task_list:
    primary key: id
    fields: assignments(array), description(string), due(string), due_date(object), id(integer), recurrence(object), start(string), task_list_template_uuid(string), tasks(array), time_frame(object), title(string)
  get_task_list_daily_summary:
    primary key: date
    fields: date(string), has_recent_task_completed(boolean), report_time(string), task_lists(array), total_completed_percentage(integer), total_in_progress_percentage(integer), total_incomplete_percentage(integer)
  get_task_list_template:
    primary key: id
    fields: activated_at(string), assignments(array), company_id(integer), created(string), description(string), due(string), id(integer), recurrence(string), status(integer), task_templates(array), time_frame(object), title(string), uuid(string)
  get_task_list_templates:
    primary key: id
    fields: activated_at(string), assignments(array), company_id(integer), created(string), description(string), due(string), id(integer), recurrence(string), status(integer), task_templates(array), time_frame(object), title(string), uuid(string)
  get_task_lists:
    primary key: id
    fields: assignments(array), description(string), due(string), due_date(object), id(integer), recurrence(object), start(string), task_list_template_uuid(string), tasks(array), time_frame(object), title(string)
  get_task_management_settings:
    primary key: id
    fields: can_use_task_management(boolean), company_id(integer), employee_login(boolean), employee_reminder(boolean), enabled(boolean), has_created_list(boolean), id(integer), overdue_alert(boolean), prompted_auto_assign_ids(boolean)
  get_time_clocking_payroll_period:
    primary key: id
    fields: closed(boolean), company_id(integer), end(string), id(integer), start(string), states(array)
  get_time_clocking_payroll_periods:
    primary key: id
    fields: all_punches_approved(boolean), closed(boolean), company_id(integer), created(string), employee_punch_approvals(string), end(string), finalized(boolean), id(integer), modified(string), start(string), states(array)
  get_time_off_list:
    primary key: id
    fields: amount_of_hours(number), category(string), comments(string), company_id(integer), created(string), from_date(string), hours(array), id(integer), partial(boolean), partial_from(string), partial_to(string), status(integer), status_action_message(string), status_action_user_id(integer), to_date(string), user_id(integer)
  get_time_off_settings:
    primary key: company_id
    fields: categories(array), company_id(integer), time_off_request_comment(boolean), time_off_request_notice(number)
  get_time_punch_by_id:
    primary key: id
    fields: approved(boolean), auto_clocked_out(boolean), breaks(array), clocked_in(string), clocked_in_offline(boolean), clocked_out(string), clocked_out_offline(boolean), company_id(integer), created(string), deleted(boolean), department_id(integer), editable_punch(boolean), hourly_wage(integer), id(integer), location_id(integer), modified(string), notes(string), pos_type(string), role_id(integer), shift_id(integer), tips(integer), user_id(integer)
  get_tip_pool_settings:
    primary key: uuid
    fields: company_id(integer), contribution_type(string), created(string), day_part_uuids(array), distribution_type(string), enabled(boolean), location_id(integer), modified(string), name(string), sales_tip_percentage(number), tip_pool_cadence_settings(object), tip_pool_contributions(array), tip_pool_stakeholders(array), unmapped_contribution_filters(array), unmapped_contribution_method(string), uuid(string)
  get_total_hours:
    primary key: user_id
    fields: category(string), hours(number), user_id(number)
  get_user:
    primary key: id
    fields: active(boolean), address(string), appear_as_employee(boolean), birth_date(string), city(string), company_id(integer), created(string), email(string), employee_id(string), first_name(string), home_number(string), hourly_wage(integer), id(integer), identity_id(integer), invite_accepted(string), invite_status(string), invited(string), is_new(boolean), language(string), last_name(string), max_weekly_hours(string), meta(object), mobile_number(string), modified(string), notes(string), notify_ot_risk(boolean), permissions_template_id(integer), photo(string), postal_zip(string), preferred_first_name(string), preferred_last_name(string), pronouns(string), prov_state(string), punch_id(string), push(boolean), reactivation_status(string), skill_level(integer), sms_me_schedules(boolean), subscribe_to_updates(boolean), timezone(string), type(string), wage_type(string)
  get_user_wages:
    primary key: stream_key
    fields: current_wages(array), stream_key(string), upcoming_wages(array), wage_type(string)
  list_availabilities:
    primary key: id
    fields: company_id(integer), created(string), fri(integer), fri_comments(string), fri_from(string), fri_reason(string), fri_to(string), id(integer), modified(string), mon(integer), mon_comments(string), mon_from(string), mon_reason(string), mon_to(string), old_approved_data(object), repeat(boolean), sat(integer), sat_comments(string), sat_from(string), sat_reason(string), sat_to(string), status(integer), status_action_message(string), status_action_user_id(integer), sun(integer), sun_comments(string), sun_from(string), sun_reason(string), sun_to(string), thu(integer), thu_comments(string), thu_from(string), thu_reason(string), thu_to(string), tue(integer), tue_comments(string), tue_from(string), tue_reason(string), tue_to(string), user_id(integer), wed(integer), wed_comments(string), wed_from(string), wed_reason(string), wed_to(string), week(string), week_to(string)
  list_availability_reasons:
    primary key: id
    fields: comments_required(boolean), company_id(integer), created(string), id(integer), modified(string), reason(string), sort(integer)
  list_company_webhooks:
    primary key: id
    fields: company_id(integer), created(string), event(string), id(integer), method(string), modified(string), url(string)
  list_department_assignments:
    primary key: location_id
    fields: appear_on_schedule(boolean), department_id(integer), location_id(integer), name(string)
  list_employment_record:
    primary key: uuid
    fields: business_entity_uuid(string), classification(string), company_id(integer), hire_date(string), locked(boolean), termination_date(string), user_id(integer), uuid(string)
  list_external_user_mappings:
    primary key: id
    fields: application_name(string), company_id(integer), created(string), external_user_id(string), id(integer), location_id(integer), modified(string), user_active(boolean), user_id(integer)
  list_inactive_reasons:
    primary key: stream_key
    fields: stream_key(string), value(string)
  list_location_assignments:
    primary key: location_id
    fields: location_id(integer), name(string)
  list_log_book_categories:
    primary key: id
    fields: col(integer), company_id(integer), created(string), field_type(string), id(integer), name(string), notify(boolean), required(boolean), sort(integer), uuid(string)
  list_log_book_comments:
    primary key: id
    fields: company_id(integer), created(string), id(integer), log_book_id(integer), message(string), user_id(integer), uuid(string)
  list_log_book_posts:
    primary key: id
    fields: attachments(array), company_id(integer), created(string), date(string), id(integer), location_id(integer), log_book_category_id(integer), log_book_comment_count(integer), message(string), user_id(integer), uuid(string)
  list_sales_receipts:
    primary key: uuid
    fields: company_id(integer), created_date(string), external_user_id(string), gross_total(integer), location_id(integer), modified_date(string), net_total(integer), pos_id(integer), receipt_close_date(string), receipt_date(string), receipt_id(string), receipt_lines(array), revenue_center(string), status(string), tip_details(array), tips(integer), total_item_discounts(integer), total_receipt_discounts(integer), uuid(string)
  list_scheduled_shifts:
    primary key: id
    fields: attendance_status(string), business_decline(boolean), close(boolean), company_id(integer), department_id(integer), end(string), id(integer), location_id(integer), notes(string), open(boolean), publish_status(string), role_id(integer), start(string), station_id(integer), station_name(string), user_id(integer)
  list_shift_feedback:
    primary key: id
    fields: comments(string), department_id(integer), dismissed(boolean), end(string), id(integer), location_id(integer), notified(boolean), rating(integer), role_id(integer), shift_id(integer), start(string), station_name(string), user_id(integer)
  list_user_contacts:
    primary key: id
    fields: company_id(integer), email(string), first_name(string), home_phone(string), id(integer), last_name(string), mobile_phone(string), photo(string), pronouns(string), type(string)
  list_users_authorized_locations:
    primary key: id
    fields: auto_send_log_book_time(string), city(string), company_id(integer), country(string), created(string), deleted(boolean), department_based_budget(boolean), formatted_address(string), fri_hours_close(string), fri_hours_open(string), fri_is_closed(boolean), hash(string), holiday_pay(boolean), id(integer), lat(number), lng(number), mapping_id(string), message(string), modified(string), mon_hours_close(string), mon_hours_open(string), mon_is_closed(boolean), name(string), place_id(string), sat_hours_close(string), sat_hours_open(string), sat_is_closed(boolean), shift_feedback(boolean), state(string), sun_hours_close(string), sun_hours_open(string), sun_is_closed(boolean), thu_hours_close(string), thu_hours_open(string), thu_is_closed(boolean), timezone(string), timezone_updated(boolean), tue_hours_close(string), tue_hours_open(string), tue_is_closed(boolean), wed_hours_close(string), wed_hours_open(string), wed_is_closed(boolean)
  retrieve_company_labor_settings:
    primary key: company_id
    fields: auto_break_enabled(boolean), auto_break_hours(number), auto_break_hours_2(number), auto_break_minutes(integer), auto_break_minutes_2(integer), company_id(integer), consecutive_days_multiplier(number), consecutive_days_threshold(integer), custom_min_wages(array), daily_overtime_multiplier(number), daily_overtime_threshold(integer), exception_cost_enabled(boolean), is_custom(boolean), is_custom_min_wage(boolean), jurisdiction(string), ot_alerts_buffer_minutes(integer), ot_alerts_enabled(boolean), overtime_by_location_enabled(boolean), overtime_enabled(boolean), premium_overtime_multiplier(number), premium_overtime_threshold(integer), regular_rate_of_pay_enabled(boolean), split_hours_on_holidays(boolean), tipped_minimum_wage_enabled(boolean), use_jurisdiction_minimum_wage_for_ot(boolean), wage_based_roles_enabled(boolean), weekly_overtime_multiplier(number), weekly_overtime_threshold(integer)
  retrieve_daily_stats:
    primary key: stream_key
    fields: intervals(array), stream_key(string), summary(object)
  retrieve_day_part_settings:
    primary key: uuid
    fields: created(string), location_id(integer), modified(string), name(string), segments(array), uuid(string)
  retrieve_department:
    primary key: id
    fields: company_id(number), created(string), default(boolean), deleted(string), id(number), location_id(number), modified(string), name(string)
  retrieve_external_user_mapping:
    primary key: id
    fields: application_name(string), company_id(integer), created(string), external_user_id(string), id(integer), location_id(integer), modified(string), user_active(boolean), user_id(integer)
  retrieve_log_book_comment:
    primary key: id
    fields: company_id(integer), created(string), id(integer), log_book_id(integer), message(string), user_id(integer), uuid(string)
  retrieve_log_book_post:
    primary key: id
    fields: attachments(array), company_id(integer), created(string), date(string), id(integer), location_id(integer), log_book_category_id(integer), log_book_comment_count(integer), message(string), user_id(integer), uuid(string)
  retrieve_receipts_summary:
    primary key: date
    fields: closed(object), date(string), deleted(object), open(object), voided(object)
  retrieve_role:
    primary key: id
    fields: color(string), color_code(string), color_dark(string), company_id(integer), created(string), department_id(integer), id(integer), is_tipped_role(boolean), job_code(string), location_id(integer), modified(string), name(string), num_stations(integer), sort(integer), stations(array)
  retrieve_shift:
    primary key: id
    fields: attendance_status(string), breaks(array), business_decline(boolean), close(boolean), created(string), custom_flag(object), deleted(boolean), department_id(integer), draft(boolean), end(string), hourly_wage(number), id(integer), late_minuets(integer), location_id(integer), modified(string), notes(string), notified(boolean), open(boolean), open_offer_type(string), publish_status(string), role_id(integer), soft_deleted(string), start(string), station(integer), station_id(integer), station_name(string), unassigned(boolean), unassigned_skill_level(integer), user_id(integer)
  retrieve_tip_pool_detailed_report:
    primary key: location_id
    fields: location_id(integer), location_name(string), report_rows(array)
  retrieve_tip_pool_summary_report:
    primary key: tip_pool_uuid
    fields: assigned_tips(array), tip_pool_name(string), tip_pool_uuid(string), total(object), unassigned_tips(object)
  retrieve_user_contact:
    primary key: id
    fields: company_id(integer), email(string), first_name(string), home_phone(string), id(integer), last_name(string), mobile_phone(string), photo(string), pronouns(string), type(string)
  view_company:
    primary key: id
    fields: converted(string), country(string), created(string), days_to_expire(number), expires(string), id(number), meta(object), modified(string), name(string), photo(string), plan_id(string), pos(string), start_week_on(number), status(string)
  who_am_i:
    primary key: identity_id
    fields: identity_id(integer), users(array)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  approve_time_off:
    endpoint: POST /v2/time_off/{{ record.time_off_id }}/approve
    required fields: time_off_id
    risk: Approve Time Off Request in the configured 7shifts account.
  clear_task:
    endpoint: POST /v2/company/{{ config.company_id }}/task_lists/{{ record.list_id }}/tasks/{{ record.task_id }}/clear
    required fields: list_id, task_id, user_id
    risk: Clear Task in the configured 7shifts account.
  complete_task:
    endpoint: POST /v2/company/{{ config.company_id }}/task_lists/{{ record.list_id }}/tasks/{{ record.task_id }}/complete
    required fields: list_id, task_id, user_id
    risk: Complete Task in the configured 7shifts account.
  create_availability:
    endpoint: POST /v2/company/{{ config.company_id }}/availabilities
    required fields: user_id, repeat, mon, mon_from, mon_to, mon_comments, mon_reason, tue, tue_from, tue_to, tue_comments, tue_reason, wed, wed_from, wed_to, wed_comments, wed_reason, thu, thu_from, thu_to, thu_comments, thu_reason, fri, fri_from, fri_to, fri_comments, fri_reason, sat, sat_from, sat_to, sat_comments, sat_reason, sun, sun_from, sun_to, sun_comments, sun_reason
    risk: Create Availability in the configured 7shifts account.
  create_availability_reason:
    endpoint: POST /v2/company/{{ config.company_id }}/availability_reasons
    required fields: reason
    risk: Create Availability Reason in the configured 7shifts account.
  create_bulk_forecast_overrides:
    endpoint: POST /v2/company/{{ config.company_id }}/location/{{ record.location_id }}/forecast_overrides
    required fields: location_id, data
    risk: Create Daily Projected Forecast Overrides in the configured 7shifts account.
  create_complete_receipt:
    endpoint: POST /v2/company/{{ config.company_id }}/receipts
    required fields: location_id, receipt_date, receipt_lines, tip_details, net_total, status
    risk: Create Receipt in the configured 7shifts account.
  create_department:
    endpoint: POST /v2/company/{{ config.company_id }}/departments
    required fields: location_id, name, default
    risk: Create Department in the configured 7shifts account.
  create_department_assignment:
    endpoint: POST /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/department_assignments
    required fields: user_id, department_id
    risk: Create Department Assignment in the configured 7shifts account.
  create_employment_record:
    endpoint: POST /v2/company/{{ config.company_id }}/employment_records
    required fields: user_id
    risk: Create Employment Record in the configured 7shifts account.
  create_event:
    endpoint: POST /v2/company/{{ config.company_id }}/events
    required fields: location_ids, start_date, start_time, end_time, end_date, title, is_multi_day
    risk: Create Event in the configured 7shifts account.
  create_external_user_mappings:
    endpoint: POST /v2/company/{{ config.company_id }}/external_user_mappings
    required fields: user_id, external_user_id
    risk: Create External User Mapping in the configured 7shifts account.
  create_forecast_override:
    endpoint: POST /v2/company/{{ config.company_id }}/location/{{ record.location_id }}/forecast_override
    required fields: location_id, date, value, report_type
    risk: Create Daily Projected Forecast Override in the configured 7shifts account.
  create_location:
    endpoint: POST /v2/company/{{ config.company_id }}/locations
    required fields: name, country
    risk: Create Location in the configured 7shifts account.
  create_location_assignment:
    endpoint: POST /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/location_assignments
    required fields: user_id, location_id
    risk: Create Location Assignments in the configured 7shifts account.
  create_log_book_category:
    endpoint: POST /v2/company/{{ config.company_id }}/log_book_categories
    required fields: name
    risk: Create Log Book Category in the configured 7shifts account.
  create_log_book_comment:
    endpoint: POST /v2/company/{{ config.company_id }}/log_book_comments
    required fields: log_book_id, message
    risk: Create Log Book Comment in the configured 7shifts account.
  create_log_book_post:
    endpoint: POST /v2/company/{{ config.company_id }}/log_book_posts
    required fields: location_id, log_book_category_id, date, message
    risk: Create Log Book Post in the configured 7shifts account.
  create_projected_sales_interval_override:
    endpoint: POST /v2/company/{{ config.company_id }}/locations/{{ record.location_id }}/forecast_override_interval
    required fields: location_id, start, end, value
    risk: Create Sales Forecast Override Interval in the configured 7shifts account.
  create_role:
    endpoint: POST /v2/company/{{ config.company_id }}/roles
    required fields: name, color, location_id, department_id
    risk: Create Role in the configured 7shifts account.
  create_role_assignment:
    endpoint: POST /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/role_assignments
    required fields: user_id, role_id
    risk: Create Role Assignment in the configured 7shifts account.
  create_task_list_template:
    endpoint: POST /v2/company/{{ config.company_id }}/task_list_templates
    required fields: title, recurrence, assignments
    risk: Create Task List Template in the configured 7shifts account.
  create_task_tags:
    endpoint: POST /v2/company/{{ config.company_id }}/task_tags
    required fields: company_id, tags
    risk: Create Task Tags in the configured 7shifts account.
  create_time_off:
    endpoint: POST /v2/time_off
    required fields: user_id, company_id, from_date, to_date, partial, status, category
    risk: Create Time Off in the configured 7shifts account.
  create_user_mappings_bulk:
    endpoint: POST /v2/company/{{ config.company_id }}/external_user_mappings_bulk
    required fields: data
    risk: Create User External Mappings in the configured 7shifts account.
  create_user_wages:
    endpoint: POST /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/wages
    required fields: user_id, effective_date, wage_type, wage_cents
    risk: Create User Wage in the configured 7shifts account.
  create_webhook:
    endpoint: POST /v2/company/{{ config.company_id }}/webhooks
    required fields: url, method, event
    risk: Create Webhook in the configured 7shifts account.
  deactivate_user:
    endpoint: DELETE /v2/company/{{ config.company_id }}/users/{{ record.identifier }}
    required fields: identifier, inactive_reason
    risk: Deactivate User in the configured 7shifts account.
  decline_time_off:
    endpoint: POST /v2/time_off/{{ record.time_off_id }}/decline
    required fields: time_off_id
    risk: Decline Time Off Request in the configured 7shifts account.
  delete_availability:
    endpoint: DELETE /v2/company/{{ config.company_id }}/availabilities/{{ record.availability_id }}
    required fields: availability_id
    risk: Delete Availability in the configured 7shifts account.
  delete_availability_reason:
    endpoint: DELETE /v2/company/{{ config.company_id }}/availability_reasons/{{ record.availability_reason_id }}
    required fields: availability_reason_id
    risk: Delete Availability Reason in the configured 7shifts account.
  delete_company_webhook:
    endpoint: DELETE /v2/company/{{ config.company_id }}/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: Delete Webhook in the configured 7shifts account.
  delete_department:
    endpoint: DELETE /v2/company/{{ config.company_id }}/departments/{{ record.department_id }}
    required fields: department_id
    risk: Delete Department in the configured 7shifts account.
  delete_department_assignment:
    endpoint: DELETE /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/department_assignments/{{ record.department_id }}
    required fields: user_id, department_id
    risk: Delete Department Assignment in the configured 7shifts account.
  delete_employment_record:
    endpoint: DELETE /v2/company/{{ config.company_id }}/employment_record/{{ record.uuid }}
    required fields: uuid
    risk: Delete Employment Record in the configured 7shifts account.
  delete_event:
    endpoint: DELETE /v2/company/{{ config.company_id }}/events/{{ record.event_id }}
    required fields: event_id
    risk: Delete Event in the configured 7shifts account.
  delete_external_user_mappings:
    endpoint: DELETE /v2/company/{{ config.company_id }}/external_user_mappings/{{ record.identifier }}
    required fields: identifier
    risk: Delete External User Mapping in the configured 7shifts account.
  delete_forecast_override:
    endpoint: DELETE /v2/company/{{ config.company_id }}/location/{{ record.location_id }}/forecast_override
    required fields: location_id, start_date, report_type
    risk: Sync Daily Projected Forecast Override in the configured 7shifts account.
  delete_location:
    endpoint: DELETE /v2/company/{{ config.company_id }}/locations/{{ record.location_id }}
    required fields: location_id
    risk: Delete Location in the configured 7shifts account.
  delete_location_assignment:
    endpoint: DELETE /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/location_assignments/{{ record.location_id }}
    required fields: user_id, location_id
    risk: Delete Location Assignment in the configured 7shifts account.
  delete_log_book_category:
    endpoint: DELETE /v2/company/{{ config.company_id }}/log_book_categories/{{ record.id }}
    required fields: id
    risk: Delete Log Book Category in the configured 7shifts account.
  delete_log_book_comment:
    endpoint: DELETE /v2/company/{{ config.company_id }}/log_book_comments/{{ record.id }}
    required fields: id
    risk: Delete Log Book Comment in the configured 7shifts account.
  delete_log_book_post:
    endpoint: DELETE /v2/company/{{ config.company_id }}/log_book_posts/{{ record.id }}
    required fields: id
    risk: Delete Log Book Post in the configured 7shifts account.
  delete_role:
    endpoint: DELETE /v2/company/{{ config.company_id }}/roles/{{ record.role_id }}
    required fields: role_id
    risk: Delete Role in the configured 7shifts account.
  delete_role_assignment:
    endpoint: DELETE /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/role_assignments/{{ record.role_id }}
    required fields: user_id, role_id
    risk: Delete Role Assignment in the configured 7shifts account.
  delete_shift:
    endpoint: DELETE /v2/company/{{ config.company_id }}/shifts/{{ record.shift_id }}
    required fields: shift_id
    risk: Delete Shift in the configured 7shifts account.
  delete_task_list_template:
    endpoint: DELETE /v2/company/{{ config.company_id }}/task_list_templates/{{ record.uuid }}
    required fields: uuid
    risk: Delete Task List Template in the configured 7shifts account.
  delete_task_tags:
    endpoint: DELETE /v2/company/{{ config.company_id }}/task_tags
    required fields: company_id, uuids
    risk: Delete Task Tags in the configured 7shifts account.
  delete_time_off:
    endpoint: DELETE /v2/time_off/{{ record.time_off_id }}
    required fields: time_off_id
    risk: Delete Time Off in the configured 7shifts account.
  delete_time_punch_by_id:
    endpoint: DELETE /v2/company/{{ config.company_id }}/time_punches/{{ record.time_punch_id }}
    required fields: time_punch_id
    risk: Delete Time Punch in the configured 7shifts account.
  edit_availability:
    endpoint: PUT /v2/company/{{ config.company_id }}/availabilities/{{ record.availability_id }}
    required fields: availability_id
    risk: Update Availability in the configured 7shifts account.
  edit_availability_reason:
    endpoint: PUT /v2/company/{{ config.company_id }}/availability_reasons/{{ record.availability_reason_id }}
    required fields: availability_reason_id, reason
    risk: Update Availability Reason in the configured 7shifts account.
  edit_company_webhook:
    endpoint: PUT /v2/company/{{ config.company_id }}/webhooks/{{ record.webhook_id }}
    required fields: webhook_id, url
    risk: Update Webhook in the configured 7shifts account.
  edit_event:
    endpoint: PATCH /v2/company/{{ config.company_id }}/events/{{ record.event_id }}
    required fields: event_id, location_ids, start_date, start_time, end_time, end_date, title, is_multi_day
    risk: Update Event in the configured 7shifts account.
  edit_task_list_template:
    endpoint: PUT /v2/company/{{ config.company_id }}/task_list_templates/{{ record.uuid }}
    required fields: uuid
    risk: Update Task List Template in the configured 7shifts account.
  edit_time_off:
    endpoint: PATCH /v2/time_off/{{ record.time_off_id }}
    required fields: time_off_id
    risk: Update Time Off in the configured 7shifts account.
  post_shift:
    endpoint: POST /v2/company/{{ config.company_id }}/shifts
    required fields: location_id, start, end
    risk: Create Shift in the configured 7shifts account.
  post_time_punch:
    endpoint: POST /v2/company/{{ config.company_id }}/time_punches
    required fields: location_id, user_id, clocked_in
    risk: Create Time Punch in the configured 7shifts account.
  post_user:
    endpoint: POST /v2/company/{{ config.company_id }}/users
    required fields: first_name, last_name, location_ids, department_ids, type
    risk: Create User in the configured 7shifts account.
  put_time_punch:
    endpoint: PUT /v2/company/{{ config.company_id }}/time_punches/{{ record.time_punch_id }}
    required fields: time_punch_id
    risk: Update Time Punch in the configured 7shifts account.
  put_user:
    endpoint: PUT /v2/company/{{ config.company_id }}/users/{{ record.identifier }}
    required fields: identifier
    risk: Update User in the configured 7shifts account.
  save_time_off_settings:
    endpoint: POST /v2/time_off_settings/{{ config.company_id }}
    risk: Create Time Off Settings in the configured 7shifts account.
  save_tip_pool_manual_entry:
    endpoint: PUT /v2/company/{{ config.company_id }}/tip_pool/{{ record.tip_pool_settings_uuid }}/manual_entry
    required fields: tip_pool_settings_uuid, data
    risk: Update Tip Pool Manual Entries in the configured 7shifts account.
  sync_overridden_projected_sales_interval:
    endpoint: DELETE /v2/company/{{ config.company_id }}/locations/{{ record.location_id }}/forecast_override_interval
    required fields: location_id, start, end
    risk: Delete Sales Forecast Override Interval in the configured 7shifts account.
  update_availability_status:
    endpoint: PUT /v2/company/{{ config.company_id }}/availabilities/{{ record.availability_id }}/status
    required fields: availability_id, status
    risk: Update Availability Status in the configured 7shifts account.
  update_company:
    endpoint: PATCH /v2/companies/{{ record.id }}
    required fields: id
    risk: Update Company in the configured 7shifts account.
  update_complete_receipt:
    endpoint: PUT /v2/company/{{ config.company_id }}/receipts/{{ record.receipt_id }}
    required fields: receipt_id, net_total
    risk: Update Receipt in the configured 7shifts account.
  update_department:
    endpoint: PUT /v2/company/{{ config.company_id }}/departments/{{ record.department_id }}
    required fields: department_id, name, default
    risk: Update Department in the configured 7shifts account.
  update_employment_record:
    endpoint: PUT /v2/company/{{ config.company_id }}/employment_record/{{ record.uuid }}
    required fields: uuid
    risk: Update Employment Record in the configured 7shifts account.
  update_external_user_mappings:
    endpoint: PUT /v2/company/{{ config.company_id }}/external_user_mappings/{{ record.identifier }}
    required fields: identifier
    risk: Update External User Mappings in the configured 7shifts account.
  update_location:
    endpoint: PUT /v2/company/{{ config.company_id }}/locations/{{ record.location_id }}
    required fields: location_id
    risk: Update Location in the configured 7shifts account.
  update_log_book_category:
    endpoint: PATCH /v2/company/{{ config.company_id }}/log_book_categories/{{ record.id }}
    required fields: id
    risk: Update Log Book Category in the configured 7shifts account.
  update_role:
    endpoint: PUT /v2/company/{{ config.company_id }}/roles/{{ record.role_id }}
    required fields: role_id
    risk: Update Role in the configured 7shifts account.
  update_role_assignment:
    endpoint: PUT /v2/company/{{ config.company_id }}/users/{{ record.user_id }}/role_assignments/{{ record.role_id }}
    required fields: user_id, role_id
    risk: Update Role Assignment in the configured 7shifts account.
  update_shift:
    endpoint: PUT /v2/company/{{ config.company_id }}/shifts/{{ record.shift_id }}
    required fields: shift_id
    risk: Update Shift in the configured 7shifts account.
  upsert_bulk_employment_records:
    endpoint: POST /v2/company/{{ config.company_id }}/bulk_employment_records
    required fields: records
    risk: Create Many Employment Records in the configured 7shifts account.

SECURITY
  read risk: external 7shifts API reads of scheduling, roster, labor, sales, task, time off, webhook, and settings data
  write risk: creates, updates, deletes, approves, declines, or otherwise mutates configured 7shifts account resources through single-request REST actions
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect 7shifts

  # Inspect as structured JSON
  pm connectors inspect 7shifts --json

AGENT WORKFLOW
  - Run pm connectors inspect 7shifts before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
