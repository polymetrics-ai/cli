# Overview

Reads broad FireHydrant REST API resources and exposes direct JSON/no-body FireHydrant mutations
through write actions.

Readable streams: `incidents`, `services`, `teams`, `environments`, `functionalities`,
`append_form_data_on_selected_value_get`, `get_ai_incident_summary_vote_status`,
`get_ai_preferences`, `get_alert`, `get_audience`, `get_audience_summary`, `get_audit_event`,
`get_aws_cloudtrail_batch`, `get_aws_connection`, `get_bootstrap`, `get_call_route`,
`get_change_event`, `get_checklist_template`, `get_comment`, `get_conference_bridge_translation`,
`get_configuration_options`, `get_current_user`, `get_environment`, `get_form_configuration`,
`get_functionality`, `get_inbound_field_map`, `get_incident`, `get_incident_channel`,
`get_incident_event`, `get_incident_relationships`, `get_incident_retrospective_field`,
`get_incident_role`, `get_incident_task`, `get_incident_type`, `get_incident_user`,
`get_integration`, `get_lifecycle_measurement_definition`, `get_mean_time_report`,
`get_member_default_audience`, `get_notification_policy`, `get_nunc_connection`,
`get_on_call_schedule_rotation`, `get_on_call_shift`, `get_options_for_field`,
`get_post_mortem_question`, `get_post_mortem_report`, `get_priority`, `get_retrospective_template`,
`get_role`, `get_runbook`, `get_runbook_action_field_options`, `get_runbook_execution`,
`get_runbook_execution_step_script`, `get_saved_search`, `get_scheduled_maintenance`, `get_service`,
`get_service_dependencies`, `get_service_dependency`, `get_severity`, `get_severity_matrix`,
`get_severity_matrix_condition`, `get_signals_alert_grouping_configuration`,
`get_signals_email_target`, `get_signals_event_source`, `get_signals_grouped_metrics`,
`get_signals_hacker_mode`, `get_signals_heartbeat_endpoint_configuration`, `get_signals_ingest_url`,
`get_signals_mttx_analytics`, `get_signals_noise_analytics`, `get_signals_timeseries_analytics`,
`get_signals_webhook_target`, `get_slack_emoji_action`, `get_status_update_template`,
`get_statuspage_connection`, `get_support_hours_schedule`, `get_task_list`, `get_team`,
`get_team_escalation_policy`, `get_team_on_call_schedule`, `get_team_signal_rule`, `get_ticket`,
`get_ticketing_field_map`, `get_ticketing_form_configuration`, `get_ticketing_priority`,
`get_ticketing_project`, `get_ticketing_project_config`, `get_user`, `get_vote_status`,
`get_webhook`, `get_zendesk_customer_support_issue`, `list_alerts`, `list_audience_summaries`,
`list_audiences`, `list_audit_events`, `list_authed_providers`, `list_available_inbound_field_maps`,
`list_available_ticketing_field_maps`, `list_aws_cloudtrail_batch_events`,
`list_aws_cloudtrail_batches`, `list_aws_connections`, `list_call_routes`, `list_change_events`,
`list_change_identities`, `list_change_types`, `list_changes`, `list_checklist_templates`,
`list_comment_reactions`, `list_comments`, `list_connection_statuses`,
`list_connection_statuses_by_slug`, `list_connection_statuses_by_slug_and_id`, `list_connections`,
`list_current_user_permissions`, `list_custom_field_definitions`,
`list_custom_field_select_options`, `list_email_subscribers`, `list_entitlements`,
`list_environment_functionalities`, `list_environment_services`, `list_field_map_available_fields`,
`list_functionality_environments`, `list_functionality_services`, `list_inbound_field_maps`,
`list_incident_alerts`, `list_incident_attachments`, `list_incident_change_events`,
`list_incident_conference_bridges`, `list_incident_events`, `list_incident_impacts`,
`list_incident_links`, `list_incident_metrics`, `list_incident_milestones`,
`list_incident_retrospectives`, `list_incident_role_assignments`, `list_incident_roles`,
`list_incident_status_pages`, `list_incident_tags`, `list_incident_tasks`, `list_incident_types`,
`list_infrastructure_metrics`, `list_infrastructure_type_metrics`, `list_infrastructures`,
`list_integrations`, `list_lifecycle_measurement_definitions`, `list_lifecycle_phases`,
`list_notification_policy_settings`, `list_nunc_connections`, `list_organization_on_call_schedules`,
`list_permissions`, `list_post_mortem_questions`, `list_post_mortem_reasons`,
`list_post_mortem_reports`, `list_priorities`, `list_processing_log_entries`,
`list_retrospective_metrics`, `list_retrospective_templates`, `list_retrospectives`, `list_roles`,
`list_runbook_actions`, `list_runbook_executions`, `list_runbooks`, `list_saved_searches`,
`list_scheduled_maintenances`, `list_schedules`, `list_service_available_downstream_dependencies`,
`list_service_available_upstream_dependencies`, `list_service_environments`, `list_severities`,
`list_severity_matrix_conditions`, `list_severity_matrix_impacts`,
`list_signals_alert_grouping_configurations`, `list_signals_email_targets`,
`list_signals_event_sources`, `list_signals_heartbeat_endpoint_configurations`,
`list_signals_transposers`, `list_signals_webhook_targets`, `list_similar_incidents`,
`list_slack_emoji_actions`, `list_slack_usergroups`, `list_slack_workspaces`,
`list_status_update_templates`, `list_statuspage_connection_pages`, `list_statuspage_connections`,
`list_task_lists`, `list_team_call_routes`, `list_team_escalation_policies`,
`list_team_on_call_schedules`, `list_team_permissions`, `list_team_signal_rules`,
`list_ticket_tags`, `list_ticketing_custom_definitions`, `list_ticketing_priorities`,
`list_ticketing_projects`, `list_tickets`, `list_transcript_entries`,
`list_user_involvement_metrics`, `list_user_notification_settings_by_user_id`,
`list_user_owned_services`, `list_users`, `list_webhook_deliveries`, `list_webhooks`,
`search_confluence_spaces`, `search_slack_channels`, `search_zendesk_tickets`.

Write actions: `archive_audience`, `bulk_update_incident_milestones`, `close_incident`,
`convert_incident_task`, `copy_on_call_schedule_rotation`, `create_audience`, `create_change`,
`create_change_event`, `create_change_identity`, `create_checklist_template`, `create_comment`,
`create_comment_reaction`, `create_connection`, `create_custom_field_definition`,
`create_email_subscriber`, `create_environment`, `create_functionality`, `create_inbound_field_map`,
`create_incident`, `create_incident_alert`, `create_incident_change_event`,
`create_incident_chat_message`, `create_incident_impact`, `create_incident_link`,
`create_incident_note`, `create_incident_retrospective`,
`create_incident_retrospective_dynamic_input`, `create_incident_retrospective_field`,
`create_incident_role`, `create_incident_role_assignment`, `create_incident_status_page`,
`create_incident_task`, `create_incident_task_list`, `create_incident_team_assignment`,
`create_incident_type`, `create_lifecycle_measurement_definition`, `create_lifecycle_milestone`,
`create_notification_policy`, `create_nunc_component_group`, `create_nunc_connection`,
`create_nunc_link`, `create_nunc_subscription`, `create_on_call_schedule_rotation`,
`create_on_call_shift`, `create_post_mortem_reason`, `create_post_mortem_report`, `create_priority`,
`create_retrospective_template`, `create_role`, `create_runbook`, `create_runbook_execution`,
`create_saved_search`, `create_scheduled_maintenance`, `create_service`,
`create_service_checklist_response`, `create_service_dependency`, `create_service_links`,
`create_severity`, `create_severity_matrix_condition`, `create_severity_matrix_impact`,
`create_signals_alert_grouping_configuration`, `create_signals_email_target`,
`create_signals_event_source`, `create_signals_heartbeat_endpoint_configuration`,
`create_signals_page`, `create_signals_webhook_target`, `create_slack_emoji_action`,
`create_status_update_template`, `create_task_list`, `create_team`, `create_team_call_route`,
`create_team_escalation_policy`, `create_team_on_call_schedule`, `create_team_signal_rule`,
`create_ticket`, `create_ticketing_custom_definition`, `create_ticketing_field_map`,
`create_ticketing_priority`, `create_ticketing_project_config`, `create_webhook`, and 164 more.

Service API documentation: https://docs.firehydrant.com/reference/firehydrant-api.

## Auth setup

Connection fields:

- `action_slug` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (action_slug).
- `alert_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (alert_id).
- `api_token` (required, secret, string); FireHydrant API token. Sent as Authorization: Bearer
  <api_token>.
- `audience_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (audience_id).
- `base_url` (optional, string); default `https://api.firehydrant.io/v1`; format `uri`; FireHydrant
  API root. Defaults to https://api.firehydrant.io/v1. Also usable as a base URL override for
  tests/proxies.
- `by_connection_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (by_connection_id).
- `change_event_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (change_event_id).
- `change_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (change_id).
- `comment_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (comment_id).
- `condition_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (condition_id).
- `config_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (config_id).
- `connection_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (connection_id).
- `conversation_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (conversation_id).
- `emoji_action_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (emoji_action_id).
- `environment_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (environment_id).
- `event_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (event_id).
- `execution_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (execution_id).
- `field` (optional, string); Path parameter used by FireHydrant detail/subresource streams (field).
- `field_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (field_id).
- `field_map_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (field_map_id).
- `functionality_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (functionality_id).
- `generated_summary_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (generated_summary_id).
- `get_zendesk_customer_support_issue_ticket_id` (optional, string); Required ticket_id query
  parameter for FireHydrant get_zendesk_customer_support_issue.
- `id` (optional, string); Path parameter used by FireHydrant detail/subresource streams (id).
- `incident_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (incident_id).
- `incident_role_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (incident_role_id).
- `infra_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (infra_id).
- `infra_type` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (infra_type).
- `integration_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (integration_id).
- `integration_slug` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (integration_slug).
- `language_code` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (language_code).
- `map_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (map_id).
- `measurement_definition_id` (optional, string); Path parameter used by FireHydrant
  detail/subresource streams (measurement_definition_id).
- `member_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (member_id).
- `nunc_connection_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (nunc_connection_id).
- `page_size` (optional, integer); default `50`; Number of records requested per page (per_page).
  Between 1 and 200.
- `priority_slug` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (priority_slug).
- `question_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (question_id).
- `report_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (report_id).
- `resource_type` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (resource_type).
- `retrospective_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (retrospective_id).
- `retrospective_template_id` (optional, string); Path parameter used by FireHydrant
  detail/subresource streams (retrospective_template_id).
- `rotation_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (rotation_id).
- `runbook_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (runbook_id).
- `saved_search_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (saved_search_id).
- `schedule_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (schedule_id).
- `scheduled_maintenance_id` (optional, string); Path parameter used by FireHydrant
  detail/subresource streams (scheduled_maintenance_id).
- `search_zendesk_tickets_query` (optional, string); Required query query parameter for FireHydrant
  search_zendesk_tickets.
- `selected_value` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (selected_value).
- `service_dependency_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (service_dependency_id).
- `service_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (service_id).
- `severity_slug` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (severity_slug).
- `slug` (optional, string); Path parameter used by FireHydrant detail/subresource streams (slug).
- `status_update_template_id` (optional, string); Path parameter used by FireHydrant
  detail/subresource streams (status_update_template_id).
