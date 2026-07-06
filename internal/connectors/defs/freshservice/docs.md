# Overview

Freshservice reads 197 stream(s), and writes through 263 action(s).

Readable streams: `tickets`, `agents`, `requesters`, `assets`, `problems`, `view_a_ticket`,
`filter_tickets`, `view_all_ticket_fields`, `get_ticket_activities`, `view_ticket_time_entry`,
`list_all_ticket_time_entries`, `view_req_items_of_sr`, `list_all_ticket_approvals`,
`view_a_ticket_approval`, `list_all_ticket_approval_groups`, `view_all_ticket_tasks`,
`view_csat_response`, `list_all_conversations`, `view_a_problem`, `view_all_problem_fields`,
`view_all_problem_notes`, `view_a_problem_time_entry`, `view_all_problem_time_entries`,
`view_all_problem_tasks`, `view_a_change`, `view_all_changes`, `view_all_change_fields`,
`list_all_change_approvals`, `view_a_change_approval`, `list_all_change_approval_groups`,
`view_all_change_notes`, `view_a_time_entry`, `view_all_time_entries`, `view_all_change_tasks`,
`view_a_release`, `filter_releases`, `view_all_release`, `view_all_release_fields`,
`view_all_release_notes`, `view_a_release_time_entry`, `view_all_release_time_entries`,
`view_all_release_tasks`, `list_all_approvals`, `view_a_requester`, `list_all_requester_fields`,
`view_an_agent`, `list_all_agent_fields`, `view_a_role`, `view_all_role`, `view_a_group`,
`view_all_group`, `view_a_requester_group`, `view_all_requester_group`,
`list_members_of_requester_group`, `view_a_location`, `list_all_locations`, `filter_locations`,
`view_a_product`, `view_all_products`, `view_a_vendor`, `view_all_vendors`, `view_alert`,
`filter_alerts`, `view_alert_logs`, `view_alert_notes`, `view_alert_note`, `view_an_asset`,
`list_all_asset_components`, `list_all_asset_requests`, `list_contracts_of_an_asset`,
`view_a_relationship`, `list_relationships`, `list_relationship_types`, `list_all_purchase_orders`,
`view_a_purchase_order`, `view_an_asset_type`, `list_all_asset_types`, `list_asset_type_fields`,
`list_users`, `list_installations`, `list_all_contract_types`, `list_all_contract_type_fields`,
`view_a_contract`, `list_all_contracts`, `list_all_associated_assets`, `list_all_attachments`,
`view_a_department`, `list_all_departments`, `filter_departments`, `list_all_department_fields`,
`view_a_business_hour`, `list_all_business_hours`, `view_a_project`, `view_all_projects`,
`view_a_project_task`, `view_all_project_tasks`, `view_a_project_newgen`,
`view_all_projects_newgen`, `view_project_fields`, `view_project_templates`,
`view_project_associations`, `view_a_project_task_newgen`, `view_all_project_tasks_newgen`,
`filter_all_project_tasks_newgen`, `view_project_task_type_fields`, `view_project_task_types`,
`view_project_task_priorities`, `view_project_task_statuses`, `view_project_versions`,
`view_project_sprints`, `view_project_memberships`, `view_project_task_associations`,
`view_notes_task_newgen`, `view_solution_category`, `view_all_solution_category`,
`view_solution_folder`, `view_solution_sub_folders`, `view_all_solution_folder`,
`view_solution_article`, `view_all_solution_article`, `view_service_item`, `list_all_service_items`,
`list_all_service_categories`, `shared_fields`, `retrieve_shared_fields_data`,
`view_an_announcement`, `list_all_announcements`, `onboarding_form`, `view_onboarding_request`,
`list_all_onboarding_requests`, `view_onboarding_tickets`, `offboarding_form`,
`view_offboarding_request`, `list_all_offboarding_requests`, `view_offboarding_tickets`,
`journey_configs`, `journey_initiator_config_form`, `view_journey_request`,
`view_all_journey_requests`, `view_journey_request_activities`, `view_all_schedules`,
`filter_schedules`, `view_a_schedule`, `view_all_shifts`, `view_a_shift`,
`view_calendar_events_user`, `view_calendar_events_schedule`, `view_calendar_events_shift`,
`view_calendar_events_shift_user`, `view_calendar_events_schedule_user`, `view_who_is_oncall`,
`view_all_ep`, `view_a_ep`, `list_all_custom_objects`, `show_custom_object`,
`list_all_custom_object_records`, `get_pir_templates_by_id`, `get_all_pir_templates`,
`list_all_sla`, `list_all_canned_response_folders`, `show_canned_response_folder`,
`list_all_canned_response_in_folder`, `list_all_canned_responses`, `show_canned_response`,
`retrieve_emails_with_id`, `retrieve_emails`, `view_incidents`, `view_incident`,
`view_incident_updates`, `list_all_incident_statuses`, `view_maintenances`,
`view_maintenance_from_change`, `view_maintenance_from_maintenance_window`,
`list_all_maintenance_statuses`, `view_maintenance_update_from_change`,
`view_maintenance_update_from_maintenance_window`, `view_status_pages`,
`identify_publishable_services_ticket`, `identify_publishable_services_change`,
`identify_publishable_services_maintenance`, `view_service_component`, `view_service_components`,
`list_all_subscribers`, `view_a_subscriber`, `view_delegation`, `view_all_physical_subtypes`,
`view_a_physical_subtype`, `view_all_devices`, `view_a_device`,
`view_all_assets_for_freshservice_itam`, `view_an_asset_for_freshservice_itam`,
`view_all_cloud_resources`, `view_a_cloud_resource`, `view_all_cloud_resource_relationships`,
`view_all_cloud_infrastructure`, `view_all_lifecycle_events`, `view_a_lifecycle_event`.

Write actions: `create_ticket`, `update_a_ticket`, `move_a_ticket`, `delete_a_ticket`,
`delete_a_ticket_attachment`, `restore_a_ticket`, `create_child_ticket`, `create_ticket_time_entry`,
`update_ticket_time_entry`, `delete_ticket_time_entry`, `create_custom_ticket_source`,
`create_service_request`, `update_req_items_of_sr`, `add_catalog_item_to_existing_sr`,
`create_ticket_approval`, `resend_reminder_ticket_approval`, `cancel_a_ticket_approval`,
`create_approval_groups`, `update_approval_groups`, `update_service_request_approval_chain_rule`,
`promote_incident_to_major`, `demote_incident_from_major`, `create_ticket_task`,
`update_a_ticket_task`, `delete_a_ticket_task`, `create_a_reply`, `create_a_note`,
`update_a_conversations`, `delete_a_conversations`, `delete_a_conversation_attachment`,
`create_problem`, `update_problem_priority`, `move_a_problem`, `delete_a_problem`,
`restore_a_problem`, `create_problem_note`, `update_a_problem_note`, `delete_a_problem_note`,
`create_problem_time_entries`, `update_problem_time_entry`, `delete_a_problem_time_entry`,
`create_problem_task`, `update_a_problem_task`, `delete_a_problem_task`, `create_change`,
`update_change_priority`, `move_a_change`, `delete_a_change`, `resend_reminder_change_approval`,
`cancel_change_approval`, `create_change_approval_groups`, `update_change_approval_groups`,
`update_approval_chain_rule_change`, `create_change_note`, `update_a_change_note`,
`delete_a_change_note`, `create_time_entries`, `update_time_entry`, `delete_a_time_entry`,
`create_change_task`, `update_a_change_task`, `delete_a_change_task`, `create_release`,
`update_release_priority`, `move_a_release`, `delete_a_release`, `restore_a_release`,
`create_release_note`, `update_a_release_note`, `delete_a_release_note`,
`create_release_time_entries`, `update_release_time_entry`, `delete_a_release_time_entry`,
`create_release_task`, `update_a_release_task`, `delete_a_release_task`, `update_a_cab`,
`delete_a_cab`, `create_a_requester`, `update_a_requester`, and 183 more.

Service API documentation: https://api.freshservice.com/.

## Auth setup

Connection fields:

- `agent_id` (optional, string); Path parameter agent_id for Freshservice stream view_an_agent.
- `alert_id` (optional, string); Path parameter alert_id for Freshservice stream view_alert.
- `announcement_id` (optional, string); Path parameter announcement_id for Freshservice stream
  view_an_announcement.
- `api_key` (required, secret, string); Freshservice API key, sent as the username of HTTP Basic
  auth (password is the literal 'X'). Never logged.
- `application_id` (optional, string); Path parameter application_id for Freshservice stream
  list_users.
- `approval_id` (optional, string); Path parameter approval_id for Freshservice stream
  view_a_ticket_approval.
- `article_id` (optional, string); Path parameter article_id for Freshservice stream
  view_solution_article.
- `asset_id` (optional, string); Path parameter asset_id for Freshservice stream
  view_an_asset_for_freshservice_itam.
- `asset_type_id` (optional, string); Path parameter asset_type_id for Freshservice stream
  view_an_asset_type.
- `business_hour_id` (optional, string); Path parameter business_hour_id for Freshservice stream
  view_a_business_hour.
- `canned_respons_id` (optional, string); Path parameter canned_respons_id for Freshservice stream
  show_canned_response.
- `category_id` (optional, string); Path parameter category_id for Freshservice stream
  view_solution_category.
- `change_id` (optional, string); Path parameter change_id for Freshservice stream view_a_change.
- `communication_id` (optional, string); Path parameter communication_id for Freshservice stream
  retrieve_emails_with_id.
- `config_id` (optional, string); Path parameter config_id for Freshservice stream
  journey_initiator_config_form.
- `contract_id` (optional, string); Path parameter contract_id for Freshservice stream
  view_a_contract.
- `contract_type_id` (optional, string); Path parameter contract_type_id for Freshservice stream
  list_all_contract_type_fields.
- `department_id` (optional, string); Path parameter department_id for Freshservice stream
  view_a_department.
- `device_id` (optional, string); Path parameter device_id for Freshservice stream view_a_device.
- `display_id` (optional, string); Path parameter display_id for Freshservice stream view_an_asset.
- `domain_name` (required, string); Freshservice account domain (e.g. acme.freshservice.com);
  combined with the fixed /api/v2 path to form the base URL.
- `ep_id` (optional, string); Path parameter ep_id for Freshservice stream view_a_ep.
- `filter_alerts_order_by` (optional, string); Query parameter order_by for Freshservice stream
  filter_alerts.
- `filter_alerts_order_type` (optional, string); Query parameter order_type for Freshservice stream
  filter_alerts.
- `filter_alerts_query` (optional, string); Query parameter query for Freshservice stream
  filter_alerts.
- `filter_all_project_tasks_newgen_query` (optional, string); Query parameter query for Freshservice
  stream filter_all_project_tasks_newgen.
- `filter_departments_query` (optional, string); Query parameter query for Freshservice stream
  filter_departments.
- `filter_locations_query` (optional, string); Query parameter query for Freshservice stream
  filter_locations.
- `filter_name` (optional, string); Query parameter filter_name for Freshservice stream
  filter_releases.
- `filter_schedules_query` (optional, string); Query parameter query for Freshservice stream
  filter_schedules.
- `filter_tickets_query` (optional, string); Query parameter query for Freshservice stream
  filter_tickets.
- `folder_id` (optional, string); Path parameter folder_id for Freshservice stream
  view_solution_folder.
- `group_id` (optional, string); Path parameter group_id for Freshservice stream view_a_group.
- `incident_id` (optional, string); Path parameter incident_id for Freshservice stream
  view_incident.
- `lifecycle_event_id` (optional, string); Path parameter lifecycle_event_id for Freshservice stream
  view_a_lifecycle_event.
- `list_all_approvals_parent` (optional, string); Query parameter parent for Freshservice stream
  list_all_approvals.
- `location_id` (optional, string); Path parameter location_id for Freshservice stream
  view_a_location.
- `maintenance_id` (optional, string); Path parameter maintenance_id for Freshservice stream
  view_maintenance_from_change.
- `maintenance_window_id` (optional, string); Path parameter maintenance_window_id for Freshservice
  stream view_maintenance_from_maintenance_window.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `module_name` (optional, string); Path parameter module_name for Freshservice stream
  view_project_associations.
- `note_id` (optional, string); Path parameter note_id for Freshservice stream view_alert_note.
- `object_id` (optional, string); Path parameter object_id for Freshservice stream
  show_custom_object.
- `offboarding_request_id` (optional, string); Path parameter offboarding_request_id for
  Freshservice stream view_offboarding_request.
- `onboarding_request_id` (optional, string); Path parameter onboarding_request_id for Freshservice
  stream view_onboarding_request.
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `physical_subtype_id` (optional, string); Path parameter physical_subtype_id for Freshservice
  stream view_a_physical_subtype.
- `problem_id` (optional, string); Path parameter problem_id for Freshservice stream view_a_problem.
- `product_id` (optional, string); Path parameter product_id for Freshservice stream view_a_product.
- `project_id` (optional, string); Path parameter project_id for Freshservice stream view_a_project.
- `purchase_order_id` (optional, string); Path parameter purchase_order_id for Freshservice stream
  view_a_purchase_order.
- `relationship_id` (optional, string); Path parameter relationship_id for Freshservice stream
  view_a_relationship.
- `release_id` (optional, string); Path parameter release_id for Freshservice stream view_a_release.
- `request_id` (optional, string); Path parameter request_id for Freshservice stream
  view_journey_request.
- `requester_group_id` (optional, string); Path parameter requester_group_id for Freshservice stream
  view_a_requester_group.
- `requester_id` (optional, string); Path parameter requester_id for Freshservice stream
  view_a_requester.
- `resource_id` (optional, string); Path parameter resource_id for Freshservice stream
  view_a_cloud_resource.
- `role_id` (optional, string); Path parameter role_id for Freshservice stream view_a_role.
- `schedule_id` (optional, string); Path parameter schedule_id for Freshservice stream
  view_a_schedule.
- `service_component_id` (optional, string); Path parameter service_component_id for Freshservice
  stream view_service_component.
- `shared_field_id` (optional, string); Path parameter shared_field_id for Freshservice stream
  retrieve_shared_fields_data.
- `shift_id` (optional, string); Path parameter shift_id for Freshservice stream view_a_shift.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only tickets updated at
  or after this time are read (tickets stream only).
- `start_token` (optional, string); Query parameter start_token for Freshservice stream
  view_alert_logs.
