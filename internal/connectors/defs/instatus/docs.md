# Overview

Reads and manages Instatus status pages, components, incidents, maintenances, templates,
subscribers, metrics, monitors, and related status-page resources through the Instatus REST API.

Readable streams: `pages`, `components`, `incidents`, `maintenances`, `workspaces`,
`component_detail`, `incident_detail`, `incident_update_detail`, `maintenance_detail`,
`maintenance_update_detail`, `templates`, `template_detail`, `teammates`, `subscribers`, `metrics`,
`metric_detail`, `user_profile`, `audience_groups`, `audience_group_detail`, `generic_notices`,
`generic_notice_detail`, `monitors`, `monitor_inserted_logs_check`, `monitor_logs`,
`monitor_alerts`, `routing_rules`, `escalation_policies`, `on_call_schedule_members`.

Write actions: `create_page`, `update_page`, `delete_page`, `create_workspace`, `delete_workspace`,
`create_component`, `update_component`, `delete_component`, `add_team_member`, `delete_team_member`,
`add_incident`, `add_incident_with_template`, `update_incident`, `delete_incident`,
`add_incident_update`, `resolve_incident_with_template`, `update_incident_update`,
`delete_incident_update`, `add_maintenance`, `update_maintenance`, `delete_maintenance`,
`add_maintenance_update`, `update_maintenance_update`, `delete_maintenance_update`, `add_template`,
`update_template`, `delete_template`, `add_generic_notice`, `update_generic_notice`,
`delete_generic_notice`, `add_subscriber`, `add_subscribers_bulk`, `delete_subscriber`,
`add_metric`, `update_metric`, `delete_metric`, `add_metric_datapoint`, `add_metric_datapoints`,
`delete_metric_datapoints`, `create_audience_group`, `update_audience_group`,
`delete_audience_group`, `regenerate_secure_link`, `create_monitor`, `update_monitor`,
`delete_monitor`, `update_monitor_group_assignment`, `create_monitor_alert`, `update_monitor_alert`,
`delete_monitor_alert`, `create_monitors_group`, `update_monitors_group`, `delete_monitors_group`,
`add_monitors_to_group`, `create_routing_rule`, `update_routing_rule`, `delete_routing_rule`,
`create_escalation_policy`, `update_escalation_policy`, `delete_escalation_policy`,
`create_on_call_schedule`, `update_on_call_schedule`, `delete_on_call_schedule`.

Service API documentation: https://instatus.com/help/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Instatus API key. Used only for Bearer auth; never logged.
- `audience_group_id` (optional, string); Audience group id used by audience_group_detail.
- `base_url` (optional, string); default `https://api.instatus.com`; format `uri`; Instatus API base
  URL override for tests or proxies.
- `component_id` (optional, string); Component id used by the component_detail stream.
- `generic_notice_id` (optional, string); General notice id used by generic_notice_detail.
- `incident_id` (optional, string); Incident id used by incident detail/update streams.
- `incident_update_id` (optional, string); Incident update id used by incident_update_detail.
- `maintenance_id` (optional, string); Maintenance id used by maintenance detail/update streams.
- `maintenance_update_id` (optional, string); Maintenance update id used by
  maintenance_update_detail.
- `metric_id` (optional, string); Metric id used by metric_detail.
- `monitor_id` (optional, string); Monitor id used by monitor_logs.
- `on_call_schedule_id` (optional, string); On-call schedule id used by on_call_schedule_members.
- `page_id` (optional, string); Status page id required by the components/incidents/maintenances
  streams (parent-scoped resources).
- `page_size` (optional, string); default `50`; Records per page (1-100).
- `subscriber_search` (optional, string); Optional subscriber search query for the subscribers
  stream.
- `template_id` (optional, string); Template id used by template_detail.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.instatus.com`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/pages`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 50.

