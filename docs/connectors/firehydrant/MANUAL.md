# pm connectors inspect firehydrant

```text
NAME
  pm connectors inspect firehydrant - FireHydrant connector manual

SYNOPSIS
  pm connectors inspect firehydrant
  pm connectors inspect firehydrant --json
  pm credentials add <name> --connector firehydrant [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads broad FireHydrant REST API resources and exposes direct JSON/no-body FireHydrant mutations through declarative write actions.

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
  action_slug
  alert_id
  audience_id
  base_url
  by_connection_id
  change_event_id
  change_id
  comment_id
  condition_id
  config_id
  connection_id
  conversation_id
  emoji_action_id
  environment_id
  event_id
  execution_id
  field
  field_id
  field_map_id
  functionality_id
  generated_summary_id
  get_zendesk_customer_support_issue_ticket_id
  id
  incident_id
  incident_role_id
  infra_id
  infra_type
  integration_id
  integration_slug
  language_code
  map_id
  measurement_definition_id
  member_id
  nunc_connection_id
  page_size
  priority_slug
  question_id
  report_id
  resource_type
  retrospective_id
  retrospective_template_id
  rotation_id
  runbook_id
  saved_search_id
  schedule_id
  scheduled_maintenance_id
  search_zendesk_tickets_query
  selected_value
  service_dependency_id
  service_id
  severity_slug
  slug
  status_update_template_id
  step_id
  task_id
  task_list_id
  team_id
  ticket_id
  ticketing_project_id
  transposer_slug
  type
  user_id
  webhook_id
  api_token (secret)

ETL STREAMS
  incidents:
    primary key: id
    cursor: updated_at
    fields: created_at(string), current_milestone(string), description(string), id(string), name(string), number(integer), priority(string), resolved_at(string), severity(string), started_at(string), summary(string), updated_at(string)
  services:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), service_tier(integer), slug(string), updated_at(string)
  teams:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), slug(string), updated_at(string)
  environments:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), updated_at(string)
  functionalities:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), slug(string), updated_at(string)
  append_form_data_on_selected_value_get:
  get_ai_incident_summary_vote_status:
  get_ai_preferences:
  get_alert:
  get_audience:
  get_audience_summary:
  get_audit_event:
  get_aws_cloudtrail_batch:
  get_aws_connection:
  get_bootstrap:
  get_call_route:
  get_change_event:
  get_checklist_template:
  get_comment:
  get_conference_bridge_translation:
  get_configuration_options:
  get_current_user:
  get_environment:
  get_form_configuration:
  get_functionality:
  get_inbound_field_map:
  get_incident:
  get_incident_channel:
  get_incident_event:
  get_incident_relationships:
  get_incident_retrospective_field:
  get_incident_role:
  get_incident_task:
  get_incident_type:
  get_incident_user:
  get_integration:
  get_lifecycle_measurement_definition:
  get_mean_time_report:
  get_member_default_audience:
  get_notification_policy:
  get_nunc_connection:
  get_on_call_schedule_rotation:
  get_on_call_shift:
  get_options_for_field:
  get_post_mortem_question:
  get_post_mortem_report:
  get_priority:
  get_retrospective_template:
  get_role:
  get_runbook:
  get_runbook_action_field_options:
  get_runbook_execution:
  get_runbook_execution_step_script:
  get_saved_search:
  get_scheduled_maintenance:
  get_service:
  get_service_dependencies:
  get_service_dependency:
  get_severity:
  get_severity_matrix:
  get_severity_matrix_condition:
  get_signals_alert_grouping_configuration:
  get_signals_email_target:
  get_signals_event_source:
  get_signals_grouped_metrics:
  get_signals_hacker_mode:
  get_signals_heartbeat_endpoint_configuration:
  get_signals_ingest_url:
  get_signals_mttx_analytics:
  get_signals_noise_analytics:
  get_signals_timeseries_analytics:
  get_signals_webhook_target:
  get_slack_emoji_action:
  get_status_update_template:
  get_statuspage_connection:
  get_support_hours_schedule:
  get_task_list:
  get_team:
  get_team_escalation_policy:
  get_team_on_call_schedule:
  get_team_signal_rule:
  get_ticket:
  get_ticketing_field_map:
  get_ticketing_form_configuration:
  get_ticketing_priority:
  get_ticketing_project:
  get_ticketing_project_config:
  get_user:
  get_vote_status:
  get_webhook:
  get_zendesk_customer_support_issue:
  list_alerts:
  list_audience_summaries:
  list_audiences:
  list_audit_events:
  list_authed_providers:
  list_available_inbound_field_maps:
  list_available_ticketing_field_maps:
  list_aws_cloudtrail_batch_events:
  list_aws_cloudtrail_batches:
  list_aws_connections:
  list_call_routes:
  list_change_events:
  list_change_identities:
  list_change_types:
  list_changes:
  list_checklist_templates:
  list_comment_reactions:
  list_comments:
  list_connection_statuses:
  list_connection_statuses_by_slug:
  list_connection_statuses_by_slug_and_id:
  list_connections:
  list_current_user_permissions:
  list_custom_field_definitions:
  list_custom_field_select_options:
  list_email_subscribers:
  list_entitlements:
  list_environment_functionalities:
  list_environment_services:
  list_field_map_available_fields:
  list_functionality_environments:
  list_functionality_services:
  list_inbound_field_maps:
  list_incident_alerts:
  list_incident_attachments:
  list_incident_change_events:
  list_incident_conference_bridges:
  list_incident_events:
  list_incident_impacts:
  list_incident_links:
  list_incident_metrics:
  list_incident_milestones:
  list_incident_retrospectives:
  list_incident_role_assignments:
  list_incident_roles:
  list_incident_status_pages:
  list_incident_tags:
  list_incident_tasks:
  list_incident_types:
  list_infrastructure_metrics:
  list_infrastructure_type_metrics:
  list_infrastructures:
  list_integrations:
  list_lifecycle_measurement_definitions:
  list_lifecycle_phases:
  list_notification_policy_settings:
  list_nunc_connections:
  list_organization_on_call_schedules:
  list_permissions:
  list_post_mortem_questions:
  list_post_mortem_reasons:
  list_post_mortem_reports:
  list_priorities:
  list_processing_log_entries:
  list_retrospective_metrics:
  list_retrospective_templates:
  list_retrospectives:
  list_roles:
  list_runbook_actions:
  list_runbook_executions:
  list_runbooks:
  list_saved_searches:
  list_scheduled_maintenances:
  list_schedules:
  list_service_available_downstream_dependencies:
  list_service_available_upstream_dependencies:
  list_service_environments:
  list_severities:
  list_severity_matrix_conditions:
  list_severity_matrix_impacts:
  list_signals_alert_grouping_configurations:
  list_signals_email_targets:
  list_signals_event_sources:
  list_signals_heartbeat_endpoint_configurations:
  list_signals_transposers:
  list_signals_webhook_targets:
  list_similar_incidents:
  list_slack_emoji_actions:
  list_slack_usergroups:
  list_slack_workspaces:
  list_status_update_templates:
  list_statuspage_connection_pages:
  list_statuspage_connections:
  list_task_lists:
  list_team_call_routes:
  list_team_escalation_policies:
  list_team_on_call_schedules:
  list_team_permissions:
  list_team_signal_rules:
  list_ticket_tags:
  list_ticketing_custom_definitions:
  list_ticketing_priorities:
  list_ticketing_projects:
  list_tickets:
  list_transcript_entries:
  list_user_involvement_metrics:
  list_user_notification_settings_by_user_id:
  list_user_owned_services:
  list_users:
  list_webhook_deliveries:
  list_webhooks:
  search_confluence_spaces:
  search_slack_channels:
  search_zendesk_tickets:

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  archive_audience:
    endpoint: DELETE /audiences/{{ record.audience_id }}
    required fields: audience_id
    risk: Archive audience; this may remove or archive FireHydrant data.
  bulk_update_incident_milestones:
    endpoint: PUT /incidents/{{ record.incident_id }}/milestones/bulk_update
    required fields: incident_id
    risk: Update milestone times through the FireHydrant API.
  close_incident:
    endpoint: PUT /incidents/{{ record.incident_id }}/close
    required fields: incident_id
    risk: Close an incident through the FireHydrant API.
  convert_incident_task:
    endpoint: POST /incidents/{{ record.incident_id }}/tasks/{{ record.task_id }}/convert
    required fields: task_id, incident_id
    risk: Convert a task to a follow-up through the FireHydrant API.
  copy_on_call_schedule_rotation:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/rotations/{{ record.rotation_id }}/copy
    required fields: rotation_id, team_id, schedule_id
    risk: Copy an on-call schedule's rotation through the FireHydrant API.
  create_audience:
    endpoint: POST /audiences
    risk: Create audience through the FireHydrant API.
  create_change:
    endpoint: POST /changes
    risk: Create a new change entry through the FireHydrant API.
  create_change_event:
    endpoint: POST /changes/events
    risk: Create a change event through the FireHydrant API.
  create_change_identity:
    endpoint: POST /changes/{{ record.change_id }}/identities
    required fields: change_id
    risk: Create an identity for a change entry through the FireHydrant API.
  create_checklist_template:
    endpoint: POST /checklist_templates
    risk: Create a checklist template through the FireHydrant API.
  create_comment:
    endpoint: POST /conversations/{{ record.conversation_id }}/comments
    required fields: conversation_id
    risk: Create a conversation comment through the FireHydrant API.
  create_comment_reaction:
    endpoint: POST /conversations/{{ record.conversation_id }}/comments/{{ record.comment_id }}/reactions
    required fields: conversation_id, comment_id
    risk: Create a reaction for a conversation comment through the FireHydrant API.
  create_connection:
    endpoint: POST /integrations/connections/{{ record.slug }}
    required fields: slug
    risk: Create a new integration connection through the FireHydrant API.
  create_custom_field_definition:
    endpoint: POST /custom_fields/definitions
    risk: Create a custom field definition through the FireHydrant API.
  create_email_subscriber:
    endpoint: POST /nunc_connections/{{ record.nunc_connection_id }}/subscribers
    required fields: nunc_connection_id
    risk: Add subscribers to a status page through the FireHydrant API.
  create_environment:
    endpoint: POST /environments
    risk: Create an environment through the FireHydrant API.
  create_functionality:
    endpoint: POST /functionalities
    risk: Create a functionality through the FireHydrant API.
  create_inbound_field_map:
    endpoint: POST /ticketing/projects/{{ record.ticketing_project_id }}/inbound_field_maps
    required fields: ticketing_project_id
    risk: Create inbound field map for a ticketing project through the FireHydrant API.
  create_incident:
    endpoint: POST /incidents
    risk: Create an incident through the FireHydrant API.
  create_incident_alert:
    endpoint: POST /incidents/{{ record.incident_id }}/alerts
    required fields: incident_id
    risk: Attach an alert to an incident through the FireHydrant API.
  create_incident_change_event:
    endpoint: POST /incidents/{{ record.incident_id }}/related_change_events
    required fields: incident_id
    risk: Add a related change to an incident through the FireHydrant API.
  create_incident_chat_message:
    endpoint: POST /incidents/{{ record.incident_id }}/generic_chat_messages
    required fields: incident_id
    risk: Add a chat message to an incident through the FireHydrant API.
  create_incident_impact:
    endpoint: POST /incidents/{{ record.incident_id }}/impact/{{ record.type }}
    required fields: incident_id, type
    risk: Add impacted infrastructure to an incident through the FireHydrant API.
  create_incident_link:
    endpoint: POST /incidents/{{ record.incident_id }}/links
    required fields: incident_id
    risk: Add a link to an incident through the FireHydrant API.
  create_incident_note:
    endpoint: POST /incidents/{{ record.incident_id }}/notes
    required fields: incident_id
    risk: Add a note to an incident through the FireHydrant API.
  create_incident_retrospective:
    endpoint: POST /incidents/{{ record.incident_id }}/retrospectives
    required fields: incident_id
    risk: Create a new retrospective on the incident using the template through the FireHydrant API.
  create_incident_retrospective_dynamic_input:
    endpoint: POST /incidents/{{ record.incident_id }}/retrospectives/{{ record.retrospective_id }}/fields/{{ record.field_id }}/inputs
    required fields: retrospective_id, field_id, incident_id
    risk: Add a new dynamic input field to a retrospective's dynamic input group field through the FireHydrant API.
  create_incident_retrospective_field:
    endpoint: PATCH /incidents/{{ record.incident_id }}/retrospectives/{{ record.retrospective_id }}/fields
    required fields: retrospective_id, incident_id
    risk: Appends a new incident retrospective field to an incident retrospective through the FireHydrant API.
  create_incident_role:
    endpoint: POST /incident_roles
    risk: Create an incident role through the FireHydrant API.
  create_incident_role_assignment:
    endpoint: POST /incidents/{{ record.incident_id }}/role_assignments
    required fields: incident_id
    risk: Assign a user to an incident through the FireHydrant API.
  create_incident_status_page:
    endpoint: POST /incidents/{{ record.incident_id }}/status_pages
    required fields: incident_id
    risk: Add a status page to an incident through the FireHydrant API.
  create_incident_task:
    endpoint: POST /incidents/{{ record.incident_id }}/tasks
    required fields: incident_id
    risk: Create an incident task through the FireHydrant API.
  create_incident_task_list:
    endpoint: POST /incidents/{{ record.incident_id }}/task_lists
    required fields: incident_id
    risk: Add tasks from a task list to an incident through the FireHydrant API.
  create_incident_team_assignment:
    endpoint: POST /incidents/{{ record.incident_id }}/team_assignments
    required fields: incident_id
    risk: Assign a team to an incident through the FireHydrant API.
  create_incident_type:
    endpoint: POST /incident_types
    risk: Create an incident type through the FireHydrant API.
  create_lifecycle_measurement_definition:
    endpoint: POST /lifecycles/measurement_definitions
    risk: Create a measurement definition through the FireHydrant API.
  create_lifecycle_milestone:
    endpoint: POST /lifecycles/milestones
    risk: Create a milestone through the FireHydrant API.
  create_notification_policy:
    endpoint: POST /signals/notification_policy_items
    risk: Create a notification policy through the FireHydrant API.
  create_nunc_component_group:
    endpoint: POST /nunc_connections/{{ record.nunc_connection_id }}/component_groups
    required fields: nunc_connection_id
    risk: Create a component group for a status page through the FireHydrant API.
  create_nunc_connection:
    endpoint: POST /nunc_connections
    risk: Create a status page through the FireHydrant API.
  create_nunc_link:
    endpoint: POST /nunc_connections/{{ record.nunc_connection_id }}/links
    required fields: nunc_connection_id
    risk: Add link to a status page through the FireHydrant API.
  create_nunc_subscription:
    endpoint: POST /nunc/subscriptions
    risk: Create a status page subscription through the FireHydrant API.
  create_on_call_schedule_rotation:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/rotations
    required fields: team_id, schedule_id
    risk: Create a new on-call rotation through the FireHydrant API.
  create_on_call_shift:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/shifts
    required fields: team_id, schedule_id
    risk: Create a shift for an on-call schedule through the FireHydrant API.
  create_post_mortem_reason:
    endpoint: POST /post_mortems/reports/{{ record.report_id }}/reasons
    required fields: report_id
    risk: Create a contributing factor for a retrospective report through the FireHydrant API.
  create_post_mortem_report:
    endpoint: POST /post_mortems/reports
    risk: Create a retrospective report through the FireHydrant API.
  create_priority:
    endpoint: POST /priorities
    risk: Create a priority through the FireHydrant API.
  create_retrospective_template:
    endpoint: POST /retrospective_templates
    risk: Create a retrospective template through the FireHydrant API.
  create_role:
    endpoint: POST /roles
    risk: Create a role through the FireHydrant API.
  create_runbook:
    endpoint: POST /runbooks
    risk: Create a runbook through the FireHydrant API.
  create_runbook_execution:
    endpoint: POST /runbooks/executions
    risk: Create a runbook execution through the FireHydrant API.
  create_saved_search:
    endpoint: POST /saved_searches/{{ record.resource_type }}
    required fields: resource_type
    risk: Create a saved search through the FireHydrant API.
  create_scheduled_maintenance:
    endpoint: POST /scheduled_maintenances
    risk: Create a scheduled maintenance event through the FireHydrant API.
  create_service:
    endpoint: POST /services
    risk: Create a service through the FireHydrant API.
  create_service_checklist_response:
    endpoint: POST /services/{{ record.service_id }}/checklist_response/{{ record.checklist_id }}
    required fields: service_id, checklist_id
    risk: Record a response for a checklist item through the FireHydrant API.
  create_service_dependency:
    endpoint: POST /service_dependencies
    risk: Create a service dependency through the FireHydrant API.
  create_service_links:
    endpoint: POST /services/service_links
    risk: Create multiple services linked to external services through the FireHydrant API.
  create_severity:
    endpoint: POST /severities
    risk: Create a severity through the FireHydrant API.
  create_severity_matrix_condition:
    endpoint: POST /severity_matrix/conditions
    risk: Create a severity matrix condition through the FireHydrant API.
  create_severity_matrix_impact:
    endpoint: POST /severity_matrix/impacts
    risk: Create a severity matrix impact through the FireHydrant API.
  create_signals_alert_grouping_configuration:
    endpoint: POST /signals/grouping
    risk: Create an alert grouping configuration. through the FireHydrant API.
  create_signals_email_target:
    endpoint: POST /signals/email_targets
    risk: Create an email target for signals through the FireHydrant API.
  create_signals_event_source:
    endpoint: PUT /signals/event_sources
    risk: Create an event source for Signals through the FireHydrant API.
  create_signals_heartbeat_endpoint_configuration:
    endpoint: POST /signals/heartbeat_endpoints
    risk: Create a heartbeat endpoint configuration through the FireHydrant API.
  create_signals_page:
    endpoint: POST /page/signals
    risk: Page a user, team, on-call schedule, or escalation policy through the FireHydrant API.
  create_signals_webhook_target:
    endpoint: POST /signals/webhook_targets
    risk: Create a webhook target through the FireHydrant API.
  create_slack_emoji_action:
    endpoint: POST /integrations/slack/connections/{{ record.connection_id }}/emoji_actions
    required fields: connection_id
    risk: Create a new Slack emoji action through the FireHydrant API.
  create_status_update_template:
    endpoint: POST /status_update_templates
    risk: Create a status update template through the FireHydrant API.
  create_task_list:
    endpoint: POST /task_lists
    risk: Create a task list through the FireHydrant API.
  create_team:
    endpoint: POST /teams
    risk: Create a team through the FireHydrant API.
  create_team_call_route:
    endpoint: POST /teams/{{ record.team_id }}/call_routes
    required fields: team_id
    risk: Create a call route for a team through the FireHydrant API.
  create_team_escalation_policy:
    endpoint: POST /teams/{{ record.team_id }}/escalation_policies
    required fields: team_id
    risk: Create an escalation policy for a team through the FireHydrant API.
  create_team_on_call_schedule:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules
    required fields: team_id
    risk: Create an on-call schedule for a team through the FireHydrant API.
  create_team_signal_rule:
    endpoint: POST /teams/{{ record.team_id }}/signal_rules
    required fields: team_id
    risk: Create a Signals rule through the FireHydrant API.
  create_ticket:
    endpoint: POST /ticketing/tickets
    risk: Create a ticket through the FireHydrant API.
  create_ticketing_custom_definition:
    endpoint: POST /ticketing/custom_fields/definitions
    risk: Create a ticketing custom field through the FireHydrant API.
  create_ticketing_field_map:
    endpoint: POST /ticketing/projects/{{ record.ticketing_project_id }}/field_maps
    required fields: ticketing_project_id
    risk: Create a field mapping for a ticketing project through the FireHydrant API.
  create_ticketing_priority:
    endpoint: POST /ticketing/priorities
    risk: Create a ticketing priority through the FireHydrant API.
  create_ticketing_project_config:
    endpoint: POST /ticketing/projects/{{ record.ticketing_project_id }}/provider_project_configurations
    required fields: ticketing_project_id
    risk: Create a ticketing project configuration through the FireHydrant API.
  create_webhook:
    endpoint: POST /webhooks
    risk: Create a webhook through the FireHydrant API.
  debug_signals_expression:
    endpoint: POST /signals/debugger
    risk: Debug Signals expressions through the FireHydrant API.
  delete_call_route:
    endpoint: DELETE /signals/call_routes/{{ record.id }}
    required fields: id
    risk: Delete a call route; this may remove or archive FireHydrant data.
  delete_change:
    endpoint: DELETE /changes/{{ record.change_id }}
    required fields: change_id
    risk: Archive a change entry; this may remove or archive FireHydrant data.
  delete_change_event:
    endpoint: DELETE /changes/events/{{ record.change_event_id }}
    required fields: change_event_id
    risk: Delete a change event; this may remove or archive FireHydrant data.
  delete_change_identity:
    endpoint: DELETE /changes/{{ record.change_id }}/identities/{{ record.identity_id }}
    required fields: identity_id, change_id
    risk: Delete an identity from a change entry; this may remove or archive FireHydrant data.
  delete_checklist_template:
    endpoint: DELETE /checklist_templates/{{ record.id }}
    required fields: id
    risk: Archive a checklist template; this may remove or archive FireHydrant data.
  delete_comment:
    endpoint: DELETE /conversations/{{ record.conversation_id }}/comments/{{ record.comment_id }}
    required fields: comment_id, conversation_id
    risk: Archive a conversation comment; this may remove or archive FireHydrant data.
  delete_comment_reaction:
    endpoint: DELETE /conversations/{{ record.conversation_id }}/comments/{{ record.comment_id }}/reactions/{{ record.reaction_id }}
    required fields: reaction_id, conversation_id, comment_id
    risk: Delete a reaction from a conversation comment; this may remove or archive FireHydrant data.
  delete_custom_field_definition:
    endpoint: DELETE /custom_fields/definitions/{{ record.field_id }}
    required fields: field_id
    risk: Delete a custom field definition; this may remove or archive FireHydrant data.
  delete_environment:
    endpoint: DELETE /environments/{{ record.environment_id }}
    required fields: environment_id
    risk: Archive an environment; this may remove or archive FireHydrant data.
  delete_functionality:
    endpoint: DELETE /functionalities/{{ record.functionality_id }}
    required fields: functionality_id
    risk: Archive a functionality; this may remove or archive FireHydrant data.
  delete_inbound_field_map:
    endpoint: DELETE /ticketing/projects/{{ record.ticketing_project_id }}/inbound_field_maps/{{ record.map_id }}
    required fields: map_id, ticketing_project_id
    risk: Archive inbound field map for a ticketing project; this may remove or archive FireHydrant data.
  delete_incident:
    endpoint: DELETE /incidents/{{ record.incident_id }}
    required fields: incident_id
    risk: Archive an incident; this may remove or archive FireHydrant data.
  delete_incident_alert:
    endpoint: DELETE /incidents/{{ record.incident_id }}/alerts/{{ record.incident_alert_id }}
    required fields: incident_alert_id, incident_id
    risk: Remove an alert from an incident; this may remove or archive FireHydrant data.
  delete_incident_chat_message:
    endpoint: DELETE /incidents/{{ record.incident_id }}/generic_chat_messages/{{ record.message_id }}
    required fields: message_id, incident_id
    risk: Delete a chat message from an incident; this may remove or archive FireHydrant data.
  delete_incident_event:
    endpoint: DELETE /incidents/{{ record.incident_id }}/events/{{ record.event_id }}
    required fields: incident_id, event_id
    risk: Delete an incident event; this may remove or archive FireHydrant data.
  delete_incident_impact:
    endpoint: DELETE /incidents/{{ record.incident_id }}/impact/{{ record.type }}/{{ record.id }}
    required fields: incident_id, type, id
    risk: Remove impacted infrastructure from an incident; this may remove or archive FireHydrant data.
  delete_incident_link:
    endpoint: DELETE /incidents/{{ record.incident_id }}/links/{{ record.link_id }}
    required fields: link_id, incident_id
    risk: Remove a link from an incident; this may remove or archive FireHydrant data.
  delete_incident_role:
    endpoint: DELETE /incident_roles/{{ record.incident_role_id }}
    required fields: incident_role_id
    risk: Archive an incident role; this may remove or archive FireHydrant data.
  delete_incident_role_assignment:
    endpoint: DELETE /incidents/{{ record.incident_id }}/role_assignments/{{ record.role_assignment_id }}
    required fields: incident_id, role_assignment_id
    risk: Unassign a user from an incident; this may remove or archive FireHydrant data.
  delete_incident_status_page:
    endpoint: DELETE /incidents/{{ record.incident_id }}/status_pages/{{ record.status_page_id }}
    required fields: status_page_id, incident_id
    risk: Remove a status page from an incident; this may remove or archive FireHydrant data.
  delete_incident_task:
    endpoint: DELETE /incidents/{{ record.incident_id }}/tasks/{{ record.task_id }}
    required fields: task_id, incident_id
    risk: Delete an incident task; this may remove or archive FireHydrant data.
  delete_incident_type:
    endpoint: DELETE /incident_types/{{ record.id }}
    required fields: id
    risk: Archive an incident type; this may remove or archive FireHydrant data.
  delete_lifecycle_measurement_definition:
    endpoint: DELETE /lifecycles/measurement_definitions/{{ record.measurement_definition_id }}
    required fields: measurement_definition_id
    risk: Archive a measurement definition; this may remove or archive FireHydrant data.
  delete_lifecycle_milestone:
    endpoint: DELETE /lifecycles/milestones/{{ record.milestone_id }}
    required fields: milestone_id
    risk: Delete a milestone; this may remove or archive FireHydrant data.
  delete_notification_policy:
    endpoint: DELETE /signals/notification_policy_items/{{ record.id }}
    required fields: id
    risk: Delete a notification policy; this may remove or archive FireHydrant data.
  delete_nunc_component_group:
    endpoint: DELETE /nunc_connections/{{ record.nunc_connection_id }}/component_groups/{{ record.group_id }}
    required fields: nunc_connection_id, group_id
    risk: Delete a status page component group; this may remove or archive FireHydrant data.
  delete_nunc_connection:
    endpoint: DELETE /nunc_connections/{{ record.nunc_connection_id }}
    required fields: nunc_connection_id
    risk: Delete a status page; this may remove or archive FireHydrant data.
  delete_nunc_image:
    endpoint: DELETE /nunc_connections/{{ record.nunc_connection_id }}/images/{{ record.type }}
    required fields: nunc_connection_id, type
    risk: Delete an image from a status page; this may remove or archive FireHydrant data.
  delete_nunc_link:
    endpoint: DELETE /nunc_connections/{{ record.nunc_connection_id }}/links/{{ record.link_id }}
    required fields: nunc_connection_id, link_id
    risk: Delete a status page link; this may remove or archive FireHydrant data.
  delete_nunc_subscription:
    endpoint: DELETE /nunc/subscriptions/{{ record.unsubscribe_token }}
    required fields: unsubscribe_token
    risk: Unsubscribe from status page notifications; this may remove or archive FireHydrant data.
  delete_on_call_schedule_rotation:
    endpoint: DELETE /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/rotations/{{ record.rotation_id }}
    required fields: rotation_id, team_id, schedule_id
    risk: Delete an on-call schedule's rotation; this may remove or archive FireHydrant data.
  delete_on_call_shift:
    endpoint: DELETE /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/shifts/{{ record.id }}
    required fields: id, team_id, schedule_id
    risk: Delete an on-call shift from a team schedule; this may remove or archive FireHydrant data.
  delete_post_mortem_reason:
    endpoint: DELETE /post_mortems/reports/{{ record.report_id }}/reasons/{{ record.reason_id }}
    required fields: report_id, reason_id
    risk: Delete a contributing factor from a retrospective report; this may remove or archive FireHydrant data.
  delete_priority:
    endpoint: DELETE /priorities/{{ record.priority_slug }}
    required fields: priority_slug
    risk: Delete a priority; this may remove or archive FireHydrant data.
  delete_retrospective_template:
    endpoint: DELETE /retrospective_templates/{{ record.retrospective_template_id }}
    required fields: retrospective_template_id
    risk: Delete a retrospective template; this may remove or archive FireHydrant data.
  delete_role:
    endpoint: DELETE /roles/{{ record.id }}
    required fields: id
    risk: Delete a role; this may remove or archive FireHydrant data.
  delete_runbook:
    endpoint: DELETE /runbooks/{{ record.runbook_id }}
    required fields: runbook_id
    risk: Delete a runbook; this may remove or archive FireHydrant data.
  delete_saved_search:
    endpoint: DELETE /saved_searches/{{ record.resource_type }}/{{ record.saved_search_id }}
    required fields: resource_type, saved_search_id
    risk: Delete a saved search; this may remove or archive FireHydrant data.
  delete_scheduled_maintenance:
    endpoint: DELETE /scheduled_maintenances/{{ record.scheduled_maintenance_id }}
    required fields: scheduled_maintenance_id
    risk: Delete a scheduled maintenance event; this may remove or archive FireHydrant data.
  delete_service:
    endpoint: DELETE /services/{{ record.service_id }}
    required fields: service_id
    risk: Delete a service; this may remove or archive FireHydrant data.
  delete_service_dependency:
    endpoint: DELETE /service_dependencies/{{ record.service_dependency_id }}
    required fields: service_dependency_id
    risk: Delete a service dependency; this may remove or archive FireHydrant data.
  delete_service_link:
    endpoint: DELETE /services/{{ record.service_id }}/service_links/{{ record.remote_id }}
    required fields: service_id, remote_id
    risk: Delete a service link; this may remove or archive FireHydrant data.
  delete_severity:
    endpoint: DELETE /severities/{{ record.severity_slug }}
    required fields: severity_slug
    risk: Delete a severity; this may remove or archive FireHydrant data.
  delete_severity_matrix_condition:
    endpoint: DELETE /severity_matrix/conditions/{{ record.condition_id }}
    required fields: condition_id
    risk: Delete a severity matrix condition; this may remove or archive FireHydrant data.
  delete_severity_matrix_impact:
    endpoint: DELETE /severity_matrix/impacts/{{ record.impact_id }}
    required fields: impact_id
    risk: Delete a severity matrix impact; this may remove or archive FireHydrant data.
  delete_signals_alert_grouping_configuration:
    endpoint: DELETE /signals/grouping/{{ record.id }}
    required fields: id
    risk: Delete an alert grouping configuration.; this may remove or archive FireHydrant data.
  delete_signals_email_target:
    endpoint: DELETE /signals/email_targets/{{ record.id }}
    required fields: id
    risk: Delete a signal email target; this may remove or archive FireHydrant data.
  delete_signals_event_source:
    endpoint: DELETE /signals/event_sources/{{ record.transposer_slug }}
    required fields: transposer_slug
    risk: Delete an event source for Signals; this may remove or archive FireHydrant data.
  delete_signals_heartbeat_endpoint_configuration:
    endpoint: DELETE /signals/heartbeat_endpoints/{{ record.id }}
    required fields: id
    risk: Delete a heartbeat endpoint configuration; this may remove or archive FireHydrant data.
  delete_signals_webhook_target:
    endpoint: DELETE /signals/webhook_targets/{{ record.id }}
    required fields: id
    risk: Delete a webhook target; this may remove or archive FireHydrant data.
  delete_slack_emoji_action:
    endpoint: DELETE /integrations/slack/connections/{{ record.connection_id }}/emoji_actions/{{ record.emoji_action_id }}
    required fields: connection_id, emoji_action_id
    risk: Delete a Slack emoji action; this may remove or archive FireHydrant data.
  delete_status_update_template:
    endpoint: DELETE /status_update_templates/{{ record.status_update_template_id }}
    required fields: status_update_template_id
    risk: Delete a status update template; this may remove or archive FireHydrant data.
  delete_statuspage_connection:
    endpoint: DELETE /integrations/statuspage/connections/{{ record.connection_id }}
    required fields: connection_id
    risk: Delete a Statuspage connection; this may remove or archive FireHydrant data.
  delete_support_hours_schedule:
    endpoint: DELETE /teams/{{ record.team_id }}/support_hours_schedule
    required fields: team_id
    risk: Delete a specific support hours schedule; this may remove or archive FireHydrant data.
  delete_task_list:
    endpoint: DELETE /task_lists/{{ record.task_list_id }}
    required fields: task_list_id
    risk: Delete a task list; this may remove or archive FireHydrant data.
  delete_team:
    endpoint: DELETE /teams/{{ record.team_id }}
    required fields: team_id
    risk: Archive a team; this may remove or archive FireHydrant data.
  delete_team_escalation_policy:
    endpoint: DELETE /teams/{{ record.team_id }}/escalation_policies/{{ record.id }}
    required fields: team_id, id
    risk: Delete an escalation policy for a team; this may remove or archive FireHydrant data.
  delete_team_on_call_schedule:
    endpoint: DELETE /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}
    required fields: team_id, schedule_id
    risk: Delete an on-call schedule for a team; this may remove or archive FireHydrant data.
  delete_team_signal_rule:
    endpoint: DELETE /teams/{{ record.team_id }}/signal_rules/{{ record.id }}
    required fields: team_id, id
    risk: Delete a Signals rule; this may remove or archive FireHydrant data.
  delete_ticket:
    endpoint: DELETE /ticketing/tickets/{{ record.ticket_id }}
    required fields: ticket_id
    risk: Archive a ticket; this may remove or archive FireHydrant data.
  delete_ticketing_custom_definition:
    endpoint: DELETE /ticketing/custom_fields/definitions/{{ record.field_id }}
    required fields: field_id
    risk: Delete a ticketing custom field; this may remove or archive FireHydrant data.
  delete_ticketing_field_map:
    endpoint: DELETE /ticketing/projects/{{ record.ticketing_project_id }}/field_maps/{{ record.map_id }}
    required fields: map_id, ticketing_project_id
    risk: Archive a field map for a ticketing project; this may remove or archive FireHydrant data.
  delete_ticketing_priority:
    endpoint: DELETE /ticketing/priorities/{{ record.id }}
    required fields: id
    risk: Delete a ticketing priority; this may remove or archive FireHydrant data.
  delete_ticketing_project_config:
    endpoint: DELETE /ticketing/projects/{{ record.ticketing_project_id }}/provider_project_configurations/{{ record.config_id }}
    required fields: ticketing_project_id, config_id
    risk: Archive a ticketing project configuration; this may remove or archive FireHydrant data.
  delete_transcript_entry:
    endpoint: DELETE /incidents/{{ record.incident_id }}/transcript/{{ record.transcript_id }}
    required fields: transcript_id, incident_id
    risk: Delete a transcript from an incident; this may remove or archive FireHydrant data.
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: Delete a webhook; this may remove or archive FireHydrant data.
  generate_audience_summary:
    endpoint: POST /audiences/{{ record.audience_id }}/summaries/{{ record.incident_id }}
    required fields: audience_id, incident_id
    risk: Generate summary (async) through the FireHydrant API.
  ingest_catalog_data:
    endpoint: POST /catalogs/{{ record.catalog_id }}/ingest
    required fields: catalog_id
    risk: Ingest service catalog data through the FireHydrant API.
  override_on_call_schedule_rotation_shifts:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/rotations/{{ record.rotation_id }}/overrides
    required fields: rotation_id, team_id, schedule_id
    risk: Override one or more shifts in an on-call rotation through the FireHydrant API.
  preview_on_call_schedule_rotation:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/rotations/preview
    required fields: team_id, schedule_id
    risk: Preview an on-call rotation through the FireHydrant API.
  preview_team_on_call_schedule:
    endpoint: POST /teams/{{ record.team_id }}/on_call_schedules/preview
    required fields: team_id
    risk: Preview a new on-call schedule for a team through the FireHydrant API.
  publish_nunc_connection:
    endpoint: POST /nunc_connections/{{ record.nunc_connection_id }}/publish
    required fields: nunc_connection_id
    risk: Publish a status page through the FireHydrant API.
  publish_post_mortem_report:
    endpoint: POST /post_mortems/reports/{{ record.report_id }}/publish
    required fields: report_id
    risk: Publish a retrospective report through the FireHydrant API.
  refresh_connection:
    endpoint: PATCH /integrations/connections/{{ record.slug }}/{{ record.connection_id }}/refresh
    required fields: slug, connection_id
    risk: Refresh an integration connection's incident role schedules through the FireHydrant API.
  reorder_post_mortem_reasons:
    endpoint: PUT /post_mortems/reports/{{ record.report_id }}/reasons/order
    required fields: report_id
    risk: Reorder a contributing factor for a retrospective report through the FireHydrant API.
  resolve_incident:
    endpoint: PUT /incidents/{{ record.incident_id }}/resolve
    required fields: incident_id
    risk: Resolve an incident through the FireHydrant API.
  restore_audience:
    endpoint: PATCH /audiences/{{ record.audience_id }}/restore
    required fields: audience_id
    risk: Restore audience through the FireHydrant API.
  set_member_default_audience:
    endpoint: PUT /audiences/member/{{ record.member_id }}/default
    required fields: member_id
    risk: Set default audience through the FireHydrant API.
  share_incident_retrospectives:
    endpoint: POST /incidents/{{ record.incident_id }}/retrospectives/share
    required fields: incident_id
    risk: Share an incident's retrospective through the FireHydrant API.
  test_slack_channel:
    endpoint: PUT /integrations/slack/channels/{{ record.id }}/test
    required fields: id
    risk: Test a Slack channel through the FireHydrant API.
  unarchive_incident:
    endpoint: POST /incidents/{{ record.incident_id }}/unarchive
    required fields: incident_id
    risk: Unarchive an incident through the FireHydrant API.
  unpublish_nunc_connection:
    endpoint: POST /nunc_connections/{{ record.nunc_connection_id }}/unpublish
    required fields: nunc_connection_id
    risk: Unpublish a status page through the FireHydrant API.
  update_ai_preferences:
    endpoint: PATCH /ai/preferences
    risk: Update AI preferences through the FireHydrant API.
  update_audience:
    endpoint: PATCH /audiences/{{ record.audience_id }}
    required fields: audience_id
    risk: Update audience through the FireHydrant API.
  update_authed_provider:
    endpoint: PATCH /integrations/authed_providers/{{ record.integration_slug }}/{{ record.connection_id }}/{{ record.authed_provider_id }}
    required fields: integration_slug, connection_id, authed_provider_id
    risk: Get an authed provider through the FireHydrant API.
  update_aws_cloudtrail_batch:
    endpoint: PATCH /integrations/aws/cloudtrail_batches/{{ record.id }}
    required fields: id
    risk: Update a CloudTrail batch through the FireHydrant API.
  update_aws_connection:
    endpoint: PATCH /integrations/aws/connections/{{ record.id }}
    required fields: id
    risk: Update an AWS connection through the FireHydrant API.
  update_call_route:
    endpoint: PATCH /signals/call_routes/{{ record.id }}
    required fields: id
    risk: Update a call route through the FireHydrant API.
  update_change:
    endpoint: PATCH /changes/{{ record.change_id }}
    required fields: change_id
    risk: Update a change entry through the FireHydrant API.
  update_change_event:
    endpoint: PATCH /changes/events/{{ record.change_event_id }}
    required fields: change_event_id
    risk: Update a change event through the FireHydrant API.
  update_change_identity:
    endpoint: PATCH /changes/{{ record.change_id }}/identities/{{ record.identity_id }}
    required fields: identity_id, change_id
    risk: Update an identity for a change entry through the FireHydrant API.
  update_checklist_template:
    endpoint: PATCH /checklist_templates/{{ record.id }}
    required fields: id
    risk: Update a checklist template through the FireHydrant API.
  update_comment:
    endpoint: PATCH /conversations/{{ record.conversation_id }}/comments/{{ record.comment_id }}
    required fields: comment_id, conversation_id
    risk: Update a conversation comment through the FireHydrant API.
  update_connection:
    endpoint: PATCH /integrations/connections/{{ record.slug }}/{{ record.connection_id }}
    required fields: slug, connection_id
    risk: Update an integration connection through the FireHydrant API.
  update_custom_field_definition:
    endpoint: PATCH /custom_fields/definitions/{{ record.field_id }}
    required fields: field_id
    risk: Update a custom field definition through the FireHydrant API.
  update_environment:
    endpoint: PATCH /environments/{{ record.environment_id }}
    required fields: environment_id
    risk: Update an environment through the FireHydrant API.
  update_field_map:
    endpoint: PATCH /integrations/field_maps/{{ record.field_map_id }}
    required fields: field_map_id
    risk: Update field mapping through the FireHydrant API.
  update_functionality:
    endpoint: PATCH /functionalities/{{ record.functionality_id }}
    required fields: functionality_id
    risk: Update a functionality through the FireHydrant API.
  update_inbound_field_map:
    endpoint: PUT /ticketing/projects/{{ record.ticketing_project_id }}/inbound_field_maps/{{ record.map_id }}
    required fields: map_id, ticketing_project_id
    risk: Update inbound field map for a ticketing project through the FireHydrant API.
  update_incident:
    endpoint: PATCH /incidents/{{ record.incident_id }}
    required fields: incident_id
    risk: Update an incident through the FireHydrant API.
  update_incident_alert_primary:
    endpoint: PATCH /incidents/{{ record.incident_id }}/alerts/{{ record.incident_alert_id }}/primary
    required fields: incident_alert_id, incident_id
    risk: Set an alert as primary for an incident through the FireHydrant API.
  update_incident_change_event:
    endpoint: PATCH /incidents/{{ record.incident_id }}/related_change_events/{{ record.related_change_event_id }}
    required fields: related_change_event_id, incident_id
    risk: Update a change attached to an incident through the FireHydrant API.
  update_incident_chat_message:
    endpoint: PATCH /incidents/{{ record.incident_id }}/generic_chat_messages/{{ record.message_id }}
    required fields: message_id, incident_id
    risk: Update a chat message on an incident through the FireHydrant API.
  update_incident_event:
    endpoint: PATCH /incidents/{{ record.incident_id }}/events/{{ record.event_id }}
    required fields: incident_id, event_id
    risk: Update an incident event through the FireHydrant API.
  update_incident_impact_patch:
    endpoint: PATCH /incidents/{{ record.incident_id }}/impact
    required fields: incident_id
    risk: Update impacts for an incident through the FireHydrant API.
  update_incident_impact_put:
    endpoint: PUT /incidents/{{ record.incident_id }}/impact
    required fields: incident_id
    risk: Update impacts for an incident through the FireHydrant API.
  update_incident_link:
    endpoint: PUT /incidents/{{ record.incident_id }}/links/{{ record.link_id }}
    required fields: link_id, incident_id
    risk: Update the external incident link through the FireHydrant API.
  update_incident_note:
    endpoint: PATCH /incidents/{{ record.incident_id }}/notes/{{ record.note_id }}
    required fields: note_id, incident_id
    risk: Update a note through the FireHydrant API.
  update_incident_retrospective:
    endpoint: PATCH /incidents/{{ record.incident_id }}/retrospectives/{{ record.retrospective_id }}
    required fields: retrospective_id, incident_id
    risk: Update a retrospective on the incident through the FireHydrant API.
  update_incident_retrospective_field:
    endpoint: PATCH /incidents/{{ record.incident_id }}/retrospectives/{{ record.retrospective_id }}/fields/{{ record.field_id }}
    required fields: retrospective_id, field_id, incident_id
    risk: Update the value on a retrospective field through the FireHydrant API.
  update_incident_role:
    endpoint: PATCH /incident_roles/{{ record.incident_role_id }}
    required fields: incident_role_id
    risk: Update an incident role through the FireHydrant API.
  update_incident_task:
    endpoint: PATCH /incidents/{{ record.incident_id }}/tasks/{{ record.task_id }}
    required fields: task_id, incident_id
    risk: Update an incident task through the FireHydrant API.
  update_incident_type:
    endpoint: PATCH /incident_types/{{ record.id }}
    required fields: id
    risk: Update an incident type through the FireHydrant API.
  update_lifecycle_measurement_definition:
    endpoint: PATCH /lifecycles/measurement_definitions/{{ record.measurement_definition_id }}
    required fields: measurement_definition_id
    risk: Update a measurement definition through the FireHydrant API.
  update_lifecycle_milestone:
    endpoint: PATCH /lifecycles/milestones/{{ record.milestone_id }}
    required fields: milestone_id
    risk: Update a milestone through the FireHydrant API.
  update_notification_policy:
    endpoint: PATCH /signals/notification_policy_items/{{ record.id }}
    required fields: id
    risk: Update a notification policy through the FireHydrant API.
  update_nunc_component_group:
    endpoint: PATCH /nunc_connections/{{ record.nunc_connection_id }}/component_groups/{{ record.group_id }}
    required fields: nunc_connection_id, group_id
    risk: Update a status page component group through the FireHydrant API.
  update_nunc_connection:
    endpoint: PUT /nunc_connections/{{ record.nunc_connection_id }}
    required fields: nunc_connection_id
    risk: Update a status page through the FireHydrant API.
  update_nunc_link:
    endpoint: PATCH /nunc_connections/{{ record.nunc_connection_id }}/links/{{ record.link_id }}
    required fields: nunc_connection_id, link_id
    risk: Update a status page link through the FireHydrant API.
  update_on_call_schedule_rotation:
    endpoint: PATCH /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/rotations/{{ record.rotation_id }}
    required fields: rotation_id, team_id, schedule_id
    risk: Update an on-call schedule's rotation through the FireHydrant API.
  update_on_call_shift:
    endpoint: PATCH /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}/shifts/{{ record.id }}
    required fields: id, team_id, schedule_id
    risk: Update an on-call shift for a team schedule through the FireHydrant API.
  update_post_mortem_field:
    endpoint: PATCH /post_mortems/reports/{{ record.report_id }}/fields/{{ record.field_id }}
    required fields: field_id, report_id
    risk: Update a retrospective field through the FireHydrant API.
  update_post_mortem_questions:
    endpoint: PUT /post_mortems/questions
    risk: Update retrospective questions through the FireHydrant API.
  update_post_mortem_reason:
    endpoint: PATCH /post_mortems/reports/{{ record.report_id }}/reasons/{{ record.reason_id }}
    required fields: report_id, reason_id
    risk: Update a contributing factor in a retrospective report through the FireHydrant API.
  update_post_mortem_report:
    endpoint: PATCH /post_mortems/reports/{{ record.report_id }}
    required fields: report_id
    risk: Update a retrospective report through the FireHydrant API.
  update_priority:
    endpoint: PATCH /priorities/{{ record.priority_slug }}
    required fields: priority_slug
    risk: Update a priority through the FireHydrant API.
  update_retrospective_template:
    endpoint: PATCH /retrospective_templates/{{ record.retrospective_template_id }}
    required fields: retrospective_template_id
    risk: Update a retrospective template through the FireHydrant API.
  update_role:
    endpoint: PATCH /roles/{{ record.id }}
    required fields: id
    risk: Update a role through the FireHydrant API.
  update_runbook:
    endpoint: PUT /runbooks/{{ record.runbook_id }}
    required fields: runbook_id
    risk: Update a runbook through the FireHydrant API.
  update_runbook_execution_step:
    endpoint: PUT /runbooks/executions/{{ record.execution_id }}/steps/{{ record.step_id }}
    required fields: execution_id, step_id
    risk: Update a runbook step execution through the FireHydrant API.
  update_runbook_execution_step_script:
    endpoint: PUT /runbooks/executions/{{ record.execution_id }}/steps/{{ record.step_id }}/script/{{ record.state }}
    required fields: execution_id, step_id, state
    risk: Update a script step's execution status through the FireHydrant API.
  update_saved_search:
    endpoint: PATCH /saved_searches/{{ record.resource_type }}/{{ record.saved_search_id }}
    required fields: resource_type, saved_search_id
    risk: Update a saved search through the FireHydrant API.
  update_scheduled_maintenance:
    endpoint: PATCH /scheduled_maintenances/{{ record.scheduled_maintenance_id }}
    required fields: scheduled_maintenance_id
    risk: Update a scheduled maintenance event through the FireHydrant API.
  update_service:
    endpoint: PATCH /services/{{ record.service_id }}
    required fields: service_id
    risk: Update a service through the FireHydrant API.
  update_service_dependency:
    endpoint: PATCH /service_dependencies/{{ record.service_dependency_id }}
    required fields: service_dependency_id
    risk: Update a service dependency through the FireHydrant API.
  update_severity:
    endpoint: PATCH /severities/{{ record.severity_slug }}
    required fields: severity_slug
    risk: Update a severity through the FireHydrant API.
  update_severity_matrix:
    endpoint: PATCH /severity_matrix
    risk: Update severity matrix through the FireHydrant API.
  update_severity_matrix_condition:
    endpoint: PATCH /severity_matrix/conditions/{{ record.condition_id }}
    required fields: condition_id
    risk: Update a severity matrix condition through the FireHydrant API.
  update_severity_matrix_impact:
    endpoint: PATCH /severity_matrix/impacts/{{ record.impact_id }}
    required fields: impact_id
    risk: Update a severity matrix impact through the FireHydrant API.
  update_signals_alert:
    endpoint: PATCH /signals/alerts/{{ record.id }}
    required fields: id
    risk: Update a Signal alert through the FireHydrant API.
  update_signals_alert_grouping_configuration:
    endpoint: PATCH /signals/grouping/{{ record.id }}
    required fields: id
    risk: Update an alert grouping configuration. through the FireHydrant API.
  update_signals_email_target:
    endpoint: PATCH /signals/email_targets/{{ record.id }}
    required fields: id
    risk: Update an email target through the FireHydrant API.
  update_signals_heartbeat_endpoint_configuration:
    endpoint: PATCH /signals/heartbeat_endpoints/{{ record.id }}
    required fields: id
    risk: Update a heartbeat endpoint configuration through the FireHydrant API.
  update_signals_webhook_target:
    endpoint: PATCH /signals/webhook_targets/{{ record.id }}
    required fields: id
    risk: Update a webhook target through the FireHydrant API.
  update_slack_emoji_action:
    endpoint: PATCH /integrations/slack/connections/{{ record.connection_id }}/emoji_actions/{{ record.emoji_action_id }}
    required fields: connection_id, emoji_action_id
    risk: Update a Slack emoji action through the FireHydrant API.
  update_status_update_template:
    endpoint: PATCH /status_update_templates/{{ record.status_update_template_id }}
    required fields: status_update_template_id
    risk: Update a status update template through the FireHydrant API.
  update_statuspage_connection:
    endpoint: PATCH /integrations/statuspage/connections/{{ record.connection_id }}
    required fields: connection_id
    risk: Update a Statuspage connection through the FireHydrant API.
  update_support_hours_schedule:
    endpoint: PATCH /teams/{{ record.team_id }}/support_hours_schedule
    required fields: team_id
    risk: Update support hours schedule through the FireHydrant API.
  update_task_list:
    endpoint: PATCH /task_lists/{{ record.task_list_id }}
    required fields: task_list_id
    risk: Update a task list through the FireHydrant API.
  update_team:
    endpoint: PATCH /teams/{{ record.team_id }}
    required fields: team_id
    risk: Update a team through the FireHydrant API.
  update_team_escalation_policy:
    endpoint: PATCH /teams/{{ record.team_id }}/escalation_policies/{{ record.id }}
    required fields: team_id, id
    risk: Update an escalation policy for a team through the FireHydrant API.
  update_team_on_call_schedule:
    endpoint: PATCH /teams/{{ record.team_id }}/on_call_schedules/{{ record.schedule_id }}
    required fields: team_id, schedule_id
    risk: Update an on-call schedule for a team through the FireHydrant API.
  update_team_signal_rule:
    endpoint: PATCH /teams/{{ record.team_id }}/signal_rules/{{ record.id }}
    required fields: team_id, id
    risk: Update a Signals rule through the FireHydrant API.
  update_ticket:
    endpoint: PATCH /ticketing/tickets/{{ record.ticket_id }}
    required fields: ticket_id
    risk: Update a ticket through the FireHydrant API.
  update_ticketing_custom_definition:
    endpoint: PATCH /ticketing/custom_fields/definitions/{{ record.field_id }}
    required fields: field_id
    risk: Update a ticketing custom field through the FireHydrant API.
  update_ticketing_field_map:
    endpoint: PATCH /ticketing/projects/{{ record.ticketing_project_id }}/field_maps/{{ record.map_id }}
    required fields: map_id, ticketing_project_id
    risk: Update a field map for a ticketing project through the FireHydrant API.
  update_ticketing_priority:
    endpoint: PATCH /ticketing/priorities/{{ record.id }}
    required fields: id
    risk: Update a ticketing priority through the FireHydrant API.
  update_ticketing_project_config:
    endpoint: PATCH /ticketing/projects/{{ record.ticketing_project_id }}/provider_project_configurations/{{ record.config_id }}
    required fields: ticketing_project_id, config_id
    risk: Update configuration for a ticketing project through the FireHydrant API.
  update_transcript_attribution:
    endpoint: PUT /incidents/{{ record.incident_id }}/transcript/attribution
    required fields: incident_id
    risk: Update the attribution of a transcript through the FireHydrant API.
  update_vote:
    endpoint: PATCH /incidents/{{ record.incident_id }}/events/{{ record.event_id }}/votes
    required fields: incident_id, event_id
    risk: Update votes through the FireHydrant API.
  update_webhook:
    endpoint: PATCH /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: Update a webhook through the FireHydrant API.
  validate_incident_tags:
    endpoint: POST /incident_tags/validate
    risk: Validate incident tags through the FireHydrant API.
  vote_ai_incident_summary:
    endpoint: PUT /ai/summarize_incident/{{ record.incident_id }}/{{ record.generated_summary_id }}/vote
    required fields: incident_id, generated_summary_id
    risk: Vote on an AI-generated incident summary through the FireHydrant API.

SECURITY
  read risk: external FireHydrant API reads across incidents, catalog, alerts, teams, runbooks, metrics, integrations, and configuration resources
  write risk: creates, updates, archives, deletes, triggers, and otherwise mutates FireHydrant resources through documented JSON/no-body REST endpoints
  approval: reverse ETL writes require plan preview and approval token before execution
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run FireHydrant's declared streams and reverse-ETL actions.
  Usage: pm firehydrant <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v1 incidents incident-id retrospectives retrospective-id fields field-id inputs - Documented DELETE /v1/incidents/{incident_id}/retrospectives/{retrospective_id}/fields/{field_id}/inputs (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.delete.v1-incidents-incident-id-retrospectives-retrospective-id-fields-field-id-inputs]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 incidents incident-id team-assignments team-assignment-id - Documented DELETE /v1/incidents/{incident_id}/team_assignments/{team_assignment_id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.delete.v1-incidents-incident-id-team-assignments-team-assignment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 nunc-connections nunc-connection-id subscribers - Documented DELETE /v1/nunc_connections/{nunc_connection_id}/subscribers (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.delete.v1-nunc-connections-nunc-connection-id-subscribers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 runbooks executions execution-id - Documented DELETE /v1/runbooks/executions/{execution_id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.delete.v1-runbooks-executions-execution-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 scim v2 groups id - Documented DELETE /v1/scim/v2/Groups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.delete.v1-scim-v2-groups-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 scim v2 users id - Documented DELETE /v1/scim/v2/Users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.delete.v1-scim-v2-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v1 catalogs catalog-id refresh - Documented GET /v1/catalogs/{catalog_id}/refresh (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-catalogs-catalog-id-refresh]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics milestone-funnel - Documented GET /v1/metrics/milestone_funnel (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-metrics-milestone-funnel]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics mttx - Documented GET /v1/metrics/mttx (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-metrics-mttx]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics ticket-funnel - Documented GET /v1/metrics/ticket_funnel (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-metrics-ticket-funnel]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 noauth ping - Documented GET /v1/noauth/ping (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-noauth-ping]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 ping - Documented GET /v1/ping (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-ping]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 runbook-audits - Documented GET /v1/runbook_audits (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-runbook-audits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 scim v2 groups - Documented GET /v1/scim/v2/Groups (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-scim-v2-groups]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 scim v2 groups id - Documented GET /v1/scim/v2/Groups/{id} (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-scim-v2-groups-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 scim v2 users - Documented GET /v1/scim/v2/Users (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-scim-v2-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 scim v2 users id - Documented GET /v1/scim/v2/Users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-scim-v2-users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 signals analytics shifts export - Documented GET /v1/signals/analytics/shifts/export (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-signals-analytics-shifts-export]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 signals heartbeat-endpoints addresses - Documented GET /v1/signals/heartbeat_endpoints/addresses (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-signals-heartbeat-endpoints-addresses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 signals heartbeat-endpoints statuses - Documented GET /v1/signals/heartbeat_endpoints/statuses (not implemented) [intent=direct_read availability=not_implemented operation=firehydrant.get.v1-signals-heartbeat-endpoints-statuses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch v1 scim v2 groups id - Documented PATCH /v1/scim/v2/Groups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.patch.v1-scim-v2-groups-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 scim v2 users id - Documented PATCH /v1/scim/v2/Users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.patch.v1-scim-v2-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 incidents incident-id attachments - Documented POST /v1/incidents/{incident_id}/attachments (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.post.v1-incidents-incident-id-attachments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 incidents incident-id retrospectives export - Documented POST /v1/incidents/{incident_id}/retrospectives/export (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.post.v1-incidents-incident-id-retrospectives-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 incidents incident-id retrospectives export-markdown - Documented POST /v1/incidents/{incident_id}/retrospectives/export_markdown (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.post.v1-incidents-incident-id-retrospectives-export-markdown]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 scim v2 groups - Documented POST /v1/scim/v2/Groups (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.post.v1-scim-v2-groups]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 scim v2 users - Documented POST /v1/scim/v2/Users (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.post.v1-scim-v2-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 nunc-connections nunc-connection-id images type - Documented PUT /v1/nunc_connections/{nunc_connection_id}/images/{type} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.put.v1-nunc-connections-nunc-connection-id-images-type]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 scim v2 groups id - Documented PUT /v1/scim/v2/Groups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.put.v1-scim-v2-groups-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 scim v2 users id - Documented PUT /v1/scim/v2/Users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=firehydrant.put.v1-scim-v2-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    append form data on selected value get list - Run the append form data on selected value get ETL stream [intent=etl availability=implemented stream=append_form_data_on_selected_value_get]
    archive audience apply - Plan and execute the archive audience reverse-ETL action [intent=reverse_etl availability=implemented write=archive_audience]; approval: requires plan, preview, approval, and execute; risk: Archive audience; this may remove or archive FireHydrant data.; flags: --audience_id (required)
    bulk update incident milestones apply - Plan and execute the bulk update incident milestones reverse-ETL action [intent=reverse_etl availability=implemented write=bulk_update_incident_milestones]; approval: requires plan, preview, approval, and execute; risk: Update milestone times through the FireHydrant API.; flags: --incident_id (required)
    close incident apply - Plan and execute the close incident reverse-ETL action [intent=reverse_etl availability=implemented write=close_incident]; approval: requires plan, preview, approval, and execute; risk: Close an incident through the FireHydrant API.; flags: --incident_id (required)
    convert incident task apply - Plan and execute the convert incident task reverse-ETL action [intent=reverse_etl availability=implemented write=convert_incident_task]; approval: requires plan, preview, approval, and execute; risk: Convert a task to a follow-up through the FireHydrant API.; flags: --incident_id (required), --task_id (required)
    copy on call schedule rotation apply - Plan and execute the copy on call schedule rotation reverse-ETL action [intent=reverse_etl availability=implemented write=copy_on_call_schedule_rotation]; approval: requires plan, preview, approval, and execute; risk: Copy an on-call schedule's rotation through the FireHydrant API.; flags: --rotation_id (required), --schedule_id (required), --team_id (required)
    create audience apply - Plan and execute the create audience reverse-ETL action [intent=reverse_etl availability=implemented write=create_audience]; approval: requires plan, preview, approval, and execute; risk: Create audience through the FireHydrant API.
    create change apply - Plan and execute the create change reverse-ETL action [intent=reverse_etl availability=implemented write=create_change]; approval: requires plan, preview, approval, and execute; risk: Create a new change entry through the FireHydrant API.
    create change event apply - Plan and execute the create change event reverse-ETL action [intent=reverse_etl availability=implemented write=create_change_event]; approval: requires plan, preview, approval, and execute; risk: Create a change event through the FireHydrant API.
    create change identity apply - Plan and execute the create change identity reverse-ETL action [intent=reverse_etl availability=implemented write=create_change_identity]; approval: requires plan, preview, approval, and execute; risk: Create an identity for a change entry through the FireHydrant API.; flags: --change_id (required)
    create checklist template apply - Plan and execute the create checklist template reverse-ETL action [intent=reverse_etl availability=implemented write=create_checklist_template]; approval: requires plan, preview, approval, and execute; risk: Create a checklist template through the FireHydrant API.
    create comment apply - Plan and execute the create comment reverse-ETL action [intent=reverse_etl availability=implemented write=create_comment]; approval: requires plan, preview, approval, and execute; risk: Create a conversation comment through the FireHydrant API.; flags: --conversation_id (required)
    create comment reaction apply - Plan and execute the create comment reaction reverse-ETL action [intent=reverse_etl availability=implemented write=create_comment_reaction]; approval: requires plan, preview, approval, and execute; risk: Create a reaction for a conversation comment through the FireHydrant API.; flags: --comment_id (required), --conversation_id (required)
    create connection apply - Plan and execute the create connection reverse-ETL action [intent=reverse_etl availability=implemented write=create_connection]; approval: requires plan, preview, approval, and execute; risk: Create a new integration connection through the FireHydrant API.; flags: --slug (required)
    create custom field definition apply - Plan and execute the create custom field definition reverse-ETL action [intent=reverse_etl availability=implemented write=create_custom_field_definition]; approval: requires plan, preview, approval, and execute; risk: Create a custom field definition through the FireHydrant API.
    create email subscriber apply - Plan and execute the create email subscriber reverse-ETL action [intent=reverse_etl availability=implemented write=create_email_subscriber]; approval: requires plan, preview, approval, and execute; risk: Add subscribers to a status page through the FireHydrant API.; flags: --nunc_connection_id (required)
    create environment apply - Plan and execute the create environment reverse-ETL action [intent=reverse_etl availability=implemented write=create_environment]; approval: requires plan, preview, approval, and execute; risk: Create an environment through the FireHydrant API.
    create functionality apply - Plan and execute the create functionality reverse-ETL action [intent=reverse_etl availability=implemented write=create_functionality]; approval: requires plan, preview, approval, and execute; risk: Create a functionality through the FireHydrant API.
    create inbound field map apply - Plan and execute the create inbound field map reverse-ETL action [intent=reverse_etl availability=implemented write=create_inbound_field_map]; approval: requires plan, preview, approval, and execute; risk: Create inbound field map for a ticketing project through the FireHydrant API.; flags: --ticketing_project_id (required)
    create incident alert apply - Plan and execute the create incident alert reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_alert]; approval: requires plan, preview, approval, and execute; risk: Attach an alert to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident apply - Plan and execute the create incident reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident]; approval: requires plan, preview, approval, and execute; risk: Create an incident through the FireHydrant API.
    create incident change event apply - Plan and execute the create incident change event reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_change_event]; approval: requires plan, preview, approval, and execute; risk: Add a related change to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident chat message apply - Plan and execute the create incident chat message reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_chat_message]; approval: requires plan, preview, approval, and execute; risk: Add a chat message to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident impact apply - Plan and execute the create incident impact reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_impact]; approval: requires plan, preview, approval, and execute; risk: Add impacted infrastructure to an incident through the FireHydrant API.; flags: --incident_id (required), --type (required)
    create incident link apply - Plan and execute the create incident link reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_link]; approval: requires plan, preview, approval, and execute; risk: Add a link to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident note apply - Plan and execute the create incident note reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_note]; approval: requires plan, preview, approval, and execute; risk: Add a note to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident retrospective apply - Plan and execute the create incident retrospective reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_retrospective]; approval: requires plan, preview, approval, and execute; risk: Create a new retrospective on the incident using the template through the FireHydrant API.; flags: --incident_id (required)
    create incident retrospective dynamic input apply - Plan and execute the create incident retrospective dynamic input reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_retrospective_dynamic_input]; approval: requires plan, preview, approval, and execute; risk: Add a new dynamic input field to a retrospective's dynamic input group field through the FireHydrant API.; flags: --field_id (required), --incident_id (required), --retrospective_id (required)
    create incident retrospective field apply - Plan and execute the create incident retrospective field reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_retrospective_field]; approval: requires plan, preview, approval, and execute; risk: Appends a new incident retrospective field to an incident retrospective through the FireHydrant API.; flags: --incident_id (required), --retrospective_id (required)
    create incident role apply - Plan and execute the create incident role reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_role]; approval: requires plan, preview, approval, and execute; risk: Create an incident role through the FireHydrant API.
    create incident role assignment apply - Plan and execute the create incident role assignment reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_role_assignment]; approval: requires plan, preview, approval, and execute; risk: Assign a user to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident status page apply - Plan and execute the create incident status page reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_status_page]; approval: requires plan, preview, approval, and execute; risk: Add a status page to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident task apply - Plan and execute the create incident task reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_task]; approval: requires plan, preview, approval, and execute; risk: Create an incident task through the FireHydrant API.; flags: --incident_id (required)
    create incident task list apply - Plan and execute the create incident task list reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_task_list]; approval: requires plan, preview, approval, and execute; risk: Add tasks from a task list to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident team assignment apply - Plan and execute the create incident team assignment reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_team_assignment]; approval: requires plan, preview, approval, and execute; risk: Assign a team to an incident through the FireHydrant API.; flags: --incident_id (required)
    create incident type apply - Plan and execute the create incident type reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident_type]; approval: requires plan, preview, approval, and execute; risk: Create an incident type through the FireHydrant API.
    create lifecycle measurement definition apply - Plan and execute the create lifecycle measurement definition reverse-ETL action [intent=reverse_etl availability=implemented write=create_lifecycle_measurement_definition]; approval: requires plan, preview, approval, and execute; risk: Create a measurement definition through the FireHydrant API.
    create lifecycle milestone apply - Plan and execute the create lifecycle milestone reverse-ETL action [intent=reverse_etl availability=implemented write=create_lifecycle_milestone]; approval: requires plan, preview, approval, and execute; risk: Create a milestone through the FireHydrant API.
    create notification policy apply - Plan and execute the create notification policy reverse-ETL action [intent=reverse_etl availability=implemented write=create_notification_policy]; approval: requires plan, preview, approval, and execute; risk: Create a notification policy through the FireHydrant API.
    create nunc component group apply - Plan and execute the create nunc component group reverse-ETL action [intent=reverse_etl availability=implemented write=create_nunc_component_group]; approval: requires plan, preview, approval, and execute; risk: Create a component group for a status page through the FireHydrant API.; flags: --nunc_connection_id (required)
    create nunc connection apply - Plan and execute the create nunc connection reverse-ETL action [intent=reverse_etl availability=implemented write=create_nunc_connection]; approval: requires plan, preview, approval, and execute; risk: Create a status page through the FireHydrant API.
    create nunc link apply - Plan and execute the create nunc link reverse-ETL action [intent=reverse_etl availability=implemented write=create_nunc_link]; approval: requires plan, preview, approval, and execute; risk: Add link to a status page through the FireHydrant API.; flags: --nunc_connection_id (required)
    create nunc subscription apply - Plan and execute the create nunc subscription reverse-ETL action [intent=reverse_etl availability=implemented write=create_nunc_subscription]; approval: requires plan, preview, approval, and execute; risk: Create a status page subscription through the FireHydrant API.
    create on call schedule rotation apply - Plan and execute the create on call schedule rotation reverse-ETL action [intent=reverse_etl availability=implemented write=create_on_call_schedule_rotation]; approval: requires plan, preview, approval, and execute; risk: Create a new on-call rotation through the FireHydrant API.; flags: --schedule_id (required), --team_id (required)
    create on call shift apply - Plan and execute the create on call shift reverse-ETL action [intent=reverse_etl availability=implemented write=create_on_call_shift]; approval: requires plan, preview, approval, and execute; risk: Create a shift for an on-call schedule through the FireHydrant API.; flags: --schedule_id (required), --team_id (required)
    create post mortem reason apply - Plan and execute the create post mortem reason reverse-ETL action [intent=reverse_etl availability=implemented write=create_post_mortem_reason]; approval: requires plan, preview, approval, and execute; risk: Create a contributing factor for a retrospective report through the FireHydrant API.; flags: --report_id (required)
    create post mortem report apply - Plan and execute the create post mortem report reverse-ETL action [intent=reverse_etl availability=implemented write=create_post_mortem_report]; approval: requires plan, preview, approval, and execute; risk: Create a retrospective report through the FireHydrant API.
    create priority apply - Plan and execute the create priority reverse-ETL action [intent=reverse_etl availability=implemented write=create_priority]; approval: requires plan, preview, approval, and execute; risk: Create a priority through the FireHydrant API.
    create retrospective template apply - Plan and execute the create retrospective template reverse-ETL action [intent=reverse_etl availability=implemented write=create_retrospective_template]; approval: requires plan, preview, approval, and execute; risk: Create a retrospective template through the FireHydrant API.
    create role apply - Plan and execute the create role reverse-ETL action [intent=reverse_etl availability=implemented write=create_role]; approval: requires plan, preview, approval, and execute; risk: Create a role through the FireHydrant API.
    create runbook apply - Plan and execute the create runbook reverse-ETL action [intent=reverse_etl availability=implemented write=create_runbook]; approval: requires plan, preview, approval, and execute; risk: Create a runbook through the FireHydrant API.
    create runbook execution apply - Plan and execute the create runbook execution reverse-ETL action [intent=reverse_etl availability=implemented write=create_runbook_execution]; approval: requires plan, preview, approval, and execute; risk: Create a runbook execution through the FireHydrant API.
    create saved search apply - Plan and execute the create saved search reverse-ETL action [intent=reverse_etl availability=implemented write=create_saved_search]; approval: requires plan, preview, approval, and execute; risk: Create a saved search through the FireHydrant API.; flags: --resource_type (required)
    create scheduled maintenance apply - Plan and execute the create scheduled maintenance reverse-ETL action [intent=reverse_etl availability=implemented write=create_scheduled_maintenance]; approval: requires plan, preview, approval, and execute; risk: Create a scheduled maintenance event through the FireHydrant API.
    create service apply - Plan and execute the create service reverse-ETL action [intent=reverse_etl availability=implemented write=create_service]; approval: requires plan, preview, approval, and execute; risk: Create a service through the FireHydrant API.
    create service checklist response apply - Plan and execute the create service checklist response reverse-ETL action [intent=reverse_etl availability=implemented write=create_service_checklist_response]; approval: requires plan, preview, approval, and execute; risk: Record a response for a checklist item through the FireHydrant API.; flags: --checklist_id (required), --service_id (required)
    create service dependency apply - Plan and execute the create service dependency reverse-ETL action [intent=reverse_etl availability=implemented write=create_service_dependency]; approval: requires plan, preview, approval, and execute; risk: Create a service dependency through the FireHydrant API.
    create service links apply - Plan and execute the create service links reverse-ETL action [intent=reverse_etl availability=implemented write=create_service_links]; approval: requires plan, preview, approval, and execute; risk: Create multiple services linked to external services through the FireHydrant API.
    create severity apply - Plan and execute the create severity reverse-ETL action [intent=reverse_etl availability=implemented write=create_severity]; approval: requires plan, preview, approval, and execute; risk: Create a severity through the FireHydrant API.
    create severity matrix condition apply - Plan and execute the create severity matrix condition reverse-ETL action [intent=reverse_etl availability=implemented write=create_severity_matrix_condition]; approval: requires plan, preview, approval, and execute; risk: Create a severity matrix condition through the FireHydrant API.
    create severity matrix impact apply - Plan and execute the create severity matrix impact reverse-ETL action [intent=reverse_etl availability=implemented write=create_severity_matrix_impact]; approval: requires plan, preview, approval, and execute; risk: Create a severity matrix impact through the FireHydrant API.
    create signals alert grouping configuration apply - Plan and execute the create signals alert grouping configuration reverse-ETL action [intent=reverse_etl availability=implemented write=create_signals_alert_grouping_configuration]; approval: requires plan, preview, approval, and execute; risk: Create an alert grouping configuration. through the FireHydrant API.
    create signals email target apply - Plan and execute the create signals email target reverse-ETL action [intent=reverse_etl availability=implemented write=create_signals_email_target]; approval: requires plan, preview, approval, and execute; risk: Create an email target for signals through the FireHydrant API.
    create signals event source apply - Plan and execute the create signals event source reverse-ETL action [intent=reverse_etl availability=implemented write=create_signals_event_source]; approval: requires plan, preview, approval, and execute; risk: Create an event source for Signals through the FireHydrant API.
    create signals heartbeat endpoint configuration apply - Plan and execute the create signals heartbeat endpoint configuration reverse-ETL action [intent=reverse_etl availability=implemented write=create_signals_heartbeat_endpoint_configuration]; approval: requires plan, preview, approval, and execute; risk: Create a heartbeat endpoint configuration through the FireHydrant API.
    create signals page apply - Plan and execute the create signals page reverse-ETL action [intent=reverse_etl availability=implemented write=create_signals_page]; approval: requires plan, preview, approval, and execute; risk: Page a user, team, on-call schedule, or escalation policy through the FireHydrant API.
    create signals webhook target apply - Plan and execute the create signals webhook target reverse-ETL action [intent=reverse_etl availability=implemented write=create_signals_webhook_target]; approval: requires plan, preview, approval, and execute; risk: Create a webhook target through the FireHydrant API.
    create slack emoji action apply - Plan and execute the create slack emoji action reverse-ETL action [intent=reverse_etl availability=implemented write=create_slack_emoji_action]; approval: requires plan, preview, approval, and execute; risk: Create a new Slack emoji action through the FireHydrant API.; flags: --connection_id (required)
    create status update template apply - Plan and execute the create status update template reverse-ETL action [intent=reverse_etl availability=implemented write=create_status_update_template]; approval: requires plan, preview, approval, and execute; risk: Create a status update template through the FireHydrant API.
    create task list apply - Plan and execute the create task list reverse-ETL action [intent=reverse_etl availability=implemented write=create_task_list]; approval: requires plan, preview, approval, and execute; risk: Create a task list through the FireHydrant API.
    create team apply - Plan and execute the create team reverse-ETL action [intent=reverse_etl availability=implemented write=create_team]; approval: requires plan, preview, approval, and execute; risk: Create a team through the FireHydrant API.
    create team call route apply - Plan and execute the create team call route reverse-ETL action [intent=reverse_etl availability=implemented write=create_team_call_route]; approval: requires plan, preview, approval, and execute; risk: Create a call route for a team through the FireHydrant API.; flags: --team_id (required)
    create team escalation policy apply - Plan and execute the create team escalation policy reverse-ETL action [intent=reverse_etl availability=implemented write=create_team_escalation_policy]; approval: requires plan, preview, approval, and execute; risk: Create an escalation policy for a team through the FireHydrant API.; flags: --team_id (required)
    create team on call schedule apply - Plan and execute the create team on call schedule reverse-ETL action [intent=reverse_etl availability=implemented write=create_team_on_call_schedule]; approval: requires plan, preview, approval, and execute; risk: Create an on-call schedule for a team through the FireHydrant API.; flags: --team_id (required)
    create team signal rule apply - Plan and execute the create team signal rule reverse-ETL action [intent=reverse_etl availability=implemented write=create_team_signal_rule]; approval: requires plan, preview, approval, and execute; risk: Create a Signals rule through the FireHydrant API.; flags: --team_id (required)
    create ticket apply - Plan and execute the create ticket reverse-ETL action [intent=reverse_etl availability=implemented write=create_ticket]; approval: requires plan, preview, approval, and execute; risk: Create a ticket through the FireHydrant API.
    create ticketing custom definition apply - Plan and execute the create ticketing custom definition reverse-ETL action [intent=reverse_etl availability=implemented write=create_ticketing_custom_definition]; approval: requires plan, preview, approval, and execute; risk: Create a ticketing custom field through the FireHydrant API.
    create ticketing field map apply - Plan and execute the create ticketing field map reverse-ETL action [intent=reverse_etl availability=implemented write=create_ticketing_field_map]; approval: requires plan, preview, approval, and execute; risk: Create a field mapping for a ticketing project through the FireHydrant API.; flags: --ticketing_project_id (required)
    create ticketing priority apply - Plan and execute the create ticketing priority reverse-ETL action [intent=reverse_etl availability=implemented write=create_ticketing_priority]; approval: requires plan, preview, approval, and execute; risk: Create a ticketing priority through the FireHydrant API.
    create ticketing project config apply - Plan and execute the create ticketing project config reverse-ETL action [intent=reverse_etl availability=implemented write=create_ticketing_project_config]; approval: requires plan, preview, approval, and execute; risk: Create a ticketing project configuration through the FireHydrant API.; flags: --ticketing_project_id (required)
    create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: Create a webhook through the FireHydrant API.
    debug signals expression apply - Plan and execute the debug signals expression reverse-ETL action [intent=reverse_etl availability=implemented write=debug_signals_expression]; approval: requires plan, preview, approval, and execute; risk: Debug Signals expressions through the FireHydrant API.
    delete call route apply - Plan and execute the delete call route reverse-ETL action [intent=reverse_etl availability=implemented write=delete_call_route]; approval: requires plan, preview, approval, and execute; risk: Delete a call route; this may remove or archive FireHydrant data.; flags: --id (required)
    delete change apply - Plan and execute the delete change reverse-ETL action [intent=reverse_etl availability=implemented write=delete_change]; approval: requires plan, preview, approval, and execute; risk: Archive a change entry; this may remove or archive FireHydrant data.; flags: --change_id (required)
    delete change event apply - Plan and execute the delete change event reverse-ETL action [intent=reverse_etl availability=implemented write=delete_change_event]; approval: requires plan, preview, approval, and execute; risk: Delete a change event; this may remove or archive FireHydrant data.; flags: --change_event_id (required)
    delete change identity apply - Plan and execute the delete change identity reverse-ETL action [intent=reverse_etl availability=implemented write=delete_change_identity]; approval: requires plan, preview, approval, and execute; risk: Delete an identity from a change entry; this may remove or archive FireHydrant data.; flags: --change_id (required), --identity_id (required)
    delete checklist template apply - Plan and execute the delete checklist template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_checklist_template]; approval: requires plan, preview, approval, and execute; risk: Archive a checklist template; this may remove or archive FireHydrant data.; flags: --id (required)
    delete comment apply - Plan and execute the delete comment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_comment]; approval: requires plan, preview, approval, and execute; risk: Archive a conversation comment; this may remove or archive FireHydrant data.; flags: --comment_id (required), --conversation_id (required)
    delete comment reaction apply - Plan and execute the delete comment reaction reverse-ETL action [intent=reverse_etl availability=implemented write=delete_comment_reaction]; approval: requires plan, preview, approval, and execute; risk: Delete a reaction from a conversation comment; this may remove or archive FireHydrant data.; flags: --comment_id (required), --conversation_id (required), --reaction_id (required)
    delete custom field definition apply - Plan and execute the delete custom field definition reverse-ETL action [intent=reverse_etl availability=implemented write=delete_custom_field_definition]; approval: requires plan, preview, approval, and execute; risk: Delete a custom field definition; this may remove or archive FireHydrant data.; flags: --field_id (required)
    delete environment apply - Plan and execute the delete environment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_environment]; approval: requires plan, preview, approval, and execute; risk: Archive an environment; this may remove or archive FireHydrant data.; flags: --environment_id (required)
    delete functionality apply - Plan and execute the delete functionality reverse-ETL action [intent=reverse_etl availability=implemented write=delete_functionality]; approval: requires plan, preview, approval, and execute; risk: Archive a functionality; this may remove or archive FireHydrant data.; flags: --functionality_id (required)
    delete inbound field map apply - Plan and execute the delete inbound field map reverse-ETL action [intent=reverse_etl availability=implemented write=delete_inbound_field_map]; approval: requires plan, preview, approval, and execute; risk: Archive inbound field map for a ticketing project; this may remove or archive FireHydrant data.; flags: --map_id (required), --ticketing_project_id (required)
    delete incident alert apply - Plan and execute the delete incident alert reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_alert]; approval: requires plan, preview, approval, and execute; risk: Remove an alert from an incident; this may remove or archive FireHydrant data.; flags: --incident_alert_id (required), --incident_id (required)
    delete incident apply - Plan and execute the delete incident reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident]; approval: requires plan, preview, approval, and execute; risk: Archive an incident; this may remove or archive FireHydrant data.; flags: --incident_id (required)
    delete incident chat message apply - Plan and execute the delete incident chat message reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_chat_message]; approval: requires plan, preview, approval, and execute; risk: Delete a chat message from an incident; this may remove or archive FireHydrant data.; flags: --incident_id (required), --message_id (required)
    delete incident event apply - Plan and execute the delete incident event reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_event]; approval: requires plan, preview, approval, and execute; risk: Delete an incident event; this may remove or archive FireHydrant data.; flags: --event_id (required), --incident_id (required)
    delete incident impact apply - Plan and execute the delete incident impact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_impact]; approval: requires plan, preview, approval, and execute; risk: Remove impacted infrastructure from an incident; this may remove or archive FireHydrant data.; flags: --id (required), --incident_id (required), --type (required)
    delete incident link apply - Plan and execute the delete incident link reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_link]; approval: requires plan, preview, approval, and execute; risk: Remove a link from an incident; this may remove or archive FireHydrant data.; flags: --incident_id (required), --link_id (required)
    delete incident role apply - Plan and execute the delete incident role reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_role]; approval: requires plan, preview, approval, and execute; risk: Archive an incident role; this may remove or archive FireHydrant data.; flags: --incident_role_id (required)
    delete incident role assignment apply - Plan and execute the delete incident role assignment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_role_assignment]; approval: requires plan, preview, approval, and execute; risk: Unassign a user from an incident; this may remove or archive FireHydrant data.; flags: --incident_id (required), --role_assignment_id (required)
    delete incident status page apply - Plan and execute the delete incident status page reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_status_page]; approval: requires plan, preview, approval, and execute; risk: Remove a status page from an incident; this may remove or archive FireHydrant data.; flags: --incident_id (required), --status_page_id (required)
    delete incident task apply - Plan and execute the delete incident task reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_task]; approval: requires plan, preview, approval, and execute; risk: Delete an incident task; this may remove or archive FireHydrant data.; flags: --incident_id (required), --task_id (required)
    delete incident type apply - Plan and execute the delete incident type reverse-ETL action [intent=reverse_etl availability=implemented write=delete_incident_type]; approval: requires plan, preview, approval, and execute; risk: Archive an incident type; this may remove or archive FireHydrant data.; flags: --id (required)
    delete lifecycle measurement definition apply - Plan and execute the delete lifecycle measurement definition reverse-ETL action [intent=reverse_etl availability=implemented write=delete_lifecycle_measurement_definition]; approval: requires plan, preview, approval, and execute; risk: Archive a measurement definition; this may remove or archive FireHydrant data.; flags: --measurement_definition_id (required)
    delete lifecycle milestone apply - Plan and execute the delete lifecycle milestone reverse-ETL action [intent=reverse_etl availability=implemented write=delete_lifecycle_milestone]; approval: requires plan, preview, approval, and execute; risk: Delete a milestone; this may remove or archive FireHydrant data.; flags: --milestone_id (required)
    delete notification policy apply - Plan and execute the delete notification policy reverse-ETL action [intent=reverse_etl availability=implemented write=delete_notification_policy]; approval: requires plan, preview, approval, and execute; risk: Delete a notification policy; this may remove or archive FireHydrant data.; flags: --id (required)
    delete nunc component group apply - Plan and execute the delete nunc component group reverse-ETL action [intent=reverse_etl availability=implemented write=delete_nunc_component_group]; approval: requires plan, preview, approval, and execute; risk: Delete a status page component group; this may remove or archive FireHydrant data.; flags: --group_id (required), --nunc_connection_id (required)
    delete nunc connection apply - Plan and execute the delete nunc connection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_nunc_connection]; approval: requires plan, preview, approval, and execute; risk: Delete a status page; this may remove or archive FireHydrant data.; flags: --nunc_connection_id (required)
    delete nunc image apply - Plan and execute the delete nunc image reverse-ETL action [intent=reverse_etl availability=implemented write=delete_nunc_image]; approval: requires plan, preview, approval, and execute; risk: Delete an image from a status page; this may remove or archive FireHydrant data.; flags: --nunc_connection_id (required), --type (required)
    delete nunc link apply - Plan and execute the delete nunc link reverse-ETL action [intent=reverse_etl availability=implemented write=delete_nunc_link]; approval: requires plan, preview, approval, and execute; risk: Delete a status page link; this may remove or archive FireHydrant data.; flags: --link_id (required), --nunc_connection_id (required)
    delete nunc subscription apply - Plan and execute the delete nunc subscription reverse-ETL action [intent=reverse_etl availability=implemented write=delete_nunc_subscription]; approval: requires plan, preview, approval, and execute; risk: Unsubscribe from status page notifications; this may remove or archive FireHydrant data.; flags: --unsubscribe_token (required)
    delete on call schedule rotation apply - Plan and execute the delete on call schedule rotation reverse-ETL action [intent=reverse_etl availability=implemented write=delete_on_call_schedule_rotation]; approval: requires plan, preview, approval, and execute; risk: Delete an on-call schedule's rotation; this may remove or archive FireHydrant data.; flags: --rotation_id (required), --schedule_id (required), --team_id (required)
    delete on call shift apply - Plan and execute the delete on call shift reverse-ETL action [intent=reverse_etl availability=implemented write=delete_on_call_shift]; approval: requires plan, preview, approval, and execute; risk: Delete an on-call shift from a team schedule; this may remove or archive FireHydrant data.; flags: --id (required), --schedule_id (required), --team_id (required)
    delete post mortem reason apply - Plan and execute the delete post mortem reason reverse-ETL action [intent=reverse_etl availability=implemented write=delete_post_mortem_reason]; approval: requires plan, preview, approval, and execute; risk: Delete a contributing factor from a retrospective report; this may remove or archive FireHydrant data.; flags: --reason_id (required), --report_id (required)
    delete priority apply - Plan and execute the delete priority reverse-ETL action [intent=reverse_etl availability=implemented write=delete_priority]; approval: requires plan, preview, approval, and execute; risk: Delete a priority; this may remove or archive FireHydrant data.; flags: --priority_slug (required)
    delete retrospective template apply - Plan and execute the delete retrospective template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_retrospective_template]; approval: requires plan, preview, approval, and execute; risk: Delete a retrospective template; this may remove or archive FireHydrant data.; flags: --retrospective_template_id (required)
    delete role apply - Plan and execute the delete role reverse-ETL action [intent=reverse_etl availability=implemented write=delete_role]; approval: requires plan, preview, approval, and execute; risk: Delete a role; this may remove or archive FireHydrant data.; flags: --id (required)
    delete runbook apply - Plan and execute the delete runbook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_runbook]; approval: requires plan, preview, approval, and execute; risk: Delete a runbook; this may remove or archive FireHydrant data.; flags: --runbook_id (required)
    delete saved search apply - Plan and execute the delete saved search reverse-ETL action [intent=reverse_etl availability=implemented write=delete_saved_search]; approval: requires plan, preview, approval, and execute; risk: Delete a saved search; this may remove or archive FireHydrant data.; flags: --resource_type (required), --saved_search_id (required)
    delete scheduled maintenance apply - Plan and execute the delete scheduled maintenance reverse-ETL action [intent=reverse_etl availability=implemented write=delete_scheduled_maintenance]; approval: requires plan, preview, approval, and execute; risk: Delete a scheduled maintenance event; this may remove or archive FireHydrant data.; flags: --scheduled_maintenance_id (required)
    delete service apply - Plan and execute the delete service reverse-ETL action [intent=reverse_etl availability=implemented write=delete_service]; approval: requires plan, preview, approval, and execute; risk: Delete a service; this may remove or archive FireHydrant data.; flags: --service_id (required)
    delete service dependency apply - Plan and execute the delete service dependency reverse-ETL action [intent=reverse_etl availability=implemented write=delete_service_dependency]; approval: requires plan, preview, approval, and execute; risk: Delete a service dependency; this may remove or archive FireHydrant data.; flags: --service_dependency_id (required)
    delete service link apply - Plan and execute the delete service link reverse-ETL action [intent=reverse_etl availability=implemented write=delete_service_link]; approval: requires plan, preview, approval, and execute; risk: Delete a service link; this may remove or archive FireHydrant data.; flags: --remote_id (required), --service_id (required)
    delete severity apply - Plan and execute the delete severity reverse-ETL action [intent=reverse_etl availability=implemented write=delete_severity]; approval: requires plan, preview, approval, and execute; risk: Delete a severity; this may remove or archive FireHydrant data.; flags: --severity_slug (required)
    delete severity matrix condition apply - Plan and execute the delete severity matrix condition reverse-ETL action [intent=reverse_etl availability=implemented write=delete_severity_matrix_condition]; approval: requires plan, preview, approval, and execute; risk: Delete a severity matrix condition; this may remove or archive FireHydrant data.; flags: --condition_id (required)
    delete severity matrix impact apply - Plan and execute the delete severity matrix impact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_severity_matrix_impact]; approval: requires plan, preview, approval, and execute; risk: Delete a severity matrix impact; this may remove or archive FireHydrant data.; flags: --impact_id (required)
    delete signals alert grouping configuration apply - Plan and execute the delete signals alert grouping configuration reverse-ETL action [intent=reverse_etl availability=implemented write=delete_signals_alert_grouping_configuration]; approval: requires plan, preview, approval, and execute; risk: Delete an alert grouping configuration.; this may remove or archive FireHydrant data.; flags: --id (required)
    delete signals email target apply - Plan and execute the delete signals email target reverse-ETL action [intent=reverse_etl availability=implemented write=delete_signals_email_target]; approval: requires plan, preview, approval, and execute; risk: Delete a signal email target; this may remove or archive FireHydrant data.; flags: --id (required)
    delete signals event source apply - Plan and execute the delete signals event source reverse-ETL action [intent=reverse_etl availability=implemented write=delete_signals_event_source]; approval: requires plan, preview, approval, and execute; risk: Delete an event source for Signals; this may remove or archive FireHydrant data.; flags: --transposer_slug (required)
    delete signals heartbeat endpoint configuration apply - Plan and execute the delete signals heartbeat endpoint configuration reverse-ETL action [intent=reverse_etl availability=implemented write=delete_signals_heartbeat_endpoint_configuration]; approval: requires plan, preview, approval, and execute; risk: Delete a heartbeat endpoint configuration; this may remove or archive FireHydrant data.; flags: --id (required)
    delete signals webhook target apply - Plan and execute the delete signals webhook target reverse-ETL action [intent=reverse_etl availability=implemented write=delete_signals_webhook_target]; approval: requires plan, preview, approval, and execute; risk: Delete a webhook target; this may remove or archive FireHydrant data.; flags: --id (required)
    delete slack emoji action apply - Plan and execute the delete slack emoji action reverse-ETL action [intent=reverse_etl availability=implemented write=delete_slack_emoji_action]; approval: requires plan, preview, approval, and execute; risk: Delete a Slack emoji action; this may remove or archive FireHydrant data.; flags: --connection_id (required), --emoji_action_id (required)
    delete status update template apply - Plan and execute the delete status update template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_status_update_template]; approval: requires plan, preview, approval, and execute; risk: Delete a status update template; this may remove or archive FireHydrant data.; flags: --status_update_template_id (required)
    delete statuspage connection apply - Plan and execute the delete statuspage connection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_statuspage_connection]; approval: requires plan, preview, approval, and execute; risk: Delete a Statuspage connection; this may remove or archive FireHydrant data.; flags: --connection_id (required)
    delete support hours schedule apply - Plan and execute the delete support hours schedule reverse-ETL action [intent=reverse_etl availability=implemented write=delete_support_hours_schedule]; approval: requires plan, preview, approval, and execute; risk: Delete a specific support hours schedule; this may remove or archive FireHydrant data.; flags: --team_id (required)
    delete task list apply - Plan and execute the delete task list reverse-ETL action [intent=reverse_etl availability=implemented write=delete_task_list]; approval: requires plan, preview, approval, and execute; risk: Delete a task list; this may remove or archive FireHydrant data.; flags: --task_list_id (required)
    delete team apply - Plan and execute the delete team reverse-ETL action [intent=reverse_etl availability=implemented write=delete_team]; approval: requires plan, preview, approval, and execute; risk: Archive a team; this may remove or archive FireHydrant data.; flags: --team_id (required)
    delete team escalation policy apply - Plan and execute the delete team escalation policy reverse-ETL action [intent=reverse_etl availability=implemented write=delete_team_escalation_policy]; approval: requires plan, preview, approval, and execute; risk: Delete an escalation policy for a team; this may remove or archive FireHydrant data.; flags: --id (required), --team_id (required)
    delete team on call schedule apply - Plan and execute the delete team on call schedule reverse-ETL action [intent=reverse_etl availability=implemented write=delete_team_on_call_schedule]; approval: requires plan, preview, approval, and execute; risk: Delete an on-call schedule for a team; this may remove or archive FireHydrant data.; flags: --schedule_id (required), --team_id (required)
    delete team signal rule apply - Plan and execute the delete team signal rule reverse-ETL action [intent=reverse_etl availability=implemented write=delete_team_signal_rule]; approval: requires plan, preview, approval, and execute; risk: Delete a Signals rule; this may remove or archive FireHydrant data.; flags: --id (required), --team_id (required)
    delete ticket apply - Plan and execute the delete ticket reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ticket]; approval: requires plan, preview, approval, and execute; risk: Archive a ticket; this may remove or archive FireHydrant data.; flags: --ticket_id (required)
    delete ticketing custom definition apply - Plan and execute the delete ticketing custom definition reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ticketing_custom_definition]; approval: requires plan, preview, approval, and execute; risk: Delete a ticketing custom field; this may remove or archive FireHydrant data.; flags: --field_id (required)
    delete ticketing field map apply - Plan and execute the delete ticketing field map reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ticketing_field_map]; approval: requires plan, preview, approval, and execute; risk: Archive a field map for a ticketing project; this may remove or archive FireHydrant data.; flags: --map_id (required), --ticketing_project_id (required)
    delete ticketing priority apply - Plan and execute the delete ticketing priority reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ticketing_priority]; approval: requires plan, preview, approval, and execute; risk: Delete a ticketing priority; this may remove or archive FireHydrant data.; flags: --id (required)
    delete ticketing project config apply - Plan and execute the delete ticketing project config reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ticketing_project_config]; approval: requires plan, preview, approval, and execute; risk: Archive a ticketing project configuration; this may remove or archive FireHydrant data.; flags: --config_id (required), --ticketing_project_id (required)
    delete transcript entry apply - Plan and execute the delete transcript entry reverse-ETL action [intent=reverse_etl availability=implemented write=delete_transcript_entry]; approval: requires plan, preview, approval, and execute; risk: Delete a transcript from an incident; this may remove or archive FireHydrant data.; flags: --incident_id (required), --transcript_id (required)
    delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: Delete a webhook; this may remove or archive FireHydrant data.; flags: --webhook_id (required)
    environments list - Run the environments ETL stream [intent=etl availability=implemented stream=environments]
    functionalities list - Run the functionalities ETL stream [intent=etl availability=implemented stream=functionalities]
    generate audience summary apply - Plan and execute the generate audience summary reverse-ETL action [intent=reverse_etl availability=implemented write=generate_audience_summary]; approval: requires plan, preview, approval, and execute; risk: Generate summary (async) through the FireHydrant API.; flags: --audience_id (required), --incident_id (required)
    get ai incident summary vote status list - Run the get ai incident summary vote status ETL stream [intent=etl availability=implemented stream=get_ai_incident_summary_vote_status]
    get ai preferences list - Run the get ai preferences ETL stream [intent=etl availability=implemented stream=get_ai_preferences]
    get alert list - Run the get alert ETL stream [intent=etl availability=implemented stream=get_alert]
    get audience list - Run the get audience ETL stream [intent=etl availability=implemented stream=get_audience]
    get audience summary list - Run the get audience summary ETL stream [intent=etl availability=implemented stream=get_audience_summary]
    get audit event list - Run the get audit event ETL stream [intent=etl availability=implemented stream=get_audit_event]
    get aws cloudtrail batch list - Run the get aws cloudtrail batch ETL stream [intent=etl availability=implemented stream=get_aws_cloudtrail_batch]
    get aws connection list - Run the get aws connection ETL stream [intent=etl availability=implemented stream=get_aws_connection]
    get bootstrap list - Run the get bootstrap ETL stream [intent=etl availability=implemented stream=get_bootstrap]
    get call route list - Run the get call route ETL stream [intent=etl availability=implemented stream=get_call_route]
    get change event list - Run the get change event ETL stream [intent=etl availability=implemented stream=get_change_event]
    get checklist template list - Run the get checklist template ETL stream [intent=etl availability=implemented stream=get_checklist_template]
    get comment list - Run the get comment ETL stream [intent=etl availability=implemented stream=get_comment]
    get conference bridge translation list - Run the get conference bridge translation ETL stream [intent=etl availability=implemented stream=get_conference_bridge_translation]
    get configuration options list - Run the get configuration options ETL stream [intent=etl availability=implemented stream=get_configuration_options]
    get current user list - Run the get current user ETL stream [intent=etl availability=implemented stream=get_current_user]
    get environment list - Run the get environment ETL stream [intent=etl availability=implemented stream=get_environment]
    get form configuration list - Run the get form configuration ETL stream [intent=etl availability=implemented stream=get_form_configuration]
    get functionality list - Run the get functionality ETL stream [intent=etl availability=implemented stream=get_functionality]
    get inbound field map list - Run the get inbound field map ETL stream [intent=etl availability=implemented stream=get_inbound_field_map]
    get incident channel list - Run the get incident channel ETL stream [intent=etl availability=implemented stream=get_incident_channel]
    get incident event list - Run the get incident event ETL stream [intent=etl availability=implemented stream=get_incident_event]
    get incident list - Run the get incident ETL stream [intent=etl availability=implemented stream=get_incident]
    get incident relationships list - Run the get incident relationships ETL stream [intent=etl availability=implemented stream=get_incident_relationships]
    get incident retrospective field list - Run the get incident retrospective field ETL stream [intent=etl availability=implemented stream=get_incident_retrospective_field]
    get incident role list - Run the get incident role ETL stream [intent=etl availability=implemented stream=get_incident_role]
    get incident task list - Run the get incident task ETL stream [intent=etl availability=implemented stream=get_incident_task]
    get incident type list - Run the get incident type ETL stream [intent=etl availability=implemented stream=get_incident_type]
    get incident user list - Run the get incident user ETL stream [intent=etl availability=implemented stream=get_incident_user]
    get integration list - Run the get integration ETL stream [intent=etl availability=implemented stream=get_integration]
    get lifecycle measurement definition list - Run the get lifecycle measurement definition ETL stream [intent=etl availability=implemented stream=get_lifecycle_measurement_definition]
    get mean time report list - Run the get mean time report ETL stream [intent=etl availability=implemented stream=get_mean_time_report]
    get member default audience list - Run the get member default audience ETL stream [intent=etl availability=implemented stream=get_member_default_audience]
    get notification policy list - Run the get notification policy ETL stream [intent=etl availability=implemented stream=get_notification_policy]
    get nunc connection list - Run the get nunc connection ETL stream [intent=etl availability=implemented stream=get_nunc_connection]
    get on call schedule rotation list - Run the get on call schedule rotation ETL stream [intent=etl availability=implemented stream=get_on_call_schedule_rotation]
    get on call shift list - Run the get on call shift ETL stream [intent=etl availability=implemented stream=get_on_call_shift]
    get options for field list - Run the get options for field ETL stream [intent=etl availability=implemented stream=get_options_for_field]
    get post mortem question list - Run the get post mortem question ETL stream [intent=etl availability=implemented stream=get_post_mortem_question]
    get post mortem report list - Run the get post mortem report ETL stream [intent=etl availability=implemented stream=get_post_mortem_report]
    get priority list - Run the get priority ETL stream [intent=etl availability=implemented stream=get_priority]
    get retrospective template list - Run the get retrospective template ETL stream [intent=etl availability=implemented stream=get_retrospective_template]
    get role list - Run the get role ETL stream [intent=etl availability=implemented stream=get_role]
    get runbook action field options list - Run the get runbook action field options ETL stream [intent=etl availability=implemented stream=get_runbook_action_field_options]
    get runbook execution list - Run the get runbook execution ETL stream [intent=etl availability=implemented stream=get_runbook_execution]
    get runbook execution step script list - Run the get runbook execution step script ETL stream [intent=etl availability=implemented stream=get_runbook_execution_step_script]
    get runbook list - Run the get runbook ETL stream [intent=etl availability=implemented stream=get_runbook]
    get saved search list - Run the get saved search ETL stream [intent=etl availability=implemented stream=get_saved_search]
    get scheduled maintenance list - Run the get scheduled maintenance ETL stream [intent=etl availability=implemented stream=get_scheduled_maintenance]
    get service dependencies list - Run the get service dependencies ETL stream [intent=etl availability=implemented stream=get_service_dependencies]
    get service dependency list - Run the get service dependency ETL stream [intent=etl availability=implemented stream=get_service_dependency]
    get service list - Run the get service ETL stream [intent=etl availability=implemented stream=get_service]
    get severity list - Run the get severity ETL stream [intent=etl availability=implemented stream=get_severity]
    get severity matrix condition list - Run the get severity matrix condition ETL stream [intent=etl availability=implemented stream=get_severity_matrix_condition]
    get severity matrix list - Run the get severity matrix ETL stream [intent=etl availability=implemented stream=get_severity_matrix]
    get signals alert grouping configuration list - Run the get signals alert grouping configuration ETL stream [intent=etl availability=implemented stream=get_signals_alert_grouping_configuration]
    get signals email target list - Run the get signals email target ETL stream [intent=etl availability=implemented stream=get_signals_email_target]
    get signals event source list - Run the get signals event source ETL stream [intent=etl availability=implemented stream=get_signals_event_source]
    get signals grouped metrics list - Run the get signals grouped metrics ETL stream [intent=etl availability=implemented stream=get_signals_grouped_metrics]
    get signals hacker mode list - Run the get signals hacker mode ETL stream [intent=etl availability=implemented stream=get_signals_hacker_mode]
    get signals heartbeat endpoint configuration list - Run the get signals heartbeat endpoint configuration ETL stream [intent=etl availability=implemented stream=get_signals_heartbeat_endpoint_configuration]
    get signals ingest url list - Run the get signals ingest url ETL stream [intent=etl availability=implemented stream=get_signals_ingest_url]
    get signals mttx analytics list - Run the get signals mttx analytics ETL stream [intent=etl availability=implemented stream=get_signals_mttx_analytics]
    get signals noise analytics list - Run the get signals noise analytics ETL stream [intent=etl availability=implemented stream=get_signals_noise_analytics]
    get signals timeseries analytics list - Run the get signals timeseries analytics ETL stream [intent=etl availability=implemented stream=get_signals_timeseries_analytics]
    get signals webhook target list - Run the get signals webhook target ETL stream [intent=etl availability=implemented stream=get_signals_webhook_target]
    get slack emoji action list - Run the get slack emoji action ETL stream [intent=etl availability=implemented stream=get_slack_emoji_action]
    get status update template list - Run the get status update template ETL stream [intent=etl availability=implemented stream=get_status_update_template]
    get statuspage connection list - Run the get statuspage connection ETL stream [intent=etl availability=implemented stream=get_statuspage_connection]
    get support hours schedule list - Run the get support hours schedule ETL stream [intent=etl availability=implemented stream=get_support_hours_schedule]
    get task list list - Run the get task list ETL stream [intent=etl availability=implemented stream=get_task_list]
    get team escalation policy list - Run the get team escalation policy ETL stream [intent=etl availability=implemented stream=get_team_escalation_policy]
    get team list - Run the get team ETL stream [intent=etl availability=implemented stream=get_team]
    get team on call schedule list - Run the get team on call schedule ETL stream [intent=etl availability=implemented stream=get_team_on_call_schedule]
    get team signal rule list - Run the get team signal rule ETL stream [intent=etl availability=implemented stream=get_team_signal_rule]
    get ticket list - Run the get ticket ETL stream [intent=etl availability=implemented stream=get_ticket]
    get ticketing field map list - Run the get ticketing field map ETL stream [intent=etl availability=implemented stream=get_ticketing_field_map]
    get ticketing form configuration list - Run the get ticketing form configuration ETL stream [intent=etl availability=implemented stream=get_ticketing_form_configuration]
    get ticketing priority list - Run the get ticketing priority ETL stream [intent=etl availability=implemented stream=get_ticketing_priority]
    get ticketing project config list - Run the get ticketing project config ETL stream [intent=etl availability=implemented stream=get_ticketing_project_config]
    get ticketing project list - Run the get ticketing project ETL stream [intent=etl availability=implemented stream=get_ticketing_project]
    get user list - Run the get user ETL stream [intent=etl availability=implemented stream=get_user]
    get vote status list - Run the get vote status ETL stream [intent=etl availability=implemented stream=get_vote_status]
    get webhook list - Run the get webhook ETL stream [intent=etl availability=implemented stream=get_webhook]
    get zendesk customer support issue list - Run the get zendesk customer support issue ETL stream [intent=etl availability=implemented stream=get_zendesk_customer_support_issue]
    incidents list - Run the incidents ETL stream [intent=etl availability=implemented stream=incidents]
    ingest catalog data apply - Plan and execute the ingest catalog data reverse-ETL action [intent=reverse_etl availability=implemented write=ingest_catalog_data]; approval: requires plan, preview, approval, and execute; risk: Ingest service catalog data through the FireHydrant API.; flags: --catalog_id (required)
    list alerts list - Run the list alerts ETL stream [intent=etl availability=implemented stream=list_alerts]
    list audience summaries list - Run the list audience summaries ETL stream [intent=etl availability=implemented stream=list_audience_summaries]
    list audiences list - Run the list audiences ETL stream [intent=etl availability=implemented stream=list_audiences]
    list audit events list - Run the list audit events ETL stream [intent=etl availability=implemented stream=list_audit_events]
    list authed providers list - Run the list authed providers ETL stream [intent=etl availability=implemented stream=list_authed_providers]
    list available inbound field maps list - Run the list available inbound field maps ETL stream [intent=etl availability=implemented stream=list_available_inbound_field_maps]
    list available ticketing field maps list - Run the list available ticketing field maps ETL stream [intent=etl availability=implemented stream=list_available_ticketing_field_maps]
    list aws cloudtrail batch events list - Run the list aws cloudtrail batch events ETL stream [intent=etl availability=implemented stream=list_aws_cloudtrail_batch_events]
    list aws cloudtrail batches list - Run the list aws cloudtrail batches ETL stream [intent=etl availability=implemented stream=list_aws_cloudtrail_batches]
    list aws connections list - Run the list aws connections ETL stream [intent=etl availability=implemented stream=list_aws_connections]
    list call routes list - Run the list call routes ETL stream [intent=etl availability=implemented stream=list_call_routes]
    list change events list - Run the list change events ETL stream [intent=etl availability=implemented stream=list_change_events]
    list change identities list - Run the list change identities ETL stream [intent=etl availability=implemented stream=list_change_identities]
    list change types list - Run the list change types ETL stream [intent=etl availability=implemented stream=list_change_types]
    list changes list - Run the list changes ETL stream [intent=etl availability=implemented stream=list_changes]
    list checklist templates list - Run the list checklist templates ETL stream [intent=etl availability=implemented stream=list_checklist_templates]
    list comment reactions list - Run the list comment reactions ETL stream [intent=etl availability=implemented stream=list_comment_reactions]
    list comments list - Run the list comments ETL stream [intent=etl availability=implemented stream=list_comments]
    list connection statuses by slug and id list - Run the list connection statuses by slug and id ETL stream [intent=etl availability=implemented stream=list_connection_statuses_by_slug_and_id]
    list connection statuses by slug list - Run the list connection statuses by slug ETL stream [intent=etl availability=implemented stream=list_connection_statuses_by_slug]
    list connection statuses list - Run the list connection statuses ETL stream [intent=etl availability=implemented stream=list_connection_statuses]
    list connections list - Run the list connections ETL stream [intent=etl availability=implemented stream=list_connections]
    list current user permissions list - Run the list current user permissions ETL stream [intent=etl availability=implemented stream=list_current_user_permissions]
    list custom field definitions list - Run the list custom field definitions ETL stream [intent=etl availability=implemented stream=list_custom_field_definitions]
    list custom field select options list - Run the list custom field select options ETL stream [intent=etl availability=implemented stream=list_custom_field_select_options]
    list email subscribers list - Run the list email subscribers ETL stream [intent=etl availability=implemented stream=list_email_subscribers]
    list entitlements list - Run the list entitlements ETL stream [intent=etl availability=implemented stream=list_entitlements]
    list environment functionalities list - Run the list environment functionalities ETL stream [intent=etl availability=implemented stream=list_environment_functionalities]
    list environment services list - Run the list environment services ETL stream [intent=etl availability=implemented stream=list_environment_services]
    list field map available fields list - Run the list field map available fields ETL stream [intent=etl availability=implemented stream=list_field_map_available_fields]
    list functionality environments list - Run the list functionality environments ETL stream [intent=etl availability=implemented stream=list_functionality_environments]
    list functionality services list - Run the list functionality services ETL stream [intent=etl availability=implemented stream=list_functionality_services]
    list inbound field maps list - Run the list inbound field maps ETL stream [intent=etl availability=implemented stream=list_inbound_field_maps]
    list incident alerts list - Run the list incident alerts ETL stream [intent=etl availability=implemented stream=list_incident_alerts]
    list incident attachments list - Run the list incident attachments ETL stream [intent=etl availability=implemented stream=list_incident_attachments]
    list incident change events list - Run the list incident change events ETL stream [intent=etl availability=implemented stream=list_incident_change_events]
    list incident conference bridges list - Run the list incident conference bridges ETL stream [intent=etl availability=implemented stream=list_incident_conference_bridges]
    list incident events list - Run the list incident events ETL stream [intent=etl availability=implemented stream=list_incident_events]
    list incident impacts list - Run the list incident impacts ETL stream [intent=etl availability=implemented stream=list_incident_impacts]
    list incident links list - Run the list incident links ETL stream [intent=etl availability=implemented stream=list_incident_links]
    list incident metrics list - Run the list incident metrics ETL stream [intent=etl availability=implemented stream=list_incident_metrics]
    list incident milestones list - Run the list incident milestones ETL stream [intent=etl availability=implemented stream=list_incident_milestones]
    list incident retrospectives list - Run the list incident retrospectives ETL stream [intent=etl availability=implemented stream=list_incident_retrospectives]
    list incident role assignments list - Run the list incident role assignments ETL stream [intent=etl availability=implemented stream=list_incident_role_assignments]
    list incident roles list - Run the list incident roles ETL stream [intent=etl availability=implemented stream=list_incident_roles]
    list incident status pages list - Run the list incident status pages ETL stream [intent=etl availability=implemented stream=list_incident_status_pages]
    list incident tags list - Run the list incident tags ETL stream [intent=etl availability=implemented stream=list_incident_tags]
    list incident tasks list - Run the list incident tasks ETL stream [intent=etl availability=implemented stream=list_incident_tasks]
    list incident types list - Run the list incident types ETL stream [intent=etl availability=implemented stream=list_incident_types]
    list infrastructure metrics list - Run the list infrastructure metrics ETL stream [intent=etl availability=implemented stream=list_infrastructure_metrics]
    list infrastructure type metrics list - Run the list infrastructure type metrics ETL stream [intent=etl availability=implemented stream=list_infrastructure_type_metrics]
    list infrastructures list - Run the list infrastructures ETL stream [intent=etl availability=implemented stream=list_infrastructures]
    list integrations list - Run the list integrations ETL stream [intent=etl availability=implemented stream=list_integrations]
    list lifecycle measurement definitions list - Run the list lifecycle measurement definitions ETL stream [intent=etl availability=implemented stream=list_lifecycle_measurement_definitions]
    list lifecycle phases list - Run the list lifecycle phases ETL stream [intent=etl availability=implemented stream=list_lifecycle_phases]
    list notification policy settings list - Run the list notification policy settings ETL stream [intent=etl availability=implemented stream=list_notification_policy_settings]
    list nunc connections list - Run the list nunc connections ETL stream [intent=etl availability=implemented stream=list_nunc_connections]
    list organization on call schedules list - Run the list organization on call schedules ETL stream [intent=etl availability=implemented stream=list_organization_on_call_schedules]
    list permissions list - Run the list permissions ETL stream [intent=etl availability=implemented stream=list_permissions]
    list post mortem questions list - Run the list post mortem questions ETL stream [intent=etl availability=implemented stream=list_post_mortem_questions]
    list post mortem reasons list - Run the list post mortem reasons ETL stream [intent=etl availability=implemented stream=list_post_mortem_reasons]
    list post mortem reports list - Run the list post mortem reports ETL stream [intent=etl availability=implemented stream=list_post_mortem_reports]
    list priorities list - Run the list priorities ETL stream [intent=etl availability=implemented stream=list_priorities]
    list processing log entries list - Run the list processing log entries ETL stream [intent=etl availability=implemented stream=list_processing_log_entries]
    list retrospective metrics list - Run the list retrospective metrics ETL stream [intent=etl availability=implemented stream=list_retrospective_metrics]
    list retrospective templates list - Run the list retrospective templates ETL stream [intent=etl availability=implemented stream=list_retrospective_templates]
    list retrospectives list - Run the list retrospectives ETL stream [intent=etl availability=implemented stream=list_retrospectives]
    list roles list - Run the list roles ETL stream [intent=etl availability=implemented stream=list_roles]
    list runbook actions list - Run the list runbook actions ETL stream [intent=etl availability=implemented stream=list_runbook_actions]
    list runbook executions list - Run the list runbook executions ETL stream [intent=etl availability=implemented stream=list_runbook_executions]
    list runbooks list - Run the list runbooks ETL stream [intent=etl availability=implemented stream=list_runbooks]
    list saved searches list - Run the list saved searches ETL stream [intent=etl availability=implemented stream=list_saved_searches]
    list scheduled maintenances list - Run the list scheduled maintenances ETL stream [intent=etl availability=implemented stream=list_scheduled_maintenances]
    list schedules list - Run the list schedules ETL stream [intent=etl availability=implemented stream=list_schedules]
    list service available downstream dependencies list - Run the list service available downstream dependencies ETL stream [intent=etl availability=implemented stream=list_service_available_downstream_dependencies]
    list service available upstream dependencies list - Run the list service available upstream dependencies ETL stream [intent=etl availability=implemented stream=list_service_available_upstream_dependencies]
    list service environments list - Run the list service environments ETL stream [intent=etl availability=implemented stream=list_service_environments]
    list severities list - Run the list severities ETL stream [intent=etl availability=implemented stream=list_severities]
    list severity matrix conditions list - Run the list severity matrix conditions ETL stream [intent=etl availability=implemented stream=list_severity_matrix_conditions]
    list severity matrix impacts list - Run the list severity matrix impacts ETL stream [intent=etl availability=implemented stream=list_severity_matrix_impacts]
    list signals alert grouping configurations list - Run the list signals alert grouping configurations ETL stream [intent=etl availability=implemented stream=list_signals_alert_grouping_configurations]
    list signals email targets list - Run the list signals email targets ETL stream [intent=etl availability=implemented stream=list_signals_email_targets]
    list signals event sources list - Run the list signals event sources ETL stream [intent=etl availability=implemented stream=list_signals_event_sources]
    list signals heartbeat endpoint configurations list - Run the list signals heartbeat endpoint configurations ETL stream [intent=etl availability=implemented stream=list_signals_heartbeat_endpoint_configurations]
    list signals transposers list - Run the list signals transposers ETL stream [intent=etl availability=implemented stream=list_signals_transposers]
    list signals webhook targets list - Run the list signals webhook targets ETL stream [intent=etl availability=implemented stream=list_signals_webhook_targets]
    list similar incidents list - Run the list similar incidents ETL stream [intent=etl availability=implemented stream=list_similar_incidents]
    list slack emoji actions list - Run the list slack emoji actions ETL stream [intent=etl availability=implemented stream=list_slack_emoji_actions]
    list slack usergroups list - Run the list slack usergroups ETL stream [intent=etl availability=implemented stream=list_slack_usergroups]
    list slack workspaces list - Run the list slack workspaces ETL stream [intent=etl availability=implemented stream=list_slack_workspaces]
    list status update templates list - Run the list status update templates ETL stream [intent=etl availability=implemented stream=list_status_update_templates]
    list statuspage connection pages list - Run the list statuspage connection pages ETL stream [intent=etl availability=implemented stream=list_statuspage_connection_pages]
    list statuspage connections list - Run the list statuspage connections ETL stream [intent=etl availability=implemented stream=list_statuspage_connections]
    list task lists list - Run the list task lists ETL stream [intent=etl availability=implemented stream=list_task_lists]
    list team call routes list - Run the list team call routes ETL stream [intent=etl availability=implemented stream=list_team_call_routes]
    list team escalation policies list - Run the list team escalation policies ETL stream [intent=etl availability=implemented stream=list_team_escalation_policies]
    list team on call schedules list - Run the list team on call schedules ETL stream [intent=etl availability=implemented stream=list_team_on_call_schedules]
    list team permissions list - Run the list team permissions ETL stream [intent=etl availability=implemented stream=list_team_permissions]
    list team signal rules list - Run the list team signal rules ETL stream [intent=etl availability=implemented stream=list_team_signal_rules]
    list ticket tags list - Run the list ticket tags ETL stream [intent=etl availability=implemented stream=list_ticket_tags]
    list ticketing custom definitions list - Run the list ticketing custom definitions ETL stream [intent=etl availability=implemented stream=list_ticketing_custom_definitions]
    list ticketing priorities list - Run the list ticketing priorities ETL stream [intent=etl availability=implemented stream=list_ticketing_priorities]
    list ticketing projects list - Run the list ticketing projects ETL stream [intent=etl availability=implemented stream=list_ticketing_projects]
    list tickets list - Run the list tickets ETL stream [intent=etl availability=implemented stream=list_tickets]
    list transcript entries list - Run the list transcript entries ETL stream [intent=etl availability=implemented stream=list_transcript_entries]
    list user involvement metrics list - Run the list user involvement metrics ETL stream [intent=etl availability=implemented stream=list_user_involvement_metrics]
    list user notification settings by user id list - Run the list user notification settings by user id ETL stream [intent=etl availability=implemented stream=list_user_notification_settings_by_user_id]
    list user owned services list - Run the list user owned services ETL stream [intent=etl availability=implemented stream=list_user_owned_services]
    list users list - Run the list users ETL stream [intent=etl availability=implemented stream=list_users]
    list webhook deliveries list - Run the list webhook deliveries ETL stream [intent=etl availability=implemented stream=list_webhook_deliveries]
    list webhooks list - Run the list webhooks ETL stream [intent=etl availability=implemented stream=list_webhooks]
    override on call schedule rotation shifts apply - Plan and execute the override on call schedule rotation shifts reverse-ETL action [intent=reverse_etl availability=implemented write=override_on_call_schedule_rotation_shifts]; approval: requires plan, preview, approval, and execute; risk: Override one or more shifts in an on-call rotation through the FireHydrant API.; flags: --rotation_id (required), --schedule_id (required), --team_id (required)
    preview on call schedule rotation apply - Plan and execute the preview on call schedule rotation reverse-ETL action [intent=reverse_etl availability=implemented write=preview_on_call_schedule_rotation]; approval: requires plan, preview, approval, and execute; risk: Preview an on-call rotation through the FireHydrant API.; flags: --schedule_id (required), --team_id (required)
    preview team on call schedule apply - Plan and execute the preview team on call schedule reverse-ETL action [intent=reverse_etl availability=implemented write=preview_team_on_call_schedule]; approval: requires plan, preview, approval, and execute; risk: Preview a new on-call schedule for a team through the FireHydrant API.; flags: --team_id (required)
    publish nunc connection apply - Plan and execute the publish nunc connection reverse-ETL action [intent=reverse_etl availability=implemented write=publish_nunc_connection]; approval: requires plan, preview, approval, and execute; risk: Publish a status page through the FireHydrant API.; flags: --nunc_connection_id (required)
    publish post mortem report apply - Plan and execute the publish post mortem report reverse-ETL action [intent=reverse_etl availability=implemented write=publish_post_mortem_report]; approval: requires plan, preview, approval, and execute; risk: Publish a retrospective report through the FireHydrant API.; flags: --report_id (required)
    refresh connection apply - Plan and execute the refresh connection reverse-ETL action [intent=reverse_etl availability=implemented write=refresh_connection]; approval: requires plan, preview, approval, and execute; risk: Refresh an integration connection's incident role schedules through the FireHydrant API.; flags: --connection_id (required), --slug (required)
    reorder post mortem reasons apply - Plan and execute the reorder post mortem reasons reverse-ETL action [intent=reverse_etl availability=implemented write=reorder_post_mortem_reasons]; approval: requires plan, preview, approval, and execute; risk: Reorder a contributing factor for a retrospective report through the FireHydrant API.; flags: --report_id (required)
    resolve incident apply - Plan and execute the resolve incident reverse-ETL action [intent=reverse_etl availability=implemented write=resolve_incident]; approval: requires plan, preview, approval, and execute; risk: Resolve an incident through the FireHydrant API.; flags: --incident_id (required)
    restore audience apply - Plan and execute the restore audience reverse-ETL action [intent=reverse_etl availability=implemented write=restore_audience]; approval: requires plan, preview, approval, and execute; risk: Restore audience through the FireHydrant API.; flags: --audience_id (required)
    search confluence spaces list - Run the search confluence spaces ETL stream [intent=etl availability=implemented stream=search_confluence_spaces]
    search slack channels list - Run the search slack channels ETL stream [intent=etl availability=implemented stream=search_slack_channels]
    search zendesk tickets list - Run the search zendesk tickets ETL stream [intent=etl availability=implemented stream=search_zendesk_tickets]
    services list - Run the services ETL stream [intent=etl availability=implemented stream=services]
    set member default audience apply - Plan and execute the set member default audience reverse-ETL action [intent=reverse_etl availability=implemented write=set_member_default_audience]; approval: requires plan, preview, approval, and execute; risk: Set default audience through the FireHydrant API.; flags: --member_id (required)
    share incident retrospectives apply - Plan and execute the share incident retrospectives reverse-ETL action [intent=reverse_etl availability=implemented write=share_incident_retrospectives]; approval: requires plan, preview, approval, and execute; risk: Share an incident's retrospective through the FireHydrant API.; flags: --incident_id (required)
    teams list - Run the teams ETL stream [intent=etl availability=implemented stream=teams]
    test slack channel apply - Plan and execute the test slack channel reverse-ETL action [intent=reverse_etl availability=implemented write=test_slack_channel]; approval: requires plan, preview, approval, and execute; risk: Test a Slack channel through the FireHydrant API.; flags: --id (required)
    unarchive incident apply - Plan and execute the unarchive incident reverse-ETL action [intent=reverse_etl availability=implemented write=unarchive_incident]; approval: requires plan, preview, approval, and execute; risk: Unarchive an incident through the FireHydrant API.; flags: --incident_id (required)
    unpublish nunc connection apply - Plan and execute the unpublish nunc connection reverse-ETL action [intent=reverse_etl availability=implemented write=unpublish_nunc_connection]; approval: requires plan, preview, approval, and execute; risk: Unpublish a status page through the FireHydrant API.; flags: --nunc_connection_id (required)
    update ai preferences apply - Plan and execute the update ai preferences reverse-ETL action [intent=reverse_etl availability=implemented write=update_ai_preferences]; approval: requires plan, preview, approval, and execute; risk: Update AI preferences through the FireHydrant API.
    update audience apply - Plan and execute the update audience reverse-ETL action [intent=reverse_etl availability=implemented write=update_audience]; approval: requires plan, preview, approval, and execute; risk: Update audience through the FireHydrant API.; flags: --audience_id (required)
    update authed provider apply - Plan and execute the update authed provider reverse-ETL action [intent=reverse_etl availability=implemented write=update_authed_provider]; approval: requires plan, preview, approval, and execute; risk: Get an authed provider through the FireHydrant API.; flags: --authed_provider_id (required), --connection_id (required), --integration_slug (required)
    update aws cloudtrail batch apply - Plan and execute the update aws cloudtrail batch reverse-ETL action [intent=reverse_etl availability=implemented write=update_aws_cloudtrail_batch]; approval: requires plan, preview, approval, and execute; risk: Update a CloudTrail batch through the FireHydrant API.; flags: --id (required)
    update aws connection apply - Plan and execute the update aws connection reverse-ETL action [intent=reverse_etl availability=implemented write=update_aws_connection]; approval: requires plan, preview, approval, and execute; risk: Update an AWS connection through the FireHydrant API.; flags: --id (required)
    update call route apply - Plan and execute the update call route reverse-ETL action [intent=reverse_etl availability=implemented write=update_call_route]; approval: requires plan, preview, approval, and execute; risk: Update a call route through the FireHydrant API.; flags: --id (required)
    update change apply - Plan and execute the update change reverse-ETL action [intent=reverse_etl availability=implemented write=update_change]; approval: requires plan, preview, approval, and execute; risk: Update a change entry through the FireHydrant API.; flags: --change_id (required)
    update change event apply - Plan and execute the update change event reverse-ETL action [intent=reverse_etl availability=implemented write=update_change_event]; approval: requires plan, preview, approval, and execute; risk: Update a change event through the FireHydrant API.; flags: --change_event_id (required)
    update change identity apply - Plan and execute the update change identity reverse-ETL action [intent=reverse_etl availability=implemented write=update_change_identity]; approval: requires plan, preview, approval, and execute; risk: Update an identity for a change entry through the FireHydrant API.; flags: --change_id (required), --identity_id (required)
    update checklist template apply - Plan and execute the update checklist template reverse-ETL action [intent=reverse_etl availability=implemented write=update_checklist_template]; approval: requires plan, preview, approval, and execute; risk: Update a checklist template through the FireHydrant API.; flags: --id (required)
    update comment apply - Plan and execute the update comment reverse-ETL action [intent=reverse_etl availability=implemented write=update_comment]; approval: requires plan, preview, approval, and execute; risk: Update a conversation comment through the FireHydrant API.; flags: --comment_id (required), --conversation_id (required)
    update connection apply - Plan and execute the update connection reverse-ETL action [intent=reverse_etl availability=implemented write=update_connection]; approval: requires plan, preview, approval, and execute; risk: Update an integration connection through the FireHydrant API.; flags: --connection_id (required), --slug (required)
    update custom field definition apply - Plan and execute the update custom field definition reverse-ETL action [intent=reverse_etl availability=implemented write=update_custom_field_definition]; approval: requires plan, preview, approval, and execute; risk: Update a custom field definition through the FireHydrant API.; flags: --field_id (required)
    update environment apply - Plan and execute the update environment reverse-ETL action [intent=reverse_etl availability=implemented write=update_environment]; approval: requires plan, preview, approval, and execute; risk: Update an environment through the FireHydrant API.; flags: --environment_id (required)
    update field map apply - Plan and execute the update field map reverse-ETL action [intent=reverse_etl availability=implemented write=update_field_map]; approval: requires plan, preview, approval, and execute; risk: Update field mapping through the FireHydrant API.; flags: --field_map_id (required)
    update functionality apply - Plan and execute the update functionality reverse-ETL action [intent=reverse_etl availability=implemented write=update_functionality]; approval: requires plan, preview, approval, and execute; risk: Update a functionality through the FireHydrant API.; flags: --functionality_id (required)
    update inbound field map apply - Plan and execute the update inbound field map reverse-ETL action [intent=reverse_etl availability=implemented write=update_inbound_field_map]; approval: requires plan, preview, approval, and execute; risk: Update inbound field map for a ticketing project through the FireHydrant API.; flags: --map_id (required), --ticketing_project_id (required)
    update incident alert primary apply - Plan and execute the update incident alert primary reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_alert_primary]; approval: requires plan, preview, approval, and execute; risk: Set an alert as primary for an incident through the FireHydrant API.; flags: --incident_alert_id (required), --incident_id (required)
    update incident apply - Plan and execute the update incident reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident]; approval: requires plan, preview, approval, and execute; risk: Update an incident through the FireHydrant API.; flags: --incident_id (required)
    update incident change event apply - Plan and execute the update incident change event reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_change_event]; approval: requires plan, preview, approval, and execute; risk: Update a change attached to an incident through the FireHydrant API.; flags: --incident_id (required), --related_change_event_id (required)
    update incident chat message apply - Plan and execute the update incident chat message reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_chat_message]; approval: requires plan, preview, approval, and execute; risk: Update a chat message on an incident through the FireHydrant API.; flags: --incident_id (required), --message_id (required)
    update incident event apply - Plan and execute the update incident event reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_event]; approval: requires plan, preview, approval, and execute; risk: Update an incident event through the FireHydrant API.; flags: --event_id (required), --incident_id (required)
    update incident impact patch apply - Plan and execute the update incident impact patch reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_impact_patch]; approval: requires plan, preview, approval, and execute; risk: Update impacts for an incident through the FireHydrant API.; flags: --incident_id (required)
    update incident impact put apply - Plan and execute the update incident impact put reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_impact_put]; approval: requires plan, preview, approval, and execute; risk: Update impacts for an incident through the FireHydrant API.; flags: --incident_id (required)
    update incident link apply - Plan and execute the update incident link reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_link]; approval: requires plan, preview, approval, and execute; risk: Update the external incident link through the FireHydrant API.; flags: --incident_id (required), --link_id (required)
    update incident note apply - Plan and execute the update incident note reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_note]; approval: requires plan, preview, approval, and execute; risk: Update a note through the FireHydrant API.; flags: --incident_id (required), --note_id (required)
    update incident retrospective apply - Plan and execute the update incident retrospective reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_retrospective]; approval: requires plan, preview, approval, and execute; risk: Update a retrospective on the incident through the FireHydrant API.; flags: --incident_id (required), --retrospective_id (required)
    update incident retrospective field apply - Plan and execute the update incident retrospective field reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_retrospective_field]; approval: requires plan, preview, approval, and execute; risk: Update the value on a retrospective field through the FireHydrant API.; flags: --field_id (required), --incident_id (required), --retrospective_id (required)
    update incident role apply - Plan and execute the update incident role reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_role]; approval: requires plan, preview, approval, and execute; risk: Update an incident role through the FireHydrant API.; flags: --incident_role_id (required)
    update incident task apply - Plan and execute the update incident task reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_task]; approval: requires plan, preview, approval, and execute; risk: Update an incident task through the FireHydrant API.; flags: --incident_id (required), --task_id (required)
    update incident type apply - Plan and execute the update incident type reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident_type]; approval: requires plan, preview, approval, and execute; risk: Update an incident type through the FireHydrant API.; flags: --id (required)
    update lifecycle measurement definition apply - Plan and execute the update lifecycle measurement definition reverse-ETL action [intent=reverse_etl availability=implemented write=update_lifecycle_measurement_definition]; approval: requires plan, preview, approval, and execute; risk: Update a measurement definition through the FireHydrant API.; flags: --measurement_definition_id (required)
    update lifecycle milestone apply - Plan and execute the update lifecycle milestone reverse-ETL action [intent=reverse_etl availability=implemented write=update_lifecycle_milestone]; approval: requires plan, preview, approval, and execute; risk: Update a milestone through the FireHydrant API.; flags: --milestone_id (required)
    update notification policy apply - Plan and execute the update notification policy reverse-ETL action [intent=reverse_etl availability=implemented write=update_notification_policy]; approval: requires plan, preview, approval, and execute; risk: Update a notification policy through the FireHydrant API.; flags: --id (required)
    update nunc component group apply - Plan and execute the update nunc component group reverse-ETL action [intent=reverse_etl availability=implemented write=update_nunc_component_group]; approval: requires plan, preview, approval, and execute; risk: Update a status page component group through the FireHydrant API.; flags: --group_id (required), --nunc_connection_id (required)
    update nunc connection apply - Plan and execute the update nunc connection reverse-ETL action [intent=reverse_etl availability=implemented write=update_nunc_connection]; approval: requires plan, preview, approval, and execute; risk: Update a status page through the FireHydrant API.; flags: --nunc_connection_id (required)
    update nunc link apply - Plan and execute the update nunc link reverse-ETL action [intent=reverse_etl availability=implemented write=update_nunc_link]; approval: requires plan, preview, approval, and execute; risk: Update a status page link through the FireHydrant API.; flags: --link_id (required), --nunc_connection_id (required)
    update on call schedule rotation apply - Plan and execute the update on call schedule rotation reverse-ETL action [intent=reverse_etl availability=implemented write=update_on_call_schedule_rotation]; approval: requires plan, preview, approval, and execute; risk: Update an on-call schedule's rotation through the FireHydrant API.; flags: --rotation_id (required), --schedule_id (required), --team_id (required)
    update on call shift apply - Plan and execute the update on call shift reverse-ETL action [intent=reverse_etl availability=implemented write=update_on_call_shift]; approval: requires plan, preview, approval, and execute; risk: Update an on-call shift for a team schedule through the FireHydrant API.; flags: --id (required), --schedule_id (required), --team_id (required)
    update post mortem field apply - Plan and execute the update post mortem field reverse-ETL action [intent=reverse_etl availability=implemented write=update_post_mortem_field]; approval: requires plan, preview, approval, and execute; risk: Update a retrospective field through the FireHydrant API.; flags: --field_id (required), --report_id (required)
    update post mortem questions apply - Plan and execute the update post mortem questions reverse-ETL action [intent=reverse_etl availability=implemented write=update_post_mortem_questions]; approval: requires plan, preview, approval, and execute; risk: Update retrospective questions through the FireHydrant API.
    update post mortem reason apply - Plan and execute the update post mortem reason reverse-ETL action [intent=reverse_etl availability=implemented write=update_post_mortem_reason]; approval: requires plan, preview, approval, and execute; risk: Update a contributing factor in a retrospective report through the FireHydrant API.; flags: --reason_id (required), --report_id (required)
    update post mortem report apply - Plan and execute the update post mortem report reverse-ETL action [intent=reverse_etl availability=implemented write=update_post_mortem_report]; approval: requires plan, preview, approval, and execute; risk: Update a retrospective report through the FireHydrant API.; flags: --report_id (required)
    update priority apply - Plan and execute the update priority reverse-ETL action [intent=reverse_etl availability=implemented write=update_priority]; approval: requires plan, preview, approval, and execute; risk: Update a priority through the FireHydrant API.; flags: --priority_slug (required)
    update retrospective template apply - Plan and execute the update retrospective template reverse-ETL action [intent=reverse_etl availability=implemented write=update_retrospective_template]; approval: requires plan, preview, approval, and execute; risk: Update a retrospective template through the FireHydrant API.; flags: --retrospective_template_id (required)
    update role apply - Plan and execute the update role reverse-ETL action [intent=reverse_etl availability=implemented write=update_role]; approval: requires plan, preview, approval, and execute; risk: Update a role through the FireHydrant API.; flags: --id (required)
    update runbook apply - Plan and execute the update runbook reverse-ETL action [intent=reverse_etl availability=implemented write=update_runbook]; approval: requires plan, preview, approval, and execute; risk: Update a runbook through the FireHydrant API.; flags: --runbook_id (required)
    update runbook execution step apply - Plan and execute the update runbook execution step reverse-ETL action [intent=reverse_etl availability=implemented write=update_runbook_execution_step]; approval: requires plan, preview, approval, and execute; risk: Update a runbook step execution through the FireHydrant API.; flags: --execution_id (required), --step_id (required)
    update runbook execution step script apply - Plan and execute the update runbook execution step script reverse-ETL action [intent=reverse_etl availability=implemented write=update_runbook_execution_step_script]; approval: requires plan, preview, approval, and execute; risk: Update a script step's execution status through the FireHydrant API.; flags: --execution_id (required), --state (required), --step_id (required)
    update saved search apply - Plan and execute the update saved search reverse-ETL action [intent=reverse_etl availability=implemented write=update_saved_search]; approval: requires plan, preview, approval, and execute; risk: Update a saved search through the FireHydrant API.; flags: --resource_type (required), --saved_search_id (required)
    update scheduled maintenance apply - Plan and execute the update scheduled maintenance reverse-ETL action [intent=reverse_etl availability=implemented write=update_scheduled_maintenance]; approval: requires plan, preview, approval, and execute; risk: Update a scheduled maintenance event through the FireHydrant API.; flags: --scheduled_maintenance_id (required)
    update service apply - Plan and execute the update service reverse-ETL action [intent=reverse_etl availability=implemented write=update_service]; approval: requires plan, preview, approval, and execute; risk: Update a service through the FireHydrant API.; flags: --service_id (required)
    update service dependency apply - Plan and execute the update service dependency reverse-ETL action [intent=reverse_etl availability=implemented write=update_service_dependency]; approval: requires plan, preview, approval, and execute; risk: Update a service dependency through the FireHydrant API.; flags: --service_dependency_id (required)
    update severity apply - Plan and execute the update severity reverse-ETL action [intent=reverse_etl availability=implemented write=update_severity]; approval: requires plan, preview, approval, and execute; risk: Update a severity through the FireHydrant API.; flags: --severity_slug (required)
    update severity matrix apply - Plan and execute the update severity matrix reverse-ETL action [intent=reverse_etl availability=implemented write=update_severity_matrix]; approval: requires plan, preview, approval, and execute; risk: Update severity matrix through the FireHydrant API.
    update severity matrix condition apply - Plan and execute the update severity matrix condition reverse-ETL action [intent=reverse_etl availability=implemented write=update_severity_matrix_condition]; approval: requires plan, preview, approval, and execute; risk: Update a severity matrix condition through the FireHydrant API.; flags: --condition_id (required)
    update severity matrix impact apply - Plan and execute the update severity matrix impact reverse-ETL action [intent=reverse_etl availability=implemented write=update_severity_matrix_impact]; approval: requires plan, preview, approval, and execute; risk: Update a severity matrix impact through the FireHydrant API.; flags: --impact_id (required)
    update signals alert apply - Plan and execute the update signals alert reverse-ETL action [intent=reverse_etl availability=implemented write=update_signals_alert]; approval: requires plan, preview, approval, and execute; risk: Update a Signal alert through the FireHydrant API.; flags: --id (required)
    update signals alert grouping configuration apply - Plan and execute the update signals alert grouping configuration reverse-ETL action [intent=reverse_etl availability=implemented write=update_signals_alert_grouping_configuration]; approval: requires plan, preview, approval, and execute; risk: Update an alert grouping configuration. through the FireHydrant API.; flags: --id (required)
    update signals email target apply - Plan and execute the update signals email target reverse-ETL action [intent=reverse_etl availability=implemented write=update_signals_email_target]; approval: requires plan, preview, approval, and execute; risk: Update an email target through the FireHydrant API.; flags: --id (required)
    update signals heartbeat endpoint configuration apply - Plan and execute the update signals heartbeat endpoint configuration reverse-ETL action [intent=reverse_etl availability=implemented write=update_signals_heartbeat_endpoint_configuration]; approval: requires plan, preview, approval, and execute; risk: Update a heartbeat endpoint configuration through the FireHydrant API.; flags: --id (required)
    update signals webhook target apply - Plan and execute the update signals webhook target reverse-ETL action [intent=reverse_etl availability=implemented write=update_signals_webhook_target]; approval: requires plan, preview, approval, and execute; risk: Update a webhook target through the FireHydrant API.; flags: --id (required)
    update slack emoji action apply - Plan and execute the update slack emoji action reverse-ETL action [intent=reverse_etl availability=implemented write=update_slack_emoji_action]; approval: requires plan, preview, approval, and execute; risk: Update a Slack emoji action through the FireHydrant API.; flags: --connection_id (required), --emoji_action_id (required)
    update status update template apply - Plan and execute the update status update template reverse-ETL action [intent=reverse_etl availability=implemented write=update_status_update_template]; approval: requires plan, preview, approval, and execute; risk: Update a status update template through the FireHydrant API.; flags: --status_update_template_id (required)
    update statuspage connection apply - Plan and execute the update statuspage connection reverse-ETL action [intent=reverse_etl availability=implemented write=update_statuspage_connection]; approval: requires plan, preview, approval, and execute; risk: Update a Statuspage connection through the FireHydrant API.; flags: --connection_id (required)
    update support hours schedule apply - Plan and execute the update support hours schedule reverse-ETL action [intent=reverse_etl availability=implemented write=update_support_hours_schedule]; approval: requires plan, preview, approval, and execute; risk: Update support hours schedule through the FireHydrant API.; flags: --team_id (required)
    update task list apply - Plan and execute the update task list reverse-ETL action [intent=reverse_etl availability=implemented write=update_task_list]; approval: requires plan, preview, approval, and execute; risk: Update a task list through the FireHydrant API.; flags: --task_list_id (required)
    update team apply - Plan and execute the update team reverse-ETL action [intent=reverse_etl availability=implemented write=update_team]; approval: requires plan, preview, approval, and execute; risk: Update a team through the FireHydrant API.; flags: --team_id (required)
    update team escalation policy apply - Plan and execute the update team escalation policy reverse-ETL action [intent=reverse_etl availability=implemented write=update_team_escalation_policy]; approval: requires plan, preview, approval, and execute; risk: Update an escalation policy for a team through the FireHydrant API.; flags: --id (required), --team_id (required)
    update team on call schedule apply - Plan and execute the update team on call schedule reverse-ETL action [intent=reverse_etl availability=implemented write=update_team_on_call_schedule]; approval: requires plan, preview, approval, and execute; risk: Update an on-call schedule for a team through the FireHydrant API.; flags: --schedule_id (required), --team_id (required)
    update team signal rule apply - Plan and execute the update team signal rule reverse-ETL action [intent=reverse_etl availability=implemented write=update_team_signal_rule]; approval: requires plan, preview, approval, and execute; risk: Update a Signals rule through the FireHydrant API.; flags: --id (required), --team_id (required)
    update ticket apply - Plan and execute the update ticket reverse-ETL action [intent=reverse_etl availability=implemented write=update_ticket]; approval: requires plan, preview, approval, and execute; risk: Update a ticket through the FireHydrant API.; flags: --ticket_id (required)
    update ticketing custom definition apply - Plan and execute the update ticketing custom definition reverse-ETL action [intent=reverse_etl availability=implemented write=update_ticketing_custom_definition]; approval: requires plan, preview, approval, and execute; risk: Update a ticketing custom field through the FireHydrant API.; flags: --field_id (required)
    update ticketing field map apply - Plan and execute the update ticketing field map reverse-ETL action [intent=reverse_etl availability=implemented write=update_ticketing_field_map]; approval: requires plan, preview, approval, and execute; risk: Update a field map for a ticketing project through the FireHydrant API.; flags: --map_id (required), --ticketing_project_id (required)
    update ticketing priority apply - Plan and execute the update ticketing priority reverse-ETL action [intent=reverse_etl availability=implemented write=update_ticketing_priority]; approval: requires plan, preview, approval, and execute; risk: Update a ticketing priority through the FireHydrant API.; flags: --id (required)
    update ticketing project config apply - Plan and execute the update ticketing project config reverse-ETL action [intent=reverse_etl availability=implemented write=update_ticketing_project_config]; approval: requires plan, preview, approval, and execute; risk: Update configuration for a ticketing project through the FireHydrant API.; flags: --config_id (required), --ticketing_project_id (required)
    update transcript attribution apply - Plan and execute the update transcript attribution reverse-ETL action [intent=reverse_etl availability=implemented write=update_transcript_attribution]; approval: requires plan, preview, approval, and execute; risk: Update the attribution of a transcript through the FireHydrant API.; flags: --incident_id (required)
    update vote apply - Plan and execute the update vote reverse-ETL action [intent=reverse_etl availability=implemented write=update_vote]; approval: requires plan, preview, approval, and execute; risk: Update votes through the FireHydrant API.; flags: --event_id (required), --incident_id (required)
    update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: Update a webhook through the FireHydrant API.; flags: --webhook_id (required)
    validate incident tags apply - Plan and execute the validate incident tags reverse-ETL action [intent=reverse_etl availability=implemented write=validate_incident_tags]; approval: requires plan, preview, approval, and execute; risk: Validate incident tags through the FireHydrant API.
    vote ai incident summary apply - Plan and execute the vote ai incident summary reverse-ETL action [intent=reverse_etl availability=implemented write=vote_ai_incident_summary]; approval: requires plan, preview, approval, and execute; risk: Vote on an AI-generated incident summary through the FireHydrant API.; flags: --generated_summary_id (required), --incident_id (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect firehydrant

  # Inspect as structured JSON
  pm connectors inspect firehydrant --json

AGENT WORKFLOW
  - Run pm connectors inspect firehydrant before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