- `status_page_id` (optional, string); Path parameter status_page_id for Freshservice stream
  view_incidents.
- `subscriber_id` (optional, string); Path parameter subscriber_id for Freshservice stream
  view_a_subscriber.
- `task_id` (optional, string); Path parameter task_id for Freshservice stream view_a_project_task.
- `template_id` (optional, string); Path parameter template_id for Freshservice stream
  get_pir_templates_by_id.
- `ticket_id` (optional, string); Path parameter ticket_id for Freshservice stream view_a_ticket.
- `time_entry_id` (optional, string); Path parameter time_entry_id for Freshservice stream
  view_ticket_time_entry.
- `type_id` (optional, string); Path parameter type_id for Freshservice stream
  view_project_task_type_fields.
- `user_id` (optional, string); Query parameter user_id for Freshservice stream
  view_calendar_events_user.
- `vendor_id` (optional, string); Path parameter vendor_id for Freshservice stream view_a_vendor.
- `view_calendar_events_schedule_end_time` (optional, string); Query parameter end_time for
  Freshservice stream view_calendar_events_schedule.
- `view_calendar_events_schedule_start_time` (optional, string); Query parameter start_time for
  Freshservice stream view_calendar_events_schedule.
- `view_calendar_events_schedule_user_end_time` (optional, string); Query parameter end_time for
  Freshservice stream view_calendar_events_schedule_user.
- `view_calendar_events_schedule_user_start_time` (optional, string); Query parameter start_time for
  Freshservice stream view_calendar_events_schedule_user.
- `view_calendar_events_shift_end_time` (optional, string); Query parameter end_time for
  Freshservice stream view_calendar_events_shift.
- `view_calendar_events_shift_start_time` (optional, string); Query parameter start_time for
  Freshservice stream view_calendar_events_shift.
- `view_calendar_events_shift_user_end_time` (optional, string); Query parameter end_time for
  Freshservice stream view_calendar_events_shift_user.
- `view_calendar_events_shift_user_start_time` (optional, string); Query parameter start_time for
  Freshservice stream view_calendar_events_shift_user.
- `view_calendar_events_user_end_time` (optional, string); Query parameter end_time for Freshservice
  stream view_calendar_events_user.
- `view_calendar_events_user_start_time` (optional, string); Query parameter start_time for
  Freshservice stream view_calendar_events_user.
- `workspace_id` (optional, string); Path parameter workspace_id for Freshservice stream
  view_all_schedules.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use base URL `https://{{ config.domain_name }}/api/v2` after applying configuration
defaults.

Connection checks call GET `/agents`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `view_a_ticket`, `view_ticket_time_entry`, `view_a_ticket_approval`,
`view_a_problem`, `view_a_problem_time_entry`, `view_a_change`, `view_a_change_approval`,
`view_a_time_entry`, `view_a_release`, `view_a_release_time_entry`, `view_a_requester`,
`view_an_agent`, `view_a_role`, `view_a_group`, `view_a_requester_group`, `view_a_location`,
`view_a_product`, `view_a_vendor`, `view_alert`, `view_alert_note`, `view_an_asset`,
`view_a_relationship`, `view_a_purchase_order`, `view_an_asset_type`, `view_a_contract`,
`view_a_department`, `view_a_business_hour`, `view_a_project`, `view_a_project_task`,
`view_a_project_newgen`, `view_project_associations`, `view_a_project_task_newgen`,
`view_project_task_associations`, `view_solution_category`, `view_solution_folder`,
`view_solution_article`, `view_service_item`, `retrieve_shared_fields_data`, `view_an_announcement`,
`view_onboarding_request`, `view_offboarding_request`, `view_journey_request`, `view_a_schedule`,
`view_a_shift`, `view_a_ep`, `show_custom_object`, `get_pir_templates_by_id`,
`show_canned_response_folder`, `show_canned_response`, `retrieve_emails_with_id`, `view_incident`,
`view_maintenance_from_change`, `view_maintenance_from_maintenance_window`,
`view_service_component`, `view_a_subscriber`, `view_delegation`, `view_a_physical_subtype`,
`view_a_device`, `view_an_asset_for_freshservice_itam`, `view_a_cloud_resource`, and 1 more;
page_number: `tickets`, `agents`, `requesters`, `assets`, `problems`, `filter_tickets`,
`view_all_ticket_fields`, `get_ticket_activities`, `list_all_ticket_time_entries`,
`view_req_items_of_sr`, `list_all_ticket_approvals`, `list_all_ticket_approval_groups`,
`view_all_ticket_tasks`, `view_csat_response`, `list_all_conversations`, `view_all_problem_fields`,
`view_all_problem_notes`, `view_all_problem_time_entries`, `view_all_problem_tasks`,
`view_all_changes`, `view_all_change_fields`, `list_all_change_approvals`,
`list_all_change_approval_groups`, `view_all_change_notes`, `view_all_time_entries`,
`view_all_change_tasks`, `filter_releases`, `view_all_release`, `view_all_release_fields`,
`view_all_release_notes`, `view_all_release_time_entries`, `view_all_release_tasks`,
`list_all_approvals`, `list_all_requester_fields`, `list_all_agent_fields`, `view_all_role`,
`view_all_group`, `view_all_requester_group`, `list_members_of_requester_group`,
`list_all_locations`, `filter_locations`, `view_all_products`, `view_all_vendors`, `filter_alerts`,
`view_alert_logs`, `view_alert_notes`, `list_all_asset_components`, `list_all_asset_requests`,
`list_contracts_of_an_asset`, `list_relationships`, `list_relationship_types`,
`list_all_purchase_orders`, `list_all_asset_types`, `list_asset_type_fields`, `list_users`,
`list_installations`, `list_all_contract_types`, `list_all_contract_type_fields`,
`list_all_contracts`, `list_all_associated_assets`, and 76 more.

- `tickets`: GET `/tickets` - records path `tickets`; query `updated_since` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100.
- `agents`: GET `/agents` - records path `agents`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `requesters`: GET `/requesters` - records path `requesters`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `assets`: GET `/assets` - records path `assets`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `problems`: GET `/problems` - records path `problems`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `view_a_ticket`: GET `/tickets/{{ config.ticket_id }}` - records path `ticket`; emits passthrough
  records.
- `filter_tickets`: GET `/tickets/filter` - records path `tickets`; query `query`=`{{
  config.filter_tickets_query }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_all_ticket_fields`: GET `/ticket_form_fields` - records path `ticket_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `get_ticket_activities`: GET `/tickets/{{ config.ticket_id }}/activities` - records path
  `activities`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_ticket_time_entry`: GET `/tickets/{{ config.ticket_id }}/time_entries/{{
  config.time_entry_id }}` - records path `time_entry`; emits passthrough records.
- `list_all_ticket_time_entries`: GET `/tickets/{{ config.ticket_id }}/time_entries` - records path
  `time_entries`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `view_req_items_of_sr`: GET `/tickets/{{ config.ticket_id }}/requested_items` - records path
  `requested_items`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `list_all_ticket_approvals`: GET `/tickets/{{ config.ticket_id }}/approvals` - records path
  `approvals`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_a_ticket_approval`: GET `/tickets/{{ config.ticket_id }}/approvals/{{ config.approval_id }}`
  - records path `approval`; emits passthrough records.
- `list_all_ticket_approval_groups`: GET `/tickets/{{ config.ticket_id }}/approval-groups` - records
  path `approval_groups`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_all_ticket_tasks`: GET `/tickets/{{ config.ticket_id }}/tasks` - records path `tasks`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_csat_response`: GET `/tickets/{{ config.ticket_id }}/csat_response` - records path
  `csat_response`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `list_all_conversations`: GET `/tickets/{{ config.ticket_id }}/conversations` - records path
  `conversations`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `view_a_problem`: GET `/problems/{{ config.problem_id }}` - records path `problem`; emits
  passthrough records.
- `view_all_problem_fields`: GET `/problem_form_fields` - records path `problem_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_all_problem_notes`: GET `/problems/{{ config.problem_id }}/notes` - records path `notes`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_problem_time_entry`: GET `/problems/{{ config.problem_id }}/time_entries/{{
  config.time_entry_id }}` - records path `time_entry`; emits passthrough records.
- `view_all_problem_time_entries`: GET `/problems/{{ config.problem_id }}/time_entries` - records
  path `time_entries`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_all_problem_tasks`: GET `/problems/{{ config.problem_id }}/tasks` - records path `tasks`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_change`: GET `/changes/{{ config.change_id }}` - records path `change`; emits passthrough
  records.
- `view_all_changes`: GET `/changes` - records path `changes`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_all_change_fields`: GET `/change_form_fields` - records path `change_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `list_all_change_approvals`: GET `/changes/{{ config.change_id }}/approvals` - records path
  `approvals`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_a_change_approval`: GET `/changes/{{ config.change_id }}/approvals/{{ config.approval_id }}`
  - records path `approval`; emits passthrough records.
- `list_all_change_approval_groups`: GET `/changes/{{ config.change_id }}/approval-groups` - records
  path `approval_groups`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_all_change_notes`: GET `/changes/{{ config.change_id }}/notes` - records path `notes`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_time_entry`: GET `/changes/{{ config.change_id }}/time_entries/{{ config.time_entry_id }}`
  - records path `time_entry`; emits passthrough records.
- `view_all_time_entries`: GET `/changes/{{ config.change_id }}/time_entries` - records path
  `time_entries`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `view_all_change_tasks`: GET `/changes/{{ config.change_id }}/tasks` - records path `tasks`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_release`: GET `/releases/{{ config.release_id }}` - records path `release`; emits
  passthrough records.
- `filter_releases`: GET `/releases` - records path `.`; query `filter_name`=`{{ config.filter_name
  }}`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; emits passthrough records.
- `view_all_release`: GET `/releases` - records path `releases`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_all_release_fields`: GET `/release_form_fields` - records path `release_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_all_release_notes`: GET `/releases/{{ config.release_id }}/notes` - records path `notes`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_release_time_entry`: GET `/releases/{{ config.release_id }}/time_entries/{{
  config.time_entry_id }}` - records path `time_entry`; emits passthrough records.
- `view_all_release_time_entries`: GET `/releases/{{ config.release_id }}/time_entries` - records
  path `time_entries`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_all_release_tasks`: GET `/releases/{{ config.release_id }}/tasks` - records path `tasks`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `list_all_approvals`: GET `/approvals` - records path `approvals`; query `parent`=`{{
  config.list_all_approvals_parent }}`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_requester`: GET `/requesters/{{ config.requester_id }}` - records path `requester`; emits
  passthrough records.
- `list_all_requester_fields`: GET `/requester_fields` - records path `requester_fields`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_an_agent`: GET `/agents/{{ config.agent_id }}` - records path `agent`; emits passthrough
  records.
- `list_all_agent_fields`: GET `/agent_fields` - records path `agent_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_a_role`: GET `/roles/{{ config.role_id }}` - records path `role`; emits passthrough records.
- `view_all_role`: GET `/roles` - records path `roles`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_group`: GET `/groups/{{ config.group_id }}` - records path `group`; emits passthrough
  records.
- `view_all_group`: GET `/groups` - records path `groups`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_requester_group`: GET `/requester_groups/{{ config.requester_group_id }}` - records path
  `requester_group`; emits passthrough records.
- `view_all_requester_group`: GET `/requester_groups` - records path `requester_groups`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `list_members_of_requester_group`: GET `/requester_groups/{{ config.requester_group_id }}/members`
  - records path `requesters`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_location`: GET `/locations/{{ config.location_id }}` - records path `locations`; emits
  passthrough records.
- `list_all_locations`: GET `/locations` - records path `locations`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `filter_locations`: GET `/locations` - records path `locations`; query `query`=`{{
  config.filter_locations_query }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_product`: GET `/products/{{ config.product_id }}` - records path `product`; emits
  passthrough records.
- `view_all_products`: GET `/products` - records path `products`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_a_vendor`: GET `/vendors/{{ config.vendor_id }}` - records path `vendor`; emits passthrough
  records.
- `view_all_vendors`: GET `/vendors` - records path `vendors`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_alert`: GET `/ams/alerts/{{ config.alert_id }}` - records path `alert`; emits passthrough
  records.
- `filter_alerts`: GET `/ams/alerts` - records path `alerts`; query `order_by`=`{{
  config.filter_alerts_order_by }}`; `order_type`=`{{ config.filter_alerts_order_type }}`;
  `query`=`{{ config.filter_alerts_query }}`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_alert_logs`: GET `/ams/alerts/{{ config.alert_id }}/logs` - records path `logs`; query
  `start_token`=`{{ config.start_token }}`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_alert_notes`: GET `/ams/alerts/{{ config.alert_id }}/notes` - records path `alert_notes`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_alert_note`: GET `/ams/alerts/{{ config.alert_id }}/notes/{{ config.note_id }}` - records
  path `alert_note`; emits passthrough records.
- `view_an_asset`: GET `/assets/{{ config.display_id }}` - records path `asset`; emits passthrough
  records.
- `list_all_asset_components`: GET `/assets/{{ config.display_id }}/components` - records path
  `components`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `list_all_asset_requests`: GET `/assets/{{ config.display_id }}/requests` - records path
  `requests`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; emits passthrough records.
- `list_contracts_of_an_asset`: GET `/assets/{{ config.display_id }}/contracts` - records path
  `contracts`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_a_relationship`: GET `/relationships/{{ config.relationship_id }}` - records path
  `relationship`; emits passthrough records.
- `list_relationships`: GET `/relationships` - records path `relationships`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `list_relationship_types`: GET `/relationship_types` - records path `relationship_types`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `list_all_purchase_orders`: GET `/purchase_orders` - records path `purchase_orders`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_a_purchase_order`: GET `/purchase_orders/{{ config.purchase_order_id }}` - records path
  `purchase_order`; emits passthrough records.
- `view_an_asset_type`: GET `/asset_types/{{ config.asset_type_id }}` - records path `asset_type`;
  emits passthrough records.
- `list_all_asset_types`: GET `/asset_types` - records path `asset_types`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `list_asset_type_fields`: GET `/asset_types/{{ config.asset_type_id }}/fields` - records path
  `asset_type_fields`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `list_users`: GET `/applications/{{ config.application_id }}/users` - records path
  `application_users`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `list_installations`: GET `/applications/{{ config.application_id }}/installations` - records path
  `installations`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `list_all_contract_types`: GET `/contract_types` - records path `contract_types`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `list_all_contract_type_fields`: GET `/contract_types/{{ config.contract_type_id }}/fields` -
  records path `contract_type_fields`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_contract`: GET `/contracts/{{ config.contract_id }}` - records path `contract`; emits
  passthrough records.