Pagination by stream: none: `component_detail`, `incident_detail`, `incident_update_detail`,
`maintenance_detail`, `maintenance_update_detail`, `template_detail`, `metric_detail`,
`user_profile`, `audience_group_detail`, `generic_notice_detail`, `monitor_inserted_logs_check`,
`routing_rules`, `escalation_policies`, `on_call_schedule_members`; page_number: `pages`,
`components`, `incidents`, `maintenances`, `workspaces`, `templates`, `teammates`, `subscribers`,
`metrics`, `audience_groups`, `generic_notices`, `monitors`, `monitor_logs`, `monitor_alerts`.

- `pages`: GET `/v2/pages` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 50.
- `components`: GET `/v2/{{ config.page_id }}/components` - records at response root; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `incidents`: GET `/v1/{{ config.page_id }}/incidents` - records at response root; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `maintenances`: GET `/v2/{{ config.page_id }}/maintenances` - records at response root;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  50.
- `workspaces`: GET `/v1/workspaces` - records at response root; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 50; emits passthrough records.
- `component_detail`: GET `/v2/{{ config.page_id }}/components/{{ config.component_id }}` - records
  at response root; emits passthrough records.
- `incident_detail`: GET `/v1/{{ config.page_id }}/incidents/{{ config.incident_id }}` - records at
  response root; emits passthrough records.
- `incident_update_detail`: GET `/v1/{{ config.page_id }}/incidents/{{ config.incident_id
  }}/incident-updates/{{ config.incident_update_id }}` - records at response root; emits passthrough
  records.
- `maintenance_detail`: GET `/v1/{{ config.page_id }}/maintenances/{{ config.maintenance_id }}` -
  records at response root; emits passthrough records.
- `maintenance_update_detail`: GET `/v1/{{ config.page_id }}/maintenances/{{ config.maintenance_id
  }}/maintenance-updates/{{ config.maintenance_update_id }}` - records at response root; emits
  passthrough records.
- `templates`: GET `/v1/{{ config.page_id }}/templates` - records at response root; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50; emits
  passthrough records.
- `template_detail`: GET `/v1/{{ config.page_id }}/templates/{{ config.template_id }}` - records at
  response root; emits passthrough records.
- `teammates`: GET `/v1/{{ config.page_id }}/team` - records at response root; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50; emits
  passthrough records.
- `subscribers`: GET `/v2/{{ config.page_id }}/subscribers` - records at response root; query
  `search` from template `{{ config.subscriber_search }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50; emits
  passthrough records.
- `metrics`: GET `/v1/{{ config.page_id }}/metrics` - records at response root; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50; emits
  passthrough records.
- `metric_detail`: GET `/v1/{{ config.page_id }}/metrics/{{ config.metric_id }}` - records at
  response root; emits passthrough records.
- `user_profile`: GET `/v1/user` - records at response root; emits passthrough records.
- `audience_groups`: GET `/v1/{{ config.page_id }}/audience-groups` - records at response root;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  50; emits passthrough records.
- `audience_group_detail`: GET `/v1/{{ config.page_id }}/audience-groups/{{ config.audience_group_id
  }}` - records at response root; emits passthrough records.
- `generic_notices`: GET `/v1/{{ config.page_id }}/generic-notices` - records at response root;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  50; emits passthrough records.
- `generic_notice_detail`: GET `/v1/{{ config.page_id }}/generic-notices/{{ config.generic_notice_id
  }}` - records at response root; emits passthrough records.
- `monitors`: GET `/{{ config.page_id }}/monitors` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 50; emits
  passthrough records.
- `monitor_inserted_logs_check`: GET `/monitors/check_inserted_logs` - records at response root;
  emits passthrough records.
- `monitor_logs`: GET `/monitors/{{ config.monitor_id }}/logs` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 50;
  emits passthrough records.
- `monitor_alerts`: GET `/{{ config.page_id }}/monitor-alerts` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 50;
  emits passthrough records.
- `routing_rules`: GET `/{{ config.page_id }}/routing-rules` - records at response root; emits
  passthrough records.
- `escalation_policies`: GET `/{{ config.page_id }}/escalation-policies` - records at response root;
  emits passthrough records.
- `on_call_schedule_members`: GET `/on-call-schedules/{{ config.on_call_schedule_id }}/members` -
  records at response root; emits passthrough records.

## Write actions & risks

Overall write risk: external Instatus API mutations for status pages, incidents, maintenances,
subscribers, metrics, monitors, routing rules, escalation policies, on-call schedules, and related
resources.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_page`: POST `/v1/pages` - kind `create`; body type `json`; risk: Mutates Instatus status
  pages; verify target ids and payload before execution.
