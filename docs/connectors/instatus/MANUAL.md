# pm connectors inspect instatus

```text
NAME
  pm connectors inspect instatus - Instatus connector manual

SYNOPSIS
  pm connectors inspect instatus
  pm connectors inspect instatus --json
  pm credentials add <name> --connector instatus [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and manages Instatus status pages, components, incidents, maintenances, templates, subscribers, metrics, monitors, and related status-page resources through the Instatus REST API.

ICON
  asset: icons/instatus.svg
  source: official
  review_status: official_verified
  review_url: https://instatus.com/help/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  audience_group_id
  base_url
  component_id
  generic_notice_id
  incident_id
  incident_update_id
  maintenance_id
  maintenance_update_id
  metric_id
  monitor_id
  on_call_schedule_id
  page_id
  page_size
  subscriber_search
  template_id
  api_key (secret)

ETL STREAMS
  pages:
    primary key: id
    fields: createdAt(), customDomain(), id(), language(), name(), publicEmail(), status(), subdomain(), updatedAt(), websiteUrl()
  components:
    primary key: id
    fields: description(), group(), id(), name(), order(), showUptime(), status(), uniqueEmail()
  incidents:
    primary key: id
    fields: createdAt(), id(), name(), resolved(), started(), status(), updatedAt()
  maintenances:
    primary key: id
    fields: autoEnd(), autoStart(), duration(), id(), name(), start(), status()
  workspaces:
    primary key: id
    fields: id()
  component_detail:
    primary key: id
    fields: id()
  incident_detail:
    primary key: id
    fields: id()
  incident_update_detail:
    primary key: id
    fields: id()
  maintenance_detail:
    primary key: id
    fields: id()
  maintenance_update_detail:
    primary key: id
    fields: id()
  templates:
    primary key: id
    fields: id()
  template_detail:
    primary key: id
    fields: id()
  teammates:
    primary key: id
    fields: id()
  subscribers:
    primary key: id
    fields: id()
  metrics:
    primary key: id
    fields: id()
  metric_detail:
    primary key: id
    fields: id()
  user_profile:
    primary key: id
    fields: id()
  audience_groups:
    primary key: id
    fields: id()
  audience_group_detail:
    primary key: id
    fields: id()
  generic_notices:
    primary key: id
    fields: id()
  generic_notice_detail:
    primary key: id
    fields: id()
  monitors:
    primary key: id
    fields: id()
  monitor_inserted_logs_check:
    fields: id()
  monitor_logs:
    primary key: id
    fields: id()
  monitor_alerts:
    primary key: id
    fields: id()
  routing_rules:
    primary key: id
    fields: id()
  escalation_policies:
    primary key: id
    fields: id()
  on_call_schedule_members:
    primary key: id
    fields: id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_page:
    endpoint: POST /v1/pages
    risk: Mutates Instatus status pages; verify target ids and payload before execution.
  update_page:
    endpoint: PUT /v2/{{ record.page_id }}
    required fields: page_id
    risk: Mutates Instatus status pages; verify target ids and payload before execution.
  delete_page:
    endpoint: DELETE /v2/{{ record.page_id }}
    required fields: page_id
    risk: Mutates Instatus status pages; verify target ids and payload before execution.
  create_workspace:
    endpoint: POST /v1/workspaces
    risk: Mutates Instatus workspaces; verify target ids and payload before execution.
  delete_workspace:
    endpoint: DELETE /v1/workspaces/{{ record.workspace_id }}
    required fields: workspace_id
    risk: Mutates Instatus workspaces; verify target ids and payload before execution.
  create_component:
    endpoint: POST /v1/{{ record.page_id }}/components
    required fields: page_id
    risk: Mutates Instatus components; verify target ids and payload before execution.
  update_component:
    endpoint: PUT /v2/{{ record.page_id }}/components/{{ record.component_id }}
    required fields: page_id, component_id
    risk: Mutates Instatus components; verify target ids and payload before execution.
  delete_component:
    endpoint: DELETE /v1/{{ record.page_id }}/components/{{ record.component_id }}
    required fields: page_id, component_id
    risk: Mutates Instatus components; verify target ids and payload before execution.
  add_team_member:
    endpoint: POST /v1/{{ record.page_id }}/team
    required fields: page_id
    risk: Mutates Instatus team members; verify target ids and payload before execution.
  delete_team_member:
    endpoint: DELETE /v1/{{ record.page_id }}/team/{{ record.member_id }}
    required fields: page_id, member_id
    risk: Mutates Instatus team members; verify target ids and payload before execution.
  add_incident:
    endpoint: POST /v1/{{ record.page_id }}/incidents
    required fields: page_id
    risk: Mutates Instatus incidents; verify target ids and payload before execution.
  add_incident_with_template:
    endpoint: POST /v1/{{ record.page_id }}/incidents/{{ record.template }}
    required fields: page_id, template
    risk: Mutates Instatus incidents; verify target ids and payload before execution.
  update_incident:
    endpoint: PUT /v1/{{ record.page_id }}/incidents/{{ record.incident_id }}
    required fields: page_id, incident_id
    risk: Mutates Instatus incidents; verify target ids and payload before execution.
  delete_incident:
    endpoint: DELETE /v1/{{ record.page_id }}/incidents/{{ record.incident_id }}
    required fields: page_id, incident_id
    risk: Mutates Instatus incidents; verify target ids and payload before execution.
  add_incident_update:
    endpoint: POST /v1/{{ record.page_id }}/incidents/{{ record.incident_id }}/incident-updates
    required fields: page_id, incident_id
    risk: Mutates Instatus incident updates; verify target ids and payload before execution.
  resolve_incident_with_template:
    endpoint: POST /v2/{{ record.page_id }}/incidents/{{ record.incident_id }}/incident-updates/{{ record.template }}
    required fields: page_id, incident_id, template
    risk: Mutates Instatus incident updates; verify target ids and payload before execution.
  update_incident_update:
    endpoint: PUT /v1/{{ record.page_id }}/incidents/{{ record.incident_id }}/incident-updates/{{ record.incident_update_id }}
    required fields: page_id, incident_id, incident_update_id
    risk: Mutates Instatus incident updates; verify target ids and payload before execution.
  delete_incident_update:
    endpoint: DELETE /v1/{{ record.page_id }}/incidents/{{ record.incident_id }}/incident-updates/{{ record.incident_update_id }}
    required fields: page_id, incident_id, incident_update_id
    risk: Mutates Instatus incident updates; verify target ids and payload before execution.
  add_maintenance:
    endpoint: POST /v1/{{ record.page_id }}/maintenances
    required fields: page_id
    risk: Mutates Instatus maintenances; verify target ids and payload before execution.
  update_maintenance:
    endpoint: PUT /v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}
    required fields: page_id, maintenance_id
    risk: Mutates Instatus maintenances; verify target ids and payload before execution.
  delete_maintenance:
    endpoint: DELETE /v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}
    required fields: page_id, maintenance_id
    risk: Mutates Instatus maintenances; verify target ids and payload before execution.
  add_maintenance_update:
    endpoint: POST /v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}/maintenance-updates
    required fields: page_id, maintenance_id
    risk: Mutates Instatus maintenance updates; verify target ids and payload before execution.
  update_maintenance_update:
    endpoint: PUT /v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}/maintenance-updates/{{ record.maintenance_update_id }}
    required fields: page_id, maintenance_id, maintenance_update_id
    risk: Mutates Instatus maintenance updates; verify target ids and payload before execution.
  delete_maintenance_update:
    endpoint: DELETE /v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}/maintenance-updates/{{ record.maintenance_update_id }}
    required fields: page_id, maintenance_id, maintenance_update_id
    risk: Mutates Instatus maintenance updates; verify target ids and payload before execution.
  add_template:
    endpoint: POST /v1/{{ record.page_id }}/templates
    required fields: page_id
    risk: Mutates Instatus templates; verify target ids and payload before execution.
  update_template:
    endpoint: PUT /v1/{{ record.page_id }}/templates/{{ record.template_id }}
    required fields: page_id, template_id
    risk: Mutates Instatus templates; verify target ids and payload before execution.
  delete_template:
    endpoint: DELETE /v1/{{ record.page_id }}/templates/{{ record.template_id }}
    required fields: page_id, template_id
    risk: Mutates Instatus templates; verify target ids and payload before execution.
  add_generic_notice:
    endpoint: POST /v1/{{ record.page_id }}/generic-notices
    required fields: page_id
    risk: Mutates Instatus general notices; verify target ids and payload before execution.
  update_generic_notice:
    endpoint: PUT /v1/{{ record.page_id }}/generic-notices/{{ record.generic_notice_id }}
    required fields: page_id, generic_notice_id
    risk: Mutates Instatus general notices; verify target ids and payload before execution.
  delete_generic_notice:
    endpoint: DELETE /v1/{{ record.page_id }}/generic-notices/{{ record.generic_notice_id }}
    required fields: page_id, generic_notice_id
    risk: Mutates Instatus general notices; verify target ids and payload before execution.
  add_subscriber:
    endpoint: POST /v1/{{ record.page_id }}/subscribers
    required fields: page_id
    risk: Mutates Instatus subscribers; verify target ids and payload before execution.
  add_subscribers_bulk:
    endpoint: POST /v1/{{ record.page_id }}/subscribers/bulk
    required fields: page_id
    risk: Mutates Instatus subscribers; verify target ids and payload before execution.
  delete_subscriber:
    endpoint: DELETE /v1/{{ record.page_id }}/subscribers/{{ record.subscriber_id }}
    required fields: page_id, subscriber_id
    risk: Mutates Instatus subscribers; verify target ids and payload before execution.
  add_metric:
    endpoint: POST /v1/{{ record.page_id }}/metrics
    required fields: page_id
    risk: Mutates Instatus metrics; verify target ids and payload before execution.
  update_metric:
    endpoint: PUT /v1/{{ record.page_id }}/metrics/{{ record.metric_id }}
    required fields: page_id, metric_id
    risk: Mutates Instatus metrics; verify target ids and payload before execution.
  delete_metric:
    endpoint: DELETE /v1/{{ record.page_id }}/metrics/{{ record.metric_id }}
    required fields: page_id, metric_id
    risk: Mutates Instatus metrics; verify target ids and payload before execution.
  add_metric_datapoint:
    endpoint: POST /v1/{{ record.page_id }}/metrics/{{ record.metric_id }}
    required fields: page_id, metric_id
    risk: Mutates Instatus metric datapoints; verify target ids and payload before execution.
  add_metric_datapoints:
    endpoint: POST /v1/{{ record.page_id }}/metrics/{{ record.metric_id }}/data
    required fields: page_id, metric_id
    risk: Mutates Instatus metric datapoints; verify target ids and payload before execution.
  delete_metric_datapoints:
    endpoint: DELETE /v1/{{ record.page_id }}/metrics/{{ record.metric_id }}/data
    required fields: page_id, metric_id
    risk: Mutates Instatus metric datapoints; verify target ids and payload before execution.
  create_audience_group:
    endpoint: POST /v1/{{ record.page_id }}/audience_groups
    required fields: page_id
    risk: Mutates Instatus audience groups; verify target ids and payload before execution.
  update_audience_group:
    endpoint: PUT /v1/{{ record.page_id }}/audience_groups/{{ record.audience_group_id }}
    required fields: page_id, audience_group_id
    risk: Mutates Instatus audience groups; verify target ids and payload before execution.
  delete_audience_group:
    endpoint: DELETE /v1/{{ record.page_id }}/audience-groups/{{ record.audience_group_id }}
    required fields: page_id, audience_group_id
    risk: Mutates Instatus audience groups; verify target ids and payload before execution.
  regenerate_secure_link:
    endpoint: POST /v1/{{ record.page_id }}/regenerate-secure-link
    required fields: page_id
    risk: Mutates Instatus private-page secure links; verify target ids and payload before execution.
  create_monitor:
    endpoint: POST /monitors
    risk: Mutates Instatus monitors; verify target ids and payload before execution.
  update_monitor:
    endpoint: PUT /monitors/{{ record.monitor_id }}
    required fields: monitor_id
    risk: Mutates Instatus monitors; verify target ids and payload before execution.
  delete_monitor:
    endpoint: DELETE /monitors/{{ record.monitor_id }}
    required fields: monitor_id
    risk: Mutates Instatus monitors; verify target ids and payload before execution.
  update_monitor_group_assignment:
    endpoint: PUT /monitors/{{ record.monitor_id }}/group
    required fields: monitor_id
    risk: Mutates Instatus monitor group assignments; verify target ids and payload before execution.
  create_monitor_alert:
    endpoint: POST /monitor-alerts
    risk: Mutates Instatus monitor alerts; verify target ids and payload before execution.
  update_monitor_alert:
    endpoint: PUT /monitor-alerts/{{ record.monitor_alert_id }}
    required fields: monitor_alert_id
    risk: Mutates Instatus monitor alerts; verify target ids and payload before execution.
  delete_monitor_alert:
    endpoint: DELETE /monitor-alerts/{{ record.monitor_alert_id }}
    required fields: monitor_alert_id
    risk: Mutates Instatus monitor alerts; verify target ids and payload before execution.
  create_monitors_group:
    endpoint: POST /monitors-groups
    risk: Mutates Instatus monitor groups; verify target ids and payload before execution.
  update_monitors_group:
    endpoint: PUT /monitors-groups/{{ record.monitors_group_id }}
    required fields: monitors_group_id
    risk: Mutates Instatus monitor groups; verify target ids and payload before execution.
  delete_monitors_group:
    endpoint: DELETE /monitors-groups/{{ record.monitors_group_id }}
    required fields: monitors_group_id
    risk: Mutates Instatus monitor groups; verify target ids and payload before execution.
  add_monitors_to_group:
    endpoint: POST /monitors-groups/{{ record.monitors_group_id }}/monitors
    required fields: monitors_group_id
    risk: Mutates Instatus monitor groups; verify target ids and payload before execution.
  create_routing_rule:
    endpoint: POST /routing-rules
    risk: Mutates Instatus routing rules; verify target ids and payload before execution.
  update_routing_rule:
    endpoint: PUT /routing-rules/{{ record.routing_rule_id }}
    required fields: routing_rule_id
    risk: Mutates Instatus routing rules; verify target ids and payload before execution.
  delete_routing_rule:
    endpoint: DELETE /routing-rules/{{ record.routing_rule_id }}
    required fields: routing_rule_id
    risk: Mutates Instatus routing rules; verify target ids and payload before execution.
  create_escalation_policy:
    endpoint: POST /escalation-policies
    risk: Mutates Instatus escalation policies; verify target ids and payload before execution.
  update_escalation_policy:
    endpoint: PUT /escalation-policies/{{ record.escalation_policy_id }}
    required fields: escalation_policy_id
    risk: Mutates Instatus escalation policies; verify target ids and payload before execution.
  delete_escalation_policy:
    endpoint: DELETE /escalation-policies/{{ record.escalation_policy_id }}
    required fields: escalation_policy_id
    risk: Mutates Instatus escalation policies; verify target ids and payload before execution.
  create_on_call_schedule:
    endpoint: POST /on-call-schedules
    risk: Mutates Instatus on-call schedules; verify target ids and payload before execution.
  update_on_call_schedule:
    endpoint: PUT /on-call-schedules/{{ record.on_call_schedule_id }}
    required fields: on_call_schedule_id
    risk: Mutates Instatus on-call schedules; verify target ids and payload before execution.
  delete_on_call_schedule:
    endpoint: DELETE /on-call-schedules/{{ record.on_call_schedule_id }}
    required fields: on_call_schedule_id
    risk: Mutates Instatus on-call schedules; verify target ids and payload before execution.

SECURITY
  read risk: external Instatus API read of status-page configuration and incident data
  write risk: external Instatus API mutations for status pages, incidents, maintenances, subscribers, metrics, monitors, routing rules, escalation policies, on-call schedules, and related resources
  approval: required for write actions; destructive delete actions are marked destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect instatus

  # Inspect as structured JSON
  pm connectors inspect instatus --json

AGENT WORKFLOW
  - Run pm connectors inspect instatus before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