- `list_all_contracts`: GET `/contracts` - records path `contracts`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `list_all_associated_assets`: GET `/contracts/{{ config.contract_id }}/associated-assets` -
  records path `associated_assets`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `list_all_attachments`: GET `/contracts/{{ config.contract_id }}/attachments` - records path
  `attachments`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_a_department`: GET `/departments/{{ config.department_id }}` - records path `department`;
  emits passthrough records.
- `list_all_departments`: GET `/departments` - records path `departments`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `filter_departments`: GET `/departments` - records path `departments`; query `query`=`{{
  config.filter_departments_query }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `list_all_department_fields`: GET `/department_fields` - records path `department_fields`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_business_hour`: GET `/business_hours/{{ config.business_hour_id }}` - records path
  `business_hours`; emits passthrough records.
- `list_all_business_hours`: GET `/business_hours` - records path `business_hours`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_a_project`: GET `/projects/{{ config.project_id }}` - records path `project`; emits
  passthrough records.
- `view_all_projects`: GET `/projects` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_project_task`: GET `/projects/{{ config.project_id }}/tasks/{{ config.task_id }}` -
  records path `task`; emits passthrough records.
- `view_all_project_tasks`: GET `/projects/{{ config.project_id }}/tasks` - records path `tasks`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_a_project_newgen`: GET `/pm/projects/{{ config.project_id }}` - records path `project`;
  emits passthrough records.
- `view_all_projects_newgen`: GET `/pm/projects` - records path `projects`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_project_fields`: GET `/pm/project-fields` - records path `project_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_project_templates`: GET `/pm/project_templates` - records path `project_templates`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_project_associations`: GET `/pm/projects/{{ config.project_id }}/{{ config.module_name }}` -
  records path `tickets`; emits passthrough records.
- `view_a_project_task_newgen`: GET `/pm/projects/{{ config.project_id }}/tasks/{{ config.task_id
  }}` - records path `task`; emits passthrough records.
- `view_all_project_tasks_newgen`: GET `/pm/projects/{{ config.project_id }}/tasks` - records path
  `tasks`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; emits passthrough records.
- `filter_all_project_tasks_newgen`: GET `/pm/projects/{{ config.project_id }}/tasks/filter` -
  records path `tasks`; query `query`=`{{ config.filter_all_project_tasks_newgen_query }}`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_project_task_type_fields`: GET `/pm/projects/{{ config.project_id }}/task-types/{{
  config.type_id }}/fields` - records path `fields`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_project_task_types`: GET `/pm/projects/{{ config.project_id }}/task-types` - records path
  `task_types`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_project_task_priorities`: GET `/pm/projects/{{ config.project_id }}/task-priorities` -
  records path `task_priorities`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_project_task_statuses`: GET `/pm/projects/{{ config.project_id }}/task-statuses` - records
  path `task_statuses`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_project_versions`: GET `/pm/projects/{{ config.project_id }}/versions` - records path
  `versions`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; emits passthrough records.
- `view_project_sprints`: GET `/pm/projects/{{ config.project_id }}/sprints` - records path
  `sprints`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; emits passthrough records.
- `view_project_memberships`: GET `/pm/projects/{{ config.project_id }}/memberships` - records path
  `memberships`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_project_task_associations`: GET `/pm/projects/{{ config.project_id }}/tasks/{{
  config.task_id }}/{{ config.module_name }}` - records path `tickets`; emits passthrough records.
- `view_notes_task_newgen`: GET `/pm/projects/{{ config.project_id }}/tasks/{{ config.task_id
  }}/notes` - records path `notes`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_solution_category`: GET `/solutions/categories/{{ config.category_id }}` - records path
  `category`; emits passthrough records.
- `view_all_solution_category`: GET `/solutions/categories` - records path `categories`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_solution_folder`: GET `/solutions/folders/{{ config.folder_id }}` - records path `folder`;
  emits passthrough records.
- `view_solution_sub_folders`: GET `/solutions/folders/{{ config.folder_id }}/sub-folders` - records
  path `sub_folders`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_all_solution_folder`: GET `/solutions/folders` - records path `folders`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_solution_article`: GET `/solutions/articles/{{ config.article_id }}` - records path
  `article`; emits passthrough records.
- `view_all_solution_article`: GET `/solutions/articles` - records path `articles`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_service_item`: GET `/service_catalog/items/{{ config.display_id }}` - records path
  `service_item`; emits passthrough records.
- `list_all_service_items`: GET `/service_catalog/items` - records path `service_items`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `list_all_service_categories`: GET `/service_catalog/categories` - records path
  `service_categories`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `shared_fields`: GET `/service-catalog/shared-fields` - records path `.`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `retrieve_shared_fields_data`: GET `/service-catalog/shared-fields/{{ config.shared_field_id }}` -
  records path `id`; emits passthrough records.
- `view_an_announcement`: GET `/announcements/{{ config.announcement_id }}` - records path
  `announcement`; emits passthrough records.
- `list_all_announcements`: GET `/announcements` - records path `announcements`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `onboarding_form`: GET `/onboarding_requests/form` - records path `fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_onboarding_request`: GET `/onboarding_requests/{{ config.onboarding_request_id }}` - records
  path `onboarding_request`; emits passthrough records.
- `list_all_onboarding_requests`: GET `/onboarding_requests` - records path `onboarding_requests`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_onboarding_tickets`: GET `/onboarding_requests/{{ config.onboarding_request_id }}/tickets` -
  records path `onboarding_tickets`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `offboarding_form`: GET `/offboarding_requests/form` - records path `fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_offboarding_request`: GET `/offboarding_requests/{{ config.offboarding_request_id }}` -
  records path `offboarding_request`; emits passthrough records.
- `list_all_offboarding_requests`: GET `/offboarding_requests` - records path
  `offboarding_requests`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_offboarding_tickets`: GET `/offboarding_requests/{{ config.offboarding_request_id
  }}/tickets` - records path `offboarding_tickets`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `journey_configs`: GET `/journeys/configs` - records path `journey_configs`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `journey_initiator_config_form`: GET `/journeys/configs/{{ config.config_id }}/data-fields` -
  records path `fields`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_journey_request`: GET `/journeys/requests/{{ config.request_id }}` - records path
  `journey_request`; emits passthrough records.
- `view_all_journey_requests`: GET `/journeys/requests` - records path `journey_requests`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_journey_request_activities`: GET `/journeys/requests/{{ config.request_id }}/activities` -
  records path `journey_request_activities`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_all_schedules`: GET `/oncall/ws/{{ config.workspace_id }}/schedules` - records path
  `schedules`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `filter_schedules`: GET `/oncall/ws/{{ config.workspace_id }}/schedules` - records path
  `schedules`; query `query`=`{{ config.filter_schedules_query }}`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_a_schedule`: GET `/oncall/ws/{{ config.workspace_id }}/schedules/{{ config.schedule_id }}` -
  records path `schedule`; emits passthrough records.
- `view_all_shifts`: GET `/oncall/ws/{{ config.workspace_id }}/schedules/{{ config.schedule_id
  }}/shifts` - records path `shifts`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_a_shift`: GET `/oncall/ws/{{ config.workspace_id }}/schedules/{{ config.schedule_id
  }}/shifts/{{ config.shift_id }}` - records path `shift`; emits passthrough records.
- `view_calendar_events_user`: GET `/oncall/shift-events` - records path `shift_events`; query
  `end_time`=`{{ config.view_calendar_events_user_end_time }}`; `start_time`=`{{
  config.view_calendar_events_user_start_time }}`; `user_id`=`{{ config.user_id }}`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_calendar_events_schedule`: GET `/oncall/shift-events` - records path `shift_events`; query
  `end_time`=`{{ config.view_calendar_events_schedule_end_time }}`; `schedule_id`=`{{
  config.schedule_id }}`; `start_time`=`{{ config.view_calendar_events_schedule_start_time }}`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_calendar_events_shift`: GET `/oncall/shift-events` - records path `shift_events`; query
  `end_time`=`{{ config.view_calendar_events_shift_end_time }}`; `schedule_id`=`{{
  config.schedule_id }}`; `shift_id`=`{{ config.shift_id }}`; `start_time`=`{{
  config.view_calendar_events_shift_start_time }}`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_calendar_events_shift_user`: GET `/oncall/shift-events` - records path `shift_events`; query
  `end_time`=`{{ config.view_calendar_events_shift_user_end_time }}`; `schedule_id`=`{{
  config.schedule_id }}`; `shift_id`=`{{ config.shift_id }}`; `start_time`=`{{
  config.view_calendar_events_shift_user_start_time }}`; `user_id`=`{{ config.user_id }}`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_calendar_events_schedule_user`: GET `/oncall/shift-events` - records path `shift_events`;
  query `end_time`=`{{ config.view_calendar_events_schedule_user_end_time }}`; `schedule_id`=`{{
  config.schedule_id }}`; `start_time`=`{{ config.view_calendar_events_schedule_user_start_time }}`;
  `user_id`=`{{ config.user_id }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_who_is_oncall`: GET `/oncall/shift-events/current` - records path `shift_events`; query
  `schedule_id`=`{{ config.schedule_id }}`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_all_ep`: GET `/oncall/ws/{{ config.workspace_id }}/schedules/{{ config.schedule_id
  }}/escalation-policies` - records path `escalation_policies`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_a_ep`: GET `/oncall/ws/{{ config.workspace_id }}/schedules/{{ config.schedule_id
  }}/escalation-policies/{{ config.ep_id }}` - records path `escalation_policy`; emits passthrough
  records.
- `list_all_custom_objects`: GET `/objects` - records path `custom_objects`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `show_custom_object`: GET `/objects/{{ config.object_id }}` - records path `custom_object`; emits
  passthrough records.
- `list_all_custom_object_records`: GET `/objects/{{ config.object_id }}/records` - records path
  `records`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; emits passthrough records.
- `get_pir_templates_by_id`: GET `/post-incident-reports/templates/{{ config.template_id }}` -
  records path `post_incident_report_template`; emits passthrough records.
- `get_all_pir_templates`: GET `/post-incident-reports/templates` - records path
  `post_incident_report_templates`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `list_all_sla`: GET `/sla_policies` - records path `sla_policies`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `list_all_canned_response_folders`: GET `/canned_response_folders` - records path
  `canned_response_folders`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `show_canned_response_folder`: GET `/canned_response_folders/{{ config.folder_id }}` - records
  path `canned_response_folder`; emits passthrough records.
- `list_all_canned_response_in_folder`: GET `/canned_response_folders/{{ config.folder_id
  }}/canned_responses` - records path `canned_responses`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `list_all_canned_responses`: GET `/canned_responses` - records path `canned_responses`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `show_canned_response`: GET `/canned_responses/{{ config.canned_respons_id }}` - records path
  `canned_responses`; emits passthrough records.
- `retrieve_emails_with_id`: GET `/tickets/{{ config.ticket_id }}/communications/{{
  config.communication_id }}` - records path `communication`; emits passthrough records.
- `retrieve_emails`: GET `/tickets/{{ config.ticket_id }}/communications` - records path
  `communications`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `view_incidents`: GET `/status/pages/{{ config.status_page_id }}/incidents` - records path
  `incidents`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_incident`: GET `/tickets/{{ config.ticket_id }}/status/pages/{{ config.status_page_id
  }}/incidents/{{ config.incident_id }}` - records path `incident`; emits passthrough records.
- `view_incident_updates`: GET `/tickets/{{ config.ticket_id }}/status/pages/{{
  config.status_page_id }}/incidents/{{ config.incident_id }}/updates` - records path
  `incident_updates`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `list_all_incident_statuses`: GET `/status/pages/{{ config.status_page_id }}/incidents/statuses` -
  records path `incident_statuses`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_maintenances`: GET `/status/pages/{{ config.status_page_id }}/maintenances` - records path
  `maintenances`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `view_maintenance_from_change`: GET `/changes/{{ config.change_id }}/status/pages/{{
  config.status_page_id }}/maintenances/{{ config.maintenance_id }}` - records path `maintenance`;
  emits passthrough records.
- `view_maintenance_from_maintenance_window`: GET `/maintenance-windows/{{
  config.maintenance_window_id }}/status/pages/{{ config.status_page_id }}/maintenances/{{
  config.maintenance_id }}` - records path `maintenance`; emits passthrough records.
- `list_all_maintenance_statuses`: GET `/status/pages/{{ config.status_page_id
  }}/maintenances/statuses` - records path `maintenance_statuses`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_maintenance_update_from_change`: GET `/changes/{{ config.change_id }}/status/pages/{{
  config.status_page_id }}/maintenances/{{ config.maintenance_id }}/updates` - records path
  `maintenance_updates`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `view_maintenance_update_from_maintenance_window`: GET `/maintenance-windows/{{
  config.maintenance_window_id }}/status/pages/{{ config.status_page_id }}/maintenances/{{
  config.maintenance_id }}/updates` - records path `maintenance_updates`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_status_pages`: GET `/status/pages` - records path `status_pages`; query `workspace_id`=`{{
  config.workspace_id }}`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; emits passthrough records.
- `identify_publishable_services_ticket`: GET `/tickets/{{ config.ticket_id }}/status/pages/{{
  config.status_page_id }}/publishable-services` - records path `service_components`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `identify_publishable_services_change`: GET `/changes/{{ config.change_id }}/status/pages/{{
  config.status_page_id }}/publishable-services` - records path `service_components`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `identify_publishable_services_maintenance`: GET `/maintenance-windows/{{
  config.maintenance_window_id }}/status/pages/{{ config.status_page_id }}/publishable-services` -
  records path `service_components`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `view_service_component`: GET `/status/pages/{{ config.status_page_id }}/service-components/{{
  config.service_component_id }}` - records path `service_component`; emits passthrough records.
- `view_service_components`: GET `/status/pages/{{ config.status_page_id }}/service-components` -
  records path `service_components`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; emits passthrough records.
- `list_all_subscribers`: GET `/status/pages/{{ config.status_page_id }}/subscribers` - records path
  `subscribers`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `view_a_subscriber`: GET `/status/pages/{{ config.status_page_id }}/subscribers/{{
  config.subscriber_id }}` - records path `subscribers`; emits passthrough records.
- `view_delegation`: GET `/users/{{ config.user_id }}/delegation` - records path `delegation`; emits
  passthrough records.
- `view_all_physical_subtypes`: GET `/itam/physical-subtypes` - records path `meta`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_a_physical_subtype`: GET `/itam/physical-subtypes/{{ config.physical_subtype_id }}` -
  records path `physical_subtype_id`; emits passthrough records.
- `view_all_devices`: GET `/itam/devices` - records path `meta`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_a_device`: GET `/itam/devices/{{ config.device_id }}` - records path `device_id`; emits
  passthrough records.
- `view_all_assets_for_freshservice_itam`: GET `/itam/assets` - records path `meta`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_an_asset_for_freshservice_itam`: GET `/itam/assets/{{ config.asset_id }}` - records path
  `name`; emits passthrough records.
- `view_all_cloud_resources`: GET `/itam/resources` - records path `meta`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `view_a_cloud_resource`: GET `/itam/resources/{{ config.resource_id }}` - records path `id`; emits
  passthrough records.
- `view_all_cloud_resource_relationships`: GET `/itam/resource_relationships` - records path `meta`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_all_cloud_infrastructure`: GET `/itam/cloud_infrastructures` - records path `meta`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `view_all_lifecycle_events`: GET `/itam/lifecycle_events` - records path `meta`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.
- `view_a_lifecycle_event`: GET `/itam/lifecycle_events/{{ config.lifecycle_event_id }}` - records
  path `lifecycle_event`; emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, restores, archives, cancels, and deletes Freshservice records
through documented REST API v2 mutation endpoints.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_ticket`: POST `/tickets` - kind `create`; body type `json`; risk: Create a Ticket through
  the Freshservice API.
- `update_a_ticket`: PUT `/tickets/{{ record.ticket_id }}` - kind `update`; body type `json`; path
  fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`; risk: Update
  a Ticket through the Freshservice API.
- `move_a_ticket`: PUT `/tickets/{{ record.ticket_id }}/move_workspace` - kind `update`; body type
  `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`;
  risk: Move a Ticket through the Freshservice API.
- `delete_a_ticket`: DELETE `/tickets/{{ record.ticket_id }}` - kind `delete`; body type `none`;
  path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Ticket.
- `delete_a_ticket_attachment`: DELETE `/tickets/{{ record.ticket_id }}/attachments/{{
  record.attachment_id }}` - kind `delete`; body type `none`; path fields `ticket_id`,
  `attachment_id`; required record fields `ticket_id`, `attachment_id`; accepted fields
  `attachment_id`, `ticket_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete a Ticket Attachment.
- `restore_a_ticket`: PUT `/tickets/{{ record.ticket_id }}/restore` - kind `update`; body type
  `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`;
  risk: Restore a Ticket through the Freshservice API.
- `create_child_ticket`: POST `/tickets/{{ record.parent_id }}/create_child_ticket` - kind `create`;
  body type `json`; path fields `parent_id`; required record fields `parent_id`; accepted fields
  `parent_id`; risk: Create a Child Ticket through the Freshservice API.
- `create_ticket_time_entry`: POST `/tickets/{{ record.ticket_id }}/time_entries` - kind `create`;
  body type `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields
  `ticket_id`; risk: Create a Time Entry through the Freshservice API.
- `update_ticket_time_entry`: PUT `/tickets/{{ record.ticket_id }}/time_entries/{{
  record.time_entry_id }}` - kind `update`; body type `json`; path fields `ticket_id`,
  `time_entry_id`; required record fields `ticket_id`, `time_entry_id`; accepted fields `ticket_id`,
  `time_entry_id`; risk: Update a Time Entry through the Freshservice API.
- `delete_ticket_time_entry`: DELETE `/tickets/{{ record.ticket_id }}/time_entries/{{
  record.time_entry_id }}` - kind `delete`; body type `none`; path fields `ticket_id`,
  `time_entry_id`; required record fields `ticket_id`, `time_entry_id`; accepted fields `ticket_id`,
  `time_entry_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Time Entry.
- `create_custom_ticket_source`: POST `/ticket_fields/sources` - kind `create`; body type `json`;
  risk: Create a Source through the Freshservice API.
- `create_service_request`: POST `/service_catalog/items/{{ record.display_id }}/place_request` -
  kind `create`; body type `json`; path fields `display_id`; required record fields `display_id`;
  accepted fields `display_id`; risk: Create a Service Request through the Freshservice API.
- `update_req_items_of_sr`: PUT `/tickets/{{ record.ticket_id }}/requested_items/{{
  record.requested_item_id }}` - kind `update`; body type `json`; path fields `ticket_id`,
  `requested_item_id`; required record fields `ticket_id`, `requested_item_id`; accepted fields
  `requested_item_id`, `ticket_id`; risk: Update Requested Items of a Service Request through the
  Freshservice API.
- `add_catalog_item_to_existing_sr`: POST `/tickets/{{ record.ticket_id }}/requested_items` - kind
  `create`; body type `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted
  fields `ticket_id`; risk: Add Catalog Item to Existing Service Request through the Freshservice
  API.
- `create_ticket_approval`: POST `/tickets/{{ record.ticket_id }}/approvals` - kind `create`; body
  type `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields
  `ticket_id`; risk: Request Approval for a Service Request through the Freshservice API.
- `resend_reminder_ticket_approval`: PUT `/tickets/{{ record.ticket_id }}/approvals/{{
  record.approval_id }}/remind` - kind `update`; body type `json`; path fields `ticket_id`,
  `approval_id`; required record fields `ticket_id`, `approval_id`; accepted fields `approval_id`,
  `ticket_id`; risk: Send Reminder for an Approval through the Freshservice API.
- `cancel_a_ticket_approval`: PUT `/tickets/{{ record.ticket_id }}/approvals/{{ record.approval_id
  }}` - kind `update`; body type `json`; path fields `ticket_id`, `approval_id`; required record
  fields `ticket_id`, `approval_id`; accepted fields `approval_id`, `ticket_id`; risk: Cancel an
  approval through the Freshservice API.
- `create_approval_groups`: POST `/tickets/{{ record.ticket_id }}/approval-groups` - kind `create`;
  body type `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields
  `ticket_id`; risk: Create Approval Groups for a Service Request through the Freshservice API.
- `update_approval_groups`: PUT `/tickets/{{ record.ticket_id }}/approval-groups/{{
  record.approval_group_id }}` - kind `update`; body type `json`; path fields `ticket_id`,
  `approval_group_id`; required record fields `ticket_id`, `approval_group_id`; accepted fields
  `approval_group_id`, `ticket_id`; risk: Update Approval Group In A Service Request through the
  Freshservice API.
- `update_service_request_approval_chain_rule`: PUT `/tickets/{{ record.ticket_id
  }}/approval-chain-rule` - kind `update`; body type `json`; path fields `ticket_id`; required
  record fields `ticket_id`; accepted fields `ticket_id`; risk: Update approval chain rule for a
  service request through the Freshservice API.
- `promote_incident_to_major`: PUT `/tickets/{{ record.ticket_id }}/promote` - kind `update`; body
  type `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields
  `ticket_id`; risk: Promote an incident to a major incident through the Freshservice API.
- `demote_incident_from_major`: PUT `/tickets/{{ record.ticket_id }}/demote` - kind `update`; body
  type `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields
  `ticket_id`; risk: Demote an incident from a major incident through the Freshservice API.
- `create_ticket_task`: POST `/tickets/{{ record.ticket_id }}/tasks` - kind `create`; body type
  `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`;
  risk: Create a Task through the Freshservice API.
- `update_a_ticket_task`: PUT `/tickets/{{ record.ticket_id }}/tasks/{{ record.task_id }}` - kind
  `update`; body type `json`; path fields `ticket_id`, `task_id`; required record fields
  `ticket_id`, `task_id`; accepted fields `task_id`, `ticket_id`; risk: Update a Task through the
  Freshservice API.
- `delete_a_ticket_task`: DELETE `/tickets/{{ record.ticket_id }}/tasks/{{ record.task_id }}` - kind
  `delete`; body type `none`; path fields `ticket_id`, `task_id`; required record fields
  `ticket_id`, `task_id`; accepted fields `task_id`, `ticket_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Task.
- `create_a_reply`: POST `/tickets/{{ record.ticket_id }}/reply` - kind `create`; body type `json`;
  path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`; risk:
  Create a Reply through the Freshservice API.
- `create_a_note`: POST `/tickets/{{ record.ticket_id }}/notes` - kind `create`; body type `json`;
  path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`; risk:
  Create a Note through the Freshservice API.
- `update_a_conversations`: PUT `/conversations/{{ record.conversation_id }}` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; risk: Update a Conversation through the Freshservice API.
- `delete_a_conversations`: DELETE `/conversations/{{ record.conversation_id }}` - kind `delete`;
  body type `none`; path fields `conversation_id`; required record fields `conversation_id`;
  accepted fields `conversation_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes Freshservice data: Delete a Conversation.
- `delete_a_conversation_attachment`: DELETE `/conversations/{{ record.conversation_id
  }}/attachments/{{ record.attachment_id }}` - kind `delete`; body type `none`; path fields
  `conversation_id`, `attachment_id`; required record fields `conversation_id`, `attachment_id`;
  accepted fields `attachment_id`, `conversation_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Conversation
  Attachment.
- `create_problem`: POST `/problems` - kind `create`; body type `json`; risk: Create a Problem
  through the Freshservice API.
- `update_problem_priority`: PUT `/problems/{{ record.problem_id }}` - kind `update`; body type
  `json`; path fields `problem_id`; required record fields `problem_id`; accepted fields
  `problem_id`; risk: Update a Problem through the Freshservice API.
- `move_a_problem`: PUT `/problems/{{ record.problem_id }}/move_workspace` - kind `update`; body
  type `json`; path fields `problem_id`; required record fields `problem_id`; accepted fields
  `problem_id`; risk: Move a Problem through the Freshservice API.
- `delete_a_problem`: DELETE `/problems/{{ record.problem_id }}` - kind `delete`; body type `none`;
  path fields `problem_id`; required record fields `problem_id`; accepted fields `problem_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Problem.
- `restore_a_problem`: PUT `/problems/{{ record.problem_id }}/restore` - kind `update`; body type
  `json`; path fields `problem_id`; required record fields `problem_id`; accepted fields
  `problem_id`; risk: Restore a Problem through the Freshservice API.
- `create_problem_note`: POST `/problems/{{ record.problem_id }}/notes` - kind `create`; body type
  `json`; path fields `problem_id`; required record fields `problem_id`; accepted fields
  `problem_id`; risk: Create a note through the Freshservice API.
- `update_a_problem_note`: PUT `/problems/{{ record.problem_id }}/notes/{{ record.note_id }}` - kind
  `update`; body type `json`; path fields `problem_id`, `note_id`; required record fields
  `problem_id`, `note_id`; accepted fields `note_id`, `problem_id`; risk: Update a note through the
  Freshservice API.
- `delete_a_problem_note`: DELETE `/problems/{{ record.problem_id }}/notes/{{ record.note_id }}` -
  kind `delete`; body type `none`; path fields `problem_id`, `note_id`; required record fields
  `problem_id`, `note_id`; accepted fields `note_id`, `problem_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  note.
- `create_problem_time_entries`: POST `/problems/{{ record.problem_id }}/time_entries` - kind
  `create`; body type `json`; path fields `problem_id`; required record fields `problem_id`;
  accepted fields `problem_id`; risk: Create a Time Entry through the Freshservice API.
- `update_problem_time_entry`: PUT `/problems/{{ record.problem_id }}/time_entries/{{
  record.time_entry_id }}` - kind `update`; body type `json`; path fields `problem_id`,
  `time_entry_id`; required record fields `problem_id`, `time_entry_id`; accepted fields
  `problem_id`, `time_entry_id`; risk: Update a Time Entry through the Freshservice API.
- `delete_a_problem_time_entry`: DELETE `/problems/{{ record.problem_id }}/time_entries/{{
  record.time_entry_id }}` - kind `delete`; body type `none`; path fields `problem_id`,
  `time_entry_id`; required record fields `problem_id`, `time_entry_id`; accepted fields
  `problem_id`, `time_entry_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete a Time Entry.
- `create_problem_task`: POST `/problems/{{ record.problem_id }}/tasks` - kind `create`; body type
  `json`; path fields `problem_id`; required record fields `problem_id`; accepted fields
  `problem_id`; risk: Create a Task through the Freshservice API.
- `update_a_problem_task`: PUT `/problems/{{ record.problem_id }}/tasks/{{ record.task_id }}` - kind
  `update`; body type `json`; path fields `problem_id`, `task_id`; required record fields
  `problem_id`, `task_id`; accepted fields `problem_id`, `task_id`; risk: Update a Task through the
  Freshservice API.
- `delete_a_problem_task`: DELETE `/problems/{{ record.problem_id }}/tasks/{{ record.task_id }}` -
  kind `delete`; body type `none`; path fields `problem_id`, `task_id`; required record fields
  `problem_id`, `task_id`; accepted fields `problem_id`, `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  Task.
- `create_change`: POST `/changes` - kind `create`; body type `json`; risk: Create a Change through
  the Freshservice API.
- `update_change_priority`: PUT `/changes/{{ record.change_id }}` - kind `update`; body type `json`;
  path fields `change_id`; required record fields `change_id`; accepted fields `change_id`; risk:
  Update a Change through the Freshservice API.
- `move_a_change`: PUT `/changes/{{ record.change_id }}/move_workspace` - kind `update`; body type
  `json`; path fields `change_id`; required record fields `change_id`; accepted fields `change_id`;
  risk: Move a Change through the Freshservice API.
- `delete_a_change`: DELETE `/changes/{{ record.change_id }}` - kind `delete`; body type `none`;
  path fields `change_id`; required record fields `change_id`; accepted fields `change_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Change.
- `resend_reminder_change_approval`: PUT `/changes/{{ record.change_id }}/approvals/{{
  record.approval_id }}/remind` - kind `update`; body type `json`; path fields `change_id`,
  `approval_id`; required record fields `change_id`, `approval_id`; accepted fields `approval_id`,
  `change_id`; risk: Send Reminder for a Change Approval through the Freshservice API.
- `cancel_change_approval`: PUT `/changes/{{ record.change_id }}/approvals/{{ record.approval_id }}`
  - kind `update`; body type `json`; path fields `change_id`, `approval_id`; required record fields
  `change_id`, `approval_id`; accepted fields `approval_id`, `change_id`; risk: Cancel a Change
  Approval through the Freshservice API.
- `create_change_approval_groups`: POST `/changes/{{ record.change_id }}/approval-groups` - kind
  `create`; body type `json`; path fields `change_id`; required record fields `change_id`; accepted
  fields `change_id`; risk: Create an Approval Group for Changes through the Freshservice API.
- `update_change_approval_groups`: PUT `/changes/{{ record.change_id }}/approval-groups/{{
  record.approval_group_id }}` - kind `update`; body type `json`; path fields `change_id`,
  `approval_group_id`; required record fields `change_id`, `approval_group_id`; accepted fields
  `approval_group_id`, `change_id`; risk: Update Approval Group In A Change Request through the
  Freshservice API.
- `update_approval_chain_rule_change`: PUT `/changes/{{ record.change_id }}/approval-chain-rule` -
  kind `update`; body type `json`; path fields `change_id`; required record fields `change_id`;
  accepted fields `change_id`; risk: Update approval chain rule for a change through the
  Freshservice API.
- `create_change_note`: POST `/changes/{{ record.change_id }}/notes` - kind `create`; body type
  `json`; path fields `change_id`; required record fields `change_id`; accepted fields `change_id`;
  risk: Create a Note through the Freshservice API.
- `update_a_change_note`: PUT `/changes/{{ record.change_id }}/notes/{{ record.note_id }}` - kind
  `update`; body type `json`; path fields `change_id`, `note_id`; required record fields
  `change_id`, `note_id`; accepted fields `change_id`, `note_id`; risk: Update a Note through the
  Freshservice API.
- `delete_a_change_note`: DELETE `/changes/{{ record.change_id }}/notes/{{ record.note_id }}` - kind
  `delete`; body type `none`; path fields `change_id`, `note_id`; required record fields
  `change_id`, `note_id`; accepted fields `change_id`, `note_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Note.
- `create_time_entries`: POST `/changes/{{ record.change_id }}/time_entries` - kind `create`; body
  type `json`; path fields `change_id`; required record fields `change_id`; accepted fields
  `change_id`; risk: Create a Time Entry through the Freshservice API.
- `update_time_entry`: PUT `/changes/{{ record.change_id }}/time_entries/{{ record.time_entry_id }}`
  - kind `update`; body type `json`; path fields `change_id`, `time_entry_id`; required record
  fields `change_id`, `time_entry_id`; accepted fields `change_id`, `time_entry_id`; risk: Update a
  Time Entry through the Freshservice API.
- `delete_a_time_entry`: DELETE `/changes/{{ record.change_id }}/time_entries/{{
  record.time_entry_id }}` - kind `delete`; body type `none`; path fields `change_id`,
  `time_entry_id`; required record fields `change_id`, `time_entry_id`; accepted fields `change_id`,
  `time_entry_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Time Entry.
- `create_change_task`: POST `/changes/{{ record.change_id }}/tasks` - kind `create`; body type
  `json`; path fields `change_id`; required record fields `change_id`; accepted fields `change_id`;
  risk: Create a Task through the Freshservice API.
- `update_a_change_task`: PUT `/changes/{{ record.change_id }}/tasks/{{ record.task_id }}` - kind
  `update`; body type `json`; path fields `change_id`, `task_id`; required record fields
  `change_id`, `task_id`; accepted fields `change_id`, `task_id`; risk: Update a Task through the
  Freshservice API.
- `delete_a_change_task`: DELETE `/changes/{{ record.change_id }}/tasks/{{ record.task_id }}` - kind
  `delete`; body type `none`; path fields `change_id`, `task_id`; required record fields
  `change_id`, `task_id`; accepted fields `change_id`, `task_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Task.
- `create_release`: POST `/releases` - kind `create`; body type `json`; risk: Create a Release
  through the Freshservice API.
- `update_release_priority`: PUT `/releases/{{ record.release_id }}` - kind `update`; body type
  `json`; path fields `release_id`; required record fields `release_id`; accepted fields
  `release_id`; risk: Update a Release through the Freshservice API.
- `move_a_release`: PUT `/releases/{{ record.release_id }}/move_workspace` - kind `update`; body
  type `json`; path fields `release_id`; required record fields `release_id`; accepted fields
  `release_id`; risk: Move a Release through the Freshservice API.
- `delete_a_release`: DELETE `/releases/{{ record.release_id }}` - kind `delete`; body type `none`;
  path fields `release_id`; required record fields `release_id`; accepted fields `release_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Release.
- `restore_a_release`: PUT `/releases/{{ record.release_id }}/restore` - kind `update`; body type
  `json`; path fields `release_id`; required record fields `release_id`; accepted fields
  `release_id`; risk: Restore a Release through the Freshservice API.
- `create_release_note`: POST `/releases/{{ record.release_id }}/notes` - kind `create`; body type
  `json`; path fields `release_id`; required record fields `release_id`; accepted fields
  `release_id`; risk: Create a note through the Freshservice API.
- `update_a_release_note`: PUT `/releases/{{ record.release_id }}/notes/{{ record.note_id }}` - kind
  `update`; body type `json`; path fields `release_id`, `note_id`; required record fields
  `release_id`, `note_id`; accepted fields `note_id`, `release_id`; risk: Update a note through the
  Freshservice API.
- `delete_a_release_note`: DELETE `/releases/{{ record.release_id }}/notes/{{ record.note_id }}` -
  kind `delete`; body type `none`; path fields `release_id`, `note_id`; required record fields
  `release_id`, `note_id`; accepted fields `note_id`, `release_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  note.
- `create_release_time_entries`: POST `/releases/{{ record.release_id }}/time_entries` - kind
  `create`; body type `json`; path fields `release_id`; required record fields `release_id`;
  accepted fields `release_id`; risk: Create a Time Entry through the Freshservice API.
- `update_release_time_entry`: PUT `/releases/{{ record.release_id }}/time_entries/{{
  record.time_entry_id }}` - kind `update`; body type `json`; path fields `release_id`,
  `time_entry_id`; required record fields `release_id`, `time_entry_id`; accepted fields
  `release_id`, `time_entry_id`; risk: Update a Time Entry through the Freshservice API.
- `delete_a_release_time_entry`: DELETE `/releases/{{ record.release_id }}/time_entries/{{
  record.time_entry_id }}` - kind `delete`; body type `none`; path fields `release_id`,
  `time_entry_id`; required record fields `release_id`, `time_entry_id`; accepted fields
  `release_id`, `time_entry_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete a Time Entry.
- `create_release_task`: POST `/releases/{{ record.release_id }}/tasks` - kind `create`; body type
  `json`; path fields `release_id`; required record fields `release_id`; accepted fields
  `release_id`; risk: Create a Task through the Freshservice API.
- `update_a_release_task`: PUT `/releases/{{ record.release_id }}/tasks/{{ record.task_id }}` - kind
  `update`; body type `json`; path fields `release_id`, `task_id`; required record fields
  `release_id`, `task_id`; accepted fields `release_id`, `task_id`; risk: Update a Task through the
  Freshservice API.
- `delete_a_release_task`: DELETE `/releases/{{ record.release_id }}/tasks/{{ record.task_id }}` -
  kind `delete`; body type `none`; path fields `release_id`, `task_id`; required record fields
  `release_id`, `task_id`; accepted fields `release_id`, `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  Task.
- `update_a_cab`: PATCH `/cabs/{{ record.cab_id }}` - kind `update`; body type `json`; path fields
  `cab_id`; required record fields `cab_id`; accepted fields `cab_id`; risk: Update a CAB through
  the Freshservice API.
- `delete_a_cab`: DELETE `/cabs/{{ record.cab_id }}` - kind `delete`; body type `none`; path fields
  `cab_id`; required record fields `cab_id`; accepted fields `cab_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  CAB.
- `create_a_requester`: POST `/requesters` - kind `create`; body type `json`; risk: Create a
  Requester/Contact through the Freshservice API.
- `update_a_requester`: PUT `/requesters/{{ record.requester_id }}` - kind `update`; body type
  `json`; path fields `requester_id`; required record fields `requester_id`; accepted fields
  `requester_id`; risk: Update a Requester/Contact through the Freshservice API.
- `deactivate_a_requester`: DELETE `/requesters/{{ record.requester_id }}` - kind `delete`; body
  type `none`; path fields `requester_id`; required record fields `requester_id`; accepted fields
  `requester_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Deactivate a Requester/Contact.
- `forget_a_requester`: DELETE `/requesters/{{ record.requester_id }}/forget` - kind `delete`; body
  type `none`; path fields `requester_id`; required record fields `requester_id`; accepted fields
  `requester_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Forget a Requester/Contact.
- `convert_to_agent`: PUT `/requesters/{{ record.requester_id }}/convert_to_agent` - kind `update`;
  body type `json`; path fields `requester_id`; required record fields `requester_id`; accepted
  fields `requester_id`; risk: Convert To Agent through the Freshservice API.
- `merge_requesters`: PUT `/requesters/{{ record.requester_id
  }}/merge?secondary_requesters=111,222,333` - kind `update`; body type `json`; path fields
  `requester_id`; required record fields `requester_id`; accepted fields `requester_id`; risk: Merge
  Requesters/Contacts through the Freshservice API.
- `create_an_agent`: POST `/agents` - kind `create`; body type `json`; risk: Create an Agent through
  the Freshservice API.
- `update_an_agent`: PUT `/agents/{{ record.agent_id }}` - kind `update`; body type `json`; path
  fields `agent_id`; required record fields `agent_id`; accepted fields `agent_id`; risk: Update an
  Agent through the Freshservice API.
- `delete_an_agent`: DELETE `/agents/{{ record.agent_id }}` - kind `delete`; body type `none`; path
  fields `agent_id`; required record fields `agent_id`; accepted fields `agent_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data:
  Deactivate an Agent.
- `forget_an_agent`: DELETE `/agents/{{ record.agent_id }}/forget` - kind `delete`; body type
  `none`; path fields `agent_id`; required record fields `agent_id`; accepted fields `agent_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Forget an Agent.
- `reactivate_an_agent`: PUT `/agents/{{ record.agent_id }}/reactivate` - kind `update`; body type
  `json`; path fields `agent_id`; required record fields `agent_id`; accepted fields `agent_id`;
  risk: Reactivate an Agent through the Freshservice API.
- `convert_an_agent_to_requester`: PUT `/agents/{{ record.agent_id }}/convert_to_requester` - kind
  `update`; body type `json`; path fields `agent_id`; required record fields `agent_id`; accepted
  fields `agent_id`; risk: Convert To Requester through the Freshservice API.
- `create_a_group`: POST `/groups` - kind `create`; body type `json`; risk: Create a Group through
  the Freshservice API.
- `update_a_group`: PUT `/groups/{{ record.group_id }}` - kind `update`; body type `json`; path
  fields `group_id`; required record fields `group_id`; accepted fields `group_id`; risk: Update a
  Group through the Freshservice API.
- `delete_a_group`: DELETE `/groups/{{ record.group_id }}` - kind `delete`; body type `none`; path
  fields `group_id`; required record fields `group_id`; accepted fields `group_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data:
  Delete a Group.
- `create_a_requester_group`: POST `/requester_groups` - kind `create`; body type `json`; risk:
  Create a Requester Group/Contact Group through the Freshservice API.
- `update_a_requester_group`: PUT `/requester_groups/{{ record.requester_group_id }}` - kind
  `update`; body type `json`; path fields `requester_group_id`; required record fields
  `requester_group_id`; accepted fields `requester_group_id`; risk: Update a Requester Group/Contact
  Group through the Freshservice API.
- `delete_a_requester_group`: DELETE `/requester_groups/{{ record.requester_group_id }}` - kind
  `delete`; body type `none`; path fields `requester_group_id`; required record fields
  `requester_group_id`; accepted fields `requester_group_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Requester
  Group/Contact Group.
- `add_member_to_requester_group`: POST `/requester_groups/{{ record.requester_group_id
  }}/members/{{ record.requester_id }}` - kind `create`; body type `json`; path fields
  `requester_group_id`, `requester_id`; required record fields `requester_group_id`, `requester_id`;
  accepted fields `requester_group_id`, `requester_id`; risk: Add Requester to Requester/Contact
  Group through the Freshservice API.
- `delete_member_from_requester_group`: DELETE `/requester_groups/{{ record.requester_group_id
  }}/members/{{ record.requester_id }}` - kind `delete`; body type `none`; path fields
  `requester_group_id`, `requester_id`; required record fields `requester_group_id`, `requester_id`;
  accepted fields `requester_group_id`, `requester_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete
  Requester/Contact from Requester/Contact Group.
- `create_a_location`: POST `/locations` - kind `create`; body type `json`; risk: Create a Location
  through the Freshservice API.
- `update_a_location`: PUT `/locations/{{ record.location_id }}` - kind `update`; body type `json`;
  path fields `location_id`; required record fields `location_id`; accepted fields `location_id`;
  risk: Update a Location through the Freshservice API.
- `delete_a_location`: DELETE `/locations/{{ record.location_id }}` - kind `delete`; body type
  `none`; path fields `location_id`; required record fields `location_id`; accepted fields
  `location_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Location.
- `create_a_product`: POST `/products` - kind `create`; body type `json`; risk: Create a Product
  through the Freshservice API.
- `update_a_product`: PUT `/products/{{ record.product_id }}` - kind `update`; body type `json`;
  path fields `product_id`; required record fields `product_id`; accepted fields `product_id`; risk:
  Update a Product through the Freshservice API.
- `delete_a_product`: DELETE `/products/{{ record.product_id }}` - kind `delete`; body type `none`;
  path fields `product_id`; required record fields `product_id`; accepted fields `product_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Product.
- `create_a_vendor`: POST `/vendors` - kind `create`; body type `json`; risk: Create a Vendor
  through the Freshservice API.
- `update_a_vendor`: PUT `/vendors/{{ record.vendor_id }}` - kind `update`; body type `json`; path
  fields `vendor_id`; required record fields `vendor_id`; accepted fields `vendor_id`; risk: Update
  a Vendor through the Freshservice API.
- `delete_a_vendor`: DELETE `/vendors/{{ record.vendor_id }}` - kind `delete`; body type `none`;
  path fields `vendor_id`; required record fields `vendor_id`; accepted fields `vendor_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Vendor.
- `acknowledge_alert`: PUT `/ams/alerts/{{ record.alert_id }}/acknowledge` - kind `update`; body
  type `json`; path fields `alert_id`; required record fields `alert_id`; accepted fields
  `alert_id`; risk: Acknowledge an alert through the Freshservice API.
- `resolve_alert`: PUT `/ams/alerts/{{ record.alert_id }}/resolve` - kind `update`; body type
  `json`; path fields `alert_id`; required record fields `alert_id`; accepted fields `alert_id`;
  risk: Resolve an alert through the Freshservice API.
- `suppress_alert`: PUT `/ams/alerts/{{ record.alert_id }}/suppress` - kind `update`; body type
  `json`; path fields `alert_id`; required record fields `alert_id`; accepted fields `alert_id`;
  risk: Suppress an alert through the Freshservice API.
- `unsuppress_alert`: PUT `/ams/alerts/{{ record.alert_id }}/unsuppress` - kind `update`; body type
  `json`; path fields `alert_id`; required record fields `alert_id`; accepted fields `alert_id`;
  risk: Unsuppress an alert through the Freshservice API.
- `delete_alert`: DELETE `/ams/alerts/{{ record.alert_id }}` - kind `delete`; body type `none`; path
  fields `alert_id`; required record fields `alert_id`; accepted fields `alert_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data:
  Delete an Alert.
- `create_alert_note`: POST `/ams/alerts/{{ record.alert_id }}/notes` - kind `create`; body type
  `json`; path fields `alert_id`; required record fields `alert_id`; accepted fields `alert_id`;
  risk: Create an Alert Note through the Freshservice API.
- `update_alert_note`: PUT `/ams/alerts/{{ record.alert_id }}/notes/{{ record.note_id }}` - kind
  `update`; body type `json`; path fields `alert_id`, `note_id`; required record fields `alert_id`,
  `note_id`; accepted fields `alert_id`, `note_id`; risk: Update an alert note through the
  Freshservice API.
- `delete_alert_note`: DELETE `/ams/alerts/{{ record.alert_id }}/notes/{{ record.note_id }}` - kind
  `delete`; body type `none`; path fields `alert_id`, `note_id`; required record fields `alert_id`,
  `note_id`; accepted fields `alert_id`, `note_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete an alert note.
- `create_an_asset`: POST `/assets` - kind `create`; body type `json`; risk: Create an Asset through
  the Freshservice API.
- `update_an_asset`: PUT `/assets/{{ record.display_id }}` - kind `update`; body type `json`; path
  fields `display_id`; required record fields `display_id`; accepted fields `display_id`; risk:
  Update an Asset through the Freshservice API.
- `delete_an_asset`: DELETE `/assets/{{ record.display_id }}` - kind `delete`; body type `none`;
  path fields `display_id`; required record fields `display_id`; accepted fields `display_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete an Asset.
- `restore_an_asset`: PUT `/assets/{{ record.display_id }}/restore` - kind `update`; body type
  `json`; path fields `display_id`; required record fields `display_id`; accepted fields
  `display_id`; risk: Restore an Asset through the Freshservice API.
- `delete_forever_an_asset`: PUT `/assets/{{ record.display_id }}/delete_forever` - kind `update`;
  body type `json`; path fields `display_id`; required record fields `display_id`; accepted fields
  `display_id`; confirmation `destructive`; risk: Delete an Asset Permanently through the
  Freshservice API.
- `create_a_component`: POST `/assets/{{ record.display_id }}/components` - kind `create`; body type
  `json`; path fields `display_id`; required record fields `display_id`; accepted fields
  `display_id`; risk: Create a Component through the Freshservice API.
- `update_a_component`: PUT `/assets/{{ record.display_id }}/components/{{ record.component_id }}` -
  kind `update`; body type `json`; path fields `display_id`, `component_id`; required record fields
  `display_id`, `component_id`; accepted fields `component_id`, `display_id`; risk: Update a
  Component through the Freshservice API.
- `create_relationships`: POST `/relationships/bulk-create` - kind `create`; body type `json`; risk:
  Create Relationships in bulk through the Freshservice API.
- `delete_relationships`: DELETE `/relationships?ids={{ record.ids }}` - kind `delete`; body type
  `none`; path fields `ids`; required record fields `ids`; accepted fields `ids`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data:
  Delete Relationships in bulk.
- `create_a_new_purchase_order`: POST `/purchase_orders` - kind `create`; body type `json`; risk:
  Create a new Purchase Order through the Freshservice API.
- `update_a_purchase_order`: PUT `/purchase_orders/{{ record.purchase_order_id }}` - kind `update`;
  body type `json`; path fields `purchase_order_id`; required record fields `purchase_order_id`;
  accepted fields `purchase_order_id`; risk: Update a Purchase Order through the Freshservice API.
- `delete_a_purchase_order`: DELETE `/purchase_orders/{{ record.purchase_order_id }}` - kind
  `delete`; body type `none`; path fields `purchase_order_id`; required record fields
  `purchase_order_id`; accepted fields `purchase_order_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Purchase
  Order.
- `create_an_asset_type`: POST `/asset_types` - kind `create`; body type `json`; risk: Create an
  Asset Type through the Freshservice API.
- `update_an_asset_type`: PUT `/asset_types/{{ record.asset_type_id }}` - kind `update`; body type
  `json`; path fields `asset_type_id`; required record fields `asset_type_id`; accepted fields
  `asset_type_id`; risk: Update an Asset Type through the Freshservice API.
- `delete_an_asset_type`: DELETE `/asset_types/{{ record.asset_type_id }}` - kind `delete`; body
  type `none`; path fields `asset_type_id`; required record fields `asset_type_id`; accepted fields
  `asset_type_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete an Asset Type.
- `create_a_software`: POST `/applications` - kind `create`; body type `json`; risk: Create a
  Software through the Freshservice API.
- `update_a_software`: PUT `/applications/{{ record.application_id }}` - kind `update`; body type
  `json`; path fields `application_id`; required record fields `application_id`; accepted fields
  `application_id`; risk: Update a Software through the Freshservice API.
- `delete_a_software`: DELETE `/applications/{{ record.application_id }}` - kind `delete`; body type
  `none`; path fields `application_id`; required record fields `application_id`; accepted fields
  `application_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Software.
- `delete_multiple_software`: DELETE `/applications` - kind `delete`; body type `none`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete multiple Software.
- `create_a_contract`: POST `/contracts` - kind `create`; body type `json`; risk: Create a Contract
  through the Freshservice API.
- `update_a_contract`: PUT `/contracts/{{ record.contract_id }}` - kind `update`; body type `json`;
  path fields `contract_id`; required record fields `contract_id`; accepted fields `contract_id`;
  risk: Update a Contract through the Freshservice API.
- `submit_contract_for_approval`: PUT `/contracts/{{ record.contract_id
  }}?operation=submit-for-approval` - kind `update`; body type `json`; path fields `contract_id`;
  required record fields `contract_id`; accepted fields `contract_id`; risk: Submit a contract for
  approval through the Freshservice API.
- `approve_contract`: PUT `/contracts/{{ record.contract_id }}?operation=approve` - kind `update`;
  body type `json`; path fields `contract_id`; required record fields `contract_id`; accepted fields
  `contract_id`; risk: Approve a Contract through the Freshservice API.
- `reject_contract`: PUT `/contracts/{{ record.contract_id }}?operation=reject` - kind `update`;
  body type `json`; path fields `contract_id`; required record fields `contract_id`; accepted fields
  `contract_id`; risk: Reject a Contract through the Freshservice API.
- `create_a_department`: POST `/departments` - kind `create`; body type `json`; risk: Create a
  Department through the Freshservice API.
- `update_a_department`: PUT `/departments/{{ record.department_id }}` - kind `update`; body type
  `json`; path fields `department_id`; required record fields `department_id`; accepted fields
  `department_id`; risk: Update a Department through the Freshservice API.
- `delete_a_department`: DELETE `/departments/{{ record.department_id }}` - kind `delete`; body type
  `none`; path fields `department_id`; required record fields `department_id`; accepted fields
  `department_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Department.
- `create_a_project`: POST `/projects` - kind `create`; body type `json`; risk: Create a Project
  through the Freshservice API.
- `update_a_project`: PUT `/projects/{{ record.project_id }}` - kind `update`; body type `json`;
  path fields `project_id`; required record fields `project_id`; accepted fields `project_id`; risk:
  Update a Project through the Freshservice API.
- `delete_a_project`: DELETE `/projects/{{ record.project_id }}` - kind `delete`; body type `none`;
  path fields `project_id`; required record fields `project_id`; accepted fields `project_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Project.
- `archive_a_project`: POST `/projects/{{ record.project_id }}/archive` - kind `update`; body type
  `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Archive a Project through the Freshservice API.
- `restore_a_project`: POST `/projects/{{ record.project_id }}/restore` - kind `update`; body type
  `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Restore a Project through the Freshservice API.
- `create_a_project_task`: POST `/projects/{{ record.project_id }}/tasks` - kind `create`; body type
  `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Create a Project Task through the Freshservice API.
- `update_a_project_task`: PUT `/projects/{{ record.project_id }}/task/{{ record.task_id }}` - kind
  `update`; body type `json`; path fields `project_id`, `task_id`; required record fields
  `project_id`, `task_id`; accepted fields `project_id`, `task_id`; risk: Update a Project Task
  through the Freshservice API.
- `delete_a_project_task`: DELETE `/projects/{{ record.project_id }}/tasks/{{ record.task_id }}` -
  kind `delete`; body type `none`; path fields `project_id`, `task_id`; required record fields
  `project_id`, `task_id`; accepted fields `project_id`, `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  Project Task.
- `create_a_project_newgen`: POST `/pm/projects` - kind `create`; body type `json`; risk: Create a
  Project through the Freshservice API.
- `update_a_project_newgen`: PUT `/pm/projects/{{ record.project_id }}` - kind `update`; body type
  `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Update a Project through the Freshservice API.
- `delete_a_project_newgen`: DELETE `/pm/projects/{{ record.project_id }}` - kind `delete`; body
  type `none`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Project.
- `archive_a_project_newgen`: POST `/pm/projects/{{ record.project_id }}/archive` - kind `update`;
  body type `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Archive a Project through the Freshservice API.
- `restore_a_project_newgen`: POST `/pm/projects/{{ record.project_id }}/restore` - kind `update`;
  body type `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Restore a Project through the Freshservice API.
- `add_project_members`: POST `/pm/projects/{{ record.project_id }}/members` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Add Members through the Freshservice API.
- `create_project_associations`: POST `/pm/projects/{{ record.project_id }}/{{ record.module_name
  }}` - kind `create`; body type `json`; path fields `project_id`, `module_name`; required record
  fields `project_id`, `module_name`; accepted fields `module_name`, `project_id`; risk: Create
  Associations through the Freshservice API.
- `delete_project_association`: DELETE `/pm/projects/{{ record.project_id }}/{{ record.module_name
  }}/{{ record.project_id_3 }}` - kind `delete`; body type `none`; path fields `project_id`,
  `module_name`, `project_id_3`; required record fields `project_id`, `module_name`, `project_id_3`;
  accepted fields `module_name`, `project_id`, `project_id_3`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete Association.
- `delete_attachment_project_newgen`: DELETE `/pm/projects/{{ record.project_id }}/attachments/{{
  record.attachment_id }}` - kind `delete`; body type `none`; path fields `project_id`,
  `attachment_id`; required record fields `project_id`, `attachment_id`; accepted fields
  `attachment_id`, `project_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete Attachment of a Project.
- `create_a_project_task_newgen`: POST `/pm/projects/{{ record.project_id }}/tasks` - kind `create`;
  body type `json`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: Create a Project Task through the Freshservice API.
- `update_a_project_task_newgen`: PUT `/pm/projects/{{ record.project_id }}/tasks/{{ record.task_id
  }}` - kind `update`; body type `json`; path fields `project_id`, `task_id`; required record fields
  `project_id`, `task_id`; accepted fields `project_id`, `task_id`; risk: Update a Project Task
  through the Freshservice API.
- `delete_a_project_task_newgen`: DELETE `/pm/projects/{{ record.project_id }}/tasks/{{
  record.task_id }}` - kind `delete`; body type `none`; path fields `project_id`, `task_id`;
  required record fields `project_id`, `task_id`; accepted fields `project_id`, `task_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a Project Task.
- `create_project_task_associations`: POST `/pm/projects/{{ record.project_id }}/tasks/{{
  record.task_id }}/{{ record.module_name }}` - kind `create`; body type `json`; path fields
  `project_id`, `task_id`, `module_name`; required record fields `project_id`, `task_id`,
  `module_name`; accepted fields `module_name`, `project_id`, `task_id`; risk: Create Associations
  through the Freshservice API.
- `delete_project_task_association`: DELETE `/pm/projects/{{ record.project_id }}/tasks/{{
  record.task_id }}/{{ record.module_name }}/{{ record.task_id_4 }}` - kind `delete`; body type
  `none`; path fields `project_id`, `task_id`, `module_name`, `task_id_4`; required record fields
  `project_id`, `task_id`, `module_name`, `task_id_4`; accepted fields `module_name`, `project_id`,
  `task_id`, `task_id_4`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete Association.
- `create_note_task_newgen`: POST `/pm/projects/{{ record.project_id }}/tasks/{{ record.task_id
  }}/notes` - kind `create`; body type `json`; path fields `project_id`, `task_id`; required record
  fields `project_id`, `task_id`; accepted fields `project_id`, `task_id`; risk: Create Note through
  the Freshservice API.
- `update_note_task_newgen`: PUT `/pm/projects/{{ record.project_id }}/tasks/{{ record.task_id
  }}/notes/{{ record.note_id }}` - kind `update`; body type `json`; path fields `project_id`,
  `task_id`, `note_id`; required record fields `project_id`, `task_id`, `note_id`; accepted fields
  `note_id`, `project_id`, `task_id`; risk: Update Note through the Freshservice API.
- `delete_note_task_newgen`: DELETE `/pm/projects/{{ record.project_id }}/tasks/{{ record.task_id
  }}/notes/{{ record.note_id }}` - kind `delete`; body type `none`; path fields `project_id`,
  `task_id`, `note_id`; required record fields `project_id`, `task_id`, `note_id`; accepted fields
  `note_id`, `project_id`, `task_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes Freshservice data: Delete Note.
- `delete_attachment_note_task_newgen`: DELETE `/pm/projects/{{ record.project_id }}/tasks/{{
  record.task_id }}/notes/{{ record.note_id }}/attachments/{{ record.attachment_id }}` - kind
  `delete`; body type `none`; path fields `project_id`, `task_id`, `note_id`, `attachment_id`;
  required record fields `project_id`, `task_id`, `note_id`, `attachment_id`; accepted fields
  `attachment_id`, `note_id`, `project_id`, `task_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete Attachment of a Note.
- `delete_attachment_task_newgen`: DELETE `/pm/projects/{{ record.project_id }}/tasks/{{
  record.task_id }}/attachments/{{ record.attachment_id }}` - kind `delete`; body type `none`; path
  fields `project_id`, `task_id`, `attachment_id`; required record fields `project_id`, `task_id`,
  `attachment_id`; accepted fields `attachment_id`, `project_id`, `task_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete
  Attachment of a Task.
- `create_solution_category`: POST `/solutions/categories` - kind `create`; body type `json`; risk:
  Create Solution Category through the Freshservice API.
- `update_solution_category`: PUT `/solutions/categories/{{ record.category_id }}` - kind `update`;
  body type `json`; path fields `category_id`; required record fields `category_id`; accepted fields
  `category_id`; risk: Update Solution Category through the Freshservice API.
- `delete_solution_category`: DELETE `/solutions/categories/{{ record.category_id }}` - kind
  `delete`; body type `none`; path fields `category_id`; required record fields `category_id`;
  accepted fields `category_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete Solution Category.
- `restore_solution_category`: PUT `/solutions/categories/{{ record.category_id }}/restore` - kind
  `update`; body type `json`; path fields `category_id`; required record fields `category_id`;
  accepted fields `category_id`; risk: Restore Solution Category through the Freshservice API.
- `delete_forever_solution_category`: DELETE `/solutions/categories/{{ record.category_id
  }}/delete_forever` - kind `delete`; body type `none`; path fields `category_id`; required record
  fields `category_id`; accepted fields `category_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Deletes Freshservice data: Permanently Delete Solution
  Category.
- `create_solution_folder`: POST `/solutions/folders` - kind `create`; body type `json`; risk:
  Create Solution Folder through the Freshservice API.
- `update_solution_folder`: PUT `/solutions/folders/{{ record.folder_id }}` - kind `update`; body
  type `json`; path fields `folder_id`; required record fields `folder_id`; accepted fields
  `folder_id`; risk: Update Solution Folder through the Freshservice API.
- `delete_solution_folder`: DELETE `/solutions/folders/{{ record.folder_id }}` - kind `delete`; body
  type `none`; path fields `folder_id`; required record fields `folder_id`; accepted fields
  `folder_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete Solution Folder.
- `restore_solution_folder`: PUT `/solutions/folders/{{ record.folder_id }}/restore` - kind
  `update`; body type `json`; path fields `folder_id`; required record fields `folder_id`; accepted
  fields `folder_id`; risk: Restore Solution Folder through the Freshservice API.
- `delete_forever_solution_folder`: DELETE `/solutions/folders/{{ record.folder_id
  }}/delete_forever` - kind `delete`; body type `none`; path fields `folder_id`; required record
  fields `folder_id`; accepted fields `folder_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Deletes Freshservice data: Permanently Delete Solution
  Folder.
- `create_solution_article`: POST `/solutions/articles` - kind `create`; body type `json`; risk:
  Create Solution Article through the Freshservice API.
- `send_article_to_approval`: PUT `/solutions/articles/{{ record.article_id }}/send_for_approval` -
  kind `update`; body type `json`; path fields `article_id`; required record fields `article_id`;
  accepted fields `article_id`; risk: Send Article to Approval through the Freshservice API.
- `publish_solution_article`: PUT `/solutions/articles/{{ record.article_id }}` - kind `update`;
  body type `json`; path fields `article_id`; required record fields `article_id`; accepted fields
  `article_id`; risk: Publish Solution Article through the Freshservice API.
- `delete_solution_article`: DELETE `/solutions/articles/{{ record.article_id }}` - kind `delete`;
  body type `none`; path fields `article_id`; required record fields `article_id`; accepted fields
  `article_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete Solution Article.
- `restore_solution_article`: PUT `/solutions/articles/{{ record.article_id }}/restore` - kind
  `update`; body type `json`; path fields `article_id`; required record fields `article_id`;
  accepted fields `article_id`; risk: Restore Solution Article through the Freshservice API.
- `delete_forever_solution_article`: DELETE `/solutions/articles/{{ record.article_id
  }}/delete_forever` - kind `delete`; body type `none`; path fields `article_id`; required record
  fields `article_id`; accepted fields `article_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Deletes Freshservice data: Permanently Delete Solution
  Article.
- `bulk_restore_solution_article`: PUT `/solutions/articles/bulk_restore` - kind `update`; body type
  `json`; risk: Bulk Restore Solution Articles through the Freshservice API.
- `update_service_item`: PUT `/service-catalog/items/{{ record.service_item_id }}` - kind `update`;
  body type `json`; path fields `service_item_id`; required record fields `service_item_id`;
  accepted fields `service_item_id`; risk: Update a Service Catalog Item through the Freshservice
  API.
- `delete_a_service_item`: DELETE `/service-catalog/items/{{ record.service_item_id }}` - kind
  `delete`; body type `none`; path fields `service_item_id`; required record fields
  `service_item_id`; accepted fields `service_item_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a Service
  Catalog Item.
- `create_shared_fields`: POST `/service-catalog/shared-fields` - kind `create`; body type `json`;
  risk: Create a shared field through the Freshservice API.
- `update_shared_fields`: PUT `/service-catalog/shared-fields/{{ record.shared_field_id }}` - kind
  `update`; body type `json`; path fields `shared_field_id`; required record fields
  `shared_field_id`; accepted fields `shared_field_id`; risk: Update shared fields through the
  Freshservice API.
- `delete_shared_fields`: DELETE `/service-catalog/shared-fields/{{ record.shared_field_id }}` -
  kind `delete`; body type `none`; path fields `shared_field_id`; required record fields
  `shared_field_id`; accepted fields `shared_field_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete shared fields.
- `archive_shared_fields`: POST `/service-catalog/shared-fields/{{ record.shared_field_id
  }}/archive` - kind `update`; body type `json`; path fields `shared_field_id`; required record
  fields `shared_field_id`; accepted fields `shared_field_id`; risk: Archive shared fields through
  the Freshservice API.
- `unarchive_shared_fields`: POST `/service-catalog/shared-fields/{{ record.shared_field_id
  }}/unarchive` - kind `update`; body type `json`; path fields `shared_field_id`; required record
  fields `shared_field_id`; accepted fields `shared_field_id`; risk: Unarchive shared fields through
  the Freshservice API.
- `create_an_announcement`: POST `/announcements` - kind `create`; body type `json`; risk: Create an
  Announcement through the Freshservice API.
- `edit_an_announcement`: PUT `/announcements/{{ record.announcement_id }}` - kind `update`; body
  type `json`; path fields `announcement_id`; required record fields `announcement_id`; accepted
  fields `announcement_id`; risk: Edit an Announcement through the Freshservice API.
- `delete_an_announcement`: DELETE `/announcements/{{ record.announcement_id }}` - kind `delete`;
  body type `none`; path fields `announcement_id`; required record fields `announcement_id`;
  accepted fields `announcement_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes Freshservice data: Delete an Announcement.
- `create_onboarding_request`: POST `/onboarding_requests` - kind `create`; body type `json`; risk:
  Create an Onboarding Request through the Freshservice API.
- `create_offboarding_request`: POST `/offboarding_requests` - kind `create`; body type `json`;
  risk: Create an Offboarding Request through the Freshservice API.
- `create_journey_request`: POST `/journeys/requests` - kind `create`; body type `json`; risk:
  Create a Journey Request through the Freshservice API.
- `filter_journey_requests`: POST `/journeys/requests/view` - kind `create`; body type `json`; risk:
  Filter Journey Requests through the Freshservice API.
- `update_journey_request`: PATCH `/journeys/requests/{{ record.request_id }}` - kind `update`; body
  type `json`; path fields `request_id`; required record fields `request_id`; accepted fields
  `request_id`; risk: Update a Journey Request through the Freshservice API.
- `cancel_journey_request`: PUT `/journeys/requests/{{ record.request_id }}/cancel` - kind `update`;
  body type `json`; path fields `request_id`; required record fields `request_id`; accepted fields
  `request_id`; risk: Cancel a Journey Request through the Freshservice API.
- `delete_journey_request`: DELETE `/journeys/requests/{{ record.request_id }}` - kind `delete`;
  body type `none`; path fields `request_id`; required record fields `request_id`; accepted fields
  `request_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a Journey Request.
- `create_schedule`: POST `/oncall/ws/{{ record.workspace_id }}/schedules` - kind `create`; body
  type `json`; path fields `workspace_id`; required record fields `workspace_id`; accepted fields
  `workspace_id`; risk: Create a schedule through the Freshservice API.
- `edit_schedule`: PUT `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id }}` -
  kind `update`; body type `json`; path fields `workspace_id`, `schedule_id`; required record fields
  `workspace_id`, `schedule_id`; accepted fields `schedule_id`, `workspace_id`; risk: Update a
  schedule through the Freshservice API.
- `delete_schedule`: DELETE `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id
  }}` - kind `delete`; body type `none`; path fields `workspace_id`, `schedule_id`; required record
  fields `workspace_id`, `schedule_id`; accepted fields `schedule_id`, `workspace_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a schedule.
- `edit_shift`: PUT `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id
  }}/shifts/{{ record.shift_id }}` - kind `update`; body type `json`; path fields `workspace_id`,
  `schedule_id`, `shift_id`; required record fields `workspace_id`, `schedule_id`, `shift_id`;
  accepted fields `schedule_id`, `shift_id`, `workspace_id`; risk: Update a shift through the
  Freshservice API.
- `delete_shift`: DELETE `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id
  }}/shifts/{{ record.shift_id }}` - kind `delete`; body type `none`; path fields `workspace_id`,
  `schedule_id`, `shift_id`; required record fields `workspace_id`, `schedule_id`, `shift_id`;
  accepted fields `schedule_id`, `shift_id`, `workspace_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a shift.