- `update_page`: PUT `/v2/{{ record.page_id }}` - kind `update`; body type `json`; path fields
  `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates Instatus
  status pages; verify target ids and payload before execution.
- `delete_page`: DELETE `/v2/{{ record.page_id }}` - kind `delete`; body type `none`; path fields
  `page_id`; required record fields `page_id`; accepted fields `page_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Instatus status pages; verify
  target ids and payload before execution.
- `create_workspace`: POST `/v1/workspaces` - kind `create`; body type `json`; risk: Mutates
  Instatus workspaces; verify target ids and payload before execution.
- `delete_workspace`: DELETE `/v1/workspaces/{{ record.workspace_id }}` - kind `delete`; body type
  `none`; path fields `workspace_id`; required record fields `workspace_id`; accepted fields
  `workspace_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Mutates Instatus workspaces; verify target ids and payload before execution.
- `create_component`: POST `/v1/{{ record.page_id }}/components` - kind `create`; body type `json`;
  path fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus components; verify target ids and payload before execution.
- `update_component`: PUT `/v2/{{ record.page_id }}/components/{{ record.component_id }}` - kind
  `update`; body type `json`; path fields `page_id`, `component_id`; required record fields
  `page_id`, `component_id`; accepted fields `component_id`, `page_id`; risk: Mutates Instatus
  components; verify target ids and payload before execution.
- `delete_component`: DELETE `/v1/{{ record.page_id }}/components/{{ record.component_id }}` - kind
  `delete`; body type `none`; path fields `page_id`, `component_id`; required record fields
  `page_id`, `component_id`; accepted fields `component_id`, `page_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Instatus components; verify
  target ids and payload before execution.
- `add_team_member`: POST `/v1/{{ record.page_id }}/team` - kind `create`; body type `json`; path
  fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus team members; verify target ids and payload before execution.
- `delete_team_member`: DELETE `/v1/{{ record.page_id }}/team/{{ record.member_id }}` - kind
  `delete`; body type `none`; path fields `page_id`, `member_id`; required record fields `page_id`,
  `member_id`; accepted fields `member_id`, `page_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Mutates Instatus team members; verify target ids and
  payload before execution.
- `add_incident`: POST `/v1/{{ record.page_id }}/incidents` - kind `create`; body type `json`; path
  fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus incidents; verify target ids and payload before execution.
- `add_incident_with_template`: POST `/v1/{{ record.page_id }}/incidents/{{ record.template }}` -
  kind `create`; body type `json`; path fields `page_id`, `template`; required record fields
  `page_id`, `template`; accepted fields `page_id`, `template`; risk: Mutates Instatus incidents;
  verify target ids and payload before execution.
- `update_incident`: PUT `/v1/{{ record.page_id }}/incidents/{{ record.incident_id }}` - kind
  `update`; body type `json`; path fields `page_id`, `incident_id`; required record fields
  `page_id`, `incident_id`; accepted fields `incident_id`, `page_id`; risk: Mutates Instatus
  incidents; verify target ids and payload before execution.
- `delete_incident`: DELETE `/v1/{{ record.page_id }}/incidents/{{ record.incident_id }}` - kind
  `delete`; body type `none`; path fields `page_id`, `incident_id`; required record fields
  `page_id`, `incident_id`; accepted fields `incident_id`, `page_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Instatus incidents; verify
  target ids and payload before execution.