- `step_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (step_id).
- `task_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (task_id).
- `task_list_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (task_list_id).
- `team_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (team_id).
- `ticket_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (ticket_id).
- `ticketing_project_id` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (ticketing_project_id).
- `transposer_slug` (optional, string); Path parameter used by FireHydrant detail/subresource
  streams (transposer_slug).
- `type` (optional, string); Path parameter used by FireHydrant detail/subresource streams (type).
- `user_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (user_id).
- `webhook_id` (optional, string); Path parameter used by FireHydrant detail/subresource streams
  (webhook_id).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.firehydrant.io/v1`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/environments` with query `per_page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `pagination.next`.

Pagination by stream: cursor: `incidents`, `services`, `teams`, `environments`, `functionalities`,
`list_alerts`, `list_audiences`, `list_authed_providers`, `list_aws_cloudtrail_batches`,
`list_aws_connections`, `list_call_routes`, `list_change_events`, `list_change_identities`,
`list_change_types`, `list_checklist_templates`, `list_connections`, `list_entitlements`,
`list_environment_functionalities`, `list_environment_services`, `list_functionality_environments`,
`list_incident_alerts`, `list_incident_attachments`, `list_incident_change_events`,
`list_incident_events`, `list_incident_impacts`, `list_incident_links`, `list_incident_milestones`,
`list_incident_retrospectives`, `list_incident_role_assignments`, `list_incident_roles`,
`list_incident_status_pages`, `list_incident_tags`, `list_incident_tasks`, `list_incident_types`,
`list_integrations`, `list_notification_policy_settings`, `list_nunc_connections`,
`list_organization_on_call_schedules`, `list_post_mortem_questions`, `list_post_mortem_reasons`,
`list_post_mortem_reports`, `list_processing_log_entries`, `list_retrospective_templates`,
`list_retrospectives`, `list_roles`, `list_runbook_actions`, `list_runbook_executions`,
`list_runbooks`, `list_schedules`, `list_service_environments`, `list_severities`,
`list_signals_alert_grouping_configurations`, `list_signals_email_targets`,
`list_signals_webhook_targets`, `list_similar_incidents`, `list_statuspage_connections`,
`list_team_call_routes`, `list_team_escalation_policies`, `list_team_on_call_schedules`,
`list_team_signal_rules`, and 6 more; none: `append_form_data_on_selected_value_get`,
`get_ai_incident_summary_vote_status`, `get_ai_preferences`, `get_alert`, `get_audience`,
`get_audience_summary`, `get_audit_event`, `get_aws_cloudtrail_batch`, `get_aws_connection`,
`get_bootstrap`, `get_call_route`, `get_change_event`, `get_checklist_template`, `get_comment`,
`get_conference_bridge_translation`, `get_configuration_options`, `get_current_user`,
`get_environment`, `get_form_configuration`, `get_functionality`, `get_inbound_field_map`,
`get_incident`, `get_incident_channel`, `get_incident_event`, `get_incident_relationships`,
`get_incident_retrospective_field`, `get_incident_role`, `get_incident_task`, `get_incident_type`,
`get_incident_user`, `get_integration`, `get_lifecycle_measurement_definition`,
`get_mean_time_report`, `get_member_default_audience`, `get_notification_policy`,
`get_nunc_connection`, `get_on_call_schedule_rotation`, `get_on_call_shift`,
`get_options_for_field`, `get_post_mortem_question`, `get_post_mortem_report`, `get_priority`,
`get_retrospective_template`, `get_role`, `get_runbook`, `get_runbook_action_field_options`,
`get_runbook_execution`, `get_runbook_execution_step_script`, `get_saved_search`,
`get_scheduled_maintenance`, `get_service`, `get_service_dependencies`, `get_service_dependency`,
`get_severity`, `get_severity_matrix`, `get_severity_matrix_condition`,
`get_signals_alert_grouping_configuration`, `get_signals_email_target`, `get_signals_event_source`,
`get_signals_grouped_metrics`, and 79 more.

- `incidents`: GET `/incidents` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`.
- `services`: GET `/services` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`.
- `teams`: GET `/teams` - records path `data`; query `per_page` from template `{{ config.page_size
  }}`, default `50`; cursor pagination; cursor parameter `page`; next token from `pagination.next`.
- `environments`: GET `/environments` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`.
- `functionalities`: GET `/functionalities` - records path `data`; query `per_page` from template
  `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token
  from `pagination.next`.
- `append_form_data_on_selected_value_get`: GET `/form_configurations/{{ config.slug
  }}/append_data_on_select/{{ config.field_id }}/{{ config.selected_value }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_ai_incident_summary_vote_status`: GET `/ai/summarize_incident/{{ config.incident_id }}/{{
  config.generated_summary_id }}/voted` - single-object response; records path `.`; emits
  passthrough records.
- `get_ai_preferences`: GET `/ai/preferences` - single-object response; records path `.`; emits
  passthrough records.
- `get_alert`: GET `/alerts/{{ config.alert_id }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_audience`: GET `/audiences/{{ config.audience_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_audience_summary`: GET `/audiences/{{ config.audience_id }}/summaries/{{ config.incident_id
  }}` - single-object response; records path `.`; emits passthrough records.
- `get_audit_event`: GET `/audit_events/{{ config.id }}` - single-object response; records path `.`;
  emits passthrough records.
- `get_aws_cloudtrail_batch`: GET `/integrations/aws/cloudtrail_batches/{{ config.id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_aws_connection`: GET `/integrations/aws/connections/{{ config.id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_bootstrap`: GET `/bootstrap` - single-object response; records path `.`; emits passthrough
  records.
- `get_call_route`: GET `/signals/call_routes/{{ config.id }}` - single-object response; records
  path `.`; emits passthrough records.
- `get_change_event`: GET `/changes/events/{{ config.change_event_id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_checklist_template`: GET `/checklist_templates/{{ config.id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_comment`: GET `/conversations/{{ config.conversation_id }}/comments/{{ config.comment_id }}`
  - single-object response; records path `.`; emits passthrough records.
- `get_conference_bridge_translation`: GET `/incidents/{{ config.incident_id
  }}/conference_bridges/{{ config.id }}/translations/{{ config.language_code }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_configuration_options`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/configuration_options` - single-object response; records path `.`; emits passthrough records.
- `get_current_user`: GET `/current_user` - single-object response; records path `.`; emits
  passthrough records.
- `get_environment`: GET `/environments/{{ config.environment_id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_form_configuration`: GET `/form_configurations/{{ config.slug }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_functionality`: GET `/functionalities/{{ config.functionality_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_inbound_field_map`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/inbound_field_maps/{{ config.map_id }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_incident`: GET `/incidents/{{ config.incident_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_incident_channel`: GET `/incidents/{{ config.incident_id }}/channel` - single-object
  response; records path `.`; emits passthrough records.
- `get_incident_event`: GET `/incidents/{{ config.incident_id }}/events/{{ config.event_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_incident_relationships`: GET `/incidents/{{ config.incident_id }}/relationships` -
  single-object response; records path `.`; emits passthrough records.
- `get_incident_retrospective_field`: GET `/incidents/{{ config.incident_id }}/retrospectives/{{
  config.retrospective_id }}/fields/{{ config.field_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_incident_role`: GET `/incident_roles/{{ config.incident_role_id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_incident_task`: GET `/incidents/{{ config.incident_id }}/tasks/{{ config.task_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_incident_type`: GET `/incident_types/{{ config.id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_incident_user`: GET `/incidents/{{ config.incident_id }}/users/{{ config.user_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_integration`: GET `/integrations/{{ config.integration_id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_lifecycle_measurement_definition`: GET `/lifecycles/measurement_definitions/{{
  config.measurement_definition_id }}` - single-object response; records path `.`; emits passthrough
  records.
- `get_mean_time_report`: GET `/reports/mean_time` - records path `data`; emits passthrough records.
- `get_member_default_audience`: GET `/audiences/member/{{ config.member_id }}/default` -
  single-object response; records path `.`; emits passthrough records.
- `get_notification_policy`: GET `/signals/notification_policy_items/{{ config.id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_nunc_connection`: GET `/nunc_connections/{{ config.nunc_connection_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_on_call_schedule_rotation`: GET `/teams/{{ config.team_id }}/on_call_schedules/{{
  config.schedule_id }}/rotations/{{ config.rotation_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_on_call_shift`: GET `/teams/{{ config.team_id }}/on_call_schedules/{{ config.schedule_id
  }}/shifts/{{ config.id }}` - single-object response; records path `.`; emits passthrough records.
- `get_options_for_field`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/configuration_options/options_for/{{ config.field_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_post_mortem_question`: GET `/post_mortems/questions/{{ config.question_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_post_mortem_report`: GET `/post_mortems/reports/{{ config.report_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_priority`: GET `/priorities/{{ config.priority_slug }}` - single-object response; records
  path `.`; emits passthrough records.
- `get_retrospective_template`: GET `/retrospective_templates/{{ config.retrospective_template_id
  }}` - single-object response; records path `.`; emits passthrough records.
- `get_role`: GET `/roles/{{ config.id }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_runbook`: GET `/runbooks/{{ config.runbook_id }}` - single-object response; records path `.`;
  emits passthrough records.
- `get_runbook_action_field_options`: GET `/runbooks/select_options/{{ config.integration_slug }}/{{
  config.action_slug }}/{{ config.field }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_runbook_execution`: GET `/runbooks/executions/{{ config.execution_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_runbook_execution_step_script`: GET `/runbooks/executions/{{ config.execution_id }}/steps/{{
  config.step_id }}/script` - single-object response; records path `.`; emits passthrough records.
- `get_saved_search`: GET `/saved_searches/{{ config.resource_type }}/{{ config.saved_search_id }}`
  - single-object response; records path `.`; emits passthrough records.
- `get_scheduled_maintenance`: GET `/scheduled_maintenances/{{ config.scheduled_maintenance_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_service`: GET `/services/{{ config.service_id }}` - single-object response; records path `.`;
  emits passthrough records.
- `get_service_dependencies`: GET `/services/{{ config.service_id }}/dependencies` - single-object
  response; records path `.`; emits passthrough records.
- `get_service_dependency`: GET `/service_dependencies/{{ config.service_dependency_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_severity`: GET `/severities/{{ config.severity_slug }}` - single-object response; records
  path `.`; emits passthrough records.
- `get_severity_matrix`: GET `/severity_matrix` - single-object response; records path `.`; emits
  passthrough records.
- `get_severity_matrix_condition`: GET `/severity_matrix/conditions/{{ config.condition_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_signals_alert_grouping_configuration`: GET `/signals/grouping/{{ config.id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_signals_email_target`: GET `/signals/email_targets/{{ config.id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_signals_event_source`: GET `/signals/event_sources/{{ config.transposer_slug }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_signals_grouped_metrics`: GET `/signals/analytics/grouped_metrics` - records path `data`;
  emits passthrough records.
- `get_signals_hacker_mode`: GET `/signals/hacker_mode` - single-object response; records path `.`;
  emits passthrough records.
- `get_signals_heartbeat_endpoint_configuration`: GET `/signals/heartbeat_endpoints/{{ config.id }}`
  - single-object response; records path `.`; emits passthrough records.
- `get_signals_ingest_url`: GET `/signals/ingest_url` - single-object response; records path `.`;
  emits passthrough records.
- `get_signals_mttx_analytics`: GET `/signals/analytics/mttx` - records path `data`; emits
  passthrough records.
- `get_signals_noise_analytics`: GET `/signals/analytics/noise/metrics` - records path `data`; emits
  passthrough records.
- `get_signals_timeseries_analytics`: GET `/signals/analytics/timeseries` - records path `data`;
  emits passthrough records.
- `get_signals_webhook_target`: GET `/signals/webhook_targets/{{ config.id }}` - single-object
  response; records path `.`; emits passthrough records.
- `get_slack_emoji_action`: GET `/integrations/slack/connections/{{ config.connection_id
  }}/emoji_actions/{{ config.emoji_action_id }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_status_update_template`: GET `/status_update_templates/{{ config.status_update_template_id
  }}` - single-object response; records path `.`; emits passthrough records.
- `get_statuspage_connection`: GET `/integrations/statuspage/connections/{{ config.connection_id }}`
  - single-object response; records path `.`; emits passthrough records.
- `get_support_hours_schedule`: GET `/teams/{{ config.team_id }}/support_hours_schedule` -
  single-object response; records path `.`; emits passthrough records.
- `get_task_list`: GET `/task_lists/{{ config.task_list_id }}` - single-object response; records
  path `.`; emits passthrough records.
- `get_team`: GET `/teams/{{ config.team_id }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_team_escalation_policy`: GET `/teams/{{ config.team_id }}/escalation_policies/{{ config.id
  }}` - single-object response; records path `.`; emits passthrough records.
- `get_team_on_call_schedule`: GET `/teams/{{ config.team_id }}/on_call_schedules/{{
  config.schedule_id }}` - single-object response; records path `.`; emits passthrough records.
- `get_team_signal_rule`: GET `/teams/{{ config.team_id }}/signal_rules/{{ config.id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_ticket`: GET `/ticketing/tickets/{{ config.ticket_id }}` - single-object response; records
  path `.`; emits passthrough records.
- `get_ticketing_field_map`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/field_maps/{{ config.map_id }}` - single-object response; records path `.`; emits passthrough
  records.
- `get_ticketing_form_configuration`: GET `/ticketing/form_configurations` - single-object response;
  records path `.`; emits passthrough records.
- `get_ticketing_priority`: GET `/ticketing/priorities/{{ config.id }}` - single-object response;
  records path `.`; emits passthrough records.
- `get_ticketing_project`: GET `/ticketing/projects/{{ config.ticketing_project_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `get_ticketing_project_config`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/provider_project_configurations/{{ config.config_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `get_user`: GET `/users/{{ config.id }}` - single-object response; records path `.`; emits
  passthrough records.
- `get_vote_status`: GET `/incidents/{{ config.incident_id }}/events/{{ config.event_id
  }}/votes/status` - single-object response; records path `.`; emits passthrough records.
- `get_webhook`: GET `/webhooks/{{ config.webhook_id }}` - single-object response; records path `.`;
  emits passthrough records.
- `get_zendesk_customer_support_issue`: GET `/integrations/zendesk/search` - single-object response;
  records path `.`; query `ticket_id`=`{{ config.get_zendesk_customer_support_issue_ticket_id }}`;
  emits passthrough records.
- `list_alerts`: GET `/alerts` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_audience_summaries`: GET `/audiences/summaries/{{ config.incident_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `list_audiences`: GET `/audiences` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_audit_events`: GET `/audit_events` - single-object response; records path `.`; emits
  passthrough records.
- `list_authed_providers`: GET `/integrations/authed_providers/{{ config.integration_slug }}/{{
  config.connection_id }}` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_available_inbound_field_maps`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/inbound_field_maps/available_fields` - single-object response; records path `.`; emits
  passthrough records.
- `list_available_ticketing_field_maps`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/field_maps/available_fields` - single-object response; records path `.`; emits passthrough
  records.
- `list_aws_cloudtrail_batch_events`: GET `/integrations/aws/cloudtrail_batches/{{ config.id
  }}/events` - single-object response; records path `.`; emits passthrough records.
- `list_aws_cloudtrail_batches`: GET `/integrations/aws/cloudtrail_batches` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_aws_connections`: GET `/integrations/aws/connections` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_call_routes`: GET `/signals/call_routes` - records path `data`; query `per_page` from
  template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next
  token from `pagination.next`; emits passthrough records.
- `list_change_events`: GET `/changes/events` - records path `data`; query `per_page` from template
  `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token
  from `pagination.next`; emits passthrough records.
- `list_change_identities`: GET `/changes/{{ config.change_id }}/identities` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_change_types`: GET `/change_types` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_changes`: GET `/changes` - single-object response; records path `.`; emits passthrough
  records.
- `list_checklist_templates`: GET `/checklist_templates` - records path `data`; query `per_page`
  from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`;
  next token from `pagination.next`; emits passthrough records.
- `list_comment_reactions`: GET `/conversations/{{ config.conversation_id }}/comments/{{
  config.comment_id }}/reactions` - single-object response; records path `.`; emits passthrough
  records.
- `list_comments`: GET `/conversations/{{ config.conversation_id }}/comments` - single-object
  response; records path `.`; emits passthrough records.
- `list_connection_statuses`: GET `/integrations/statuses` - single-object response; records path
  `.`; emits passthrough records.
- `list_connection_statuses_by_slug`: GET `/integrations/statuses/{{ config.slug }}` - single-object
  response; records path `.`; emits passthrough records.
- `list_connection_statuses_by_slug_and_id`: GET `/integrations/statuses/{{ config.slug }}/{{
  config.by_connection_id }}` - single-object response; records path `.`; emits passthrough records.
- `list_connections`: GET `/integrations/connections` - records path `data`; query `per_page` from
  template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next
  token from `pagination.next`; emits passthrough records.
- `list_current_user_permissions`: GET `/permissions/current_user` - records path `data`; emits
  passthrough records.
- `list_custom_field_definitions`: GET `/custom_fields/definitions` - single-object response;
  records path `.`; emits passthrough records.
- `list_custom_field_select_options`: GET `/custom_fields/definitions/{{ config.field_id
  }}/select_options` - single-object response; records path `.`; emits passthrough records.
- `list_email_subscribers`: GET `/nunc_connections/{{ config.nunc_connection_id }}/subscribers` -
  single-object response; records path `.`; emits passthrough records.
- `list_entitlements`: GET `/entitlements` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_environment_functionalities`: GET `/environments/{{ config.environment_id
  }}/functionalities` - records path `data`; query `per_page` from template `{{ config.page_size
  }}`, default `50`; cursor pagination; cursor parameter `page`; next token from `pagination.next`;
  emits passthrough records.
- `list_environment_services`: GET `/environments/{{ config.environment_id }}/services` - records
  path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_field_map_available_fields`: GET `/integrations/field_maps/{{ config.field_map_id
  }}/available_fields` - single-object response; records path `.`; emits passthrough records.
- `list_functionality_environments`: GET `/functionalities/{{ config.functionality_id
  }}/environments` - records path `data`; query `per_page` from template `{{ config.page_size }}`,
  default `50`; cursor pagination; cursor parameter `page`; next token from `pagination.next`; emits
  passthrough records.
- `list_functionality_services`: GET `/functionalities/{{ config.functionality_id }}/services` -
  single-object response; records path `.`; emits passthrough records.
- `list_inbound_field_maps`: GET `/ticketing/projects/{{ config.ticketing_project_id
  }}/inbound_field_maps` - single-object response; records path `.`; emits passthrough records.
- `list_incident_alerts`: GET `/incidents/{{ config.incident_id }}/alerts` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_attachments`: GET `/incidents/{{ config.incident_id }}/attachments` - records path
  `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination;
  cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_change_events`: GET `/incidents/{{ config.incident_id }}/related_change_events` -
  records path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_conference_bridges`: GET `/incidents/{{ config.incident_id }}/conference_bridges` -
  single-object response; records path `.`; emits passthrough records.
- `list_incident_events`: GET `/incidents/{{ config.incident_id }}/events` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_impacts`: GET `/incidents/{{ config.incident_id }}/impact/{{ config.type }}` -
  records path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_links`: GET `/incidents/{{ config.incident_id }}/links` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_metrics`: GET `/metrics/incidents` - single-object response; records path `.`;
  emits passthrough records.
- `list_incident_milestones`: GET `/incidents/{{ config.incident_id }}/milestones` - records path
  `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination;
  cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_retrospectives`: GET `/incidents/{{ config.incident_id }}/retrospectives` - records
  path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_role_assignments`: GET `/incidents/{{ config.incident_id }}/role_assignments` -
  records path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_roles`: GET `/incident_roles` - records path `data`; query `per_page` from template
  `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token
  from `pagination.next`; emits passthrough records.
- `list_incident_status_pages`: GET `/incidents/{{ config.incident_id }}/status_pages` - records
  path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_tags`: GET `/incident_tags` - records path `data`; query `per_page` from template
  `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token
  from `pagination.next`; emits passthrough records.
- `list_incident_tasks`: GET `/incidents/{{ config.incident_id }}/tasks` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_incident_types`: GET `/incident_types` - records path `data`; query `per_page` from template
  `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token
  from `pagination.next`; emits passthrough records.
- `list_infrastructure_metrics`: GET `/metrics/{{ config.infra_type }}/{{ config.infra_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `list_infrastructure_type_metrics`: GET `/metrics/{{ config.infra_type }}` - records path `data`;
  emits passthrough records.
- `list_infrastructures`: GET `/infrastructures` - single-object response; records path `.`; emits
  passthrough records.
- `list_integrations`: GET `/integrations` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_lifecycle_measurement_definitions`: GET `/lifecycles/measurement_definitions` -
  single-object response; records path `.`; emits passthrough records.
- `list_lifecycle_phases`: GET `/lifecycles/phases` - records path `data`; emits passthrough
  records.
- `list_notification_policy_settings`: GET `/signals/notification_policy_items` - records path
  `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination;
  cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_nunc_connections`: GET `/nunc_connections` - records path `data`; query `per_page` from
  template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next
  token from `pagination.next`; emits passthrough records.
- `list_organization_on_call_schedules`: GET `/signals_on_call` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_permissions`: GET `/permissions` - records path `data`; emits passthrough records.
- `list_post_mortem_questions`: GET `/post_mortems/questions` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_post_mortem_reasons`: GET `/post_mortems/reports/{{ config.report_id }}/reasons` - records
  path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_post_mortem_reports`: GET `/post_mortems/reports` - records path `data`; query `per_page`
  from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`;
  next token from `pagination.next`; emits passthrough records.
- `list_priorities`: GET `/priorities` - single-object response; records path `.`; emits passthrough
  records.
- `list_processing_log_entries`: GET `/processing_log_entries` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_retrospective_metrics`: GET `/metrics/retrospectives` - records path `data`; emits
  passthrough records.
- `list_retrospective_templates`: GET `/retrospective_templates` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_retrospectives`: GET `/retrospectives` - records path `data`; query `per_page` from template
  `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token
  from `pagination.next`; emits passthrough records.
- `list_roles`: GET `/roles` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_runbook_actions`: GET `/runbooks/actions` - records path `data`; query `per_page` from
  template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next
  token from `pagination.next`; emits passthrough records.
- `list_runbook_executions`: GET `/runbooks/executions` - records path `data`; query `per_page` from
  template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next
  token from `pagination.next`; emits passthrough records.
- `list_runbooks`: GET `/runbooks` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_saved_searches`: GET `/saved_searches/{{ config.resource_type }}` - single-object response;
  records path `.`; emits passthrough records.
- `list_scheduled_maintenances`: GET `/scheduled_maintenances` - single-object response; records
  path `.`; emits passthrough records.
- `list_schedules`: GET `/schedules` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_service_available_downstream_dependencies`: GET `/services/{{ config.service_id
  }}/available_downstream_dependencies` - single-object response; records path `.`; emits
  passthrough records.
- `list_service_available_upstream_dependencies`: GET `/services/{{ config.service_id
  }}/available_upstream_dependencies` - single-object response; records path `.`; emits passthrough
  records.
- `list_service_environments`: GET `/services/{{ config.service_id }}/environments` - records path
  `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination;
  cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_severities`: GET `/severities` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_severity_matrix_conditions`: GET `/severity_matrix/conditions` - single-object response;
  records path `.`; emits passthrough records.
- `list_severity_matrix_impacts`: GET `/severity_matrix/impacts` - single-object response; records
  path `.`; emits passthrough records.
- `list_signals_alert_grouping_configurations`: GET `/signals/grouping` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_signals_email_targets`: GET `/signals/email_targets` - records path `data`; query `per_page`
  from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`;
  next token from `pagination.next`; emits passthrough records.
- `list_signals_event_sources`: GET `/signals/event_sources` - records path `data`; emits
  passthrough records.
- `list_signals_heartbeat_endpoint_configurations`: GET `/signals/heartbeat_endpoints` -
  single-object response; records path `.`; emits passthrough records.
- `list_signals_transposers`: GET `/signals/transposers` - records path `data`; emits passthrough
  records.
- `list_signals_webhook_targets`: GET `/signals/webhook_targets` - records path `data`; query
  `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_similar_incidents`: GET `/incidents/{{ config.incident_id }}/similar` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_slack_emoji_actions`: GET `/integrations/slack/connections/{{ config.connection_id
  }}/emoji_actions` - single-object response; records path `.`; emits passthrough records.
- `list_slack_usergroups`: GET `/integrations/slack/usergroups` - single-object response; records
  path `.`; emits passthrough records.
- `list_slack_workspaces`: GET `/integrations/slack/connections/{{ config.connection_id
  }}/workspaces` - single-object response; records path `.`; emits passthrough records.
- `list_status_update_templates`: GET `/status_update_templates` - single-object response; records
  path `.`; emits passthrough records.
- `list_statuspage_connection_pages`: GET `/integrations/statuspage/connections/{{
  config.connection_id }}/pages` - single-object response; records path `.`; emits passthrough
  records.
- `list_statuspage_connections`: GET `/integrations/statuspage/connections` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_task_lists`: GET `/task_lists` - single-object response; records path `.`; emits passthrough
  records.
- `list_team_call_routes`: GET `/teams/{{ config.team_id }}/call_routes` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_team_escalation_policies`: GET `/teams/{{ config.team_id }}/escalation_policies` - records
  path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_team_on_call_schedules`: GET `/teams/{{ config.team_id }}/on_call_schedules` - records path
  `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination;
  cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_team_permissions`: GET `/permissions/team` - records path `data`; emits passthrough records.
- `list_team_signal_rules`: GET `/teams/{{ config.team_id }}/signal_rules` - records path `data`;
  query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination; cursor
  parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_ticket_tags`: GET `/ticketing/ticket_tags` - records path `data`; query `per_page` from
  template `{{ config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next
  token from `pagination.next`; emits passthrough records.
- `list_ticketing_custom_definitions`: GET `/ticketing/custom_fields/definitions` - single-object
  response; records path `.`; emits passthrough records.
- `list_ticketing_priorities`: GET `/ticketing/priorities` - single-object response; records path
  `.`; emits passthrough records.
- `list_ticketing_projects`: GET `/ticketing/projects` - single-object response; records path `.`;
  emits passthrough records.
- `list_tickets`: GET `/ticketing/tickets` - single-object response; records path `.`; emits
  passthrough records.
- `list_transcript_entries`: GET `/incidents/{{ config.incident_id }}/transcript` - single-object
  response; records path `.`; emits passthrough records.
- `list_user_involvement_metrics`: GET `/metrics/user_involvements` - single-object response;
  records path `.`; emits passthrough records.
- `list_user_notification_settings_by_user_id`: GET `/signals/users/{{ config.user_id
  }}/notification_settings` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_user_owned_services`: GET `/users/{{ config.id }}/services` - records path `.`; emits
  passthrough records.
- `list_users`: GET `/users` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `list_webhook_deliveries`: GET `/webhooks/{{ config.webhook_id }}/deliveries` - records path
  `data`; query `per_page` from template `{{ config.page_size }}`, default `50`; cursor pagination;
  cursor parameter `page`; next token from `pagination.next`; emits passthrough records.
- `list_webhooks`: GET `/webhooks` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `50`; cursor pagination; cursor parameter `page`; next token from
  `pagination.next`; emits passthrough records.
- `search_confluence_spaces`: GET `/integrations/confluence_cloud/connections/{{ config.id
  }}/space/search` - single-object response; records path `.`; emits passthrough records.
- `search_slack_channels`: GET `/integrations/slack/channels` - single-object response; records path
  `.`; emits passthrough records.
- `search_zendesk_tickets`: GET `/integrations/zendesk/{{ config.connection_id }}/tickets/search` -
  records path `data`; query `per_page` from template `{{ config.page_size }}`, default `50`;
  `query`=`{{ config.search_zendesk_tickets_query }}`; cursor pagination; cursor parameter `page`;
  next token from `pagination.next`; emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, archives, deletes, triggers, and otherwise mutates FireHydrant
resources through documented JSON/no-body REST endpoints.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `archive_audience`: DELETE `/audiences/{{ record.audience_id }}` - kind `delete`; body type
  `none`; path fields `audience_id`; required record fields `audience_id`; accepted fields
  `audience_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Archive audience; this may remove or archive FireHydrant data.
- `bulk_update_incident_milestones`: PUT `/incidents/{{ record.incident_id
  }}/milestones/bulk_update` - kind `update`; body type `json`; path fields `incident_id`; required
  record fields `incident_id`; accepted fields `incident_id`; risk: Update milestone times through
  the FireHydrant API.
- `close_incident`: PUT `/incidents/{{ record.incident_id }}/close` - kind `update`; body type
  `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; confirmation `destructive`; risk: Close an incident through the FireHydrant API.
- `convert_incident_task`: POST `/incidents/{{ record.incident_id }}/tasks/{{ record.task_id
  }}/convert` - kind `create`; body type `json`; path fields `task_id`, `incident_id`; required
  record fields `task_id`, `incident_id`; accepted fields `incident_id`, `task_id`; risk: Convert a
  task to a follow-up through the FireHydrant API.
- `copy_on_call_schedule_rotation`: POST `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}/rotations/{{ record.rotation_id }}/copy` - kind `custom`; body type `json`;
  path fields `rotation_id`, `team_id`, `schedule_id`; required record fields `rotation_id`,
  `team_id`, `schedule_id`; accepted fields `rotation_id`, `schedule_id`, `team_id`; risk: Copy an
  on-call schedule's rotation through the FireHydrant API.
- `create_audience`: POST `/audiences` - kind `create`; body type `json`; risk: Create audience
  through the FireHydrant API.
- `create_change`: POST `/changes` - kind `create`; body type `json`; risk: Create a new change
  entry through the FireHydrant API.
- `create_change_event`: POST `/changes/events` - kind `create`; body type `json`; risk: Create a
  change event through the FireHydrant API.
- `create_change_identity`: POST `/changes/{{ record.change_id }}/identities` - kind `create`; body
  type `json`; path fields `change_id`; required record fields `change_id`; accepted fields
  `change_id`; risk: Create an identity for a change entry through the FireHydrant API.
- `create_checklist_template`: POST `/checklist_templates` - kind `create`; body type `json`; risk:
  Create a checklist template through the FireHydrant API.
- `create_comment`: POST `/conversations/{{ record.conversation_id }}/comments` - kind `create`;
  body type `json`; path fields `conversation_id`; required record fields `conversation_id`;
  accepted fields `conversation_id`; risk: Create a conversation comment through the FireHydrant
  API.
- `create_comment_reaction`: POST `/conversations/{{ record.conversation_id }}/comments/{{
  record.comment_id }}/reactions` - kind `create`; body type `json`; path fields `conversation_id`,
  `comment_id`; required record fields `conversation_id`, `comment_id`; accepted fields
  `comment_id`, `conversation_id`; risk: Create a reaction for a conversation comment through the
  FireHydrant API.
- `create_connection`: POST `/integrations/connections/{{ record.slug }}` - kind `create`; body type
  `json`; path fields `slug`; required record fields `slug`; accepted fields `slug`; risk: Create a
  new integration connection through the FireHydrant API.
- `create_custom_field_definition`: POST `/custom_fields/definitions` - kind `create`; body type
  `json`; risk: Create a custom field definition through the FireHydrant API.
- `create_email_subscriber`: POST `/nunc_connections/{{ record.nunc_connection_id }}/subscribers` -
  kind `create`; body type `json`; path fields `nunc_connection_id`; required record fields
  `nunc_connection_id`; accepted fields `nunc_connection_id`; risk: Add subscribers to a status page
  through the FireHydrant API.
- `create_environment`: POST `/environments` - kind `create`; body type `json`; risk: Create an
  environment through the FireHydrant API.
- `create_functionality`: POST `/functionalities` - kind `create`; body type `json`; risk: Create a
  functionality through the FireHydrant API.
- `create_inbound_field_map`: POST `/ticketing/projects/{{ record.ticketing_project_id
  }}/inbound_field_maps` - kind `create`; body type `json`; path fields `ticketing_project_id`;
  required record fields `ticketing_project_id`; accepted fields `ticketing_project_id`; risk:
  Create inbound field map for a ticketing project through the FireHydrant API.
- `create_incident`: POST `/incidents` - kind `create`; body type `json`; risk: Create an incident
  through the FireHydrant API.
- `create_incident_alert`: POST `/incidents/{{ record.incident_id }}/alerts` - kind `create`; body
  type `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; risk: Attach an alert to an incident through the FireHydrant API.
- `create_incident_change_event`: POST `/incidents/{{ record.incident_id }}/related_change_events` -
  kind `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Add a related change to an incident through the FireHydrant
  API.
- `create_incident_chat_message`: POST `/incidents/{{ record.incident_id }}/generic_chat_messages` -
  kind `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Add a chat message to an incident through the FireHydrant
  API.
- `create_incident_impact`: POST `/incidents/{{ record.incident_id }}/impact/{{ record.type }}` -
  kind `create`; body type `json`; path fields `incident_id`, `type`; required record fields
  `incident_id`, `type`; accepted fields `incident_id`, `type`; risk: Add impacted infrastructure to
  an incident through the FireHydrant API.
- `create_incident_link`: POST `/incidents/{{ record.incident_id }}/links` - kind `create`; body
  type `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; risk: Add a link to an incident through the FireHydrant API.
- `create_incident_note`: POST `/incidents/{{ record.incident_id }}/notes` - kind `create`; body
  type `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; risk: Add a note to an incident through the FireHydrant API.
- `create_incident_retrospective`: POST `/incidents/{{ record.incident_id }}/retrospectives` - kind
  `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Create a new retrospective on the incident using the template
  through the FireHydrant API.
- `create_incident_retrospective_dynamic_input`: POST `/incidents/{{ record.incident_id
  }}/retrospectives/{{ record.retrospective_id }}/fields/{{ record.field_id }}/inputs` - kind
  `create`; body type `json`; path fields `retrospective_id`, `field_id`, `incident_id`; required
  record fields `retrospective_id`, `field_id`, `incident_id`; accepted fields `field_id`,
  `incident_id`, `retrospective_id`; risk: Add a new dynamic input field to a retrospective's
  dynamic input group field through the FireHydrant API.
- `create_incident_retrospective_field`: PATCH `/incidents/{{ record.incident_id
  }}/retrospectives/{{ record.retrospective_id }}/fields` - kind `create`; body type `json`; path
  fields `retrospective_id`, `incident_id`; required record fields `retrospective_id`,
  `incident_id`; accepted fields `incident_id`, `retrospective_id`; risk: Appends a new incident
  retrospective field to an incident retrospective through the FireHydrant API.
- `create_incident_role`: POST `/incident_roles` - kind `create`; body type `json`; risk: Create an
  incident role through the FireHydrant API.
- `create_incident_role_assignment`: POST `/incidents/{{ record.incident_id }}/role_assignments` -
  kind `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Assign a user to an incident through the FireHydrant API.
- `create_incident_status_page`: POST `/incidents/{{ record.incident_id }}/status_pages` - kind
  `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; confirmation `destructive`; risk: Add a status page to an incident
  through the FireHydrant API.
- `create_incident_task`: POST `/incidents/{{ record.incident_id }}/tasks` - kind `create`; body
  type `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; risk: Create an incident task through the FireHydrant API.
- `create_incident_task_list`: POST `/incidents/{{ record.incident_id }}/task_lists` - kind
  `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Add tasks from a task list to an incident through the
  FireHydrant API.
- `create_incident_team_assignment`: POST `/incidents/{{ record.incident_id }}/team_assignments` -
  kind `create`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Assign a team to an incident through the FireHydrant API.
- `create_incident_type`: POST `/incident_types` - kind `create`; body type `json`; risk: Create an
  incident type through the FireHydrant API.
- `create_lifecycle_measurement_definition`: POST `/lifecycles/measurement_definitions` - kind
  `create`; body type `json`; risk: Create a measurement definition through the FireHydrant API.
- `create_lifecycle_milestone`: POST `/lifecycles/milestones` - kind `create`; body type `json`;
  risk: Create a milestone through the FireHydrant API.
- `create_notification_policy`: POST `/signals/notification_policy_items` - kind `create`; body type
  `json`; risk: Create a notification policy through the FireHydrant API.
- `create_nunc_component_group`: POST `/nunc_connections/{{ record.nunc_connection_id
  }}/component_groups` - kind `create`; body type `json`; path fields `nunc_connection_id`; required
  record fields `nunc_connection_id`; accepted fields `nunc_connection_id`; risk: Create a component
  group for a status page through the FireHydrant API.
- `create_nunc_connection`: POST `/nunc_connections` - kind `create`; body type `json`; risk: Create
  a status page through the FireHydrant API.
- `create_nunc_link`: POST `/nunc_connections/{{ record.nunc_connection_id }}/links` - kind
  `create`; body type `json`; path fields `nunc_connection_id`; required record fields
  `nunc_connection_id`; accepted fields `nunc_connection_id`; risk: Add link to a status page
  through the FireHydrant API.
- `create_nunc_subscription`: POST `/nunc/subscriptions` - kind `create`; body type `json`; risk:
  Create a status page subscription through the FireHydrant API.
- `create_on_call_schedule_rotation`: POST `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}/rotations` - kind `create`; body type `json`; path fields `team_id`,
  `schedule_id`; required record fields `team_id`, `schedule_id`; accepted fields `schedule_id`,
  `team_id`; risk: Create a new on-call rotation through the FireHydrant API.
- `create_on_call_shift`: POST `/teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id
  }}/shifts` - kind `create`; body type `json`; path fields `team_id`, `schedule_id`; required
  record fields `team_id`, `schedule_id`; accepted fields `schedule_id`, `team_id`; risk: Create a
  shift for an on-call schedule through the FireHydrant API.
- `create_post_mortem_reason`: POST `/post_mortems/reports/{{ record.report_id }}/reasons` - kind
  `create`; body type `json`; path fields `report_id`; required record fields `report_id`; accepted
  fields `report_id`; risk: Create a contributing factor for a retrospective report through the
  FireHydrant API.
- `create_post_mortem_report`: POST `/post_mortems/reports` - kind `create`; body type `json`; risk:
  Create a retrospective report through the FireHydrant API.
- `create_priority`: POST `/priorities` - kind `create`; body type `json`; risk: Create a priority
  through the FireHydrant API.
- `create_retrospective_template`: POST `/retrospective_templates` - kind `create`; body type
  `json`; risk: Create a retrospective template through the FireHydrant API.
- `create_role`: POST `/roles` - kind `create`; body type `json`; risk: Create a role through the
  FireHydrant API.
- `create_runbook`: POST `/runbooks` - kind `create`; body type `json`; risk: Create a runbook
  through the FireHydrant API.
- `create_runbook_execution`: POST `/runbooks/executions` - kind `create`; body type `json`; risk:
  Create a runbook execution through the FireHydrant API.
- `create_saved_search`: POST `/saved_searches/{{ record.resource_type }}` - kind `create`; body
  type `json`; path fields `resource_type`; required record fields `resource_type`; accepted fields
  `resource_type`; risk: Create a saved search through the FireHydrant API.
- `create_scheduled_maintenance`: POST `/scheduled_maintenances` - kind `create`; body type `json`;
  risk: Create a scheduled maintenance event through the FireHydrant API.
- `create_service`: POST `/services` - kind `create`; body type `json`; risk: Create a service
  through the FireHydrant API.
- `create_service_checklist_response`: POST `/services/{{ record.service_id }}/checklist_response/{{
  record.checklist_id }}` - kind `create`; body type `json`; path fields `service_id`,
  `checklist_id`; required record fields `service_id`, `checklist_id`; accepted fields
  `checklist_id`, `service_id`; risk: Record a response for a checklist item through the FireHydrant
  API.
- `create_service_dependency`: POST `/service_dependencies` - kind `create`; body type `json`; risk:
  Create a service dependency through the FireHydrant API.
- `create_service_links`: POST `/services/service_links` - kind `create`; body type `json`; risk:
  Create multiple services linked to external services through the FireHydrant API.
- `create_severity`: POST `/severities` - kind `create`; body type `json`; risk: Create a severity
  through the FireHydrant API.
- `create_severity_matrix_condition`: POST `/severity_matrix/conditions` - kind `create`; body type
  `json`; risk: Create a severity matrix condition through the FireHydrant API.
- `create_severity_matrix_impact`: POST `/severity_matrix/impacts` - kind `create`; body type
  `json`; risk: Create a severity matrix impact through the FireHydrant API.
- `create_signals_alert_grouping_configuration`: POST `/signals/grouping` - kind `create`; body type
  `json`; risk: Create an alert grouping configuration. through the FireHydrant API.
- `create_signals_email_target`: POST `/signals/email_targets` - kind `create`; body type `json`;
  risk: Create an email target for signals through the FireHydrant API.
- `create_signals_event_source`: PUT `/signals/event_sources` - kind `create`; body type `json`;
  risk: Create an event source for Signals through the FireHydrant API.
- `create_signals_heartbeat_endpoint_configuration`: POST `/signals/heartbeat_endpoints` - kind
  `create`; body type `json`; risk: Create a heartbeat endpoint configuration through the
  FireHydrant API.
- `create_signals_page`: POST `/page/signals` - kind `create`; body type `json`; confirmation
  `destructive`; risk: Page a user, team, on-call schedule, or escalation policy through the
  FireHydrant API.
- `create_signals_webhook_target`: POST `/signals/webhook_targets` - kind `create`; body type
  `json`; risk: Create a webhook target through the FireHydrant API.
- `create_slack_emoji_action`: POST `/integrations/slack/connections/{{ record.connection_id
  }}/emoji_actions` - kind `create`; body type `json`; path fields `connection_id`; required record
  fields `connection_id`; accepted fields `connection_id`; risk: Create a new Slack emoji action
  through the FireHydrant API.
- `create_status_update_template`: POST `/status_update_templates` - kind `create`; body type
  `json`; risk: Create a status update template through the FireHydrant API.
- `create_task_list`: POST `/task_lists` - kind `create`; body type `json`; risk: Create a task list
  through the FireHydrant API.
- `create_team`: POST `/teams` - kind `create`; body type `json`; risk: Create a team through the
  FireHydrant API.
- `create_team_call_route`: POST `/teams/{{ record.team_id }}/call_routes` - kind `create`; body
  type `json`; path fields `team_id`; required record fields `team_id`; accepted fields `team_id`;
  risk: Create a call route for a team through the FireHydrant API.
- `create_team_escalation_policy`: POST `/teams/{{ record.team_id }}/escalation_policies` - kind
  `create`; body type `json`; path fields `team_id`; required record fields `team_id`; accepted
  fields `team_id`; risk: Create an escalation policy for a team through the FireHydrant API.
- `create_team_on_call_schedule`: POST `/teams/{{ record.team_id }}/on_call_schedules` - kind
  `create`; body type `json`; path fields `team_id`; required record fields `team_id`; accepted
  fields `team_id`; risk: Create an on-call schedule for a team through the FireHydrant API.
- `create_team_signal_rule`: POST `/teams/{{ record.team_id }}/signal_rules` - kind `create`; body
  type `json`; path fields `team_id`; required record fields `team_id`; accepted fields `team_id`;
  risk: Create a Signals rule through the FireHydrant API.
- `create_ticket`: POST `/ticketing/tickets` - kind `create`; body type `json`; risk: Create a
  ticket through the FireHydrant API.
- `create_ticketing_custom_definition`: POST `/ticketing/custom_fields/definitions` - kind `create`;
  body type `json`; risk: Create a ticketing custom field through the FireHydrant API.
- `create_ticketing_field_map`: POST `/ticketing/projects/{{ record.ticketing_project_id
  }}/field_maps` - kind `create`; body type `json`; path fields `ticketing_project_id`; required
  record fields `ticketing_project_id`; accepted fields `ticketing_project_id`; risk: Create a field
  mapping for a ticketing project through the FireHydrant API.
- `create_ticketing_priority`: POST `/ticketing/priorities` - kind `create`; body type `json`; risk:
  Create a ticketing priority through the FireHydrant API.
- `create_ticketing_project_config`: POST `/ticketing/projects/{{ record.ticketing_project_id
  }}/provider_project_configurations` - kind `create`; body type `json`; path fields
  `ticketing_project_id`; required record fields `ticketing_project_id`; accepted fields
  `ticketing_project_id`; risk: Create a ticketing project configuration through the FireHydrant
  API.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; risk: Create a webhook
  through the FireHydrant API.
- `debug_signals_expression`: POST `/signals/debugger` - kind `custom`; body type `json`; risk:
  Debug Signals expressions through the FireHydrant API.
- `delete_call_route`: DELETE `/signals/call_routes/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a call route; this
  may remove or archive FireHydrant data.
- `delete_change`: DELETE `/changes/{{ record.change_id }}` - kind `delete`; body type `none`; path
  fields `change_id`; required record fields `change_id`; accepted fields `change_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Archive a change
  entry; this may remove or archive FireHydrant data.
- `delete_change_event`: DELETE `/changes/events/{{ record.change_event_id }}` - kind `delete`; body
  type `none`; path fields `change_event_id`; required record fields `change_event_id`; accepted
  fields `change_event_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a change event; this may remove or archive FireHydrant data.
- `delete_change_identity`: DELETE `/changes/{{ record.change_id }}/identities/{{ record.identity_id
  }}` - kind `delete`; body type `none`; path fields `identity_id`, `change_id`; required record
  fields `identity_id`, `change_id`; accepted fields `change_id`, `identity_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete an identity from a
  change entry; this may remove or archive FireHydrant data.
- `delete_checklist_template`: DELETE `/checklist_templates/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Archive a checklist
  template; this may remove or archive FireHydrant data.
- `delete_comment`: DELETE `/conversations/{{ record.conversation_id }}/comments/{{
  record.comment_id }}` - kind `delete`; body type `none`; path fields `comment_id`,
  `conversation_id`; required record fields `comment_id`, `conversation_id`; accepted fields
  `comment_id`, `conversation_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Archive a conversation comment; this may remove or archive FireHydrant data.
- `delete_comment_reaction`: DELETE `/conversations/{{ record.conversation_id }}/comments/{{
  record.comment_id }}/reactions/{{ record.reaction_id }}` - kind `delete`; body type `none`; path
  fields `reaction_id`, `conversation_id`, `comment_id`; required record fields `reaction_id`,
  `conversation_id`, `comment_id`; accepted fields `comment_id`, `conversation_id`, `reaction_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  reaction from a conversation comment; this may remove or archive FireHydrant data.
- `delete_custom_field_definition`: DELETE `/custom_fields/definitions/{{ record.field_id }}` - kind
  `delete`; body type `none`; path fields `field_id`; required record fields `field_id`; accepted
  fields `field_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a custom field definition; this may remove or archive FireHydrant
  data.
- `delete_environment`: DELETE `/environments/{{ record.environment_id }}` - kind `delete`; body
  type `none`; path fields `environment_id`; required record fields `environment_id`; accepted
  fields `environment_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Archive an environment; this may remove or archive FireHydrant data.
- `delete_functionality`: DELETE `/functionalities/{{ record.functionality_id }}` - kind `delete`;
  body type `none`; path fields `functionality_id`; required record fields `functionality_id`;
  accepted fields `functionality_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Archive a functionality; this may remove or archive FireHydrant
  data.
- `delete_inbound_field_map`: DELETE `/ticketing/projects/{{ record.ticketing_project_id
  }}/inbound_field_maps/{{ record.map_id }}` - kind `delete`; body type `none`; path fields
  `map_id`, `ticketing_project_id`; required record fields `map_id`, `ticketing_project_id`;
  accepted fields `map_id`, `ticketing_project_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Archive inbound field map for a ticketing project; this
  may remove or archive FireHydrant data.
- `delete_incident`: DELETE `/incidents/{{ record.incident_id }}` - kind `delete`; body type `none`;
  path fields `incident_id`; required record fields `incident_id`; accepted fields `incident_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Archive an
  incident; this may remove or archive FireHydrant data.
- `delete_incident_alert`: DELETE `/incidents/{{ record.incident_id }}/alerts/{{
  record.incident_alert_id }}` - kind `delete`; body type `none`; path fields `incident_alert_id`,
  `incident_id`; required record fields `incident_alert_id`, `incident_id`; accepted fields
  `incident_alert_id`, `incident_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Remove an alert from an incident; this may remove or archive
  FireHydrant data.
- `delete_incident_chat_message`: DELETE `/incidents/{{ record.incident_id
  }}/generic_chat_messages/{{ record.message_id }}` - kind `delete`; body type `none`; path fields
  `message_id`, `incident_id`; required record fields `message_id`, `incident_id`; accepted fields
  `incident_id`, `message_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a chat message from an incident; this may remove or archive
  FireHydrant data.
- `delete_incident_event`: DELETE `/incidents/{{ record.incident_id }}/events/{{ record.event_id }}`
  - kind `delete`; body type `none`; path fields `incident_id`, `event_id`; required record fields
  `incident_id`, `event_id`; accepted fields `event_id`, `incident_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete an incident event; this may
  remove or archive FireHydrant data.
- `delete_incident_impact`: DELETE `/incidents/{{ record.incident_id }}/impact/{{ record.type }}/{{
  record.id }}` - kind `delete`; body type `none`; path fields `incident_id`, `type`, `id`; required
  record fields `incident_id`, `type`, `id`; accepted fields `id`, `incident_id`, `type`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Remove impacted
  infrastructure from an incident; this may remove or archive FireHydrant data.
- `delete_incident_link`: DELETE `/incidents/{{ record.incident_id }}/links/{{ record.link_id }}` -
  kind `delete`; body type `none`; path fields `link_id`, `incident_id`; required record fields
  `link_id`, `incident_id`; accepted fields `incident_id`, `link_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Remove a link from an incident; this
  may remove or archive FireHydrant data.
- `delete_incident_role`: DELETE `/incident_roles/{{ record.incident_role_id }}` - kind `delete`;
  body type `none`; path fields `incident_role_id`; required record fields `incident_role_id`;
  accepted fields `incident_role_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Archive an incident role; this may remove or archive FireHydrant
  data.
- `delete_incident_role_assignment`: DELETE `/incidents/{{ record.incident_id }}/role_assignments/{{
  record.role_assignment_id }}` - kind `delete`; body type `none`; path fields `incident_id`,
  `role_assignment_id`; required record fields `incident_id`, `role_assignment_id`; accepted fields
  `incident_id`, `role_assignment_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Unassign a user from an incident; this may remove or archive
  FireHydrant data.
- `delete_incident_status_page`: DELETE `/incidents/{{ record.incident_id }}/status_pages/{{
  record.status_page_id }}` - kind `delete`; body type `none`; path fields `status_page_id`,
  `incident_id`; required record fields `status_page_id`, `incident_id`; accepted fields
  `incident_id`, `status_page_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Remove a status page from an incident; this may remove or archive FireHydrant
  data.
- `delete_incident_task`: DELETE `/incidents/{{ record.incident_id }}/tasks/{{ record.task_id }}` -
  kind `delete`; body type `none`; path fields `task_id`, `incident_id`; required record fields
  `task_id`, `incident_id`; accepted fields `incident_id`, `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete an incident task; this may
  remove or archive FireHydrant data.
- `delete_incident_type`: DELETE `/incident_types/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Archive an incident type;
  this may remove or archive FireHydrant data.
- `delete_lifecycle_measurement_definition`: DELETE `/lifecycles/measurement_definitions/{{
  record.measurement_definition_id }}` - kind `delete`; body type `none`; path fields
  `measurement_definition_id`; required record fields `measurement_definition_id`; accepted fields
  `measurement_definition_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Archive a measurement definition; this may remove or archive FireHydrant
  data.
- `delete_lifecycle_milestone`: DELETE `/lifecycles/milestones/{{ record.milestone_id }}` - kind
  `delete`; body type `none`; path fields `milestone_id`; required record fields `milestone_id`;
  accepted fields `milestone_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a milestone; this may remove or archive FireHydrant data.
- `delete_notification_policy`: DELETE `/signals/notification_policy_items/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  notification policy; this may remove or archive FireHydrant data.
- `delete_nunc_component_group`: DELETE `/nunc_connections/{{ record.nunc_connection_id
  }}/component_groups/{{ record.group_id }}` - kind `delete`; body type `none`; path fields
  `nunc_connection_id`, `group_id`; required record fields `nunc_connection_id`, `group_id`;
  accepted fields `group_id`, `nunc_connection_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete a status page component group; this may remove or
  archive FireHydrant data.
- `delete_nunc_connection`: DELETE `/nunc_connections/{{ record.nunc_connection_id }}` - kind
  `delete`; body type `none`; path fields `nunc_connection_id`; required record fields
  `nunc_connection_id`; accepted fields `nunc_connection_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a status page; this may remove or archive
  FireHydrant data.
- `delete_nunc_image`: DELETE `/nunc_connections/{{ record.nunc_connection_id }}/images/{{
  record.type }}` - kind `delete`; body type `none`; path fields `nunc_connection_id`, `type`;
  required record fields `nunc_connection_id`, `type`; accepted fields `nunc_connection_id`, `type`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete an
  image from a status page; this may remove or archive FireHydrant data.
- `delete_nunc_link`: DELETE `/nunc_connections/{{ record.nunc_connection_id }}/links/{{
  record.link_id }}` - kind `delete`; body type `none`; path fields `nunc_connection_id`, `link_id`;
  required record fields `nunc_connection_id`, `link_id`; accepted fields `link_id`,
  `nunc_connection_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a status page link; this may remove or archive FireHydrant data.
- `delete_nunc_subscription`: DELETE `/nunc/subscriptions/{{ record.unsubscribe_token }}` - kind
  `delete`; body type `none`; path fields `unsubscribe_token`; required record fields
  `unsubscribe_token`; accepted fields `unsubscribe_token`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Unsubscribe from status page notifications; this
  may remove or archive FireHydrant data.
- `delete_on_call_schedule_rotation`: DELETE `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}/rotations/{{ record.rotation_id }}` - kind `delete`; body type `none`; path
  fields `rotation_id`, `team_id`, `schedule_id`; required record fields `rotation_id`, `team_id`,
  `schedule_id`; accepted fields `rotation_id`, `schedule_id`, `team_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete an on-call schedule's rotation;
  this may remove or archive FireHydrant data.
- `delete_on_call_shift`: DELETE `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}/shifts/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`,
  `team_id`, `schedule_id`; required record fields `id`, `team_id`, `schedule_id`; accepted fields
  `id`, `schedule_id`, `team_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete an on-call shift from a team schedule; this may remove or archive
  FireHydrant data.
- `delete_post_mortem_reason`: DELETE `/post_mortems/reports/{{ record.report_id }}/reasons/{{
  record.reason_id }}` - kind `delete`; body type `none`; path fields `report_id`, `reason_id`;
  required record fields `report_id`, `reason_id`; accepted fields `reason_id`, `report_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  contributing factor from a retrospective report; this may remove or archive FireHydrant data.
- `delete_priority`: DELETE `/priorities/{{ record.priority_slug }}` - kind `delete`; body type
  `none`; path fields `priority_slug`; required record fields `priority_slug`; accepted fields
  `priority_slug`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a priority; this may remove or archive FireHydrant data.
- `delete_retrospective_template`: DELETE `/retrospective_templates/{{
  record.retrospective_template_id }}` - kind `delete`; body type `none`; path fields
  `retrospective_template_id`; required record fields `retrospective_template_id`; accepted fields
  `retrospective_template_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a retrospective template; this may remove or archive FireHydrant data.
- `delete_role`: DELETE `/roles/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a role; this may remove or archive
  FireHydrant data.
- `delete_runbook`: DELETE `/runbooks/{{ record.runbook_id }}` - kind `delete`; body type `none`;
  path fields `runbook_id`; required record fields `runbook_id`; accepted fields `runbook_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  runbook; this may remove or archive FireHydrant data.
- `delete_saved_search`: DELETE `/saved_searches/{{ record.resource_type }}/{{
  record.saved_search_id }}` - kind `delete`; body type `none`; path fields `resource_type`,
  `saved_search_id`; required record fields `resource_type`, `saved_search_id`; accepted fields
  `resource_type`, `saved_search_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a saved search; this may remove or archive FireHydrant
  data.
- `delete_scheduled_maintenance`: DELETE `/scheduled_maintenances/{{ record.scheduled_maintenance_id
  }}` - kind `delete`; body type `none`; path fields `scheduled_maintenance_id`; required record
  fields `scheduled_maintenance_id`; accepted fields `scheduled_maintenance_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a scheduled
  maintenance event; this may remove or archive FireHydrant data.
- `delete_service`: DELETE `/services/{{ record.service_id }}` - kind `delete`; body type `none`;
  path fields `service_id`; required record fields `service_id`; accepted fields `service_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  service; this may remove or archive FireHydrant data.
- `delete_service_dependency`: DELETE `/service_dependencies/{{ record.service_dependency_id }}` -
  kind `delete`; body type `none`; path fields `service_dependency_id`; required record fields
  `service_dependency_id`; accepted fields `service_dependency_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a service dependency; this may
  remove or archive FireHydrant data.
- `delete_service_link`: DELETE `/services/{{ record.service_id }}/service_links/{{ record.remote_id
  }}` - kind `delete`; body type `none`; path fields `service_id`, `remote_id`; required record
  fields `service_id`, `remote_id`; accepted fields `remote_id`, `service_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a service link; this
  may remove or archive FireHydrant data.
- `delete_severity`: DELETE `/severities/{{ record.severity_slug }}` - kind `delete`; body type
  `none`; path fields `severity_slug`; required record fields `severity_slug`; accepted fields
  `severity_slug`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a severity; this may remove or archive FireHydrant data.
- `delete_severity_matrix_condition`: DELETE `/severity_matrix/conditions/{{ record.condition_id }}`
  - kind `delete`; body type `none`; path fields `condition_id`; required record fields
  `condition_id`; accepted fields `condition_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete a severity matrix condition; this may remove or
  archive FireHydrant data.
- `delete_severity_matrix_impact`: DELETE `/severity_matrix/impacts/{{ record.impact_id }}` - kind
  `delete`; body type `none`; path fields `impact_id`; required record fields `impact_id`; accepted
  fields `impact_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a severity matrix impact; this may remove or archive FireHydrant data.
- `delete_signals_alert_grouping_configuration`: DELETE `/signals/grouping/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete an
  alert grouping configuration.; this may remove or archive FireHydrant data.
- `delete_signals_email_target`: DELETE `/signals/email_targets/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Delete a signal
  email target; this may remove or archive FireHydrant data.
- `delete_signals_event_source`: DELETE `/signals/event_sources/{{ record.transposer_slug }}` - kind
  `delete`; body type `none`; path fields `transposer_slug`; required record fields
  `transposer_slug`; accepted fields `transposer_slug`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete an event source for Signals; this may
  remove or archive FireHydrant data.
- `delete_signals_heartbeat_endpoint_configuration`: DELETE `/signals/heartbeat_endpoints/{{
  record.id }}` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a heartbeat endpoint configuration; this may remove or archive
  FireHydrant data.
- `delete_signals_webhook_target`: DELETE `/signals/webhook_targets/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  webhook target; this may remove or archive FireHydrant data.
- `delete_slack_emoji_action`: DELETE `/integrations/slack/connections/{{ record.connection_id
  }}/emoji_actions/{{ record.emoji_action_id }}` - kind `delete`; body type `none`; path fields
  `connection_id`, `emoji_action_id`; required record fields `connection_id`, `emoji_action_id`;
  accepted fields `connection_id`, `emoji_action_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete a Slack emoji action; this may remove or archive
  FireHydrant data.
- `delete_status_update_template`: DELETE `/status_update_templates/{{
  record.status_update_template_id }}` - kind `delete`; body type `none`; path fields
  `status_update_template_id`; required record fields `status_update_template_id`; accepted fields
  `status_update_template_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a status update template; this may remove or archive FireHydrant data.
- `delete_statuspage_connection`: DELETE `/integrations/statuspage/connections/{{
  record.connection_id }}` - kind `delete`; body type `none`; path fields `connection_id`; required
  record fields `connection_id`; accepted fields `connection_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a Statuspage connection; this may
  remove or archive FireHydrant data.
- `delete_support_hours_schedule`: DELETE `/teams/{{ record.team_id }}/support_hours_schedule` -
  kind `delete`; body type `none`; path fields `team_id`; required record fields `team_id`; accepted
  fields `team_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a specific support hours schedule; this may remove or archive FireHydrant data.
- `delete_task_list`: DELETE `/task_lists/{{ record.task_list_id }}` - kind `delete`; body type
  `none`; path fields `task_list_id`; required record fields `task_list_id`; accepted fields
  `task_list_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a task list; this may remove or archive FireHydrant data.
- `delete_team`: DELETE `/teams/{{ record.team_id }}` - kind `delete`; body type `none`; path fields
  `team_id`; required record fields `team_id`; accepted fields `team_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Archive a team; this may remove or
  archive FireHydrant data.
- `delete_team_escalation_policy`: DELETE `/teams/{{ record.team_id }}/escalation_policies/{{
  record.id }}` - kind `delete`; body type `none`; path fields `team_id`, `id`; required record
  fields `team_id`, `id`; accepted fields `id`, `team_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete an escalation policy for a team; this may
  remove or archive FireHydrant data.
- `delete_team_on_call_schedule`: DELETE `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}` - kind `delete`; body type `none`; path fields `team_id`, `schedule_id`;
  required record fields `team_id`, `schedule_id`; accepted fields `schedule_id`, `team_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Delete an on-call
  schedule for a team; this may remove or archive FireHydrant data.
- `delete_team_signal_rule`: DELETE `/teams/{{ record.team_id }}/signal_rules/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `team_id`, `id`; required record fields `team_id`,
  `id`; accepted fields `id`, `team_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a Signals rule; this may remove or archive FireHydrant
  data.
- `delete_ticket`: DELETE `/ticketing/tickets/{{ record.ticket_id }}` - kind `delete`; body type
  `none`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Archive a
  ticket; this may remove or archive FireHydrant data.
- `delete_ticketing_custom_definition`: DELETE `/ticketing/custom_fields/definitions/{{
  record.field_id }}` - kind `delete`; body type `none`; path fields `field_id`; required record
  fields `field_id`; accepted fields `field_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete a ticketing custom field; this may remove or
  archive FireHydrant data.
- `delete_ticketing_field_map`: DELETE `/ticketing/projects/{{ record.ticketing_project_id
  }}/field_maps/{{ record.map_id }}` - kind `delete`; body type `none`; path fields `map_id`,
  `ticketing_project_id`; required record fields `map_id`, `ticketing_project_id`; accepted fields
  `map_id`, `ticketing_project_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Archive a field map for a ticketing project; this may remove or
  archive FireHydrant data.
- `delete_ticketing_priority`: DELETE `/ticketing/priorities/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a ticketing
  priority; this may remove or archive FireHydrant data.
- `delete_ticketing_project_config`: DELETE `/ticketing/projects/{{ record.ticketing_project_id
  }}/provider_project_configurations/{{ record.config_id }}` - kind `delete`; body type `none`; path
  fields `ticketing_project_id`, `config_id`; required record fields `ticketing_project_id`,
  `config_id`; accepted fields `config_id`, `ticketing_project_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Archive a ticketing project
  configuration; this may remove or archive FireHydrant data.
- `delete_transcript_entry`: DELETE `/incidents/{{ record.incident_id }}/transcript/{{
  record.transcript_id }}` - kind `delete`; body type `none`; path fields `transcript_id`,
  `incident_id`; required record fields `transcript_id`, `incident_id`; accepted fields
  `incident_id`, `transcript_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a transcript from an incident; this may remove or archive FireHydrant
  data.
- `delete_webhook`: DELETE `/webhooks/{{ record.webhook_id }}` - kind `delete`; body type `none`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  webhook; this may remove or archive FireHydrant data.
- `generate_audience_summary`: POST `/audiences/{{ record.audience_id }}/summaries/{{
  record.incident_id }}` - kind `create`; body type `json`; path fields `audience_id`,
  `incident_id`; required record fields `audience_id`, `incident_id`; accepted fields `audience_id`,
  `incident_id`; risk: Generate summary (async) through the FireHydrant API.
- `ingest_catalog_data`: POST `/catalogs/{{ record.catalog_id }}/ingest` - kind `create`; body type
  `json`; path fields `catalog_id`; required record fields `catalog_id`; accepted fields
  `catalog_id`; risk: Ingest service catalog data through the FireHydrant API.
- `override_on_call_schedule_rotation_shifts`: POST `/teams/{{ record.team_id
  }}/on_call_schedules/{{ record.schedule_id }}/rotations/{{ record.rotation_id }}/overrides` - kind
  `custom`; body type `json`; path fields `rotation_id`, `team_id`, `schedule_id`; required record
  fields `rotation_id`, `team_id`, `schedule_id`; accepted fields `rotation_id`, `schedule_id`,
  `team_id`; risk: Override one or more shifts in an on-call rotation through the FireHydrant API.
- `preview_on_call_schedule_rotation`: POST `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}/rotations/preview` - kind `custom`; body type `json`; path fields `team_id`,
  `schedule_id`; required record fields `team_id`, `schedule_id`; accepted fields `schedule_id`,
  `team_id`; risk: Preview an on-call rotation through the FireHydrant API.
- `preview_team_on_call_schedule`: POST `/teams/{{ record.team_id }}/on_call_schedules/preview` -
  kind `custom`; body type `json`; path fields `team_id`; required record fields `team_id`; accepted
  fields `team_id`; risk: Preview a new on-call schedule for a team through the FireHydrant API.
- `publish_nunc_connection`: POST `/nunc_connections/{{ record.nunc_connection_id }}/publish` - kind
  `update`; body type `json`; path fields `nunc_connection_id`; required record fields
  `nunc_connection_id`; accepted fields `nunc_connection_id`; confirmation `destructive`; risk:
  Publish a status page through the FireHydrant API.
- `publish_post_mortem_report`: POST `/post_mortems/reports/{{ record.report_id }}/publish` - kind
  `update`; body type `json`; path fields `report_id`; required record fields `report_id`; accepted
  fields `report_id`; confirmation `destructive`; risk: Publish a retrospective report through the
  FireHydrant API.
- `refresh_connection`: PATCH `/integrations/connections/{{ record.slug }}/{{ record.connection_id
  }}/refresh` - kind `custom`; body type `json`; path fields `slug`, `connection_id`; required
  record fields `slug`, `connection_id`; accepted fields `connection_id`, `slug`; risk: Refresh an
  integration connection's incident role schedules through the FireHydrant API.
- `reorder_post_mortem_reasons`: PUT `/post_mortems/reports/{{ record.report_id }}/reasons/order` -
  kind `update`; body type `json`; path fields `report_id`; required record fields `report_id`;
  accepted fields `report_id`; risk: Reorder a contributing factor for a retrospective report
  through the FireHydrant API.
- `resolve_incident`: PUT `/incidents/{{ record.incident_id }}/resolve` - kind `update`; body type
  `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; confirmation `destructive`; risk: Resolve an incident through the FireHydrant API.
- `restore_audience`: PATCH `/audiences/{{ record.audience_id }}/restore` - kind `update`; body type
  `json`; path fields `audience_id`; required record fields `audience_id`; accepted fields
  `audience_id`; risk: Restore audience through the FireHydrant API.
- `set_member_default_audience`: PUT `/audiences/member/{{ record.member_id }}/default` - kind
  `update`; body type `json`; path fields `member_id`; required record fields `member_id`; accepted
  fields `member_id`; risk: Set default audience through the FireHydrant API.
- `share_incident_retrospectives`: POST `/incidents/{{ record.incident_id }}/retrospectives/share` -
  kind `custom`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Share an incident's retrospective through the FireHydrant
  API.
- `test_slack_channel`: PUT `/integrations/slack/channels/{{ record.id }}/test` - kind `custom`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Test
  a Slack channel through the FireHydrant API.
- `unarchive_incident`: POST `/incidents/{{ record.incident_id }}/unarchive` - kind `update`; body
  type `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; confirmation `destructive`; risk: Unarchive an incident through the FireHydrant
  API.
- `unpublish_nunc_connection`: POST `/nunc_connections/{{ record.nunc_connection_id }}/unpublish` -
  kind `update`; body type `json`; path fields `nunc_connection_id`; required record fields
  `nunc_connection_id`; accepted fields `nunc_connection_id`; confirmation `destructive`; risk:
  Unpublish a status page through the FireHydrant API.
- `update_ai_preferences`: PATCH `/ai/preferences` - kind `update`; body type `json`; risk: Update
  AI preferences through the FireHydrant API.
- `update_audience`: PATCH `/audiences/{{ record.audience_id }}` - kind `update`; body type `json`;
  path fields `audience_id`; required record fields `audience_id`; accepted fields `audience_id`;
  risk: Update audience through the FireHydrant API.
- `update_authed_provider`: PATCH `/integrations/authed_providers/{{ record.integration_slug }}/{{
  record.connection_id }}/{{ record.authed_provider_id }}` - kind `update`; body type `json`; path
  fields `integration_slug`, `connection_id`, `authed_provider_id`; required record fields
  `integration_slug`, `connection_id`, `authed_provider_id`; accepted fields `authed_provider_id`,
  `connection_id`, `integration_slug`; risk: Get an authed provider through the FireHydrant API.
- `update_aws_cloudtrail_batch`: PATCH `/integrations/aws/cloudtrail_batches/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Update a CloudTrail batch through the FireHydrant API.
- `update_aws_connection`: PATCH `/integrations/aws/connections/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Update an AWS connection through the FireHydrant API.
- `update_call_route`: PATCH `/signals/call_routes/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Update a call
  route through the FireHydrant API.
- `update_change`: PATCH `/changes/{{ record.change_id }}` - kind `update`; body type `json`; path
  fields `change_id`; required record fields `change_id`; accepted fields `change_id`; risk: Update
  a change entry through the FireHydrant API.
- `update_change_event`: PATCH `/changes/events/{{ record.change_event_id }}` - kind `update`; body
  type `json`; path fields `change_event_id`; required record fields `change_event_id`; accepted
  fields `change_event_id`; risk: Update a change event through the FireHydrant API.
- `update_change_identity`: PATCH `/changes/{{ record.change_id }}/identities/{{ record.identity_id
  }}` - kind `update`; body type `json`; path fields `identity_id`, `change_id`; required record
  fields `identity_id`, `change_id`; accepted fields `change_id`, `identity_id`; risk: Update an
  identity for a change entry through the FireHydrant API.
- `update_checklist_template`: PATCH `/checklist_templates/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Update a
  checklist template through the FireHydrant API.
- `update_comment`: PATCH `/conversations/{{ record.conversation_id }}/comments/{{ record.comment_id
  }}` - kind `update`; body type `json`; path fields `comment_id`, `conversation_id`; required
  record fields `comment_id`, `conversation_id`; accepted fields `comment_id`, `conversation_id`;
  risk: Update a conversation comment through the FireHydrant API.
- `update_connection`: PATCH `/integrations/connections/{{ record.slug }}/{{ record.connection_id
  }}` - kind `update`; body type `json`; path fields `slug`, `connection_id`; required record fields
  `slug`, `connection_id`; accepted fields `connection_id`, `slug`; risk: Update an integration
  connection through the FireHydrant API.
- `update_custom_field_definition`: PATCH `/custom_fields/definitions/{{ record.field_id }}` - kind
  `update`; body type `json`; path fields `field_id`; required record fields `field_id`; accepted
  fields `field_id`; risk: Update a custom field definition through the FireHydrant API.
- `update_environment`: PATCH `/environments/{{ record.environment_id }}` - kind `update`; body type
  `json`; path fields `environment_id`; required record fields `environment_id`; accepted fields
  `environment_id`; risk: Update an environment through the FireHydrant API.
- `update_field_map`: PATCH `/integrations/field_maps/{{ record.field_map_id }}` - kind `update`;
  body type `json`; path fields `field_map_id`; required record fields `field_map_id`; accepted
  fields `field_map_id`; risk: Update field mapping through the FireHydrant API.
- `update_functionality`: PATCH `/functionalities/{{ record.functionality_id }}` - kind `update`;
  body type `json`; path fields `functionality_id`; required record fields `functionality_id`;
  accepted fields `functionality_id`; risk: Update a functionality through the FireHydrant API.
- `update_inbound_field_map`: PUT `/ticketing/projects/{{ record.ticketing_project_id
  }}/inbound_field_maps/{{ record.map_id }}` - kind `update`; body type `json`; path fields
  `map_id`, `ticketing_project_id`; required record fields `map_id`, `ticketing_project_id`;
  accepted fields `map_id`, `ticketing_project_id`; risk: Update inbound field map for a ticketing
  project through the FireHydrant API.
- `update_incident`: PATCH `/incidents/{{ record.incident_id }}` - kind `update`; body type `json`;
  path fields `incident_id`; required record fields `incident_id`; accepted fields `incident_id`;
  risk: Update an incident through the FireHydrant API.
- `update_incident_alert_primary`: PATCH `/incidents/{{ record.incident_id }}/alerts/{{
  record.incident_alert_id }}/primary` - kind `update`; body type `json`; path fields
  `incident_alert_id`, `incident_id`; required record fields `incident_alert_id`, `incident_id`;
  accepted fields `incident_alert_id`, `incident_id`; risk: Set an alert as primary for an incident
  through the FireHydrant API.
- `update_incident_change_event`: PATCH `/incidents/{{ record.incident_id
  }}/related_change_events/{{ record.related_change_event_id }}` - kind `update`; body type `json`;
  path fields `related_change_event_id`, `incident_id`; required record fields
  `related_change_event_id`, `incident_id`; accepted fields `incident_id`,
  `related_change_event_id`; risk: Update a change attached to an incident through the FireHydrant
  API.
- `update_incident_chat_message`: PATCH `/incidents/{{ record.incident_id
  }}/generic_chat_messages/{{ record.message_id }}` - kind `update`; body type `json`; path fields
  `message_id`, `incident_id`; required record fields `message_id`, `incident_id`; accepted fields
  `incident_id`, `message_id`; risk: Update a chat message on an incident through the FireHydrant
  API.
- `update_incident_event`: PATCH `/incidents/{{ record.incident_id }}/events/{{ record.event_id }}`
  - kind `update`; body type `json`; path fields `incident_id`, `event_id`; required record fields
  `incident_id`, `event_id`; accepted fields `event_id`, `incident_id`; risk: Update an incident
  event through the FireHydrant API.
- `update_incident_impact_patch`: PATCH `/incidents/{{ record.incident_id }}/impact` - kind
  `update`; body type `json`; path fields `incident_id`; required record fields `incident_id`;
  accepted fields `incident_id`; risk: Update impacts for an incident through the FireHydrant API.
- `update_incident_impact_put`: PUT `/incidents/{{ record.incident_id }}/impact` - kind `update`;
  body type `json`; path fields `incident_id`; required record fields `incident_id`; accepted fields
  `incident_id`; risk: Update impacts for an incident through the FireHydrant API.
- `update_incident_link`: PUT `/incidents/{{ record.incident_id }}/links/{{ record.link_id }}` -
  kind `update`; body type `json`; path fields `link_id`, `incident_id`; required record fields
  `link_id`, `incident_id`; accepted fields `incident_id`, `link_id`; risk: Update the external
  incident link through the FireHydrant API.
- `update_incident_note`: PATCH `/incidents/{{ record.incident_id }}/notes/{{ record.note_id }}` -
  kind `update`; body type `json`; path fields `note_id`, `incident_id`; required record fields
  `note_id`, `incident_id`; accepted fields `incident_id`, `note_id`; risk: Update a note through
  the FireHydrant API.
- `update_incident_retrospective`: PATCH `/incidents/{{ record.incident_id }}/retrospectives/{{
  record.retrospective_id }}` - kind `update`; body type `json`; path fields `retrospective_id`,
  `incident_id`; required record fields `retrospective_id`, `incident_id`; accepted fields
  `incident_id`, `retrospective_id`; risk: Update a retrospective on the incident through the
  FireHydrant API.
- `update_incident_retrospective_field`: PATCH `/incidents/{{ record.incident_id
  }}/retrospectives/{{ record.retrospective_id }}/fields/{{ record.field_id }}` - kind `update`;
  body type `json`; path fields `retrospective_id`, `field_id`, `incident_id`; required record
  fields `retrospective_id`, `field_id`, `incident_id`; accepted fields `field_id`, `incident_id`,
  `retrospective_id`; risk: Update the value on a retrospective field through the FireHydrant API.
- `update_incident_role`: PATCH `/incident_roles/{{ record.incident_role_id }}` - kind `update`;
  body type `json`; path fields `incident_role_id`; required record fields `incident_role_id`;
  accepted fields `incident_role_id`; risk: Update an incident role through the FireHydrant API.
- `update_incident_task`: PATCH `/incidents/{{ record.incident_id }}/tasks/{{ record.task_id }}` -
  kind `update`; body type `json`; path fields `task_id`, `incident_id`; required record fields
  `task_id`, `incident_id`; accepted fields `incident_id`, `task_id`; risk: Update an incident task
  through the FireHydrant API.
- `update_incident_type`: PATCH `/incident_types/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: Update an incident type
  through the FireHydrant API.
- `update_lifecycle_measurement_definition`: PATCH `/lifecycles/measurement_definitions/{{
  record.measurement_definition_id }}` - kind `update`; body type `json`; path fields
  `measurement_definition_id`; required record fields `measurement_definition_id`; accepted fields
  `measurement_definition_id`; risk: Update a measurement definition through the FireHydrant API.
- `update_lifecycle_milestone`: PATCH `/lifecycles/milestones/{{ record.milestone_id }}` - kind
  `update`; body type `json`; path fields `milestone_id`; required record fields `milestone_id`;
  accepted fields `milestone_id`; risk: Update a milestone through the FireHydrant API.
- `update_notification_policy`: PATCH `/signals/notification_policy_items/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Update a notification policy through the FireHydrant API.
- `update_nunc_component_group`: PATCH `/nunc_connections/{{ record.nunc_connection_id
  }}/component_groups/{{ record.group_id }}` - kind `update`; body type `json`; path fields
  `nunc_connection_id`, `group_id`; required record fields `nunc_connection_id`, `group_id`;
  accepted fields `group_id`, `nunc_connection_id`; risk: Update a status page component group
  through the FireHydrant API.
- `update_nunc_connection`: PUT `/nunc_connections/{{ record.nunc_connection_id }}` - kind `update`;
  body type `json`; path fields `nunc_connection_id`; required record fields `nunc_connection_id`;
  accepted fields `nunc_connection_id`; risk: Update a status page through the FireHydrant API.
- `update_nunc_link`: PATCH `/nunc_connections/{{ record.nunc_connection_id }}/links/{{
  record.link_id }}` - kind `update`; body type `json`; path fields `nunc_connection_id`, `link_id`;
  required record fields `nunc_connection_id`, `link_id`; accepted fields `link_id`,
  `nunc_connection_id`; risk: Update a status page link through the FireHydrant API.
- `update_on_call_schedule_rotation`: PATCH `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}/rotations/{{ record.rotation_id }}` - kind `update`; body type `json`; path
  fields `rotation_id`, `team_id`, `schedule_id`; required record fields `rotation_id`, `team_id`,
  `schedule_id`; accepted fields `rotation_id`, `schedule_id`, `team_id`; risk: Update an on-call
  schedule's rotation through the FireHydrant API.
- `update_on_call_shift`: PATCH `/teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id
  }}/shifts/{{ record.id }}` - kind `update`; body type `json`; path fields `id`, `team_id`,
  `schedule_id`; required record fields `id`, `team_id`, `schedule_id`; accepted fields `id`,
  `schedule_id`, `team_id`; risk: Update an on-call shift for a team schedule through the
  FireHydrant API.
- `update_post_mortem_field`: PATCH `/post_mortems/reports/{{ record.report_id }}/fields/{{
  record.field_id }}` - kind `update`; body type `json`; path fields `field_id`, `report_id`;
  required record fields `field_id`, `report_id`; accepted fields `field_id`, `report_id`; risk:
  Update a retrospective field through the FireHydrant API.
- `update_post_mortem_questions`: PUT `/post_mortems/questions` - kind `update`; body type `json`;
  risk: Update retrospective questions through the FireHydrant API.
- `update_post_mortem_reason`: PATCH `/post_mortems/reports/{{ record.report_id }}/reasons/{{
  record.reason_id }}` - kind `update`; body type `json`; path fields `report_id`, `reason_id`;
  required record fields `report_id`, `reason_id`; accepted fields `reason_id`, `report_id`; risk:
  Update a contributing factor in a retrospective report through the FireHydrant API.
- `update_post_mortem_report`: PATCH `/post_mortems/reports/{{ record.report_id }}` - kind `update`;
  body type `json`; path fields `report_id`; required record fields `report_id`; accepted fields
  `report_id`; risk: Update a retrospective report through the FireHydrant API.
- `update_priority`: PATCH `/priorities/{{ record.priority_slug }}` - kind `update`; body type
  `json`; path fields `priority_slug`; required record fields `priority_slug`; accepted fields
  `priority_slug`; risk: Update a priority through the FireHydrant API.
- `update_retrospective_template`: PATCH `/retrospective_templates/{{
  record.retrospective_template_id }}` - kind `update`; body type `json`; path fields
  `retrospective_template_id`; required record fields `retrospective_template_id`; accepted fields
  `retrospective_template_id`; risk: Update a retrospective template through the FireHydrant API.
- `update_role`: PATCH `/roles/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `id`; risk: Update a role through the FireHydrant
  API.
- `update_runbook`: PUT `/runbooks/{{ record.runbook_id }}` - kind `update`; body type `json`; path
  fields `runbook_id`; required record fields `runbook_id`; accepted fields `runbook_id`; risk:
  Update a runbook through the FireHydrant API.
- `update_runbook_execution_step`: PUT `/runbooks/executions/{{ record.execution_id }}/steps/{{
  record.step_id }}` - kind `update`; body type `json`; path fields `execution_id`, `step_id`;
  required record fields `execution_id`, `step_id`; accepted fields `execution_id`, `step_id`; risk:
  Update a runbook step execution through the FireHydrant API.
- `update_runbook_execution_step_script`: PUT `/runbooks/executions/{{ record.execution_id
  }}/steps/{{ record.step_id }}/script/{{ record.state }}` - kind `update`; body type `json`; path
  fields `execution_id`, `step_id`, `state`; required record fields `execution_id`, `step_id`,
  `state`; accepted fields `execution_id`, `state`, `step_id`; risk: Update a script step's
  execution status through the FireHydrant API.
- `update_saved_search`: PATCH `/saved_searches/{{ record.resource_type }}/{{ record.saved_search_id
  }}` - kind `update`; body type `json`; path fields `resource_type`, `saved_search_id`; required
  record fields `resource_type`, `saved_search_id`; accepted fields `resource_type`,
  `saved_search_id`; risk: Update a saved search through the FireHydrant API.
- `update_scheduled_maintenance`: PATCH `/scheduled_maintenances/{{ record.scheduled_maintenance_id
  }}` - kind `update`; body type `json`; path fields `scheduled_maintenance_id`; required record
  fields `scheduled_maintenance_id`; accepted fields `scheduled_maintenance_id`; risk: Update a
  scheduled maintenance event through the FireHydrant API.
- `update_service`: PATCH `/services/{{ record.service_id }}` - kind `update`; body type `json`;
  path fields `service_id`; required record fields `service_id`; accepted fields `service_id`; risk:
  Update a service through the FireHydrant API.
- `update_service_dependency`: PATCH `/service_dependencies/{{ record.service_dependency_id }}` -
  kind `update`; body type `json`; path fields `service_dependency_id`; required record fields
  `service_dependency_id`; accepted fields `service_dependency_id`; risk: Update a service
  dependency through the FireHydrant API.
- `update_severity`: PATCH `/severities/{{ record.severity_slug }}` - kind `update`; body type
  `json`; path fields `severity_slug`; required record fields `severity_slug`; accepted fields
  `severity_slug`; risk: Update a severity through the FireHydrant API.
- `update_severity_matrix`: PATCH `/severity_matrix` - kind `update`; body type `json`; risk: Update
  severity matrix through the FireHydrant API.
- `update_severity_matrix_condition`: PATCH `/severity_matrix/conditions/{{ record.condition_id }}`
  - kind `update`; body type `json`; path fields `condition_id`; required record fields
  `condition_id`; accepted fields `condition_id`; risk: Update a severity matrix condition through
  the FireHydrant API.
- `update_severity_matrix_impact`: PATCH `/severity_matrix/impacts/{{ record.impact_id }}` - kind
  `update`; body type `json`; path fields `impact_id`; required record fields `impact_id`; accepted
  fields `impact_id`; risk: Update a severity matrix impact through the FireHydrant API.
- `update_signals_alert`: PATCH `/signals/alerts/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: Update a Signal alert
  through the FireHydrant API.
- `update_signals_alert_grouping_configuration`: PATCH `/signals/grouping/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Update an alert grouping configuration. through the FireHydrant API.
- `update_signals_email_target`: PATCH `/signals/email_targets/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Update an email target through the FireHydrant API.
- `update_signals_heartbeat_endpoint_configuration`: PATCH `/signals/heartbeat_endpoints/{{
  record.id }}` - kind `update`; body type `json`; path fields `id`; required record fields `id`;
  accepted fields `id`; risk: Update a heartbeat endpoint configuration through the FireHydrant API.
- `update_signals_webhook_target`: PATCH `/signals/webhook_targets/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Update a webhook target through the FireHydrant API.
- `update_slack_emoji_action`: PATCH `/integrations/slack/connections/{{ record.connection_id
  }}/emoji_actions/{{ record.emoji_action_id }}` - kind `update`; body type `json`; path fields
  `connection_id`, `emoji_action_id`; required record fields `connection_id`, `emoji_action_id`;
  accepted fields `connection_id`, `emoji_action_id`; risk: Update a Slack emoji action through the
  FireHydrant API.
- `update_status_update_template`: PATCH `/status_update_templates/{{
  record.status_update_template_id }}` - kind `update`; body type `json`; path fields
  `status_update_template_id`; required record fields `status_update_template_id`; accepted fields
  `status_update_template_id`; risk: Update a status update template through the FireHydrant API.
- `update_statuspage_connection`: PATCH `/integrations/statuspage/connections/{{
  record.connection_id }}` - kind `update`; body type `json`; path fields `connection_id`; required
  record fields `connection_id`; accepted fields `connection_id`; confirmation `destructive`; risk:
  Update a Statuspage connection through the FireHydrant API.
- `update_support_hours_schedule`: PATCH `/teams/{{ record.team_id }}/support_hours_schedule` - kind
  `update`; body type `json`; path fields `team_id`; required record fields `team_id`; accepted
  fields `team_id`; risk: Update support hours schedule through the FireHydrant API.
- `update_task_list`: PATCH `/task_lists/{{ record.task_list_id }}` - kind `update`; body type
  `json`; path fields `task_list_id`; required record fields `task_list_id`; accepted fields
  `task_list_id`; risk: Update a task list through the FireHydrant API.
- `update_team`: PATCH `/teams/{{ record.team_id }}` - kind `update`; body type `json`; path fields
  `team_id`; required record fields `team_id`; accepted fields `team_id`; risk: Update a team
  through the FireHydrant API.
- `update_team_escalation_policy`: PATCH `/teams/{{ record.team_id }}/escalation_policies/{{
  record.id }}` - kind `update`; body type `json`; path fields `team_id`, `id`; required record
  fields `team_id`, `id`; accepted fields `id`, `team_id`; risk: Update an escalation policy for a
  team through the FireHydrant API.
- `update_team_on_call_schedule`: PATCH `/teams/{{ record.team_id }}/on_call_schedules/{{
  record.schedule_id }}` - kind `update`; body type `json`; path fields `team_id`, `schedule_id`;
  required record fields `team_id`, `schedule_id`; accepted fields `schedule_id`, `team_id`; risk:
  Update an on-call schedule for a team through the FireHydrant API.
- `update_team_signal_rule`: PATCH `/teams/{{ record.team_id }}/signal_rules/{{ record.id }}` - kind
  `update`; body type `json`; path fields `team_id`, `id`; required record fields `team_id`, `id`;
  accepted fields `id`, `team_id`; risk: Update a Signals rule through the FireHydrant API.
- `update_ticket`: PATCH `/ticketing/tickets/{{ record.ticket_id }}` - kind `update`; body type
  `json`; path fields `ticket_id`; required record fields `ticket_id`; accepted fields `ticket_id`;
  risk: Update a ticket through the FireHydrant API.
- `update_ticketing_custom_definition`: PATCH `/ticketing/custom_fields/definitions/{{
  record.field_id }}` - kind `update`; body type `json`; path fields `field_id`; required record
  fields `field_id`; accepted fields `field_id`; risk: Update a ticketing custom field through the
  FireHydrant API.
- `update_ticketing_field_map`: PATCH `/ticketing/projects/{{ record.ticketing_project_id
  }}/field_maps/{{ record.map_id }}` - kind `update`; body type `json`; path fields `map_id`,
  `ticketing_project_id`; required record fields `map_id`, `ticketing_project_id`; accepted fields
  `map_id`, `ticketing_project_id`; risk: Update a field map for a ticketing project through the
  FireHydrant API.
- `update_ticketing_priority`: PATCH `/ticketing/priorities/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Update a
  ticketing priority through the FireHydrant API.
- `update_ticketing_project_config`: PATCH `/ticketing/projects/{{ record.ticketing_project_id
  }}/provider_project_configurations/{{ record.config_id }}` - kind `update`; body type `json`; path
  fields `ticketing_project_id`, `config_id`; required record fields `ticketing_project_id`,
  `config_id`; accepted fields `config_id`, `ticketing_project_id`; risk: Update configuration for a
  ticketing project through the FireHydrant API.
- `update_transcript_attribution`: PUT `/incidents/{{ record.incident_id }}/transcript/attribution`
  - kind `update`; body type `json`; path fields `incident_id`; required record fields
  `incident_id`; accepted fields `incident_id`; risk: Update the attribution of a transcript through
  the FireHydrant API.
- `update_vote`: PATCH `/incidents/{{ record.incident_id }}/events/{{ record.event_id }}/votes` -
  kind `update`; body type `json`; path fields `incident_id`, `event_id`; required record fields
  `incident_id`, `event_id`; accepted fields `event_id`, `incident_id`; risk: Update votes through
  the FireHydrant API.
- `update_webhook`: PATCH `/webhooks/{{ record.webhook_id }}` - kind `update`; body type `json`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`; risk:
  Update a webhook through the FireHydrant API.
- `validate_incident_tags`: POST `/incident_tags/validate` - kind `create`; body type `json`; risk:
  Validate incident tags through the FireHydrant API.
- `vote_ai_incident_summary`: PUT `/ai/summarize_incident/{{ record.incident_id }}/{{
  record.generated_summary_id }}/vote` - kind `update`; body type `json`; path fields `incident_id`,
  `generated_summary_id`; required record fields `incident_id`, `generated_summary_id`; accepted
  fields `generated_summary_id`, `incident_id`; risk: Vote on an AI-generated incident summary
  through the FireHydrant API.

## Known limits

- Published rate limit metadata: requests_per_minute=300.
- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 205 stream-backed endpoint group(s), 244 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=11, deprecated=1, non_data_endpoint=3, out_of_scope=3, requires_elevated_scope=12.