- `override`: PUT `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id }}/shifts/{{
  record.shift_id }}/rosters/override` - kind `update`; body type `json`; path fields
  `workspace_id`, `schedule_id`, `shift_id`; required record fields `workspace_id`, `schedule_id`,
  `shift_id`; accepted fields `schedule_id`, `shift_id`, `workspace_id`; risk: Create/Update/Delete
  an override through the Freshservice API.
- `create_ep`: POST `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id
  }}/escalation-policies` - kind `create`; body type `json`; path fields `workspace_id`,
  `schedule_id`; required record fields `workspace_id`, `schedule_id`; accepted fields
  `schedule_id`, `workspace_id`; risk: Create an escalation policy through the Freshservice API.
- `delete_ep`: DELETE `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id
  }}/escalation-policies/{{ record.ep_id }}` - kind `delete`; body type `none`; path fields
  `workspace_id`, `schedule_id`, `ep_id`; required record fields `workspace_id`, `schedule_id`,
  `ep_id`; accepted fields `ep_id`, `schedule_id`, `workspace_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete an
  escalation policy.
- `reorder_ep`: PUT `/oncall/ws/{{ record.workspace_id }}/schedules/{{ record.schedule_id
  }}/escalation-policies/reorder` - kind `update`; body type `json`; path fields `workspace_id`,
  `schedule_id`; required record fields `workspace_id`, `schedule_id`; accepted fields
  `schedule_id`, `workspace_id`; risk: Reorder an escalation policy through the Freshservice API.
- `create_custom_object_record`: POST `/objects/{{ record.object_id }}/records` - kind `create`;
  body type `json`; path fields `object_id`; required record fields `object_id`; accepted fields
  `object_id`; risk: Create new Custom Object Record through the Freshservice API.
- `put_custom_object_record`: PUT `/objects/{{ record.object_id }}/records/{{ record.record_id }}` -
  kind `update`; body type `json`; path fields `object_id`, `record_id`; required record fields
  `object_id`, `record_id`; accepted fields `object_id`, `record_id`; risk: Update Custom Object
  Record through the Freshservice API.
- `delete_custom_object_record`: DELETE `/objects/{{ record.object_id }}/records/{{ record.record_id
  }}` - kind `delete`; body type `none`; path fields `object_id`, `record_id`; required record
  fields `object_id`, `record_id`; accepted fields `object_id`, `record_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete
  Custom Object Record.