- `add_incident_update`: POST `/v1/{{ record.page_id }}/incidents/{{ record.incident_id
  }}/incident-updates` - kind `create`; body type `json`; path fields `page_id`, `incident_id`;
  required record fields `page_id`, `incident_id`; accepted fields `incident_id`, `page_id`; risk:
  Mutates Instatus incident updates; verify target ids and payload before execution.
- `resolve_incident_with_template`: POST `/v2/{{ record.page_id }}/incidents/{{ record.incident_id
  }}/incident-updates/{{ record.template }}` - kind `custom`; body type `json`; path fields
  `page_id`, `incident_id`, `template`; required record fields `page_id`, `incident_id`, `template`;
  accepted fields `incident_id`, `page_id`, `template`; risk: Mutates Instatus incident updates;
  verify target ids and payload before execution.
- `update_incident_update`: PUT `/v1/{{ record.page_id }}/incidents/{{ record.incident_id
  }}/incident-updates/{{ record.incident_update_id }}` - kind `update`; body type `json`; path
  fields `page_id`, `incident_id`, `incident_update_id`; required record fields `page_id`,
  `incident_id`, `incident_update_id`; accepted fields `incident_id`, `incident_update_id`,
  `page_id`; risk: Mutates Instatus incident updates; verify target ids and payload before
  execution.
- `delete_incident_update`: DELETE `/v1/{{ record.page_id }}/incidents/{{ record.incident_id
  }}/incident-updates/{{ record.incident_update_id }}` - kind `delete`; body type `none`; path
  fields `page_id`, `incident_id`, `incident_update_id`; required record fields `page_id`,
  `incident_id`, `incident_update_id`; accepted fields `incident_id`, `incident_update_id`,
  `page_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Mutates Instatus incident updates; verify target ids and payload before execution.
- `add_maintenance`: POST `/v1/{{ record.page_id }}/maintenances` - kind `create`; body type `json`;
  path fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus maintenances; verify target ids and payload before execution.
- `update_maintenance`: PUT `/v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}` -
  kind `update`; body type `json`; path fields `page_id`, `maintenance_id`; required record fields
  `page_id`, `maintenance_id`; accepted fields `maintenance_id`, `page_id`; risk: Mutates Instatus
  maintenances; verify target ids and payload before execution.
- `delete_maintenance`: DELETE `/v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id }}` -
  kind `delete`; body type `none`; path fields `page_id`, `maintenance_id`; required record fields
  `page_id`, `maintenance_id`; accepted fields `maintenance_id`, `page_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Mutates Instatus maintenances;
  verify target ids and payload before execution.
- `add_maintenance_update`: POST `/v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id
  }}/maintenance-updates` - kind `create`; body type `json`; path fields `page_id`,
  `maintenance_id`; required record fields `page_id`, `maintenance_id`; accepted fields
  `maintenance_id`, `page_id`; risk: Mutates Instatus maintenance updates; verify target ids and
  payload before execution.
- `update_maintenance_update`: PUT `/v1/{{ record.page_id }}/maintenances/{{ record.maintenance_id
  }}/maintenance-updates/{{ record.maintenance_update_id }}` - kind `update`; body type `json`; path
  fields `page_id`, `maintenance_id`, `maintenance_update_id`; required record fields `page_id`,
  `maintenance_id`, `maintenance_update_id`; accepted fields `maintenance_id`,
  `maintenance_update_id`, `page_id`; risk: Mutates Instatus maintenance updates; verify target ids
  and payload before execution.
