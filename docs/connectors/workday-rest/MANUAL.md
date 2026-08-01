# pm connectors inspect workday-rest

```text
NAME
  pm connectors inspect workday-rest - Workday REST connector manual

SYNOPSIS
  pm connectors inspect workday-rest
  pm connectors inspect workday-rest --json
  pm credentials add <name> --connector workday-rest [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Workday REST 2026.30 production service resources, exposes bounded provider direct reads for values/search endpoints, and provides typed non-binary reverse ETL write actions. Fixture-backed; not live-certified.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  activity_id
  base_url
  business_object_id
  business_object_resource
  custom_object_alias
  custom_object_id
  data_change_id
  file_container_id
  id
  subresource_id
  tenant
  access_token (secret)

ETL STREAMS
  absence_management_v5_absence_management_v5_workers_id_leaves_of_absence_subresource_id:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_balances_id:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id_eligible_absence_types:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id_leaves_of_absence:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_balances:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id_time_off_details_subresource_id:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id_time_off_details:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id_valid_time_off_dates:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers:
    primary key: id
    fields: descriptor(), id()
  absence_management_v5_absence_management_v5_workers_id_eligible_absence_types_subresource_id:
    primary key: id
    fields: descriptor(), id()
  accounts_payable_v1_accounts_payable_v1_supplier_invoice_requests:
    primary key: id
    fields: descriptor(), id()
  accounts_payable_v1_accounts_payable_v1_supplier_invoice_requests_id_lines:
    primary key: id
    fields: descriptor(), id()
  accounts_payable_v1_accounts_payable_v1_supplier_invoice_requests_id:
    primary key: id
    fields: descriptor(), id()
  accounts_payable_v1_accounts_payable_v1_supplier_invoice_requests_id_lines_subresource_id:
    primary key: id
    fields: descriptor(), id()
  asor_v1_asor_v1_agent_definition:
    primary key: id
    fields: descriptor(), id()
  asor_v1_asor_v1_registration:
    primary key: id
    fields: descriptor(), id()
  asor_v1_asor_v1_registration_id:
    primary key: id
    fields: descriptor(), id()
  benefit_enrollment_event_offerings_v1_benefit_enrollment_event_offerings_v1_employee_enrollment_event_id:
    primary key: id
    fields: descriptor(), id()
  benefit_enrollment_event_offerings_v1_benefit_enrollment_event_offerings_v1_employee_enrollment_event:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_event_steps:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_events:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_events_id_completed_steps:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_types:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_types_id:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_event_steps_id:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_events_id_comments:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_events_id_in_progress_steps:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_events_id_remaining_steps:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_types_id_attachment_categories:
    primary key: id
    fields: descriptor(), id()
  business_process_v1_business_process_v1_events_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_time_off_plans_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_job_change_reasons:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_organizations:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_business_title_changes_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_business_title_changes_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_currencies_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_inbox_tasks:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_currencies:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_customers_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_organizations_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_history:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_organization_types:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_history_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_organizations:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_supervisory_organizations_managed:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_time_off_entries_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_pay_slips:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_direct_reports:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_job_change_reasons_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_supervisory_organizations:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_pay_slips_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_supervisory_organizations_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_customers_id_activities_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_inbox_tasks_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_supervisory_organizations_id_workers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_direct_reports_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_time_off_plans:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_supervisory_organizations_id_workers:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_business_title_changes:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_organizations_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_customers_id_activities:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_supervisory_organizations_managed_subresource_id:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_workers_id_time_off_entries:
    primary key: id
    fields: descriptor(), id()
  common_v1_api_common_v1_organization_types_id:
    primary key: id
    fields: descriptor(), id()
  compensation_v3_compensation_v3_scorecard_results:
    primary key: id
    fields: descriptor(), id()
  compensation_v3_compensation_v3_scorecards:
    primary key: id
    fields: descriptor(), id()
  compensation_v3_compensation_v3_workers_id:
    primary key: id
    fields: descriptor(), id()
  compensation_v3_compensation_v3_scorecard_results_id:
    primary key: id
    fields: descriptor(), id()
  compensation_v3_compensation_v3_workers:
    primary key: id
    fields: descriptor(), id()
  compensation_v3_compensation_v3_scorecards_id:
    primary key: id
    fields: descriptor(), id()
  connect_v2_connect_v2_message_templates_id:
    primary key: id
    fields: descriptor(), id()
  connect_v2_connect_v2_notification_types:
    primary key: id
    fields: descriptor(), id()
  connect_v2_connect_v2_notification_types_id:
    primary key: id
    fields: descriptor(), id()
  connect_v2_connect_v2_message_templates:
    primary key: id
    fields: descriptor(), id()
  contract_compliance_v1_contract_compliance_v1_supplier_contracts:
    primary key: id
    fields: descriptor(), id()
  contract_compliance_v1_contract_compliance_v1_supplier_contracts_id:
    primary key: id
    fields: descriptor(), id()
  core_accounting_v1_core_accounting_v1_currencies:
    primary key: id
    fields: descriptor(), id()
  core_accounting_v1_core_accounting_v1_currencies_id:
    primary key: id
    fields: descriptor(), id()
  core_accounting_v1_core_accounting_v1_evaluate_account_posting_rules:
    primary key: id
    fields: descriptor(), id()
  custom_object_data_multi_instance_v2_business_object_resource_business_object_id_custom_objects_custom_object_alias:
    primary key: id
    fields: descriptor(), id()
  custom_object_data_multi_instance_v2_custom_objects_custom_object_alias_custom_object_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_data_single_instance_v2_custom_objects_custom_object_alias_custom_object_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id_validations:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id_validations_subresource_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_field_types:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_field_types_id_list_values_subresource_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id_condition_rules:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id_condition_rules_subresource_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_field_types_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_field_types_id_list_values:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id_fields_subresource_id:
    primary key: id
    fields: descriptor(), id()
  custom_object_definition_v1_custom_object_definition_v1_definitions_id_fields:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_payments_id_remittance_details_subresource_id:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_customers_id:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_payments_id:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_invoices_id:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_invoices:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_invoices_id_print_runs_subresource_id:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_invoices_id_print_runs:
    primary key: id
    fields: descriptor(), id()
  customer_accounts_v1_customer_accounts_v1_customers:
    primary key: id
    fields: descriptor(), id()
  expense_v1_expense_v1_expense_items_id:
    primary key: id
    fields: descriptor(), id()
  expense_v1_expense_v1_reports_id:
    primary key: id
    fields: descriptor(), id()
  expense_v1_expense_v1_entries_id:
    primary key: id
    fields: descriptor(), id()
  expense_v1_expense_v1_reports:
    primary key: id
    fields: descriptor(), id()
  expense_v1_expense_v1_entries:
    primary key: id
    fields: descriptor(), id()
  expense_v1_expense_v1_expense_items:
    primary key: id
    fields: descriptor(), id()
  global_payroll_v1_global_payroll_v1_pay_groups_id:
    primary key: id
    fields: descriptor(), id()
  global_payroll_v1_global_payroll_v1_pay_groups_id_periods:
    primary key: id
    fields: descriptor(), id()
  global_payroll_v1_global_payroll_v1_event_driven_integration_vendor_response_id:
    primary key: id
    fields: descriptor(), id()
  global_payroll_v1_global_payroll_v1_pay_groups_id_periods_subresource_id:
    primary key: id
    fields: descriptor(), id()
  global_payroll_v1_global_payroll_v1_pay_groups:
    primary key: id
    fields: descriptor(), id()
  global_payroll_v1_global_payroll_v1_effective_changes_id:
    primary key: id
    fields: descriptor(), id()
  graph_v1_graph_v1_versions:
    primary key: id
    fields: descriptor(), id()
  graph_v1_graph_v1_versions_id:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_versions_id:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_versions:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_versions_id_approval_decision:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_versions_id_approval_decision_subresource_id:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_versions_id_approval_request_subresource_id:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_statuses:
    primary key: id
    fields: descriptor(), id()
  help_article_v1_help_article_v1_article_statuses_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_service_categories_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id_timeline:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id_comment_subresource_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_case_suggestions:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_configuration:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_case_suggestions_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_external_creators:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id_timeline_subresource_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id_internal_note_timeline_subresource_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id_comment:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id_internal_note_timeline:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_external_creators_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_case_types_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_cases_id:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_case_types:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_service_categories:
    primary key: id
    fields: descriptor(), id()
  help_case_v4_help_case_v4_configuration_id:
    primary key: id
    fields: descriptor(), id()
  holiday_v1_holiday_v1_holiday_events:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_content:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_content_id:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_records:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_enrollments_id_lesson_trackings:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_content_id_lessons:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_records_id:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_enrollments_id:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_enrollments:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_enrollments_id_lesson_trackings_subresource_id:
    primary key: id
    fields: descriptor(), id()
  learning_v1_learning_v1_content_id_lessons_subresource_id:
    primary key: id
    fields: descriptor(), id()
  o_auth_client_v1_o_auth_client_v1_client_details:
    primary key: id
    fields: descriptor(), id()
  o_auth_client_v1_o_auth_client_v1_client_details_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_jobs_id_pay_group_subresource_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_pay_groups_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_jobs_id_pay_group:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_pay_group_details_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_jobs:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_minimum_wage_rates:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_jobs_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_pay_group_details:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_payroll_inputs_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_minimum_wage_rates_id:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_payroll_inputs:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_tax_rates:
    primary key: id
    fields: descriptor(), id()
  payroll_v2_payroll_v2_pay_groups:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_feedback_badges:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_requested_feedback_on_self_events_subresource_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_give_requested_feedback_events:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_feedback_badges_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_goals:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_requested_feedback_on_self_events:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_requested_feedback_on_worker_events_subresource_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_development_items:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_requested_feedback_on_worker_events:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_worker_goal_events:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_development_items_subresource_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_goals_subresource_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_anytime_feedback_events_subresource_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_workers_id_anytime_feedback_events:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_give_requested_feedback_events_id:
    primary key: id
    fields: descriptor(), id()
  performance_enablement_v5_performance_enablement_v5_worker_goal_events_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_personal_information:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_countries_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_web_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_additional_names_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_phone_numbers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_audio_name_pronunciation:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_personal_information_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_instant_messengers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_phone_numbers:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_photos:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_preferred_name:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_countries_id_address_components:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_photos_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_instant_messengers:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_legal_name_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_emails:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_phones_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_instant_messengers:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_email_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_countries_id_name_components:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_instant_messengers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_instant_messengers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_phones:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_web_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_audio_name_pronunciation_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_web_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_phone_numbers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_public_contact_information_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_phones:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_public_contact_information:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_legal_name:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_web_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_preferred_name_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_instant_messengers:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_additional_names:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_web_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_phones_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_emails_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_web_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_instant_messengers_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_work_emails_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_email_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_phone_numbers:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_email_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_web_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_email_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_instant_messengers:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_people_id_home_emails:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_web_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_home_contact_information_changes_id_addresses_subresource_id:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_work_contact_information_changes_id_addresses:
    primary key: id
    fields: descriptor(), id()
  person_v4_person_v4_countries:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_tables:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_tables_id:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_buckets:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_buckets_id:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_data_changes:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_data_changes_data_change_id:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_data_changes_data_change_id_activities_activity_id:
    primary key: id
    fields: descriptor(), id()
  prism_analytics_v3_data_changes_data_change_id_validate:
    primary key: id
    fields: descriptor(), id()
  privacy_v1_privacy_v1_activity_logging:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_purchase_orders_id:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisitions_id_related_purchase_orders_subresource_id:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisitions_id_requisition_lines_subresource_id:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisition_templates_id:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisitions:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisitions_id_related_purchase_orders:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisition_templates:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisitions_id:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_purchase_orders:
    primary key: id
    fields: descriptor(), id()
  procurement_v5_procurement_v5_requisitions_id_requisition_lines:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_resource_forecast_lines_id_allocations:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_resource_forecast_lines_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_resource_plan_lines_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_projects:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_projects_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_ad_hoc_project_time_transactions:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_resource_forecast_lines_id_allocations_subresource_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_plan_tasks:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_task_resources:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_resource_plan_lines:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_task_resources_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_plan_phases_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_plan_tasks_id:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_resource_forecast_lines:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_plan_phases:
    primary key: id
    fields: descriptor(), id()
  projects_v3_projects_v3_ad_hoc_project_time_transactions_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_job_postings_id_questionnaire_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_interviews_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_interviews:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_languages:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_skills:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_experiences_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_job_postings_id_candidate_availability_template:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_job_postings:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_languages_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_educations:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_interviews_id_feedback_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_skills_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_job_postings_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_job_postings_id_candidate_availability_template_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_experiences:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_interviews_id_feedback:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_job_postings_id_questionnaire:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id_educations_subresource_id:
    primary key: id
    fields: descriptor(), id()
  recruiting_v4_recruiting_v4_prospects_id:
    primary key: id
    fields: descriptor(), id()
  request_v2_request_v2_requests_id:
    primary key: id
    fields: descriptor(), id()
  request_v2_request_v2_types:
    primary key: id
    fields: descriptor(), id()
  request_v2_request_v2_requests:
    primary key: id
    fields: descriptor(), id()
  request_v2_request_v2_types_id:
    primary key: id
    fields: descriptor(), id()
  revenue_v1_revenue_v1_billable_transactions_id_billing_rate_application:
    primary key: id
    fields: descriptor(), id()
  revenue_v1_revenue_v1_billable_transactions_id:
    primary key: id
    fields: descriptor(), id()
  revenue_v1_revenue_v1_billable_transactions_id_billing_rate_application_subresource_id:
    primary key: id
    fields: descriptor(), id()
  revenue_v1_revenue_v1_billable_transactions:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_location:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_skill_items:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_region:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_start_details_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_administrative:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_job_profile_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_location_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_check_in_topics:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_comment:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_custom_organizations_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_business_unit_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_supervisory_organizations_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_position:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_start_details:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_supervisory_organizations_id_org_chart:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_skill_items_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_external_skill_level_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_start_details_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_move_team_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_position_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_comment:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_profiles:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_business_title_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_check_ins_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_job_classification:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_administrative_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_move_team:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_opening:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_explicit_skills:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_business_title:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_explicit_skills_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_company_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_families:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_costing_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_check_in_topics_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_home_contact_information_changes_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_jobs_id_workspace:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_contract:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_start_details:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_region_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_work_contact_information_changes_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_jobs_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_check_ins:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_families_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_custom_organizations:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_job_classification_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_supervisory_organizations_id_members_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_opening_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_supervisory_organizations:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_costing:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_service_dates_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_cost_center_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_comment_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_service_dates:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_contract_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_jobs:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_job_profile:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_profiles_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_workers_id_external_skill_level:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_supervisory_organizations_id_members:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_jobs_id_workspace_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_business_unit:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_company:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_supervisory_organizations_id_org_chart_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_job_changes_id_comment_subresource_id:
    primary key: id
    fields: descriptor(), id()
  staffing_v7_staffing_v7_organization_assignment_changes_id_cost_center:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_programs_of_study_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_units_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_calendars_id_academic_years_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_educational_credentials:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_calendars:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_periods_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_calendars_id_academic_years:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_levels_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_educational_credentials_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_calendars_id:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_units:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_periods:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_academic_levels:
    primary key: id
    fields: descriptor(), id()
  student_academic_foundation_v1_student_academic_foundation_v1_programs_of_study:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id_residencies:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id_apply_hold_events:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_holds_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id_immigration_data_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id_immigration_data:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id_apply_hold_events_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id_immigration_pages:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id_immigration_pages_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_holds:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id_dependent_immigration_data_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_immigration_events_id_dependent_immigration_data:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id_immigration_events:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id_immigration_events_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_core_v1_student_core_v1_students_id_residencies_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_curriculum_v1_student_curriculum_v1_courses_id:
    primary key: id
    fields: descriptor(), id()
  student_curriculum_v1_student_curriculum_v1_course_sections:
    primary key: id
    fields: descriptor(), id()
  student_curriculum_v1_student_curriculum_v1_course_sections_id:
    primary key: id
    fields: descriptor(), id()
  student_curriculum_v1_student_curriculum_v1_courses:
    primary key: id
    fields: descriptor(), id()
  student_curriculum_v1_student_curriculum_v1_course_subjects:
    primary key: id
    fields: descriptor(), id()
  student_curriculum_v1_student_curriculum_v1_course_subjects_id:
    primary key: id
    fields: descriptor(), id()
  student_engagement_v1_student_engagement_v1_students_id_holds:
    primary key: id
    fields: descriptor(), id()
  student_engagement_v1_student_engagement_v1_students:
    primary key: id
    fields: descriptor(), id()
  student_engagement_v1_student_engagement_v1_students_id:
    primary key: id
    fields: descriptor(), id()
  student_engagement_v1_student_engagement_v1_students_id_holds_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_finance_v1_student_finance_v1_students:
    primary key: id
    fields: descriptor(), id()
  student_finance_v1_student_finance_v1_students_id:
    primary key: id
    fields: descriptor(), id()
  student_finance_v1_student_finance_v1_students_id_payments_subresource_id:
    primary key: id
    fields: descriptor(), id()
  student_finance_v1_student_finance_v1_students_id_payments:
    primary key: id
    fields: descriptor(), id()
  system_metrics_v1_system_metrics_v1_active_tasks_id:
    primary key: id
    fields: descriptor(), id()
  system_metrics_v1_system_metrics_v1_active_user_sessions:
    primary key: id
    fields: descriptor(), id()
  system_metrics_v1_system_metrics_v1_active_tasks:
    primary key: id
    fields: descriptor(), id()
  system_metrics_v1_system_metrics_v1_active_user_sessions_id:
    primary key: id
    fields: descriptor(), id()
  system_metrics_v1_system_metrics_v1_system_metrics_overview:
    primary key: id
    fields: descriptor(), id()
  system_metrics_v1_system_metrics_v1_system_metrics_overview_id:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_succession_plans_id:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_succession_plan_events_id_candidates:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_mentorships:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_succession_plan_events_id_candidates_subresource_id:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_succession_plan_events_id:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_mentorships_id:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_succession_plans:
    primary key: id
    fields: descriptor(), id()
  talent_management_v2_talent_management_v2_succession_plan_events:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_worker_time_attestation_id_followup_worker_time_attestation:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_time_validations:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_workers:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_workers_id_time_totals:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_worker_time_blocks_id:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_time_clock_events:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_worker_time_attestation_id:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_workers_id_time_totals_subresource_id:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_time_clock_events_id:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_workers_id_period:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_workers_id:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_worker_time_attestation_id_followup_time_attestation_prompt:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_worker_time_blocks:
    primary key: id
    fields: descriptor(), id()
  time_tracking_v5_time_tracking_v5_time_attestation_prompts:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data_sources_id:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data_sources_id_data_source_filters_subresource_id:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data_sources_id_data_source_filters:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data_sources:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data_sources_id_fields_subresource_id:
    primary key: id
    fields: descriptor(), id()
  wql_v1_wql_v1_data_sources_id_fields:
    primary key: id
    fields: descriptor(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_absence_management_v5_absence_management_v5_workers_id_correct_time_off_entry:
    endpoint: POST /absenceManagement/v5/workers/{{ record.id }}/correctTimeOffEntry
    required fields: id, json_data
    risk: POST /absenceManagement/v5/workers/{ID}/correctTimeOffEntry against Workday REST Absence Management v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/absenceManagement_v5_20260727_oas2.json
  create_absence_management_v5_absence_management_v5_workers_id_request_time_off:
    endpoint: POST /absenceManagement/v5/workers/{{ record.id }}/requestTimeOff
    required fields: id, json_data
    risk: POST /absenceManagement/v5/workers/{ID}/requestTimeOff against Workday REST Absence Management v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/absenceManagement_v5_20260727_oas2.json
  create_accounts_payable_v1_accounts_payable_v1_supplier_invoice_requests_id_submit:
    endpoint: POST /accountsPayable/v1/supplierInvoiceRequests/{{ record.id }}/submit
    required fields: id
    risk: POST /accountsPayable/v1/supplierInvoiceRequests/{ID}/submit against Workday REST Accounts Payable v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/accountsPayable_v1_20260727_oas2.json
  create_accounts_payable_v1_accounts_payable_v1_supplier_invoice_requests:
    endpoint: POST /accountsPayable/v1/supplierInvoiceRequests
    required fields: company, invoiceDate, supplier
    risk: POST /accountsPayable/v1/supplierInvoiceRequests against Workday REST Accounts Payable v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/accountsPayable_v1_20260727_oas2.json
  create_asor_v1_asor_v1_agent_definition:
    endpoint: POST /asor/v1/agentDefinition
    required fields: description, name, platform, provider, skills, url, version
    risk: POST /asor/v1/agentDefinition against Workday REST Asor v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/asor_v1_20260727_oas2.json
  create_benefit_partner_v1_benefit_partner_v1_programs:
    endpoint: POST /benefitPartner/v1/programs
    required fields: json_data
    risk: POST /benefitPartner/v1/programs against Workday REST Benefit Partner v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/benefitPartner_v1_20260727_oas2.json
  create_budgets_v1_budgets_v1_run_budget_check:
    endpoint: POST /budgets/v1/runBudgetCheck
    required fields: company, inflightTransactionDate
    risk: POST /budgets/v1/runBudgetCheck against Workday REST Budgets v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/budgets_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_event_steps_id_to_do:
    endpoint: POST /businessProcess/v1/eventSteps/{{ record.id }}/toDo
    required fields: id, stepAction
    risk: POST /businessProcess/v1/eventSteps/{ID}/toDo against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_event_steps_id_reassign:
    endpoint: POST /businessProcess/v1/eventSteps/{{ record.id }}/reassign
    required fields: id
    risk: POST /businessProcess/v1/eventSteps/{ID}/reassign against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_events_id_rescind:
    endpoint: POST /businessProcess/v1/events/{{ record.id }}/rescind
    required fields: id
    risk: POST /businessProcess/v1/events/{ID}/rescind against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_event_steps_id_deny:
    endpoint: POST /businessProcess/v1/eventSteps/{{ record.id }}/deny
    required fields: id
    risk: POST /businessProcess/v1/eventSteps/{ID}/deny against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_event_steps_id_questionnaire:
    endpoint: POST /businessProcess/v1/eventSteps/{{ record.id }}/questionnaire
    required fields: id, json_data
    risk: POST /businessProcess/v1/eventSteps/{ID}/questionnaire against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_events_id_cancel:
    endpoint: POST /businessProcess/v1/events/{{ record.id }}/cancel
    required fields: id
    risk: POST /businessProcess/v1/events/{ID}/cancel against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_event_steps_id_approve:
    endpoint: POST /businessProcess/v1/eventSteps/{{ record.id }}/approve
    required fields: id
    risk: POST /businessProcess/v1/eventSteps/{ID}/approve against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_business_process_v1_business_process_v1_event_steps_id_send_back:
    endpoint: POST /businessProcess/v1/eventSteps/{{ record.id }}/sendBack
    required fields: id, reason, to
    risk: POST /businessProcess/v1/eventSteps/{ID}/sendBack against Workday REST Business Process v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/businessProcess_v1_20260727_oas2.json
  create_common_v1_api_common_v1_workers_id_job_changes:
    endpoint: POST /api/common/v1/workers/{{ record.id }}/jobChanges
    required fields: id
    risk: POST /api/common/v1/workers/{ID}/jobChanges against Workday REST Common v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/common_v1_20260727_oas2.json
  create_common_v1_api_common_v1_validate_worktags:
    endpoint: POST /api/common/v1/validateWorktags
    risk: POST /api/common/v1/validateWorktags against Workday REST Common v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/common_v1_20260727_oas2.json
  create_common_v1_api_common_v1_workers_id_business_title_changes_2:
    endpoint: POST /api/common/v1/workers/{{ record.id }}/businessTitleChanges
    required fields: id
    risk: POST /api/common/v1/workers/{ID}/businessTitleChanges against Workday REST Common v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/common_v1_20260727_oas2.json
  create_communications_v1_communications_v1_managed_recipient:
    endpoint: POST /communications/v1/managedRecipient
    required fields: channel, messagingContactable_ID, optInPreference, phoneNumber
    risk: POST /communications/v1/managedRecipient against Workday REST Communications v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/communications_v1_20260727_oas2.json
  create_compensation_v3_compensation_v3_scorecard_results:
    endpoint: POST /compensation/v3/scorecardResults
    required fields: evaluationDate, scorecardID
    risk: POST /compensation/v3/scorecardResults against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  create_compensation_v3_compensation_v3_scorecards:
    endpoint: POST /compensation/v3/scorecards
    required fields: defaultScorecardGoals, effectiveDate, scorecardName
    risk: POST /compensation/v3/scorecards against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  delete_compensation_v3_compensation_v3_scorecard_results_id:
    endpoint: DELETE /compensation/v3/scorecardResults/{{ record.id }}
    required fields: id
    risk: DELETE /compensation/v3/scorecardResults/{ID} against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  create_compensation_v3_compensation_v3_workers_id_request_one_time_payment:
    endpoint: POST /compensation/v3/workers/{{ record.id }}/requestOneTimePayment
    required fields: id
    risk: POST /compensation/v3/workers/{ID}/requestOneTimePayment against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  update_compensation_v3_compensation_v3_scorecard_results_id_scores_subresource_id:
    endpoint: PATCH /compensation/v3/scorecardResults/{{ record.id }}/scores/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /compensation/v3/scorecardResults/{ID}/scores/{subresourceID} against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  update_compensation_v3_compensation_v3_scorecards_id:
    endpoint: PUT /compensation/v3/scorecards/{{ record.id }}
    required fields: id, defaultScorecardGoals, effectiveDate, scorecardName
    risk: PUT /compensation/v3/scorecards/{ID} against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  delete_compensation_v3_compensation_v3_scorecards_id:
    endpoint: DELETE /compensation/v3/scorecards/{{ record.id }}
    required fields: id
    risk: DELETE /compensation/v3/scorecards/{ID} against Workday REST Compensation v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/compensation_v3_20260727_oas2.json
  update_connect_v2_connect_v2_message_templates_id:
    endpoint: PUT /connect/v2/messageTemplates/{{ record.id }}
    required fields: id, name, notificationType
    risk: PUT /connect/v2/messageTemplates/{ID} against Workday REST Connect v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/connect_v2_20260727_oas2.json
  update_connect_v2_connect_v2_message_templates_id_2:
    endpoint: PATCH /connect/v2/messageTemplates/{{ record.id }}
    required fields: id, name, notificationType
    risk: PATCH /connect/v2/messageTemplates/{ID} against Workday REST Connect v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/connect_v2_20260727_oas2.json
  create_connect_v2_connect_v2_send_message:
    endpoint: POST /connect/v2/sendMessage
    risk: POST /connect/v2/sendMessage against Workday REST Connect v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/connect_v2_20260727_oas2.json
  create_connect_v2_connect_v2_message_templates:
    endpoint: POST /connect/v2/messageTemplates
    required fields: name, notificationType
    risk: POST /connect/v2/messageTemplates against Workday REST Connect v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/connect_v2_20260727_oas2.json
  create_custom_object_data_multi_instance_v2_custom_objects_custom_object_alias:
    endpoint: POST /customObjects/{{ record.custom_object_alias }}
    required fields: custom_object_alias
    risk: POST /customObjects/{customObjectAlias} against Workday REST Custom Object Data (multi-instance) v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDataMultiInstance_v2_20230712_oas3.json
  update_custom_object_data_multi_instance_v2_custom_objects_custom_object_alias_custom_object_id:
    endpoint: PUT /customObjects/{{ record.custom_object_alias }}/{{ record.custom_object_id }}
    required fields: custom_object_alias, custom_object_id
    risk: PUT /customObjects/{customObjectAlias}/{customObjectID} against Workday REST Custom Object Data (multi-instance) v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDataMultiInstance_v2_20230712_oas3.json
  delete_custom_object_data_multi_instance_v2_custom_objects_custom_object_alias_custom_object_id:
    endpoint: DELETE /customObjects/{{ record.custom_object_alias }}/{{ record.custom_object_id }}
    required fields: custom_object_alias, custom_object_id
    risk: DELETE /customObjects/{customObjectAlias}/{customObjectID} against Workday REST Custom Object Data (multi-instance) v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDataMultiInstance_v2_20230712_oas3.json
  create_custom_object_data_single_instance_v2_custom_objects_custom_object_alias:
    endpoint: POST /customObjects/{{ record.custom_object_alias }}
    required fields: custom_object_alias
    risk: POST /customObjects/{customObjectAlias} against Workday REST Custom Object Data (single-instance) v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDataSingleInstance_v2_20230712_oas3.json
  update_custom_object_data_single_instance_v2_custom_objects_custom_object_alias_custom_object_id:
    endpoint: PUT /customObjects/{{ record.custom_object_alias }}/{{ record.custom_object_id }}
    required fields: custom_object_alias, custom_object_id
    risk: PUT /customObjects/{customObjectAlias}/{customObjectID} against Workday REST Custom Object Data (single-instance) v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDataSingleInstance_v2_20230712_oas3.json
  delete_custom_object_data_single_instance_v2_custom_objects_custom_object_alias_custom_object_id:
    endpoint: DELETE /customObjects/{{ record.custom_object_alias }}/{{ record.custom_object_id }}
    required fields: custom_object_alias, custom_object_id
    risk: DELETE /customObjects/{customObjectAlias}/{customObjectID} against Workday REST Custom Object Data (single-instance) v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDataSingleInstance_v2_20230712_oas3.json
  create_custom_object_definition_v1_custom_object_definition_v1_definitions_id_validations:
    endpoint: POST /customObjectDefinition/v1/definitions/{{ record.id }}/validations
    required fields: id, conditionRule, customField, messageText, name, onlyOnOk, severityLevel
    risk: POST /customObjectDefinition/v1/definitions/{ID}/validations against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  update_custom_object_definition_v1_custom_object_definition_v1_definitions_id_validations_subresource_id:
    endpoint: PUT /customObjectDefinition/v1/definitions/{{ record.id }}/validations/{{ record.subresource_id }}
    required fields: id, subresource_id, conditionRule, customField, messageText, name, onlyOnOk, severityLevel
    risk: PUT /customObjectDefinition/v1/definitions/{ID}/validations/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  delete_custom_object_definition_v1_custom_object_definition_v1_definitions_id_validations_subresource_id:
    endpoint: DELETE /customObjectDefinition/v1/definitions/{{ record.id }}/validations/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /customObjectDefinition/v1/definitions/{ID}/validations/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_custom_object_definition_v1_custom_object_definition_v1_field_types:
    endpoint: POST /customObjectDefinition/v1/fieldTypes
    required fields: alias, name
    risk: POST /customObjectDefinition/v1/fieldTypes against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_custom_object_definition_v1_custom_object_definition_v1_definitions_id_activate:
    endpoint: POST /customObjectDefinition/v1/definitions/{{ record.id }}/activate
    required fields: id
    risk: POST /customObjectDefinition/v1/definitions/{ID}/activate against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  update_custom_object_definition_v1_custom_object_definition_v1_definitions_id:
    endpoint: PUT /customObjectDefinition/v1/definitions/{{ record.id }}
    required fields: id, alias, domains, multiInstance, name
    risk: PUT /customObjectDefinition/v1/definitions/{ID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  delete_custom_object_definition_v1_custom_object_definition_v1_definitions_id:
    endpoint: DELETE /customObjectDefinition/v1/definitions/{{ record.id }}
    required fields: id
    risk: DELETE /customObjectDefinition/v1/definitions/{ID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_custom_object_definition_v1_custom_object_definition_v1_definitions:
    endpoint: POST /customObjectDefinition/v1/definitions
    required fields: alias, domains, multiInstance, name
    risk: POST /customObjectDefinition/v1/definitions against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  update_custom_object_definition_v1_custom_object_definition_v1_field_types_id_list_values_subresource_id:
    endpoint: PUT /customObjectDefinition/v1/fieldTypes/{{ record.id }}/listValues/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PUT /customObjectDefinition/v1/fieldTypes/{ID}/listValues/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_custom_object_definition_v1_custom_object_definition_v1_definitions_id_condition_rules:
    endpoint: POST /customObjectDefinition/v1/definitions/{{ record.id }}/conditionRules
    required fields: id, conditionItems, ruleDescription
    risk: POST /customObjectDefinition/v1/definitions/{ID}/conditionRules against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  update_custom_object_definition_v1_custom_object_definition_v1_definitions_id_condition_rules_subresource_id:
    endpoint: PUT /customObjectDefinition/v1/definitions/{{ record.id }}/conditionRules/{{ record.subresource_id }}
    required fields: id, subresource_id, conditionItems, ruleDescription
    risk: PUT /customObjectDefinition/v1/definitions/{ID}/conditionRules/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  delete_custom_object_definition_v1_custom_object_definition_v1_definitions_id_condition_rules_subresource_id:
    endpoint: DELETE /customObjectDefinition/v1/definitions/{{ record.id }}/conditionRules/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /customObjectDefinition/v1/definitions/{ID}/conditionRules/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  update_custom_object_definition_v1_custom_object_definition_v1_field_types_id:
    endpoint: PUT /customObjectDefinition/v1/fieldTypes/{{ record.id }}
    required fields: id, alias, name
    risk: PUT /customObjectDefinition/v1/fieldTypes/{ID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_custom_object_definition_v1_custom_object_definition_v1_field_types_id_list_values:
    endpoint: POST /customObjectDefinition/v1/fieldTypes/{{ record.id }}/listValues
    required fields: id
    risk: POST /customObjectDefinition/v1/fieldTypes/{ID}/listValues against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  update_custom_object_definition_v1_custom_object_definition_v1_definitions_id_fields_subresource_id:
    endpoint: PUT /customObjectDefinition/v1/definitions/{{ record.id }}/fields/{{ record.subresource_id }}
    required fields: id, subresource_id, alias, authorizedUsages, categories, fieldType, name
    risk: PUT /customObjectDefinition/v1/definitions/{ID}/fields/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  delete_custom_object_definition_v1_custom_object_definition_v1_definitions_id_fields_subresource_id:
    endpoint: DELETE /customObjectDefinition/v1/definitions/{{ record.id }}/fields/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /customObjectDefinition/v1/definitions/{ID}/fields/{subresourceID} against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_custom_object_definition_v1_custom_object_definition_v1_definitions_id_fields:
    endpoint: POST /customObjectDefinition/v1/definitions/{{ record.id }}/fields
    required fields: id, alias, authorizedUsages, categories, fieldType, name
    risk: POST /customObjectDefinition/v1/definitions/{ID}/fields against Workday REST Custom Object Definition v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customObjectDefinition_v1_20260727_oas2.json
  create_customer_accounts_v1_customer_accounts_v1_payments:
    endpoint: POST /customerAccounts/v1/payments
    risk: POST /customerAccounts/v1/payments against Workday REST Customer Accounts v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customerAccounts_v1_20260727_oas2.json
  create_customer_accounts_v1_customer_accounts_v1_payments_id_remittance_details:
    endpoint: POST /customerAccounts/v1/payments/{{ record.id }}/remittanceDetails
    required fields: id
    risk: POST /customerAccounts/v1/payments/{ID}/remittanceDetails against Workday REST Customer Accounts v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/customerAccounts_v1_20260727_oas2.json
  update_expense_v1_expense_v1_entries_id:
    endpoint: PUT /expense/v1/entries/{{ record.id }}
    required fields: id, json_data
    risk: PUT /expense/v1/entries/{ID} against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  delete_expense_v1_expense_v1_entries_id:
    endpoint: DELETE /expense/v1/entries/{{ record.id }}
    required fields: id
    risk: DELETE /expense/v1/entries/{ID} against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  update_expense_v1_expense_v1_entries_id_2:
    endpoint: PATCH /expense/v1/entries/{{ record.id }}
    required fields: id, json_data
    risk: PATCH /expense/v1/entries/{ID} against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  create_expense_v1_expense_v1_reports:
    endpoint: POST /expense/v1/reports
    risk: POST /expense/v1/reports against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  create_expense_v1_expense_v1_reports_id_lines:
    endpoint: POST /expense/v1/reports/{{ record.id }}/lines
    required fields: id
    risk: POST /expense/v1/reports/{ID}/lines against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  create_expense_v1_expense_v1_reports_id_submit:
    endpoint: POST /expense/v1/reports/{{ record.id }}/submit
    required fields: id
    risk: POST /expense/v1/reports/{ID}/submit against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  create_expense_v1_expense_v1_entries:
    endpoint: POST /expense/v1/entries
    required fields: json_data
    risk: POST /expense/v1/entries against Workday REST Expense v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/expense_v1_20260727_oas2.json
  create_fin_tax_public_v1_fin_tax_public_v1_electronic_reporting_runs:
    endpoint: POST /finTaxPublic/v1/electronicReportingRuns
    risk: POST /finTaxPublic/v1/electronicReportingRuns against Workday REST Fin Tax Public v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/finTaxPublic_v1_20260727_oas2.json
  create_global_payroll_v1_global_payroll_v1_event_driven_integration_vendor_response:
    endpoint: POST /globalPayroll/v1/eventDrivenIntegrationVendorResponse
    required fields: id, overallStatus, relaunchable, setLsrd, skipReview
    risk: POST /globalPayroll/v1/eventDrivenIntegrationVendorResponse against Workday REST Global Payroll v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/globalPayroll_v1_20260727_oas2.json
  create_global_payroll_v1_global_payroll_v1_authorizations:
    endpoint: POST /globalPayroll/v1/authorizations
    required fields: featureId, subjectId, targets
    risk: POST /globalPayroll/v1/authorizations against Workday REST Global Payroll v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/globalPayroll_v1_20260727_oas2.json
  update_global_payroll_v1_global_payroll_v1_pay_groups_id_periods_subresource_id:
    endpoint: PATCH /globalPayroll/v1/payGroups/{{ record.id }}/periods/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /globalPayroll/v1/payGroups/{ID}/periods/{subresourceID} against Workday REST Global Payroll v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/globalPayroll_v1_20260727_oas2.json
  create_global_payroll_v1_global_payroll_v1_notifications:
    endpoint: POST /globalPayroll/v1/notifications
    required fields: message, recipient
    risk: POST /globalPayroll/v1/notifications against Workday REST Global Payroll v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/globalPayroll_v1_20260727_oas2.json
  create_global_payroll_v1_global_payroll_v1_effective_changes:
    endpoint: POST /globalPayroll/v1/effectiveChanges
    risk: POST /globalPayroll/v1/effectiveChanges against Workday REST Global Payroll v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/globalPayroll_v1_20260727_oas2.json
  delete_help_article_v1_help_article_v1_article_versions_id_article_effective_date_subresource_id:
    endpoint: DELETE /helpArticle/v1/articleVersions/{{ record.id }}/articleEffectiveDate/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /helpArticle/v1/articleVersions/{ID}/articleEffectiveDate/{subresourceID} against Workday REST Help Article v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpArticle_v1_20260727_oas2.json
  create_help_article_v1_help_article_v1_article_versions_id_approval_request:
    endpoint: POST /helpArticle/v1/articleVersions/{{ record.id }}/approvalRequest
    required fields: id
    risk: POST /helpArticle/v1/articleVersions/{ID}/approvalRequest against Workday REST Help Article v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpArticle_v1_20260727_oas2.json
  update_help_article_v1_help_article_v1_article_versions_id_approval_decision_subresource_id:
    endpoint: PATCH /helpArticle/v1/articleVersions/{{ record.id }}/approvalDecision/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /helpArticle/v1/articleVersions/{ID}/approvalDecision/{subresourceID} against Workday REST Help Article v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpArticle_v1_20260727_oas2.json
  update_help_article_v1_help_article_v1_article_versions_id_approval_request_subresource_id:
    endpoint: PATCH /helpArticle/v1/articleVersions/{{ record.id }}/approvalRequest/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /helpArticle/v1/articleVersions/{ID}/approvalRequest/{subresourceID} against Workday REST Help Article v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpArticle_v1_20260727_oas2.json
  create_help_article_v1_help_article_v1_article_versions_id_article_effective_date:
    endpoint: POST /helpArticle/v1/articleVersions/{{ record.id }}/articleEffectiveDate
    required fields: id
    risk: POST /helpArticle/v1/articleVersions/{ID}/articleEffectiveDate against Workday REST Help Article v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpArticle_v1_20260727_oas2.json
  update_help_case_v4_help_case_v4_external_records_id:
    endpoint: PUT /helpCase/v4/externalRecords/{{ record.id }}
    required fields: id
    risk: PUT /helpCase/v4/externalRecords/{ID} against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_help_case_v4_help_case_v4_cases_id_reopen:
    endpoint: POST /helpCase/v4/cases/{{ record.id }}/reopen
    required fields: id
    risk: POST /helpCase/v4/cases/{ID}/reopen against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_help_case_v4_help_case_v4_external_creators:
    endpoint: POST /helpCase/v4/externalCreators
    risk: POST /helpCase/v4/externalCreators against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_help_case_v4_help_case_v4_cases:
    endpoint: POST /helpCase/v4/cases
    required fields: json_data
    risk: POST /helpCase/v4/cases against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  delete_help_case_v4_help_case_v4_cases_id_internal_note_timeline_subresource_id:
    endpoint: DELETE /helpCase/v4/cases/{{ record.id }}/internalNoteTimeline/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /helpCase/v4/cases/{ID}/internalNoteTimeline/{subresourceID} against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  update_help_case_v4_help_case_v4_cases_id_internal_note_timeline_subresource_id:
    endpoint: PATCH /helpCase/v4/cases/{{ record.id }}/internalNoteTimeline/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /helpCase/v4/cases/{ID}/internalNoteTimeline/{subresourceID} against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_help_case_v4_help_case_v4_cases_id_comment:
    endpoint: POST /helpCase/v4/cases/{{ record.id }}/comment
    required fields: id, json_data
    risk: POST /helpCase/v4/cases/{ID}/comment against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_help_case_v4_help_case_v4_cases_id_internal_note_timeline:
    endpoint: POST /helpCase/v4/cases/{{ record.id }}/internalNoteTimeline
    required fields: id, json_data
    risk: POST /helpCase/v4/cases/{ID}/internalNoteTimeline against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  update_help_case_v4_help_case_v4_cases_id:
    endpoint: PATCH /helpCase/v4/cases/{{ record.id }}
    required fields: id
    risk: PATCH /helpCase/v4/cases/{ID} against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_help_case_v4_help_case_v4_external_records:
    endpoint: POST /helpCase/v4/externalRecords
    risk: POST /helpCase/v4/externalRecords against Workday REST Help Case v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/helpCase_v4_20260727_oas2.json
  create_journeys_v1_journeys_v1_distribution_requests:
    endpoint: POST /journeys/v1/distributionRequests
    risk: POST /journeys/v1/distributionRequests against Workday REST Journeys v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/journeys_v1_20260727_oas2.json
  create_learning_v1_learning_v1_manage_enrollments:
    endpoint: POST /learning/v1/manageEnrollments
    required fields: content, learner
    risk: POST /learning/v1/manageEnrollments against Workday REST Learning v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/learning_v1_20260727_oas2.json
  create_learning_v1_learning_v1_enrollments_id_lesson_trackings:
    endpoint: POST /learning/v1/enrollments/{{ record.id }}/lessonTrackings
    required fields: id, completionStatus, lessonIdentifier
    risk: POST /learning/v1/enrollments/{ID}/lessonTrackings against Workday REST Learning v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/learning_v1_20260727_oas2.json
  create_learning_v1_learning_v1_manage_digital_courses:
    endpoint: POST /learning/v1/manageDigitalCourses
    required fields: availabilityStatus, description, lessons, title, topics
    risk: POST /learning/v1/manageDigitalCourses against Workday REST Learning v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/learning_v1_20260727_oas2.json
  update_learning_v1_learning_v1_enrollments_id:
    endpoint: PATCH /learning/v1/enrollments/{{ record.id }}
    required fields: id, attendanceStatus
    risk: PATCH /learning/v1/enrollments/{ID} against Workday REST Learning v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/learning_v1_20260727_oas2.json
  update_learning_v1_learning_v1_enrollments_id_lesson_trackings_subresource_id:
    endpoint: PATCH /learning/v1/enrollments/{{ record.id }}/lessonTrackings/{{ record.subresource_id }}
    required fields: id, subresource_id, completionStatus, lessonIdentifier
    risk: PATCH /learning/v1/enrollments/{ID}/lessonTrackings/{subresourceID} against Workday REST Learning v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/learning_v1_20260727_oas2.json
  delete_payroll_v2_payroll_v2_payroll_inputs_id:
    endpoint: DELETE /payroll/v2/payrollInputs/{{ record.id }}
    required fields: id
    risk: DELETE /payroll/v2/payrollInputs/{ID} against Workday REST Payroll v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/payroll_v2_20260727_oas2.json
  update_payroll_v2_payroll_v2_payroll_inputs_id:
    endpoint: PATCH /payroll/v2/payrollInputs/{{ record.id }}
    required fields: id
    risk: PATCH /payroll/v2/payrollInputs/{ID} against Workday REST Payroll v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/payroll_v2_20260727_oas2.json
  create_payroll_v2_payroll_v2_payroll_inputs:
    endpoint: POST /payroll/v2/payrollInputs
    risk: POST /payroll/v2/payrollInputs against Workday REST Payroll v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/payroll_v2_20260727_oas2.json
  create_payroll_v2_payroll_v2_tax_rates:
    endpoint: POST /payroll/v2/taxRates
    required fields: companyInstance, startDate, stateInstance, taxCode
    risk: POST /payroll/v2/taxRates against Workday REST Payroll v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/payroll_v2_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_worker_goal_events_id_submit:
    endpoint: POST /performanceEnablement/v5/workerGoalEvents/{{ record.id }}/submit
    required fields: id, json_data
    risk: POST /performanceEnablement/v5/workerGoalEvents/{ID}/submit against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_workers_id_requested_feedback_on_self_events:
    endpoint: POST /performanceEnablement/v5/workers/{{ record.id }}/requestedFeedbackOnSelfEvents
    required fields: id, json_data
    risk: POST /performanceEnablement/v5/workers/{ID}/requestedFeedbackOnSelfEvents against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_organization_goals:
    endpoint: POST /performanceEnablement/v5/organizationGoals
    required fields: goalName, organization, planPeriod
    risk: POST /performanceEnablement/v5/organizationGoals against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_workers_id_requested_feedback_on_worker_events:
    endpoint: POST /performanceEnablement/v5/workers/{{ record.id }}/requestedFeedbackOnWorkerEvents
    required fields: id, json_data
    risk: POST /performanceEnablement/v5/workers/{ID}/requestedFeedbackOnWorkerEvents against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_cascaded_goal_event:
    endpoint: POST /performanceEnablement/v5/cascadedGoalEvent
    risk: POST /performanceEnablement/v5/cascadedGoalEvent against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_worker_goal_events:
    endpoint: POST /performanceEnablement/v5/workerGoalEvents
    required fields: worker
    risk: POST /performanceEnablement/v5/workerGoalEvents against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_worker_goal_events_id_goals:
    endpoint: POST /performanceEnablement/v5/workerGoalEvents/{{ record.id }}/goals
    required fields: id, name
    risk: POST /performanceEnablement/v5/workerGoalEvents/{ID}/goals against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  update_performance_enablement_v5_performance_enablement_v5_workers_id_development_items_subresource_id:
    endpoint: PATCH /performanceEnablement/v5/workers/{{ record.id }}/developmentItems/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /performanceEnablement/v5/workers/{ID}/developmentItems/{subresourceID} against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  update_performance_enablement_v5_performance_enablement_v5_organization_goals_id:
    endpoint: PATCH /performanceEnablement/v5/organizationGoals/{{ record.id }}
    required fields: id, goalName, organization, planPeriod
    risk: PATCH /performanceEnablement/v5/organizationGoals/{ID} against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_my_performance_reviews:
    endpoint: POST /performanceEnablement/v5/myPerformanceReviews
    required fields: businessProcessParameters, reviewTemplate
    risk: POST /performanceEnablement/v5/myPerformanceReviews against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_performance_enablement_v5_performance_enablement_v5_workers_id_anytime_feedback_events:
    endpoint: POST /performanceEnablement/v5/workers/{{ record.id }}/anytimeFeedbackEvents
    required fields: id, json_data
    risk: POST /performanceEnablement/v5/workers/{ID}/anytimeFeedbackEvents against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  update_performance_enablement_v5_performance_enablement_v5_give_requested_feedback_events_id:
    endpoint: PATCH /performanceEnablement/v5/giveRequestedFeedbackEvents/{{ record.id }}
    required fields: id, json_data
    risk: PATCH /performanceEnablement/v5/giveRequestedFeedbackEvents/{ID} against Workday REST Performance Enablement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/performanceEnablement_v5_20260727_oas2.json
  create_person_v4_person_v4_phone_validation:
    endpoint: POST /person/v4/phoneValidation
    risk: POST /person/v4/phoneValidation against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_home_contact_information_changes_id_web_addresses:
    endpoint: POST /person/v4/homeContactInformationChanges/{{ record.id }}/webAddresses
    required fields: id
    risk: POST /person/v4/homeContactInformationChanges/{ID}/webAddresses against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_work_contact_information_changes_id_phone_numbers_subresource_id:
    endpoint: DELETE /person/v4/workContactInformationChanges/{{ record.id }}/phoneNumbers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/workContactInformationChanges/{ID}/phoneNumbers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_work_contact_information_changes_id_phone_numbers_subresource_id:
    endpoint: PATCH /person/v4/workContactInformationChanges/{{ record.id }}/phoneNumbers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/workContactInformationChanges/{ID}/phoneNumbers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_work_contact_information_changes_id_instant_messengers_subresource_id:
    endpoint: DELETE /person/v4/workContactInformationChanges/{{ record.id }}/instantMessengers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/workContactInformationChanges/{ID}/instantMessengers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_work_contact_information_changes_id_instant_messengers_subresource_id:
    endpoint: PATCH /person/v4/workContactInformationChanges/{{ record.id }}/instantMessengers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/workContactInformationChanges/{ID}/instantMessengers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_home_contact_information_changes_id_phone_numbers:
    endpoint: POST /person/v4/homeContactInformationChanges/{{ record.id }}/phoneNumbers
    required fields: id
    risk: POST /person/v4/homeContactInformationChanges/{ID}/phoneNumbers against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_work_contact_information_changes_id_submit:
    endpoint: POST /person/v4/workContactInformationChanges/{{ record.id }}/submit
    required fields: id, json_data
    risk: POST /person/v4/workContactInformationChanges/{ID}/submit against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_work_contact_information_changes_id:
    endpoint: PATCH /person/v4/workContactInformationChanges/{{ record.id }}
    required fields: id
    risk: PATCH /person/v4/workContactInformationChanges/{ID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_work_contact_information_changes_id_instant_messengers:
    endpoint: POST /person/v4/workContactInformationChanges/{{ record.id }}/instantMessengers
    required fields: id
    risk: POST /person/v4/workContactInformationChanges/{ID}/instantMessengers against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_home_contact_information_changes_id_email_addresses:
    endpoint: POST /person/v4/homeContactInformationChanges/{{ record.id }}/emailAddresses
    required fields: id
    risk: POST /person/v4/homeContactInformationChanges/{ID}/emailAddresses against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_home_contact_information_changes_id_addresses:
    endpoint: POST /person/v4/homeContactInformationChanges/{{ record.id }}/addresses
    required fields: id
    risk: POST /person/v4/homeContactInformationChanges/{ID}/addresses against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_home_contact_information_changes_id_instant_messengers_subresource_id:
    endpoint: DELETE /person/v4/homeContactInformationChanges/{{ record.id }}/instantMessengers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/homeContactInformationChanges/{ID}/instantMessengers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_home_contact_information_changes_id_instant_messengers_subresource_id:
    endpoint: PATCH /person/v4/homeContactInformationChanges/{{ record.id }}/instantMessengers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/homeContactInformationChanges/{ID}/instantMessengers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_home_contact_information_changes_id_web_addresses_subresource_id:
    endpoint: DELETE /person/v4/homeContactInformationChanges/{{ record.id }}/webAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/homeContactInformationChanges/{ID}/webAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_home_contact_information_changes_id_web_addresses_subresource_id:
    endpoint: PATCH /person/v4/homeContactInformationChanges/{{ record.id }}/webAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/homeContactInformationChanges/{ID}/webAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_home_contact_information_changes_id_phone_numbers_subresource_id:
    endpoint: DELETE /person/v4/homeContactInformationChanges/{{ record.id }}/phoneNumbers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/homeContactInformationChanges/{ID}/phoneNumbers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_home_contact_information_changes_id_phone_numbers_subresource_id:
    endpoint: PATCH /person/v4/homeContactInformationChanges/{{ record.id }}/phoneNumbers/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/homeContactInformationChanges/{ID}/phoneNumbers/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_home_contact_information_changes_id_submit:
    endpoint: POST /person/v4/homeContactInformationChanges/{{ record.id }}/submit
    required fields: id, json_data
    risk: POST /person/v4/homeContactInformationChanges/{ID}/submit against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_home_contact_information_changes_id_instant_messengers:
    endpoint: POST /person/v4/homeContactInformationChanges/{{ record.id }}/instantMessengers
    required fields: id
    risk: POST /person/v4/homeContactInformationChanges/{ID}/instantMessengers against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_work_contact_information_changes_id_web_addresses_subresource_id:
    endpoint: DELETE /person/v4/workContactInformationChanges/{{ record.id }}/webAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/workContactInformationChanges/{ID}/webAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_work_contact_information_changes_id_web_addresses_subresource_id:
    endpoint: PATCH /person/v4/workContactInformationChanges/{{ record.id }}/webAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/workContactInformationChanges/{ID}/webAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_home_contact_information_changes_id_email_addresses_subresource_id:
    endpoint: DELETE /person/v4/homeContactInformationChanges/{{ record.id }}/emailAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/homeContactInformationChanges/{ID}/emailAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_home_contact_information_changes_id_email_addresses_subresource_id:
    endpoint: PATCH /person/v4/homeContactInformationChanges/{{ record.id }}/emailAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/homeContactInformationChanges/{ID}/emailAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_work_contact_information_changes_id_phone_numbers:
    endpoint: POST /person/v4/workContactInformationChanges/{{ record.id }}/phoneNumbers
    required fields: id
    risk: POST /person/v4/workContactInformationChanges/{ID}/phoneNumbers against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_work_contact_information_changes_id_email_addresses:
    endpoint: POST /person/v4/workContactInformationChanges/{{ record.id }}/emailAddresses
    required fields: id
    risk: POST /person/v4/workContactInformationChanges/{ID}/emailAddresses against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_work_contact_information_changes_id_email_addresses_subresource_id:
    endpoint: DELETE /person/v4/workContactInformationChanges/{{ record.id }}/emailAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/workContactInformationChanges/{ID}/emailAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_work_contact_information_changes_id_email_addresses_subresource_id:
    endpoint: PATCH /person/v4/workContactInformationChanges/{{ record.id }}/emailAddresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /person/v4/workContactInformationChanges/{ID}/emailAddresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_work_contact_information_changes_id_web_addresses:
    endpoint: POST /person/v4/workContactInformationChanges/{{ record.id }}/webAddresses
    required fields: id
    risk: POST /person/v4/workContactInformationChanges/{ID}/webAddresses against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_work_contact_information_changes_id_addresses_subresource_id:
    endpoint: PUT /person/v4/workContactInformationChanges/{{ record.id }}/addresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PUT /person/v4/workContactInformationChanges/{ID}/addresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_work_contact_information_changes_id_addresses_subresource_id:
    endpoint: DELETE /person/v4/workContactInformationChanges/{{ record.id }}/addresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/workContactInformationChanges/{ID}/addresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  update_person_v4_person_v4_home_contact_information_changes_id_addresses_subresource_id:
    endpoint: PUT /person/v4/homeContactInformationChanges/{{ record.id }}/addresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PUT /person/v4/homeContactInformationChanges/{ID}/addresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  delete_person_v4_person_v4_home_contact_information_changes_id_addresses_subresource_id:
    endpoint: DELETE /person/v4/homeContactInformationChanges/{{ record.id }}/addresses/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /person/v4/homeContactInformationChanges/{ID}/addresses/{subresourceID} against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_person_v4_person_v4_work_contact_information_changes_id_addresses:
    endpoint: POST /person/v4/workContactInformationChanges/{{ record.id }}/addresses
    required fields: id
    risk: POST /person/v4/workContactInformationChanges/{ID}/addresses against Workday REST Person v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/person_v4_20260727_oas2.json
  create_prism_analytics_v3_tables:
    endpoint: POST /tables
    risk: POST /tables against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  update_prism_analytics_v3_tables_id:
    endpoint: PUT /tables/{{ record.id }}
    required fields: id
    risk: PUT /tables/{id} against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  update_prism_analytics_v3_tables_id_2:
    endpoint: PATCH /tables/{{ record.id }}
    required fields: id
    risk: PATCH /tables/{id} against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  create_prism_analytics_v3_buckets:
    endpoint: POST /buckets
    risk: POST /buckets against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  update_prism_analytics_v3_buckets_id:
    endpoint: PUT /buckets/{{ record.id }}
    required fields: id
    risk: PUT /buckets/{id} against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  create_prism_analytics_v3_buckets_id_complete:
    endpoint: POST /buckets/{{ record.id }}/complete
    required fields: id
    risk: POST /buckets/{id}/complete against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  create_prism_analytics_v3_data_changes_data_change_id_activities:
    endpoint: POST /dataChanges/{{ record.data_change_id }}/activities
    required fields: data_change_id
    risk: POST /dataChanges/{dataChangeID}/activities against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  create_prism_analytics_v3_data_changes_data_change_id_cancel_activity_id:
    endpoint: POST /dataChanges/{{ record.data_change_id }}/cancel/{{ record.activity_id }}
    required fields: data_change_id, activity_id
    risk: POST /dataChanges/{dataChangeID}/cancel/{activityID} against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  create_prism_analytics_v3_file_containers:
    endpoint: POST /fileContainers
    risk: POST /fileContainers against Workday REST Prism Analytics v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/prismAnalytics_v3_20231120_oas3.json
  delete_procurement_v5_procurement_v5_requisitions_id_requisition_lines_subresource_id:
    endpoint: DELETE /procurement/v5/requisitions/{{ record.id }}/requisitionLines/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /procurement/v5/requisitions/{ID}/requisitionLines/{subresourceID} against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  update_procurement_v5_procurement_v5_requisitions_id_requisition_lines_subresource_id:
    endpoint: PATCH /procurement/v5/requisitions/{{ record.id }}/requisitionLines/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /procurement/v5/requisitions/{ID}/requisitionLines/{subresourceID} against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  create_procurement_v5_procurement_v5_requisitions_id_close:
    endpoint: POST /procurement/v5/requisitions/{{ record.id }}/close
    required fields: id
    risk: POST /procurement/v5/requisitions/{ID}/close against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  create_procurement_v5_procurement_v5_requisitions:
    endpoint: POST /procurement/v5/requisitions
    risk: POST /procurement/v5/requisitions against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  create_procurement_v5_procurement_v5_requisitions_id_cancel:
    endpoint: POST /procurement/v5/requisitions/{{ record.id }}/cancel
    required fields: id
    risk: POST /procurement/v5/requisitions/{ID}/cancel against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  update_procurement_v5_procurement_v5_requisitions_id:
    endpoint: PATCH /procurement/v5/requisitions/{{ record.id }}
    required fields: id
    risk: PATCH /procurement/v5/requisitions/{ID} against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  create_procurement_v5_procurement_v5_requisitions_id_requisition_events:
    endpoint: POST /procurement/v5/requisitions/{{ record.id }}/requisitionEvents
    required fields: id
    risk: POST /procurement/v5/requisitions/{ID}/requisitionEvents against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  create_procurement_v5_procurement_v5_requisitions_id_requisition_lines:
    endpoint: POST /procurement/v5/requisitions/{{ record.id }}/requisitionLines
    required fields: id
    risk: POST /procurement/v5/requisitions/{ID}/requisitionLines against Workday REST Procurement v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/procurement_v5_20260727_oas2.json
  create_projects_v3_projects_v3_resource_forecast_lines_id_allocations:
    endpoint: POST /projects/v3/resourceForecastLines/{{ record.id }}/allocations
    required fields: id, date, forecastedHours
    risk: POST /projects/v3/resourceForecastLines/{ID}/allocations against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_resource_plan_lines_id_edit:
    endpoint: POST /projects/v3/resourcePlanLines/{{ record.id }}/edit
    required fields: id
    risk: POST /projects/v3/resourcePlanLines/{ID}/edit against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_projects:
    endpoint: POST /projects/v3/projects
    risk: POST /projects/v3/projects against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_ad_hoc_project_time_transactions:
    endpoint: POST /projects/v3/adHocProjectTimeTransactions
    risk: POST /projects/v3/adHocProjectTimeTransactions against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  update_projects_v3_projects_v3_resource_forecast_lines_id_allocations_subresource_id:
    endpoint: PATCH /projects/v3/resourceForecastLines/{{ record.id }}/allocations/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /projects/v3/resourceForecastLines/{ID}/allocations/{subresourceID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_plan_tasks:
    endpoint: POST /projects/v3/planTasks
    risk: POST /projects/v3/planTasks against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_projects_id_edit:
    endpoint: POST /projects/v3/projects/{{ record.id }}/edit
    required fields: id
    risk: POST /projects/v3/projects/{ID}/edit against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_task_resources:
    endpoint: POST /projects/v3/taskResources
    risk: POST /projects/v3/taskResources against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_resource_plan_lines:
    endpoint: POST /projects/v3/resourcePlanLines
    risk: POST /projects/v3/resourcePlanLines against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  delete_projects_v3_projects_v3_task_resources_id:
    endpoint: DELETE /projects/v3/taskResources/{{ record.id }}
    required fields: id
    risk: DELETE /projects/v3/taskResources/{ID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  update_projects_v3_projects_v3_task_resources_id:
    endpoint: PATCH /projects/v3/taskResources/{{ record.id }}
    required fields: id
    risk: PATCH /projects/v3/taskResources/{ID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  delete_projects_v3_projects_v3_plan_phases_id:
    endpoint: DELETE /projects/v3/planPhases/{{ record.id }}
    required fields: id
    risk: DELETE /projects/v3/planPhases/{ID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  update_projects_v3_projects_v3_plan_phases_id:
    endpoint: PATCH /projects/v3/planPhases/{{ record.id }}
    required fields: id
    risk: PATCH /projects/v3/planPhases/{ID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  delete_projects_v3_projects_v3_plan_tasks_id:
    endpoint: DELETE /projects/v3/planTasks/{{ record.id }}
    required fields: id
    risk: DELETE /projects/v3/planTasks/{ID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  update_projects_v3_projects_v3_plan_tasks_id:
    endpoint: PATCH /projects/v3/planTasks/{{ record.id }}
    required fields: id
    risk: PATCH /projects/v3/planTasks/{ID} against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_projects_v3_projects_v3_plan_phases:
    endpoint: POST /projects/v3/planPhases
    risk: POST /projects/v3/planPhases against Workday REST Projects v3 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/projects_v3_20260727_oas2.json
  create_recruiting_v4_recruiting_v4_prospects:
    endpoint: POST /recruiting/v4/prospects
    risk: POST /recruiting/v4/prospects against Workday REST Recruiting v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/recruiting_v4_20260727_oas2.json
  create_recruiting_v4_recruiting_v4_prospects_id_languages:
    endpoint: POST /recruiting/v4/prospects/{{ record.id }}/languages
    required fields: id
    risk: POST /recruiting/v4/prospects/{ID}/languages against Workday REST Recruiting v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/recruiting_v4_20260727_oas2.json
  create_recruiting_v4_recruiting_v4_prospects_id_skills:
    endpoint: POST /recruiting/v4/prospects/{{ record.id }}/skills
    required fields: id
    risk: POST /recruiting/v4/prospects/{ID}/skills against Workday REST Recruiting v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/recruiting_v4_20260727_oas2.json
  create_recruiting_v4_recruiting_v4_prospects_id_educations:
    endpoint: POST /recruiting/v4/prospects/{{ record.id }}/educations
    required fields: id
    risk: POST /recruiting/v4/prospects/{ID}/educations against Workday REST Recruiting v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/recruiting_v4_20260727_oas2.json
  create_recruiting_v4_recruiting_v4_prospects_id_experiences:
    endpoint: POST /recruiting/v4/prospects/{{ record.id }}/experiences
    required fields: id
    risk: POST /recruiting/v4/prospects/{ID}/experiences against Workday REST Recruiting v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/recruiting_v4_20260727_oas2.json
  create_recruiting_v4_recruiting_v4_interviews_id_feedback:
    endpoint: POST /recruiting/v4/interviews/{{ record.id }}/feedback
    required fields: id
    risk: POST /recruiting/v4/interviews/{ID}/feedback against Workday REST Recruiting v4 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/recruiting_v4_20260727_oas2.json
  create_request_v2_request_v2_requests_id_close:
    endpoint: POST /request/v2/requests/{{ record.id }}/close
    required fields: id
    risk: POST /request/v2/requests/{ID}/close against Workday REST Request v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/request_v2_20260727_oas2.json
  create_request_v2_request_v2_requests:
    endpoint: POST /request/v2/requests
    required fields: json_data
    risk: POST /request/v2/requests against Workday REST Request v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/request_v2_20260727_oas2.json
  update_revenue_v1_revenue_v1_billable_transactions_id:
    endpoint: PATCH /revenue/v1/billableTransactions/{{ record.id }}
    required fields: id
    risk: PATCH /revenue/v1/billableTransactions/{ID} against Workday REST Revenue v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/revenue_v1_20260727_oas2.json
  create_skill_v1_skill_v1_ml_skills:
    endpoint: POST /skill/v1/mlSkills
    required fields: skillID
    risk: POST /skill/v1/mlSkills against Workday REST Skill v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/skill_v1_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_organization_assignment_changes:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/organizationAssignmentChanges
    required fields: id, date
    risk: POST /staffing/v7/workers/{ID}/organizationAssignmentChanges against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_skill_items:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/skillItems
    required fields: id
    risk: POST /staffing/v7/workers/{ID}/skillItems against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_start_details_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/startDetails/{{ record.subresource_id }}
    required fields: id, subresource_id, date
    risk: PATCH /staffing/v7/jobChanges/{ID}/startDetails/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_work_contact_information_changes:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/workContactInformationChanges
    required fields: id
    risk: POST /staffing/v7/workers/{ID}/workContactInformationChanges against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_job_profile_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/jobProfile/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/jobProfile/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_location_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/location/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/location/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_check_in_topics:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/checkInTopics
    required fields: id, json_data
    risk: POST /staffing/v7/workers/{ID}/checkInTopics against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_custom_organizations_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/customOrganizations/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/customOrganizations/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_business_unit_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/businessUnit/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/businessUnit/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_workers_id_external_skill_level_subresource_id:
    endpoint: PUT /staffing/v7/workers/{{ record.id }}/externalSkillLevel/{{ record.subresource_id }}
    required fields: id, subresource_id, externalSkillId, externalSkillLevel
    risk: PUT /staffing/v7/workers/{ID}/externalSkillLevel/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_start_details_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/startDetails/{{ record.subresource_id }}
    required fields: id, subresource_id, position
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/startDetails/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_move_team_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/moveTeam/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/moveTeam/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_home_contact_information_changes:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/homeContactInformationChanges
    required fields: id
    risk: POST /staffing/v7/workers/{ID}/homeContactInformationChanges against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_position_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/position/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/position/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_business_title_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/businessTitle/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/businessTitle/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  delete_staffing_v7_staffing_v7_workers_id_check_ins_subresource_id:
    endpoint: DELETE /staffing/v7/workers/{{ record.id }}/checkIns/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /staffing/v7/workers/{ID}/checkIns/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_workers_id_check_ins_subresource_id:
    endpoint: PATCH /staffing/v7/workers/{{ record.id }}/checkIns/{{ record.subresource_id }}
    required fields: id, subresource_id, json_data
    risk: PATCH /staffing/v7/workers/{ID}/checkIns/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_administrative_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/administrative/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/administrative/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_explicit_skills:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/explicitSkills
    required fields: id
    risk: POST /staffing/v7/workers/{ID}/explicitSkills against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_company_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/company/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/company/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_costing_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/costing/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/costing/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  delete_staffing_v7_staffing_v7_workers_id_check_in_topics_subresource_id:
    endpoint: DELETE /staffing/v7/workers/{{ record.id }}/checkInTopics/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /staffing/v7/workers/{ID}/checkInTopics/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_workers_id_check_in_topics_subresource_id_2:
    endpoint: PATCH /staffing/v7/workers/{{ record.id }}/checkInTopics/{{ record.subresource_id }}
    required fields: id, subresource_id, json_data
    risk: PATCH /staffing/v7/workers/{ID}/checkInTopics/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_region_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/region/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/region/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_check_ins:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/checkIns
    required fields: id, json_data
    risk: POST /staffing/v7/workers/{ID}/checkIns against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_job_changes_id_submit:
    endpoint: POST /staffing/v7/jobChanges/{{ record.id }}/submit
    required fields: id, json_data
    risk: POST /staffing/v7/jobChanges/{ID}/submit against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_job_classification_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/jobClassification/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/jobClassification/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_job_changes:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/jobChanges
    required fields: id, date
    risk: POST /staffing/v7/workers/{ID}/jobChanges against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_opening_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/opening/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/opening/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_organization_assignment_changes_id_submit:
    endpoint: POST /staffing/v7/organizationAssignmentChanges/{{ record.id }}/submit
    required fields: id, json_data
    risk: POST /staffing/v7/organizationAssignmentChanges/{ID}/submit against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_cost_center_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/costCenter/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/costCenter/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_organization_assignment_changes_id_comment_subresource_id:
    endpoint: PATCH /staffing/v7/organizationAssignmentChanges/{{ record.id }}/comment/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/organizationAssignmentChanges/{ID}/comment/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_contract_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/contract/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/contract/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_workers_id_external_skill_level:
    endpoint: POST /staffing/v7/workers/{{ record.id }}/externalSkillLevel
    required fields: id, externalSkillId, externalSkillLevel
    risk: POST /staffing/v7/workers/{ID}/externalSkillLevel against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_staffing_v7_staffing_v7_organization_assignment_changes:
    endpoint: POST /staffing/v7/organizationAssignmentChanges
    required fields: date, position
    risk: POST /staffing/v7/organizationAssignmentChanges against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  update_staffing_v7_staffing_v7_job_changes_id_comment_subresource_id:
    endpoint: PATCH /staffing/v7/jobChanges/{{ record.id }}/comment/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /staffing/v7/jobChanges/{ID}/comment/{subresourceID} against Workday REST Staffing v7 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/staffing_v7_20260727_oas2.json
  create_student_advising_v1_student_advising_v1_hypothetical_academic_progress:
    endpoint: POST /studentAdvising/v1/hypotheticalAcademicProgress
    required fields: programOfStudy, startDate, student
    risk: POST /studentAdvising/v1/hypotheticalAcademicProgress against Workday REST Student Advising v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentAdvising_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_immigration_events_id_submit:
    endpoint: POST /studentCore/v1/immigrationEvents/{{ record.id }}/submit
    required fields: id
    risk: POST /studentCore/v1/immigrationEvents/{ID}/submit against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  update_student_core_v1_student_core_v1_immigration_events_id_immigration_data_subresource_id:
    endpoint: PUT /studentCore/v1/immigrationEvents/{{ record.id }}/immigrationData/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PUT /studentCore/v1/immigrationEvents/{ID}/immigrationData/{subresourceID} against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_holds_id_override_hold:
    endpoint: POST /studentCore/v1/holds/{{ record.id }}/overrideHold
    required fields: id, endDate, holdTypes, startDate
    risk: POST /studentCore/v1/holds/{ID}/overrideHold against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_students_id_apply_hold:
    endpoint: POST /studentCore/v1/students/{{ record.id }}/applyHold
    required fields: id, reason, typeContext
    risk: POST /studentCore/v1/students/{ID}/applyHold against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_immigration_events_id_immigration_data:
    endpoint: POST /studentCore/v1/immigrationEvents/{{ record.id }}/immigrationData
    required fields: id
    risk: POST /studentCore/v1/immigrationEvents/{ID}/immigrationData against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_holds_id_update_hold:
    endpoint: POST /studentCore/v1/holds/{{ record.id }}/updateHold
    required fields: id, typeContexts
    risk: POST /studentCore/v1/holds/{ID}/updateHold against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  update_student_core_v1_student_core_v1_immigration_events_id_dependent_immigration_data_subresource_id:
    endpoint: PUT /studentCore/v1/immigrationEvents/{{ record.id }}/dependentImmigrationData/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PUT /studentCore/v1/immigrationEvents/{ID}/dependentImmigrationData/{subresourceID} against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  delete_student_core_v1_student_core_v1_immigration_events_id_dependent_immigration_data_subresource_id:
    endpoint: DELETE /studentCore/v1/immigrationEvents/{{ record.id }}/dependentImmigrationData/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /studentCore/v1/immigrationEvents/{ID}/dependentImmigrationData/{subresourceID} against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_immigration_events_id_dependent_immigration_data:
    endpoint: POST /studentCore/v1/immigrationEvents/{{ record.id }}/dependentImmigrationData
    required fields: id
    risk: POST /studentCore/v1/immigrationEvents/{ID}/dependentImmigrationData against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  create_student_core_v1_student_core_v1_holds_id_remove_hold:
    endpoint: POST /studentCore/v1/holds/{{ record.id }}/removeHold
    required fields: id
    risk: POST /studentCore/v1/holds/{ID}/removeHold against Workday REST Student Core v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentCore_v1_20260727_oas2.json
  update_student_finance_v1_student_finance_v1_students_id_payments_subresource_id:
    endpoint: PATCH /studentFinance/v1/students/{{ record.id }}/payments/{{ record.subresource_id }}
    required fields: id, subresource_id, reason
    risk: PATCH /studentFinance/v1/students/{ID}/payments/{subresourceID} against Workday REST Student Finance v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentFinance_v1_20260727_oas2.json
  create_student_finance_v1_student_finance_v1_students_id_payments:
    endpoint: POST /studentFinance/v1/students/{{ record.id }}/payments
    required fields: id, academicPeriod, amount, institutionalAcademicUnit, paymentType
    risk: POST /studentFinance/v1/students/{ID}/payments against Workday REST Student Finance v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentFinance_v1_20260727_oas2.json
  create_student_recruiting_v1_student_recruiting_v1_academic_requirement_evaluation:
    endpoint: POST /studentRecruiting/v1/academicRequirementEvaluation
    required fields: programOfStudy, startDate
    risk: POST /studentRecruiting/v1/academicRequirementEvaluation against Workday REST Student Recruiting v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/studentRecruiting_v1_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_succession_plan_events_id_submit:
    endpoint: POST /talentManagement/v2/successionPlanEvents/{{ record.id }}/submit
    required fields: id, json_data
    risk: POST /talentManagement/v2/successionPlanEvents/{ID}/submit against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_succession_plan_events_id_candidates:
    endpoint: POST /talentManagement/v2/successionPlanEvents/{{ record.id }}/candidates
    required fields: id, readiness, successionPlanCandidate
    risk: POST /talentManagement/v2/successionPlanEvents/{ID}/candidates against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_create_mentorship_for_worker:
    endpoint: POST /talentManagement/v2/createMentorshipForWorker
    required fields: mentee, mentor, mentorType, startDate
    risk: POST /talentManagement/v2/createMentorshipForWorker against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  delete_talent_management_v2_talent_management_v2_succession_plan_events_id_candidates_subresource_id:
    endpoint: DELETE /talentManagement/v2/successionPlanEvents/{{ record.id }}/candidates/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /talentManagement/v2/successionPlanEvents/{ID}/candidates/{subresourceID} against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  update_talent_management_v2_talent_management_v2_succession_plan_events_id_candidates_subresource_id:
    endpoint: PATCH /talentManagement/v2/successionPlanEvents/{{ record.id }}/candidates/{{ record.subresource_id }}
    required fields: id, subresource_id, readiness, successionPlanCandidate
    risk: PATCH /talentManagement/v2/successionPlanEvents/{ID}/candidates/{subresourceID} against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_mentorships_id_close:
    endpoint: POST /talentManagement/v2/mentorships/{{ record.id }}/close
    required fields: id, closeMentorshipReason, endDate, startDate
    risk: POST /talentManagement/v2/mentorships/{ID}/close against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_mentorships_id_edit:
    endpoint: POST /talentManagement/v2/mentorships/{{ record.id }}/edit
    required fields: id, startDate
    risk: POST /talentManagement/v2/mentorships/{ID}/edit against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_succession_plans:
    endpoint: POST /talentManagement/v2/successionPlans
    required fields: position
    risk: POST /talentManagement/v2/successionPlans against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_succession_plan_events:
    endpoint: POST /talentManagement/v2/successionPlanEvents
    required fields: json_data
    risk: POST /talentManagement/v2/successionPlanEvents against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_talent_management_v2_talent_management_v2_create_mentorship_for_me:
    endpoint: POST /talentManagement/v2/createMentorshipForMe
    required fields: mentor, mentorType, startDate
    risk: POST /talentManagement/v2/createMentorshipForMe against Workday REST Talent Management v2 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/talentManagement_v2_20260727_oas2.json
  create_time_tracking_v5_time_tracking_v5_worker_time_attestation:
    endpoint: POST /timeTracking/v5/workerTimeAttestation
    risk: POST /timeTracking/v5/workerTimeAttestation against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  create_time_tracking_v5_time_tracking_v5_workers_id_worker_time_block:
    endpoint: POST /timeTracking/v5/workers/{{ record.id }}/workerTimeBlock
    required fields: id
    risk: POST /timeTracking/v5/workers/{ID}/workerTimeBlock against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  delete_time_tracking_v5_time_tracking_v5_workers_id_worker_time_block_subresource_id:
    endpoint: DELETE /timeTracking/v5/workers/{{ record.id }}/workerTimeBlock/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: DELETE /timeTracking/v5/workers/{ID}/workerTimeBlock/{subresourceID} against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  update_time_tracking_v5_time_tracking_v5_workers_id_worker_time_block_subresource_id:
    endpoint: PATCH /timeTracking/v5/workers/{{ record.id }}/workerTimeBlock/{{ record.subresource_id }}
    required fields: id, subresource_id
    risk: PATCH /timeTracking/v5/workers/{ID}/workerTimeBlock/{subresourceID} against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  create_time_tracking_v5_time_tracking_v5_workers_id_time_review_events:
    endpoint: POST /timeTracking/v5/workers/{{ record.id }}/timeReviewEvents
    required fields: id
    risk: POST /timeTracking/v5/workers/{ID}/timeReviewEvents against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  create_time_tracking_v5_time_tracking_v5_time_clock_events:
    endpoint: POST /timeTracking/v5/timeClockEvents
    risk: POST /timeTracking/v5/timeClockEvents against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  update_time_tracking_v5_time_tracking_v5_time_clock_events_id:
    endpoint: PUT /timeTracking/v5/timeClockEvents/{{ record.id }}
    required fields: id
    risk: PUT /timeTracking/v5/timeClockEvents/{ID} against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  delete_time_tracking_v5_time_tracking_v5_time_clock_events_id:
    endpoint: DELETE /timeTracking/v5/timeClockEvents/{{ record.id }}
    required fields: id
    risk: DELETE /timeTracking/v5/timeClockEvents/{ID} against Workday REST Time Tracking v5 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/timeTracking_v5_20260727_oas2.json
  create_worktag_v1_worktag_v1_validate_worktags:
    endpoint: POST /worktag/v1/validateWorktags
    risk: POST /worktag/v1/validateWorktags against Workday REST Worktag v1 mutates tenant data; use reverse ETL plan, preview, explicit approval, and least-privilege Workday scopes. Source: https://community.workday.com/sites/default/files/file-hosting/restapi/worktag_v1_20260727_oas2.json

SECURITY
  read risk: external Workday REST reads can include HR, recruiting, procurement, and financial tenant data; fixtures are synthetic and live reads require caller-provided credentials
  write risk: typed Workday REST reverse ETL actions can create, update, submit, or delete tenant business objects; every execution must go through plan, preview, explicit approval, and least-privilege Workday scopes
  approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive deletes require destructive confirmation and treat documented 404 as idempotent missing only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Workday REST provider commands are definition-owned and bounded; no raw API passthrough is available.
  Usage: pm workday-rest <command> [flags]
  Source CLI: Workday REST API Directory (https://community.workday.com/sites/default/files/file-hosting/restapi/index.html)
  Bounded direct reads for Workday values/search endpoints
    direct_absence_management_v5_absence_management_v5_values_leave_status - The status of the event that tracks the requested time off. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_absence_management_v5_absence_management_v5_values_time_off_status - The possible statuses for a worker's time off entry request event. They include: * Approved, 0391102bd1b542538d996936c8fa2fa7 * Not Submitted, eddf5968e6d4430ca4bce9a4cfaba337 * Se [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_asor_v1_asor_v1_agent_resource_search_id - Agent Resource Search APi [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB; flags: --id
    direct_asor_v1_asor_v1_agent_resource_search - Agent Resource Search APi [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB; flags: --operation-type, --protocol, --search-string, --tool-type
    direct_business_process_v1_business_process_v1_values_send_back_to - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_compensation_v3_compensation_v3_values_one_time_payment_plan_group_one_time_paym - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_connect_v2_connect_v2_values_audience_prompt_group_audience_type - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_connect_v2_connect_v2_values_audience_prompt_group_selection - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_contract_compliance_v1_contract_compliance_v1_values_contract_compliance_group_companie - Retrieves all companies or company hierarchies the user has access to. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_contract_compliance_v1_contract_compliance_v1_values_contract_compliance_group_contract - Retrieves all contract types the user has access to. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_expense_v1_expense_v1_values_bespoke_prompt_expense_group - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_expense_v1_expense_v1_values_bespoke_prompt_expense_item - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_expense_v1_expense_v1_values_bespoke_prompt_currency - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_help_article_v1_help_article_v1_values_common_audiences - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_help_case_v4_help_case_v4_values_external_records_source - Get all sources for External Records [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_help_case_v4_help_case_v4_values_case_statuses_group_case_statuses - Get all \~case\~ statuses. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_help_case_v4_help_case_v4_values_service_teams_group_service_teams - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_help_case_v4_help_case_v4_values_case_label_categories_group_case_label_categ - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_help_case_v4_help_case_v4_values_case_flags_group_case_flags - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_journeys_v1_journeys_v1_values_related_role_role - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_learning_v1_learning_v1_values_tracking_enrollment_data_attendance_status - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_learning_v1_learning_v1_values_tracking_enrollment_data_grade - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_learning_v1_learning_v1_values_tracking_enrollment_data_duration_unit - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_tax_rates_group_company_instances - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_payroll_inputs_group_worktags - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_payroll_inputs_group_run_categories - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_payroll_inputs_group_positions - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_minimum_wage_rates_group_tax_authorities - Retrieves a list of tax authorities available for use as query parameters in the /minimumWageRates endpoint. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_tax_rates_group_state_instances - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_minimum_wage_rates_group_countries - Retrieves a list of all countries available as query parameters in the /minimumWageRates endpoint. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_payroll_v2_payroll_v2_values_payroll_inputs_group_pay_components - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_workers_to_notify_workers_to_no - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_manage_goals_categories - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_feedback_template_feedback_temp - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_feedback_on_worker_feedback_on - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_feedback_responder_feedback_res - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_organization_goal_organizations - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_organization_goal_supporting_in - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_review_template_template_for_my - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_cascade_goal_talent_pools - Talent Pools the processing user has view and edit access for. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_organization_goal_goal_periods - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_manage_goals_status - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_performance_enablement_v5_performance_enablement_v5_values_relates_to_relates_to - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_hereditary - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_religious - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_honorary - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_royal - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_title - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_common_phone_country_phone_codes - Exposes prompting for country phone codes, which are used during the collection of phone numbers. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_country_components_country_city - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
  Other Commands
    direct_person_v4_person_v4_values_name_components_salutation - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_marital_status_active_martial_statuses - Marital Statuses that are Active and configured for the Country or Country Region. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_personal_information_country_populated_country - The set of countries a person has populated with country specific data. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_social - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_professional - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_country_components_country - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_name_components_academic - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_personal_information_country_allowed_country - The set of countries a person is allowed to populated with country specific data. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_country_components_country_region - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_person_v4_person_v4_values_common_phone_phone_device_types - Exposes prompting for phone device types, which are used during the collection of phone numbers. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_resolved_worktags - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_requesters - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_companies - Retrieves instances that can be used as values for other endpoint parameters in this service. Note: This prompt value endpoint also returns the Company_Reference_ID response field. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_order_from_connection - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_ship_to_address - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_resource_provider - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_deliver_to_location - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_spend_category - Retrieves instances that can be used as values for other endpoint parameters in this service. Note: This prompt value endpoint also returns the Spend_Category_ID response field. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_supplier_contract - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_unit_of_measure - Retrieves instances that can be used as values for other endpoint parameters in this service. Note: This prompt value endpoint also returns the unCefactCommonCode response field. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_sourcing_buyer - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_inventory_site - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_line_company - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_currencies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_requesting_entity - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_requisition_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_par_location - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_commodity_codes - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_procurement_v5_procurement_v5_values_requisitions_group_worktags - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_worker_groups - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_currencies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_worktag_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_unnamed_resources - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_worker_to_replace_unnamed_resou - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_success_ratings - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_groups - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_project_plan_project_phases - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_roles - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_hierarchies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_customers - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_worktags - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_role_categories - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_project_plan_project_tasks - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_owners - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_workers - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_statuses - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_optional_hierarchies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_priorities - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_resource_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_project_plan_project_plan_phases - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_cost_rate_currencies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_requirements - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_requirement_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_companies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_risk_levels - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_resource_plan_booking_status - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_project_plan_project_plan_tasks - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_importance_ratings - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_projects - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_project_states - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_projects_v3_projects_v3_values_common_project_dependencies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_recruiting_v4_recruiting_v4_values_common_countries - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_regions - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_reason - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_customs - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_supervisory_organization - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_program - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_jobs - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_funds - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_workers_compensation_code_o - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_time_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_work_shifts - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_cost_ce - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_busines - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_frequencies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_employee_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_pay_rate_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_templates - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_workers - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_contingent_worker_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_currencies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_proposed_position - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_jobs - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_company_insider_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_headcount_options - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_grants - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_worker_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_workers - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_work_study_awards - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_locations - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_positio - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_job_requisitions - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_job_profiles - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_compani - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_work_spaces - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_assignment_types - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_job_changes_group_job_classifications - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_staffing_v7_staffing_v7_values_organization_assignment_changes_group_gifts - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_advising_v1_student_advising_v1_values_hypothetical_academic_progress_progra - Retrieves all programs of study. You can filter the programs of study by academic unit, academic level, program type, and effective date. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_advising_v1_student_advising_v1_values_hypothetical_academic_progress_studen - Retrieves a student with the specified username or universal ID. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_advising_v1_student_advising_v1_values_hypothetical_academic_progress_academ - Retrieves an academic record with the specified ID. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_award_years - Valid financial aid award years. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_academic_records - Valid academic records filtered by student and hold reason. Call the GET /values/holds/academicRecords?student=id&reason=id endpoint. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_reasons - All of the active hold reasons, optionally filtered by student and academic unit query parms. Call these endpoints for valid filter values: - GET /values/holds/reasons?student={id} [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_award_items - Valid student award items for the student and academic unit specified. Call the GET /values/holds/awardItems?student={id}&academicUnit={id} endpoint. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_institution - The institution of the hold reason. Call the following endpoint for the correct institution: GET /values/holds/institution?reason={id} [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_school_codes - All federal school codes, optionally filtered by student and hold reason query parms. Call these endpoints for valid filter values: - GET /values/holds/schoolCodes?student={id} - G [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_core_v1_student_core_v1_values_holds_academic_periods - Valid academic periods filtered by student, academic record, or academic unit. Call these endpoints for valid filter values: GET /values/holds/academicPeriods?student={id} GET /val [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_recruiting_v1_student_recruiting_v1_values_academic_requirement_evaluation_uni - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_recruiting_v1_student_recruiting_v1_values_academic_requirement_evaluation_edu - Retrieves all educational institutions. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_recruiting_v1_student_recruiting_v1_values_academic_requirement_evaluation_ext - Retrieves the external courses for the educational institution you specify. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_recruiting_v1_student_recruiting_v1_values_academic_requirement_evaluation_ext_2 - Retrieves the educational institution grading schemes for the educational institution you specify. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_student_recruiting_v1_student_recruiting_v1_values_academic_requirement_evaluation_pro - Retrieves all programs of study. You can filter the programs of study by academic unit, academic level, program type, and effective date. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_talent_management_v2_talent_management_v2_values_succession_plan_members - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_talent_management_v2_talent_management_v2_values_succession_plan_strategies - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_talent_management_v2_talent_management_v2_values_succession_plan_readiness - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_types_time_entry_codes - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_values_positions - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_enter_time_by_type_time_type_nested - Provides a nested prompt for Time Tracking Time Types, similar prompt structure to the one provided in UI prompts. Currently does not support Time Offs. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_types_default_time_entry_code - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_values_worker_time_zone - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_types_project_plan_tasks - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_types_projects - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
    direct_time_tracking_v5_time_tracking_v5_values_time_values_out_reason - Retrieves instances that can be used as values for other endpoint parameters in this service. [intent=direct_read availability=implemented]; risk: bounded read of Workday data; response is JSON-redacted and capped at 1 MiB
  Help topics:
    workday-rest-auth - Use a Workday REST bearer token supplied through credentials; tokens are never stored in fixtures.
    workday-rest-safety - Writes require reverse ETL plan, preview, explicit approval, and destructive confirmation where applicable.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect workday-rest

  # Inspect as structured JSON
  pm connectors inspect workday-rest --json

AGENT WORKFLOW
  - Run pm connectors inspect workday-rest before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