- `create_new_template`: POST `/post-incident-reports/templates` - kind `create`; body type `json`;
  risk: Create new templates for post-incident reports through the Freshservice API.
- `enable_pir_template`: PUT `/post-incident-reports/templates/{{ record.template_id }}` - kind
  `update`; body type `json`; path fields `template_id`; required record fields `template_id`;
  accepted fields `template_id`; risk: Enable template for post-incident reports through the
  Freshservice API.
- `mark_primary_pir_template`: PUT `/post-incident-reports/templates/{{ record.template_id
  }}/mark-as-primary` - kind `update`; body type `json`; path fields `template_id`; required record
  fields `template_id`; accepted fields `template_id`; risk: Set template as primary for
  post-incident reports through the Freshservice API.
- `delete_pir_template`: DELETE `/post-incident-reports/templates/{{ record.template_id }}` - kind
  `delete`; body type `none`; path fields `template_id`; required record fields `template_id`;
  accepted fields `template_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete an existing post-incident report template.
- `trigger_email`: POST `/tickets/{{ record.ticket_id }}/communications` - kind `create`; body type
  `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`;
  risk: Trigger email for a major incident through the Freshservice API.
- `create_incident`: POST `/tickets/{{ record.ticket_id }}/status/pages/{{ record.status_page_id
  }}/incidents` - kind `create`; body type `json`; path fields `ticket_id`, `status_page_id`;
  required record fields `ticket_id`, `status_page_id`; accepted fields `status_page_id`,
  `ticket_id`; risk: Create an incident through the Freshservice API.
- `update_incident`: PUT `/tickets/{{ record.ticket_id }}/status/pages/{{ record.status_page_id
  }}/incidents/{{ record.incident_id }}` - kind `update`; body type `json`; path fields `ticket_id`,
  `status_page_id`, `incident_id`; required record fields `ticket_id`, `status_page_id`,
  `incident_id`; accepted fields `incident_id`, `status_page_id`, `ticket_id`; risk: Update an
  incident through the Freshservice API.
- `delete_incident`: DELETE `/tickets/{{ record.ticket_id }}/status/pages/{{ record.status_page_id
  }}/incidents/{{ record.incident_id }}` - kind `delete`; body type `none`; path fields `ticket_id`,
  `status_page_id`, `incident_id`; required record fields `ticket_id`, `status_page_id`,
  `incident_id`; accepted fields `incident_id`, `status_page_id`, `ticket_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data:
  Delete an incident.