- `delete_maintenance_update`: DELETE `/v1/{{ record.page_id }}/maintenances/{{
  record.maintenance_id }}/maintenance-updates/{{ record.maintenance_update_id }}` - kind `delete`;
  body type `none`; path fields `page_id`, `maintenance_id`, `maintenance_update_id`; required
  record fields `page_id`, `maintenance_id`, `maintenance_update_id`; accepted fields
  `maintenance_id`, `maintenance_update_id`, `page_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Mutates Instatus maintenance updates; verify
  target ids and payload before execution.
- `add_template`: POST `/v1/{{ record.page_id }}/templates` - kind `create`; body type `json`; path
  fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus templates; verify target ids and payload before execution.
- `update_template`: PUT `/v1/{{ record.page_id }}/templates/{{ record.template_id }}` - kind
  `update`; body type `json`; path fields `page_id`, `template_id`; required record fields
  `page_id`, `template_id`; accepted fields `page_id`, `template_id`; risk: Mutates Instatus
  templates; verify target ids and payload before execution.
- `delete_template`: DELETE `/v1/{{ record.page_id }}/templates/{{ record.template_id }}` - kind
  `delete`; body type `none`; path fields `page_id`, `template_id`; required record fields
  `page_id`, `template_id`; accepted fields `page_id`, `template_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Instatus templates; verify
  target ids and payload before execution.
- `add_generic_notice`: POST `/v1/{{ record.page_id }}/generic-notices` - kind `create`; body type
  `json`; path fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk:
  Mutates Instatus general notices; verify target ids and payload before execution.
- `update_generic_notice`: PUT `/v1/{{ record.page_id }}/generic-notices/{{ record.generic_notice_id
  }}` - kind `update`; body type `json`; path fields `page_id`, `generic_notice_id`; required record
  fields `page_id`, `generic_notice_id`; accepted fields `generic_notice_id`, `page_id`; risk:
  Mutates Instatus general notices; verify target ids and payload before execution.
- `delete_generic_notice`: DELETE `/v1/{{ record.page_id }}/generic-notices/{{
  record.generic_notice_id }}` - kind `delete`; body type `none`; path fields `page_id`,
  `generic_notice_id`; required record fields `page_id`, `generic_notice_id`; accepted fields
  `generic_notice_id`, `page_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Mutates Instatus general notices; verify target ids and payload before
  execution.
- `add_subscriber`: POST `/v1/{{ record.page_id }}/subscribers` - kind `create`; body type `json`;
  path fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus subscribers; verify target ids and payload before execution.
- `add_subscribers_bulk`: POST `/v1/{{ record.page_id }}/subscribers/bulk` - kind `create`; body
  type `json`; path fields `page_id`; required record fields `page_id`; accepted fields `page_id`;
  risk: Mutates Instatus subscribers; verify target ids and payload before execution.
- `delete_subscriber`: DELETE `/v1/{{ record.page_id }}/subscribers/{{ record.subscriber_id }}` -
  kind `delete`; body type `none`; path fields `page_id`, `subscriber_id`; required record fields
  `page_id`, `subscriber_id`; accepted fields `page_id`, `subscriber_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Mutates Instatus subscribers; verify
  target ids and payload before execution.
- `add_metric`: POST `/v1/{{ record.page_id }}/metrics` - kind `create`; body type `json`; path
  fields `page_id`; required record fields `page_id`; accepted fields `page_id`; risk: Mutates
  Instatus metrics; verify target ids and payload before execution.
- `update_metric`: PUT `/v1/{{ record.page_id }}/metrics/{{ record.metric_id }}` - kind `update`;
  body type `json`; path fields `page_id`, `metric_id`; required record fields `page_id`,
  `metric_id`; accepted fields `metric_id`, `page_id`; risk: Mutates Instatus metrics; verify target
  ids and payload before execution.
