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
  company_id
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
  access_token (secret)

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

COMMAND SURFACE
  Run 7shifts's declared streams and reverse-ETL actions.
  Usage: pm 7shifts <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v2 company company-id location location-id external-user-mappings identifier - Documented DELETE /v2/company/{company_id}/location/{location_id}/external_user_mappings/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.delete.v2-company-company-id-location-location-id-external-user-mappings-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v2 company company-id location location-id external-user-mappings - Documented GET /v2/company/{company_id}/location/{location_id}/external_user_mappings (not implemented) [intent=direct_read availability=not_implemented operation=7shifts.get.v2-company-company-id-location-location-id-external-user-mappings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 company company-id location location-id external-user-mappings identifier - Documented GET /v2/company/{company_id}/location/{location_id}/external_user_mappings/{identifier} (not implemented) [intent=direct_read availability=not_implemented operation=7shifts.get.v2-company-company-id-location-location-id-external-user-mappings-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 company company-id test-webhook - Documented GET /v2/company/{company_id}/test_webhook (not implemented) [intent=direct_read availability=not_implemented operation=7shifts.get.v2-company-company-id-test-webhook]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post oauth2 token - Documented POST /oauth2/token (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.oauth2-token]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 batch company company-id location location-id forecast-override - Documented POST /v2/batch/company/{company_id}/location/{location_id}/forecast_override (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-batch-company-company-id-location-location-id-forecast-override]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 batch company company-id users role-assignments - Documented POST /v2/batch/company/{company_id}/users/role_assignments (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-batch-company-company-id-users-role-assignments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 company company-id create-many-users - Documented POST /v2/company/{company_id}/create_many_users (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-company-company-id-create-many-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 company company-id location location-id external-user-mappings - Documented POST /v2/company/{company_id}/location/{location_id}/external_user_mappings (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-company-company-id-location-location-id-external-user-mappings]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 company company-id location location-id external-user-mappings-bulk - Documented POST /v2/company/{company_id}/location/{location_id}/external_user_mappings_bulk (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-company-company-id-location-location-id-external-user-mappings-bulk]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 company company-id locations location-id forecast-overrides-intervals - Documented POST /v2/company/{company_id}/locations/{location_id}/forecast_overrides_intervals (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-company-company-id-locations-location-id-forecast-overrides-intervals]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 partner-company-creation - Documented POST /v2/partner_company_creation (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.post.v2-partner-company-creation]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v2 company company-id location location-id external-user-mappings identifier - Documented PUT /v2/company/{company_id}/location/{location_id}/external_user_mappings/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=7shifts.put.v2-company-company-id-location-location-id-external-user-mappings-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    approve time off apply - Plan and execute the approve time off reverse-ETL action [intent=reverse_etl availability=implemented write=approve_time_off]; approval: requires plan, preview, approval, and execute; risk: Approve Time Off Request in the configured 7shifts account.; flags: --time_off_id (required)
    clear task apply - Plan and execute the clear task reverse-ETL action [intent=reverse_etl availability=implemented write=clear_task]; approval: requires plan, preview, approval, and execute; risk: Clear Task in the configured 7shifts account.; flags: --list_id (required), --task_id (required), --user_id (required)
    companies list - Run the companies ETL stream [intent=etl availability=implemented stream=companies]
    complete task apply - Plan and execute the complete task reverse-ETL action [intent=reverse_etl availability=implemented write=complete_task]; approval: requires plan, preview, approval, and execute; risk: Complete Task in the configured 7shifts account.; flags: --list_id (required), --task_id (required), --user_id (required)
    create availability apply - Plan and execute the create availability reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_availability]; approval: requires plan, preview, approval, and execute; risk: Create Availability in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create availability reason apply - Plan and execute the create availability reason reverse-ETL action [intent=reverse_etl availability=implemented write=create_availability_reason]; approval: requires plan, preview, approval, and execute; risk: Create Availability Reason in the configured 7shifts account.; flags: --reason (required)
    create bulk forecast overrides apply - Plan and execute the create bulk forecast overrides reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_bulk_forecast_overrides]; approval: requires plan, preview, approval, and execute; risk: Create Daily Projected Forecast Overrides in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create complete receipt apply - Plan and execute the create complete receipt reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_complete_receipt]; approval: requires plan, preview, approval, and execute; risk: Create Receipt in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create department apply - Plan and execute the create department reverse-ETL action [intent=reverse_etl availability=implemented write=create_department]; approval: requires plan, preview, approval, and execute; risk: Create Department in the configured 7shifts account.; flags: --default (required), --location_id (required), --name (required)
    create department assignment apply - Plan and execute the create department assignment reverse-ETL action [intent=reverse_etl availability=implemented write=create_department_assignment]; approval: requires plan, preview, approval, and execute; risk: Create Department Assignment in the configured 7shifts account.; flags: --department_id (required), --user_id (required)
    create employment record apply - Plan and execute the create employment record reverse-ETL action [intent=reverse_etl availability=implemented write=create_employment_record]; approval: requires plan, preview, approval, and execute; risk: Create Employment Record in the configured 7shifts account.; flags: --user_id (required)
    create event apply - Plan and execute the create event reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_event]; approval: requires plan, preview, approval, and execute; risk: Create Event in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create external user mappings apply - Plan and execute the create external user mappings reverse-ETL action [intent=reverse_etl availability=implemented write=create_external_user_mappings]; approval: requires plan, preview, approval, and execute; risk: Create External User Mapping in the configured 7shifts account.; flags: --external_user_id (required), --user_id (required)
    create forecast override apply - Plan and execute the create forecast override reverse-ETL action [intent=reverse_etl availability=implemented write=create_forecast_override]; approval: requires plan, preview, approval, and execute; risk: Create Daily Projected Forecast Override in the configured 7shifts account.; flags: --date (required), --location_id (required), --report_type (required), --value (required)
    create location apply - Plan and execute the create location reverse-ETL action [intent=reverse_etl availability=implemented write=create_location]; approval: requires plan, preview, approval, and execute; risk: Create Location in the configured 7shifts account.; flags: --country (required), --name (required)
    create location assignment apply - Plan and execute the create location assignment reverse-ETL action [intent=reverse_etl availability=implemented write=create_location_assignment]; approval: requires plan, preview, approval, and execute; risk: Create Location Assignments in the configured 7shifts account.; flags: --location_id (required), --user_id (required)
    create log book category apply - Plan and execute the create log book category reverse-ETL action [intent=reverse_etl availability=implemented write=create_log_book_category]; approval: requires plan, preview, approval, and execute; risk: Create Log Book Category in the configured 7shifts account.; flags: --name (required)
    create log book comment apply - Plan and execute the create log book comment reverse-ETL action [intent=reverse_etl availability=implemented write=create_log_book_comment]; approval: requires plan, preview, approval, and execute; risk: Create Log Book Comment in the configured 7shifts account.; flags: --log_book_id (required), --message (required)
    create log book post apply - Plan and execute the create log book post reverse-ETL action [intent=reverse_etl availability=implemented write=create_log_book_post]; approval: requires plan, preview, approval, and execute; risk: Create Log Book Post in the configured 7shifts account.; flags: --date (required), --location_id (required), --log_book_category_id (required), --message (required)
    create projected sales interval override apply - Plan and execute the create projected sales interval override reverse-ETL action [intent=reverse_etl availability=implemented write=create_projected_sales_interval_override]; approval: requires plan, preview, approval, and execute; risk: Create Sales Forecast Override Interval in the configured 7shifts account.; flags: --end (required), --location_id (required), --start (required), --value (required)
    create role apply - Plan and execute the create role reverse-ETL action [intent=reverse_etl availability=implemented write=create_role]; approval: requires plan, preview, approval, and execute; risk: Create Role in the configured 7shifts account.; flags: --color (required), --department_id (required), --location_id (required), --name (required)
    create role assignment apply - Plan and execute the create role assignment reverse-ETL action [intent=reverse_etl availability=implemented write=create_role_assignment]; approval: requires plan, preview, approval, and execute; risk: Create Role Assignment in the configured 7shifts account.; flags: --role_id (required), --user_id (required)
    create task list template apply - Plan and execute the create task list template reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_task_list_template]; approval: requires plan, preview, approval, and execute; risk: Create Task List Template in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create task tags apply - Plan and execute the create task tags reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_task_tags]; approval: requires plan, preview, approval, and execute; risk: Create Task Tags in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create time off apply - Plan and execute the create time off reverse-ETL action [intent=reverse_etl availability=implemented write=create_time_off]; approval: requires plan, preview, approval, and execute; risk: Create Time Off in the configured 7shifts account.; flags: --category (required), --company_id (required), --from_date (required), --partial (required), --status (required), --to_date (required), --user_id (required)
    create user mappings bulk apply - Plan and execute the create user mappings bulk reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_user_mappings_bulk]; approval: requires plan, preview, approval, and execute; risk: Create User External Mappings in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create user wages apply - Plan and execute the create user wages reverse-ETL action [intent=reverse_etl availability=implemented write=create_user_wages]; approval: requires plan, preview, approval, and execute; risk: Create User Wage in the configured 7shifts account.; flags: --effective_date (required), --user_id (required), --wage_cents (required), --wage_type (required)
    create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: Create Webhook in the configured 7shifts account.; flags: --event (required), --method (required), --url (required)
    deactivate user apply - Plan and execute the deactivate user reverse-ETL action [intent=reverse_etl availability=not_implemented write=deactivate_user]; approval: requires plan, preview, approval, and execute; risk: Deactivate User in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    decline time off apply - Plan and execute the decline time off reverse-ETL action [intent=reverse_etl availability=implemented write=decline_time_off]; approval: requires plan, preview, approval, and execute; risk: Decline Time Off Request in the configured 7shifts account.; flags: --time_off_id (required)
    delete availability apply - Plan and execute the delete availability reverse-ETL action [intent=reverse_etl availability=implemented write=delete_availability]; approval: requires plan, preview, approval, and execute; risk: Delete Availability in the configured 7shifts account.; flags: --availability_id (required)
    delete availability reason apply - Plan and execute the delete availability reason reverse-ETL action [intent=reverse_etl availability=implemented write=delete_availability_reason]; approval: requires plan, preview, approval, and execute; risk: Delete Availability Reason in the configured 7shifts account.; flags: --availability_reason_id (required)
    delete company webhook apply - Plan and execute the delete company webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_company_webhook]; approval: requires plan, preview, approval, and execute; risk: Delete Webhook in the configured 7shifts account.; flags: --webhook_id (required)
    delete department apply - Plan and execute the delete department reverse-ETL action [intent=reverse_etl availability=implemented write=delete_department]; approval: requires plan, preview, approval, and execute; risk: Delete Department in the configured 7shifts account.; flags: --department_id (required)
    delete department assignment apply - Plan and execute the delete department assignment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_department_assignment]; approval: requires plan, preview, approval, and execute; risk: Delete Department Assignment in the configured 7shifts account.; flags: --department_id (required), --user_id (required)
    delete employment record apply - Plan and execute the delete employment record reverse-ETL action [intent=reverse_etl availability=implemented write=delete_employment_record]; approval: requires plan, preview, approval, and execute; risk: Delete Employment Record in the configured 7shifts account.; flags: --uuid (required)
    delete event apply - Plan and execute the delete event reverse-ETL action [intent=reverse_etl availability=implemented write=delete_event]; approval: requires plan, preview, approval, and execute; risk: Delete Event in the configured 7shifts account.; flags: --event_id (required)
    delete external user mappings apply - Plan and execute the delete external user mappings reverse-ETL action [intent=reverse_etl availability=implemented write=delete_external_user_mappings]; approval: requires plan, preview, approval, and execute; risk: Delete External User Mapping in the configured 7shifts account.; flags: --identifier (required)
    delete forecast override apply - Plan and execute the delete forecast override reverse-ETL action [intent=reverse_etl availability=implemented write=delete_forecast_override]; approval: requires plan, preview, approval, and execute; risk: Sync Daily Projected Forecast Override in the configured 7shifts account.; flags: --location_id (required), --report_type (required), --start_date (required)
    delete location apply - Plan and execute the delete location reverse-ETL action [intent=reverse_etl availability=implemented write=delete_location]; approval: requires plan, preview, approval, and execute; risk: Delete Location in the configured 7shifts account.; flags: --location_id (required)
    delete location assignment apply - Plan and execute the delete location assignment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_location_assignment]; approval: requires plan, preview, approval, and execute; risk: Delete Location Assignment in the configured 7shifts account.; flags: --location_id (required), --user_id (required)
    delete log book category apply - Plan and execute the delete log book category reverse-ETL action [intent=reverse_etl availability=implemented write=delete_log_book_category]; approval: requires plan, preview, approval, and execute; risk: Delete Log Book Category in the configured 7shifts account.; flags: --id (required)
    delete log book comment apply - Plan and execute the delete log book comment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_log_book_comment]; approval: requires plan, preview, approval, and execute; risk: Delete Log Book Comment in the configured 7shifts account.; flags: --id (required)
    delete log book post apply - Plan and execute the delete log book post reverse-ETL action [intent=reverse_etl availability=implemented write=delete_log_book_post]; approval: requires plan, preview, approval, and execute; risk: Delete Log Book Post in the configured 7shifts account.; flags: --id (required)
    delete role apply - Plan and execute the delete role reverse-ETL action [intent=reverse_etl availability=implemented write=delete_role]; approval: requires plan, preview, approval, and execute; risk: Delete Role in the configured 7shifts account.; flags: --role_id (required)
    delete role assignment apply - Plan and execute the delete role assignment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_role_assignment]; approval: requires plan, preview, approval, and execute; risk: Delete Role Assignment in the configured 7shifts account.; flags: --role_id (required), --user_id (required)
    delete shift apply - Plan and execute the delete shift reverse-ETL action [intent=reverse_etl availability=implemented write=delete_shift]; approval: requires plan, preview, approval, and execute; risk: Delete Shift in the configured 7shifts account.; flags: --shift_id (required)
    delete task list template apply - Plan and execute the delete task list template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_task_list_template]; approval: requires plan, preview, approval, and execute; risk: Delete Task List Template in the configured 7shifts account.; flags: --uuid (required)
    delete task tags apply - Plan and execute the delete task tags reverse-ETL action [intent=reverse_etl availability=implemented write=delete_task_tags]; approval: requires plan, preview, approval, and execute; risk: Delete Task Tags in the configured 7shifts account.; flags: --company_id (required), --uuids (required)
    delete time off apply - Plan and execute the delete time off reverse-ETL action [intent=reverse_etl availability=implemented write=delete_time_off]; approval: requires plan, preview, approval, and execute; risk: Delete Time Off in the configured 7shifts account.; flags: --time_off_id (required)
    delete time punch by id apply - Plan and execute the delete time punch by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_time_punch_by_id]; approval: requires plan, preview, approval, and execute; risk: Delete Time Punch in the configured 7shifts account.; flags: --time_punch_id (required)
    departments list - Run the departments ETL stream [intent=etl availability=implemented stream=departments]
    edit availability apply - Plan and execute the edit availability reverse-ETL action [intent=reverse_etl availability=implemented write=edit_availability]; approval: requires plan, preview, approval, and execute; risk: Update Availability in the configured 7shifts account.; flags: --availability_id (required)
    edit availability reason apply - Plan and execute the edit availability reason reverse-ETL action [intent=reverse_etl availability=implemented write=edit_availability_reason]; approval: requires plan, preview, approval, and execute; risk: Update Availability Reason in the configured 7shifts account.; flags: --availability_reason_id (required), --reason (required)
    edit company webhook apply - Plan and execute the edit company webhook reverse-ETL action [intent=reverse_etl availability=implemented write=edit_company_webhook]; approval: requires plan, preview, approval, and execute; risk: Update Webhook in the configured 7shifts account.; flags: --url (required), --webhook_id (required)
    edit event apply - Plan and execute the edit event reverse-ETL action [intent=reverse_etl availability=not_implemented write=edit_event]; approval: requires plan, preview, approval, and execute; risk: Update Event in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    edit task list template apply - Plan and execute the edit task list template reverse-ETL action [intent=reverse_etl availability=implemented write=edit_task_list_template]; approval: requires plan, preview, approval, and execute; risk: Update Task List Template in the configured 7shifts account.; flags: --uuid (required)
    edit time off apply - Plan and execute the edit time off reverse-ETL action [intent=reverse_etl availability=implemented write=edit_time_off]; approval: requires plan, preview, approval, and execute; risk: Update Time Off in the configured 7shifts account.; flags: --time_off_id (required)
    fetch tip pool manual entry list - Run the fetch tip pool manual entry ETL stream [intent=etl availability=implemented stream=fetch_tip_pool_manual_entry]
    find by id list - Run the find by id ETL stream [intent=etl availability=implemented stream=find_by_id]
    get assignments list - Run the get assignments ETL stream [intent=etl availability=implemented stream=get_assignments]
    get availability by id list - Run the get availability by id ETL stream [intent=etl availability=implemented stream=get_availability_by_id]
    get daily sales and labor list - Run the get daily sales and labor ETL stream [intent=etl availability=implemented stream=get_daily_sales_and_labor]
    get engage overview by location id list - Run the get engage overview by location id ETL stream [intent=etl availability=implemented stream=get_engage_overview_by_location_id]
    get event list - Run the get event ETL stream [intent=etl availability=implemented stream=get_event]
    get events list - Run the get events ETL stream [intent=etl availability=implemented stream=get_events]
    get hours and wages list - Run the get hours and wages ETL stream [intent=etl availability=implemented stream=get_hours_and_wages]
    get location by id list - Run the get location by id ETL stream [intent=etl availability=implemented stream=get_location_by_id]
    get receipt list - Run the get receipt ETL stream [intent=etl availability=implemented stream=get_receipt]
    get role assignments list - Run the get role assignments ETL stream [intent=etl availability=implemented stream=get_role_assignments]
    get task list daily summary list - Run the get task list daily summary ETL stream [intent=etl availability=implemented stream=get_task_list_daily_summary]
    get task list list - Run the get task list ETL stream [intent=etl availability=implemented stream=get_task_list]
    get task list template list - Run the get task list template ETL stream [intent=etl availability=implemented stream=get_task_list_template]
    get task list templates list - Run the get task list templates ETL stream [intent=etl availability=implemented stream=get_task_list_templates]
    get task lists list - Run the get task lists ETL stream [intent=etl availability=implemented stream=get_task_lists]
    get task management settings list - Run the get task management settings ETL stream [intent=etl availability=implemented stream=get_task_management_settings]
    get time clocking payroll period list - Run the get time clocking payroll period ETL stream [intent=etl availability=implemented stream=get_time_clocking_payroll_period]
    get time clocking payroll periods list - Run the get time clocking payroll periods ETL stream [intent=etl availability=implemented stream=get_time_clocking_payroll_periods]
    get time off list list - Run the get time off list ETL stream [intent=etl availability=implemented stream=get_time_off_list]
    get time off settings list - Run the get time off settings ETL stream [intent=etl availability=implemented stream=get_time_off_settings]
    get time punch by id list - Run the get time punch by id ETL stream [intent=etl availability=implemented stream=get_time_punch_by_id]
    get tip pool settings list - Run the get tip pool settings ETL stream [intent=etl availability=implemented stream=get_tip_pool_settings]
    get total hours list - Run the get total hours ETL stream [intent=etl availability=implemented stream=get_total_hours]
    get user list - Run the get user ETL stream [intent=etl availability=implemented stream=get_user]
    get user wages list - Run the get user wages ETL stream [intent=etl availability=implemented stream=get_user_wages]
    list availabilities list - Run the list availabilities ETL stream [intent=etl availability=implemented stream=list_availabilities]
    list availability reasons list - Run the list availability reasons ETL stream [intent=etl availability=implemented stream=list_availability_reasons]
    list company webhooks list - Run the list company webhooks ETL stream [intent=etl availability=implemented stream=list_company_webhooks]
    list department assignments list - Run the list department assignments ETL stream [intent=etl availability=implemented stream=list_department_assignments]
    list employment record list - Run the list employment record ETL stream [intent=etl availability=implemented stream=list_employment_record]
    list external user mappings list - Run the list external user mappings ETL stream [intent=etl availability=implemented stream=list_external_user_mappings]
    list inactive reasons list - Run the list inactive reasons ETL stream [intent=etl availability=implemented stream=list_inactive_reasons]
    list location assignments list - Run the list location assignments ETL stream [intent=etl availability=implemented stream=list_location_assignments]
    list log book categories list - Run the list log book categories ETL stream [intent=etl availability=implemented stream=list_log_book_categories]
    list log book comments list - Run the list log book comments ETL stream [intent=etl availability=implemented stream=list_log_book_comments]
    list log book posts list - Run the list log book posts ETL stream [intent=etl availability=implemented stream=list_log_book_posts]
    list sales receipts list - Run the list sales receipts ETL stream [intent=etl availability=implemented stream=list_sales_receipts]
    list scheduled shifts list - Run the list scheduled shifts ETL stream [intent=etl availability=implemented stream=list_scheduled_shifts]
    list shift feedback list - Run the list shift feedback ETL stream [intent=etl availability=implemented stream=list_shift_feedback]
    list user contacts list - Run the list user contacts ETL stream [intent=etl availability=implemented stream=list_user_contacts]
    list users authorized locations list - Run the list users authorized locations ETL stream [intent=etl availability=implemented stream=list_users_authorized_locations]
    locations list - Run the locations ETL stream [intent=etl availability=implemented stream=locations]
    post shift apply - Plan and execute the post shift reverse-ETL action [intent=reverse_etl availability=implemented write=post_shift]; approval: requires plan, preview, approval, and execute; risk: Create Shift in the configured 7shifts account.; flags: --end (required), --location_id (required), --start (required)
    post time punch apply - Plan and execute the post time punch reverse-ETL action [intent=reverse_etl availability=implemented write=post_time_punch]; approval: requires plan, preview, approval, and execute; risk: Create Time Punch in the configured 7shifts account.; flags: --clocked_in (required), --location_id (required), --user_id (required)
    post user apply - Plan and execute the post user reverse-ETL action [intent=reverse_etl availability=not_implemented write=post_user]; approval: requires plan, preview, approval, and execute; risk: Create User in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    put time punch apply - Plan and execute the put time punch reverse-ETL action [intent=reverse_etl availability=implemented write=put_time_punch]; approval: requires plan, preview, approval, and execute; risk: Update Time Punch in the configured 7shifts account.; flags: --time_punch_id (required)
    put user apply - Plan and execute the put user reverse-ETL action [intent=reverse_etl availability=implemented write=put_user]; approval: requires plan, preview, approval, and execute; risk: Update User in the configured 7shifts account.; flags: --identifier (required)
    retrieve company labor settings list - Run the retrieve company labor settings ETL stream [intent=etl availability=implemented stream=retrieve_company_labor_settings]
    retrieve daily stats list - Run the retrieve daily stats ETL stream [intent=etl availability=implemented stream=retrieve_daily_stats]
    retrieve day part settings list - Run the retrieve day part settings ETL stream [intent=etl availability=implemented stream=retrieve_day_part_settings]
    retrieve department list - Run the retrieve department ETL stream [intent=etl availability=implemented stream=retrieve_department]
    retrieve external user mapping list - Run the retrieve external user mapping ETL stream [intent=etl availability=implemented stream=retrieve_external_user_mapping]
    retrieve log book comment list - Run the retrieve log book comment ETL stream [intent=etl availability=implemented stream=retrieve_log_book_comment]
    retrieve log book post list - Run the retrieve log book post ETL stream [intent=etl availability=implemented stream=retrieve_log_book_post]
    retrieve receipts summary list - Run the retrieve receipts summary ETL stream [intent=etl availability=implemented stream=retrieve_receipts_summary]
    retrieve role list - Run the retrieve role ETL stream [intent=etl availability=implemented stream=retrieve_role]
    retrieve shift list - Run the retrieve shift ETL stream [intent=etl availability=implemented stream=retrieve_shift]
    retrieve tip pool detailed report list - Run the retrieve tip pool detailed report ETL stream [intent=etl availability=implemented stream=retrieve_tip_pool_detailed_report]
    retrieve tip pool summary report list - Run the retrieve tip pool summary report ETL stream [intent=etl availability=implemented stream=retrieve_tip_pool_summary_report]
    retrieve user contact list - Run the retrieve user contact ETL stream [intent=etl availability=implemented stream=retrieve_user_contact]
    roles list - Run the roles ETL stream [intent=etl availability=implemented stream=roles]
    save time off settings apply - Plan and execute the save time off settings reverse-ETL action [intent=reverse_etl availability=implemented write=save_time_off_settings]; approval: requires plan, preview, approval, and execute; risk: Create Time Off Settings in the configured 7shifts account.
    save tip pool manual entry apply - Plan and execute the save tip pool manual entry reverse-ETL action [intent=reverse_etl availability=not_implemented write=save_tip_pool_manual_entry]; approval: requires plan, preview, approval, and execute; risk: Update Tip Pool Manual Entries in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    shifts list - Run the shifts ETL stream [intent=etl availability=implemented stream=shifts]
    sync overridden projected sales interval apply - Plan and execute the sync overridden projected sales interval reverse-ETL action [intent=reverse_etl availability=implemented write=sync_overridden_projected_sales_interval]; approval: requires plan, preview, approval, and execute; risk: Delete Sales Forecast Override Interval in the configured 7shifts account.; flags: --end (required), --location_id (required), --start (required)
    time punches list - Run the time punches ETL stream [intent=etl availability=implemented stream=time_punches]
    update availability status apply - Plan and execute the update availability status reverse-ETL action [intent=reverse_etl availability=implemented write=update_availability_status]; approval: requires plan, preview, approval, and execute; risk: Update Availability Status in the configured 7shifts account.; flags: --availability_id (required), --status (required)
    update company apply - Plan and execute the update company reverse-ETL action [intent=reverse_etl availability=implemented write=update_company]; approval: requires plan, preview, approval, and execute; risk: Update Company in the configured 7shifts account.; flags: --id (required)
    update complete receipt apply - Plan and execute the update complete receipt reverse-ETL action [intent=reverse_etl availability=implemented write=update_complete_receipt]; approval: requires plan, preview, approval, and execute; risk: Update Receipt in the configured 7shifts account.; flags: --net_total (required), --receipt_id (required)
    update department apply - Plan and execute the update department reverse-ETL action [intent=reverse_etl availability=implemented write=update_department]; approval: requires plan, preview, approval, and execute; risk: Update Department in the configured 7shifts account.; flags: --default (required), --department_id (required), --name (required)
    update employment record apply - Plan and execute the update employment record reverse-ETL action [intent=reverse_etl availability=implemented write=update_employment_record]; approval: requires plan, preview, approval, and execute; risk: Update Employment Record in the configured 7shifts account.; flags: --uuid (required)
    update external user mappings apply - Plan and execute the update external user mappings reverse-ETL action [intent=reverse_etl availability=implemented write=update_external_user_mappings]; approval: requires plan, preview, approval, and execute; risk: Update External User Mappings in the configured 7shifts account.; flags: --identifier (required)
    update location apply - Plan and execute the update location reverse-ETL action [intent=reverse_etl availability=implemented write=update_location]; approval: requires plan, preview, approval, and execute; risk: Update Location in the configured 7shifts account.; flags: --location_id (required)
    update log book category apply - Plan and execute the update log book category reverse-ETL action [intent=reverse_etl availability=implemented write=update_log_book_category]; approval: requires plan, preview, approval, and execute; risk: Update Log Book Category in the configured 7shifts account.; flags: --id (required)
    update role apply - Plan and execute the update role reverse-ETL action [intent=reverse_etl availability=implemented write=update_role]; approval: requires plan, preview, approval, and execute; risk: Update Role in the configured 7shifts account.; flags: --role_id (required)
    update role assignment apply - Plan and execute the update role assignment reverse-ETL action [intent=reverse_etl availability=implemented write=update_role_assignment]; approval: requires plan, preview, approval, and execute; risk: Update Role Assignment in the configured 7shifts account.; flags: --role_id (required), --user_id (required)
    update shift apply - Plan and execute the update shift reverse-ETL action [intent=reverse_etl availability=implemented write=update_shift]; approval: requires plan, preview, approval, and execute; risk: Update Shift in the configured 7shifts account.; flags: --shift_id (required)
    upsert bulk employment records apply - Plan and execute the upsert bulk employment records reverse-ETL action [intent=reverse_etl availability=not_implemented write=upsert_bulk_employment_records]; approval: requires plan, preview, approval, and execute; risk: Create Many Employment Records in the configured 7shifts account.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
    view company list - Run the view company ETL stream [intent=etl availability=implemented stream=view_company]
    who am i list - Run the who am i ETL stream [intent=etl availability=implemented stream=who_am_i]

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