- `create_incident_update`: POST `/tickets/{{ record.ticket_id }}/status/pages/{{
  record.status_page_id }}/incidents/{{ record.incident_id }}/updates` - kind `update`; body type
  `json`; path fields `ticket_id`, `status_page_id`, `incident_id`; required record fields
  `ticket_id`, `status_page_id`, `incident_id`; accepted fields `incident_id`, `status_page_id`,
  `ticket_id`; risk: Create incident update through the Freshservice API.
- `update_incident_update`: PUT `/tickets/{{ record.ticket_id }}/status/pages/{{
  record.status_page_id }}/incidents/{{ record.incident_id }}/updates/{{ record.update_id }}` - kind
  `update`; body type `json`; path fields `ticket_id`, `status_page_id`, `incident_id`, `update_id`;
  required record fields `ticket_id`, `status_page_id`, `incident_id`, `update_id`; accepted fields
  `incident_id`, `status_page_id`, `ticket_id`, `update_id`; risk: Edit an incident update through
  the Freshservice API.
- `delete_incident_update`: DELETE `/tickets/{{ record.ticket_id }}/status/pages/{{
  record.status_page_id }}/incidents/{{ record.incident_id }}/updates/{{ record.update_id }}` - kind
  `delete`; body type `none`; path fields `ticket_id`, `status_page_id`, `incident_id`, `update_id`;
  required record fields `ticket_id`, `status_page_id`, `incident_id`, `update_id`; accepted fields
  `incident_id`, `status_page_id`, `ticket_id`, `update_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete an incident
  update.
- `create_maintenance_from_change`: POST `/changes/{{ record.change_id }}/status/pages/{{
  record.status_page_id }}/maintenances` - kind `create`; body type `json`; path fields `change_id`,
  `status_page_id`; required record fields `change_id`, `status_page_id`; accepted fields
  `change_id`, `status_page_id`; risk: Create a maintenance from a change through the Freshservice
  API.
- `update_maintenance_from_change`: PUT `/changes/{{ record.change_id }}/status/pages/{{
  record.status_page_id }}/maintenances/{{ record.maintenance_id }}` - kind `update`; body type
  `json`; path fields `change_id`, `status_page_id`, `maintenance_id`; required record fields
  `change_id`, `status_page_id`, `maintenance_id`; accepted fields `change_id`, `maintenance_id`,
  `status_page_id`; risk: Update a maintenance from a change through the Freshservice API.
- `delete_maintenance_from_change`: DELETE `/changes/{{ record.change_id }}/status/pages/{{
  record.status_page_id }}/maintenances/{{ record.maintenance_id }}` - kind `delete`; body type
  `none`; path fields `change_id`, `status_page_id`, `maintenance_id`; required record fields
  `change_id`, `status_page_id`, `maintenance_id`; accepted fields `change_id`, `maintenance_id`,
  `status_page_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a maintenance from a change.