- `delete_metric`: DELETE `/v1/{{ record.page_id }}/metrics/{{ record.metric_id }}` - kind `delete`;
  body type `none`; path fields `page_id`, `metric_id`; required record fields `page_id`,
  `metric_id`; accepted fields `metric_id`, `page_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Mutates Instatus metrics; verify target ids and payload
  before execution.
- `add_metric_datapoint`: POST `/v1/{{ record.page_id }}/metrics/{{ record.metric_id }}` - kind
  `create`; body type `json`; path fields `page_id`, `metric_id`; required record fields `page_id`,
  `metric_id`; accepted fields `metric_id`, `page_id`; risk: Mutates Instatus metric datapoints;
  verify target ids and payload before execution.
- `add_metric_datapoints`: POST `/v1/{{ record.page_id }}/metrics/{{ record.metric_id }}/data` -
  kind `create`; body type `json`; path fields `page_id`, `metric_id`; required record fields
  `page_id`, `metric_id`; accepted fields `metric_id`, `page_id`; risk: Mutates Instatus metric
  datapoints; verify target ids and payload before execution.
- `delete_metric_datapoints`: DELETE `/v1/{{ record.page_id }}/metrics/{{ record.metric_id }}/data`
  - kind `delete`; body type `json`; path fields `page_id`, `metric_id`; required record fields
  `page_id`, `metric_id`; accepted fields `metric_id`, `page_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Mutates Instatus metric datapoints; verify
  target ids and payload before execution.
- `create_audience_group`: POST `/v1/{{ record.page_id }}/audience_groups` - kind `create`; body
  type `json`; path fields `page_id`; required record fields `page_id`; accepted fields `page_id`;
  risk: Mutates Instatus audience groups; verify target ids and payload before execution.
- `update_audience_group`: PUT `/v1/{{ record.page_id }}/audience_groups/{{ record.audience_group_id
  }}` - kind `update`; body type `json`; path fields `page_id`, `audience_group_id`; required record
  fields `page_id`, `audience_group_id`; accepted fields `audience_group_id`, `page_id`; risk:
  Mutates Instatus audience groups; verify target ids and payload before execution.
- `delete_audience_group`: DELETE `/v1/{{ record.page_id }}/audience-groups/{{
  record.audience_group_id }}` - kind `delete`; body type `none`; path fields `page_id`,
  `audience_group_id`; required record fields `page_id`, `audience_group_id`; accepted fields
  `audience_group_id`, `page_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Mutates Instatus audience groups; verify target ids and payload before
  execution.
- `regenerate_secure_link`: POST `/v1/{{ record.page_id }}/regenerate-secure-link` - kind `custom`;
  body type `none`; path fields `page_id`; required record fields `page_id`; accepted fields
  `page_id`; risk: Mutates Instatus private-page secure links; verify target ids and payload before
  execution.
- `create_monitor`: POST `/monitors` - kind `create`; body type `json`; risk: Mutates Instatus
  monitors; verify target ids and payload before execution.
- `update_monitor`: PUT `/monitors/{{ record.monitor_id }}` - kind `update`; body type `json`; path
  fields `monitor_id`; required record fields `monitor_id`; accepted fields `monitor_id`; risk:
  Mutates Instatus monitors; verify target ids and payload before execution.
- `delete_monitor`: DELETE `/monitors/{{ record.monitor_id }}` - kind `delete`; body type `none`;
  path fields `monitor_id`; required record fields `monitor_id`; accepted fields `monitor_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Mutates
  Instatus monitors; verify target ids and payload before execution.
- `update_monitor_group_assignment`: PUT `/monitors/{{ record.monitor_id }}/group` - kind `update`;
  body type `json`; path fields `monitor_id`; required record fields `monitor_id`; accepted fields
  `monitor_id`; risk: Mutates Instatus monitor group assignments; verify target ids and payload
  before execution.
- `create_monitor_alert`: POST `/monitor-alerts` - kind `create`; body type `json`; risk: Mutates
  Instatus monitor alerts; verify target ids and payload before execution.
- `update_monitor_alert`: PUT `/monitor-alerts/{{ record.monitor_alert_id }}` - kind `update`; body
  type `json`; path fields `monitor_alert_id`; required record fields `monitor_alert_id`; accepted
  fields `monitor_alert_id`; risk: Mutates Instatus monitor alerts; verify target ids and payload
  before execution.