- `create_maintenance_from_maintenance_window`: POST `/maintenance-windows/{{
  record.maintenance_window_id }}/status/pages/{{ record.status_page_id }}/maintenances` - kind
  `create`; body type `json`; path fields `maintenance_window_id`, `status_page_id`; required record
  fields `maintenance_window_id`, `status_page_id`; accepted fields `maintenance_window_id`,
  `status_page_id`; risk: Create a maintenance from a maintenance window through the Freshservice
  API.
- `update_maintenance_from_maintenance_window`: PUT `/maintenance-windows/{{
  record.maintenance_window_id }}/status/pages/{{ record.status_page_id }}/maintenances/{{
  record.maintenance_id }}` - kind `update`; body type `json`; path fields `maintenance_window_id`,
  `status_page_id`, `maintenance_id`; required record fields `maintenance_window_id`,
  `status_page_id`, `maintenance_id`; accepted fields `maintenance_id`, `maintenance_window_id`,
  `status_page_id`; risk: Update a maintenance from a maintenance window through the Freshservice
  API.
- `delete_maintenance_from_maintenance_window`: DELETE `/maintenance-windows/{{
  record.maintenance_window_id }}/status/pages/{{ record.status_page_id }}/maintenances/{{
  record.maintenance_id }}` - kind `delete`; body type `none`; path fields `maintenance_window_id`,
  `status_page_id`, `maintenance_id`; required record fields `maintenance_window_id`,
  `status_page_id`, `maintenance_id`; accepted fields `maintenance_id`, `maintenance_window_id`,
  `status_page_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a maintenance from a maintenance window.
- `create_maintenance_update_from_change`: POST `/changes/{{ record.change_id }}/status/pages/{{
  record.status_page_id }}/maintenances/{{ record.maintenance_id }}/updates` - kind `update`; body
  type `json`; path fields `change_id`, `status_page_id`, `maintenance_id`; required record fields
  `change_id`, `status_page_id`, `maintenance_id`; accepted fields `change_id`, `maintenance_id`,
  `status_page_id`; risk: Create a maintenance update from a change through the Freshservice API.
- `update_maintenance_update_from_change`: PUT `/changes/{{ record.change_id }}/status/pages/{{
  record.status_page_id }}/maintenances/{{ record.maintenance_id }}/updates/{{ record.update_id }}`
  - kind `update`; body type `json`; path fields `change_id`, `status_page_id`, `maintenance_id`,
  `update_id`; required record fields `change_id`, `status_page_id`, `maintenance_id`, `update_id`;
  accepted fields `change_id`, `maintenance_id`, `status_page_id`, `update_id`; risk: Update a
  maintenance update from a change through the Freshservice API.
- `delete_maintenance_update_from_change`: DELETE `/changes/{{ record.change_id }}/status/pages/{{
  record.status_page_id }}/maintenances/{{ record.maintenance_id }}/updates/{{ record.update_id }}`
  - kind `delete`; body type `none`; path fields `change_id`, `status_page_id`, `maintenance_id`,
  `update_id`; required record fields `change_id`, `status_page_id`, `maintenance_id`, `update_id`;
  accepted fields `change_id`, `maintenance_id`, `status_page_id`, `update_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data:
  Delete a maintenance update from a change.