- `delete_monitor_alert`: DELETE `/monitor-alerts/{{ record.monitor_alert_id }}` - kind `delete`;
  body type `none`; path fields `monitor_alert_id`; required record fields `monitor_alert_id`;
  accepted fields `monitor_alert_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Mutates Instatus monitor alerts; verify target ids and payload
  before execution.
- `create_monitors_group`: POST `/monitors-groups` - kind `create`; body type `json`; risk: Mutates
  Instatus monitor groups; verify target ids and payload before execution.
- `update_monitors_group`: PUT `/monitors-groups/{{ record.monitors_group_id }}` - kind `update`;
  body type `json`; path fields `monitors_group_id`; required record fields `monitors_group_id`;
  accepted fields `monitors_group_id`; risk: Mutates Instatus monitor groups; verify target ids and
  payload before execution.
- `delete_monitors_group`: DELETE `/monitors-groups/{{ record.monitors_group_id }}` - kind `delete`;
  body type `none`; path fields `monitors_group_id`; required record fields `monitors_group_id`;
  accepted fields `monitors_group_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Mutates Instatus monitor groups; verify target ids and payload
  before execution.
- `add_monitors_to_group`: POST `/monitors-groups/{{ record.monitors_group_id }}/monitors` - kind
  `custom`; body type `json`; path fields `monitors_group_id`; required record fields
  `monitors_group_id`; accepted fields `monitors_group_id`; risk: Mutates Instatus monitor groups;
  verify target ids and payload before execution.
- `create_routing_rule`: POST `/routing-rules` - kind `create`; body type `json`; risk: Mutates
  Instatus routing rules; verify target ids and payload before execution.
- `update_routing_rule`: PUT `/routing-rules/{{ record.routing_rule_id }}` - kind `update`; body
  type `json`; path fields `routing_rule_id`; required record fields `routing_rule_id`; accepted
  fields `routing_rule_id`; risk: Mutates Instatus routing rules; verify target ids and payload
  before execution.
- `delete_routing_rule`: DELETE `/routing-rules/{{ record.routing_rule_id }}` - kind `delete`; body
  type `none`; path fields `routing_rule_id`; required record fields `routing_rule_id`; accepted
  fields `routing_rule_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Mutates Instatus routing rules; verify target ids and payload before
  execution.
- `create_escalation_policy`: POST `/escalation-policies` - kind `create`; body type `json`; risk:
  Mutates Instatus escalation policies; verify target ids and payload before execution.
- `update_escalation_policy`: PUT `/escalation-policies/{{ record.escalation_policy_id }}` - kind
  `update`; body type `json`; path fields `escalation_policy_id`; required record fields
  `escalation_policy_id`; accepted fields `escalation_policy_id`; risk: Mutates Instatus escalation
  policies; verify target ids and payload before execution.
- `delete_escalation_policy`: DELETE `/escalation-policies/{{ record.escalation_policy_id }}` - kind
  `delete`; body type `none`; path fields `escalation_policy_id`; required record fields
  `escalation_policy_id`; accepted fields `escalation_policy_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Mutates Instatus escalation policies; verify
  target ids and payload before execution.
- `create_on_call_schedule`: POST `/on-call-schedules` - kind `create`; body type `json`; risk:
  Mutates Instatus on-call schedules; verify target ids and payload before execution.
- `update_on_call_schedule`: PUT `/on-call-schedules/{{ record.on_call_schedule_id }}` - kind
  `update`; body type `json`; path fields `on_call_schedule_id`; required record fields
  `on_call_schedule_id`; accepted fields `on_call_schedule_id`; risk: Mutates Instatus on-call
  schedules; verify target ids and payload before execution.
- `delete_on_call_schedule`: DELETE `/on-call-schedules/{{ record.on_call_schedule_id }}` - kind
  `delete`; body type `none`; path fields `on_call_schedule_id`; required record fields
  `on_call_schedule_id`; accepted fields `on_call_schedule_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Mutates Instatus on-call schedules; verify
  target ids and payload before execution.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=50.
- API coverage includes 28 stream-backed endpoint group(s), 63 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=2, out_of_scope=1.