- `create_maintenance_update_from_maintenance_window`: POST `/maintenance-windows/{{
  record.maintenance_window_id }}/status/pages/{{ record.status_page_id }}/maintenances/{{
  record.maintenance_id }}/updates` - kind `update`; body type `json`; path fields
  `maintenance_window_id`, `status_page_id`, `maintenance_id`; required record fields
  `maintenance_window_id`, `status_page_id`, `maintenance_id`; accepted fields `maintenance_id`,
  `maintenance_window_id`, `status_page_id`; risk: Create a maintenance update from a maintenance
  window through the Freshservice API.
- `update_maintenance_update_from_maintenance_window`: PUT `/maintenance-windows/{{
  record.maintenance_window_id }}/status/pages/{{ record.status_page_id }}/maintenances/{{
  record.maintenance_id }}/updates/{{ record.update_id }}` - kind `update`; body type `json`; path
  fields `maintenance_window_id`, `status_page_id`, `maintenance_id`, `update_id`; required record
  fields `maintenance_window_id`, `status_page_id`, `maintenance_id`, `update_id`; accepted fields
  `maintenance_id`, `maintenance_window_id`, `status_page_id`, `update_id`; risk: Update a
  maintenance update from a maintenance window through the Freshservice API.
- `delete_maintenance_update_from_maintenance_window`: DELETE `/maintenance-windows/{{
  record.maintenance_window_id }}/status/pages/{{ record.status_page_id }}/maintenances/{{
  record.maintenance_id }}/updates/{{ record.update_id }}` - kind `delete`; body type `none`; path
  fields `maintenance_window_id`, `status_page_id`, `maintenance_id`, `update_id`; required record
  fields `maintenance_window_id`, `status_page_id`, `maintenance_id`, `update_id`; accepted fields
  `maintenance_id`, `maintenance_window_id`, `status_page_id`, `update_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a
  maintenance update from a maintenance window.
- `create_subscriber`: POST `/status/pages/{{ record.status_page_id }}/subscribers` - kind `create`;
  body type `json`; path fields `status_page_id`; required record fields `status_page_id`; accepted
  fields `status_page_id`; risk: Create a subscriber through the Freshservice API.
- `update_subscriber`: PUT `/status/pages/{{ record.status_page_id }}/subscribers/{{
  record.subscriber_id }}` - kind `update`; body type `json`; path fields `status_page_id`,
  `subscriber_id`; required record fields `status_page_id`, `subscriber_id`; accepted fields
  `status_page_id`, `subscriber_id`; risk: Update a subscriber through the Freshservice API.
- `delete_subscriber`: DELETE `/status/pages/{{ record.status_page_id }}/subscribers/{{
  record.subscriber_id }}` - kind `delete`; body type `none`; path fields `status_page_id`,
  `subscriber_id`; required record fields `status_page_id`, `subscriber_id`; accepted fields
  `status_page_id`, `subscriber_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes Freshservice data: Delete a subscriber.
- `create_delegation`: POST `/users/{{ record.user_id }}/delegation` - kind `create`; body type
  `json`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`; risk:
  Create a delegation through the Freshservice API.
- `update_delegation`: PUT `/users/{{ record.user_id }}/delegation` - kind `update`; body type
  `json`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`; risk:
  Update a delegation through the Freshservice API.
- `delete_delegation`: DELETE `/users/{{ record.user_id }}/delegation` - kind `delete`; body type
  `none`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a delegation.
- `create_physical_subtype`: POST `/itam/physical-subtypes` - kind `update`; body type `json`; risk:
  Create or update physical subtypes through the Freshservice API.
- `delete_physical_subtype`: DELETE `/itam/physical-subtypes/{{ record.physical_subtype_id }}` -
  kind `delete`; body type `none`; path fields `physical_subtype_id`; required record fields
  `physical_subtype_id`; accepted fields `physical_subtype_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a physical
  subtype.
- `create_device`: POST `/itam/devices` - kind `create`; body type `json`; risk: Create a device
  through the Freshservice API.
- `update_device`: PUT `/itam/devices/{{ record.device_id }}` - kind `update`; body type `json`;
  path fields `device_id`; required record fields `device_id`; accepted fields `device_id`; risk:
  Update a device through the Freshservice API.
- `update_custom_field_of_devices`: PUT `/itam/custom_fields/devices` - kind `upsert`; body type
  `json`; risk: Create or update custom field through the Freshservice API.
- `delete_device`: DELETE `/itam/devices/{{ record.device_id }}` - kind `delete`; body type `none`;
  path fields `device_id`; required record fields `device_id`; accepted fields `device_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Freshservice data: Delete a device.
- `create_an_asset_for_freshservice_itam`: POST `/itam/assets` - kind `create`; body type `json`;
  risk: Create an asset through the Freshservice API.
- `update_an_asset_for_freshservice_itam`: PUT `/itam/assets` - kind `update`; body type `json`;
  risk: Update an asset through the Freshservice API.
- `update_an_asset_by_id_for_freshservice_itam`: PUT `/itam/assets/{{ record.asset_id }}` - kind
  `update`; body type `json`; path fields `asset_id`; required record fields `asset_id`; accepted
  fields `asset_id`; risk: Update an asset by ID through the Freshservice API.
- `delete_an_asset_for_freshservice_itam`: DELETE `/itam/assets/{{ record.asset_id }}` - kind
  `delete`; body type `none`; path fields `asset_id`; required record fields `asset_id`; accepted
  fields `asset_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete an asset.
- `create_or_update_custom_field_of_assets_for_freshservice_itam`: PUT `/itam/custom_fields/assets`
  - kind `upsert`; body type `json`; risk: Create or update custom field through the Freshservice
  API.
- `update_cloud_resource`: POST `/itam/resources` - kind `update`; body type `json`; risk: Update a
  cloud resource through the Freshservice API.
- `delete_cloud_resource`: DELETE `/itam/resources/{{ record.resource_id }}` - kind `delete`; body
  type `none`; path fields `resource_id`; required record fields `resource_id`; accepted fields
  `resource_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes Freshservice data: Delete a cloud resource.
- `update_custom_field_of_resources`: PUT `/itam/custom_fields/resources` - kind `upsert`; body type
  `json`; risk: Create or update custom field through the Freshservice API.
- `delete_cloud_resource_relationship`: DELETE `/itam/resource_relationships/{{
  record.resource_relationship_id }}` - kind `delete`; body type `none`; path fields
  `resource_relationship_id`; required record fields `resource_relationship_id`; accepted fields
  `resource_relationship_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes Freshservice data: Delete a cloud resource relationship.
- `update_cloud_infrastructure`: POST `/itam/cloud_infrastructures` - kind `update`; body type
  `json`; risk: Update cloud details through the Freshservice API.
- `update_custom_field_of_cloud_infrastructure`: PUT `/itam/custom_fields/cloudinfrastructures` -
  kind `upsert`; body type `json`; risk: Create or update custom field through the Freshservice API.
- `update_lifecycle_events`: PUT `/itam/lifecycle_events` - kind `upsert`; body type `json`; risk:
  Create or update lifecycle events through the Freshservice API.
- `delete_lifecycle_event`: DELETE `/itam/lifecycle_events/{{ record.lifecycle_event_id }}` - kind
  `delete`; body type `none`; path fields `lifecycle_event_id`; required record fields
  `lifecycle_event_id`; accepted fields `lifecycle_event_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Deletes Freshservice data: Delete a lifecycle
  event.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 197 stream-backed endpoint group(s), 263 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, non_data_endpoint=12.
