# Overview

Reads cost, budget, resource-management, segment, notification, and integration data from the
Vantage API and writes budgets, folders, dashboards, cost reports, resource reports, saved filters,
workspaces, teams, cost alerts, budget alerts, anomaly alerts, business metrics, virtual tag
configs, segments, report notifications, recommendation views, network flow reports, Kubernetes
efficiency reports, anomaly notifications, and canvases.

Readable streams: `costs`, `cost_reports`, `budgets`, `folders`, `dashboards`, `business_metrics`,
`resource_reports`, `recommendations`, `teams`, `saved_filters`, `workspaces`,
`virtual_tag_configs`, `tags`, `cost_alerts`, `budget_alerts`, `anomaly_alerts`, `managed_accounts`,
`financial_commitments`, `segments`, `report_notifications`, `recommendation_views`,
`network_flow_reports`, `kubernetes_efficiency_reports`, `anomaly_notifications`, `canvases`,
`invoices`, `integrations`.

Write actions: `create_budget`, `update_budget`, `delete_budget`, `create_folder`, `update_folder`,
`delete_folder`, `create_dashboard`, `update_dashboard`, `delete_dashboard`, `create_cost_report`,
`update_cost_report`, `delete_cost_report`, `create_saved_filter`, `delete_saved_filter`,
`create_workspace`, `delete_workspace`, `create_team`, `delete_team`, `create_cost_alert`,
`update_cost_alert`, `delete_cost_alert`, `create_budget_alert`, `delete_budget_alert`,
`create_business_metric`, `update_business_metric`, `delete_business_metric`,
`delete_virtual_tag_config`, `update_saved_filter`, `update_team`, `update_workspace`,
`update_anomaly_alert`, `update_budget_alert`, `update_virtual_tag_config`, `create_segment`,
`update_segment`, `delete_segment`, `create_report_notification`, `update_report_notification`,
`delete_report_notification`, `create_recommendation_view`, `update_recommendation_view`,
`delete_recommendation_view`, `create_network_flow_report`, `update_network_flow_report`,
`delete_network_flow_report`, `create_kubernetes_efficiency_report`,
`update_kubernetes_efficiency_report`, `delete_kubernetes_efficiency_report`,
`create_anomaly_notification`, `update_anomaly_notification`, `delete_anomaly_notification`,
`create_canvas`, `update_canvas`, `delete_canvas`, `create_resource_report`,
`update_resource_report`, `delete_resource_report`, `create_virtual_tag_config`.

Service API documentation: https://docs.vantage.sh/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Vantage API bearer access token. Used only for Bearer
  auth; never logged.
- `base_url` (optional, string); default `https://api.vantage.sh`; format `uri`; Vantage API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.vantage.sh`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/costs`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `cost_reports`, `budgets`, `folders`, `dashboards`,
`business_metrics`, `resource_reports`, `recommendations`, `teams`, `saved_filters`, `workspaces`,
`virtual_tag_configs`, `tags`, `cost_alerts`, `budget_alerts`, `anomaly_alerts`, `managed_accounts`,
`financial_commitments`, `segments`, `report_notifications`, `recommendation_views`,
`network_flow_reports`, `kubernetes_efficiency_reports`, `anomaly_notifications`, `canvases`,
`invoices`, `integrations`; none: `costs`.

- `costs`: GET `/v2/costs` - records path `costs`; computed output fields `amount`, `date`, `id`,
  `service`.
- `cost_reports`: GET `/v2/cost_reports` - records path `cost_reports`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host.
- `budgets`: GET `/v2/budgets` - records path `budgets`; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host.
- `folders`: GET `/v2/folders` - records path `folders`; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host.
- `dashboards`: GET `/v2/dashboards` - records path `dashboards`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host.
- `business_metrics`: GET `/v2/business_metrics` - records path `business_metrics`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host.
- `resource_reports`: GET `/v2/resource_reports` - records path `resource_reports`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host.
- `recommendations`: GET `/v2/recommendations` - records path `recommendations`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host.
- `teams`: GET `/v2/teams` - records path `teams`; follows a next-page URL from the response body;
  URL path `links.next`; next URLs stay on the configured API host.
- `saved_filters`: GET `/v2/saved_filters` - records path `saved_filters`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host.
- `workspaces`: GET `/v2/workspaces` - records path `workspaces`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host.
- `virtual_tag_configs`: GET `/v2/virtual_tag_configs` - records path `virtual_tag_configs`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host.
- `tags`: GET `/v2/tags` - records path `tags`; follows a next-page URL from the response body; URL
  path `links.next`; next URLs stay on the configured API host.
- `cost_alerts`: GET `/v2/cost_alerts` - records path `cost_alerts`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host.
- `budget_alerts`: GET `/v2/budget_alerts` - records path `budget_alerts`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host.
- `anomaly_alerts`: GET `/v2/anomaly_alerts` - records path `anomaly_alerts`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host.
- `managed_accounts`: GET `/v2/managed_accounts` - records path `managed_accounts`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host.
- `financial_commitments`: GET `/v2/financial_commitments` - records path `financial_commitments`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `segments`: GET `/v2/segments` - records path `segments`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host.
- `report_notifications`: GET `/v2/report_notifications` - records path `report_notifications`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `recommendation_views`: GET `/v2/recommendation_views` - records path `recommendation_views`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `network_flow_reports`: GET `/v2/network_flow_reports` - records path `network_flow_reports`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `kubernetes_efficiency_reports`: GET `/v2/kubernetes_efficiency_reports` - records path
  `kubernetes_efficiency_reports`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host.
- `anomaly_notifications`: GET `/v2/anomaly_notifications` - records path `anomaly_notifications`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host.
- `canvases`: GET `/v2/canvases` - records path `canvases`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host.
- `invoices`: GET `/v2/invoices` - records path `invoices`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host.
- `integrations`: GET `/v2/integrations` - records path `integrations`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host.

## Write actions & risks

Overall write risk: external mutation of Vantage cost-management configuration (budgets, folders,
dashboards, cost reports, resource reports, saved filters, workspaces, teams, alerts, business
metrics, virtual tag configs, segments, report notifications, recommendation views, network flow
reports, Kubernetes efficiency reports, anomaly notifications, canvases); approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_budget`: POST `/v2/budgets` - kind `create`; body type `json`; required record fields
  `name`, `period_amount`; accepted fields `name`, `period_amount`, `period_duration`,
  `workspace_token`; risk: external mutation; approval required.
- `update_budget`: PUT `/v2/budgets/{{ record.budget_token }}` - kind `update`; body type `json`;
  path fields `budget_token`; required record fields `budget_token`; accepted fields `budget_token`,
  `name`, `period_amount`; risk: external mutation; approval required.
- `delete_budget`: DELETE `/v2/budgets/{{ record.budget_token }}` - kind `delete`; body type `none`;
  path fields `budget_token`; required record fields `budget_token`; accepted fields `budget_token`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; approval required.
- `create_folder`: POST `/v2/folders` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `parent_folder_token`, `title`, `workspace_token`; risk: external
  mutation; approval required.
- `update_folder`: PUT `/v2/folders/{{ record.folder_token }}` - kind `update`; body type `json`;
  path fields `folder_token`; required record fields `folder_token`; accepted fields `folder_token`,
  `title`; risk: external mutation; approval required.
- `delete_folder`: DELETE `/v2/folders/{{ record.folder_token }}` - kind `delete`; body type `none`;
  path fields `folder_token`; required record fields `folder_token`; accepted fields `folder_token`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; approval required.
- `create_dashboard`: POST `/v2/dashboards` - kind `create`; body type `json`; required record
  fields `title`; accepted fields `date_interval`, `title`, `workspace_token`; risk: external
  mutation; approval required.
- `update_dashboard`: PUT `/v2/dashboards/{{ record.dashboard_token }}` - kind `update`; body type
  `json`; path fields `dashboard_token`; required record fields `dashboard_token`; accepted fields
  `dashboard_token`, `title`; risk: external mutation; approval required.
- `delete_dashboard`: DELETE `/v2/dashboards/{{ record.dashboard_token }}` - kind `delete`; body
  type `none`; path fields `dashboard_token`; required record fields `dashboard_token`; accepted
  fields `dashboard_token`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: destructive external mutation; approval required.
- `create_cost_report`: POST `/v2/cost_reports` - kind `create`; body type `json`; required record
  fields `title`; accepted fields `filter`, `folder_token`, `title`, `workspace_token`; risk:
  external mutation; approval required.
- `update_cost_report`: PUT `/v2/cost_reports/{{ record.cost_report_token }}` - kind `update`; body
  type `json`; path fields `cost_report_token`; required record fields `cost_report_token`; accepted
  fields `cost_report_token`, `filter`, `title`; risk: external mutation; approval required.
- `delete_cost_report`: DELETE `/v2/cost_reports/{{ record.cost_report_token }}` - kind `delete`;
  body type `none`; path fields `cost_report_token`; required record fields `cost_report_token`;
  accepted fields `cost_report_token`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external mutation; approval required.
- `create_saved_filter`: POST `/v2/saved_filters` - kind `create`; body type `json`; required record
  fields `title`, `filter`; accepted fields `filter`, `title`, `workspace_token`; risk: external
  mutation; approval required.
- `delete_saved_filter`: DELETE `/v2/saved_filters/{{ record.saved_filter_token }}` - kind `delete`;
  body type `none`; path fields `saved_filter_token`; required record fields `saved_filter_token`;
  accepted fields `saved_filter_token`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external mutation; approval required.
- `create_workspace`: POST `/v2/workspaces` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `name`; risk: external mutation; approval required.
- `delete_workspace`: DELETE `/v2/workspaces/{{ record.workspace_token }}` - kind `delete`; body
  type `none`; path fields `workspace_token`; required record fields `workspace_token`; accepted
  fields `workspace_token`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: destructive external mutation; approval required.
- `create_team`: POST `/v2/teams` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: external mutation; approval required.
- `delete_team`: DELETE `/v2/teams/{{ record.team_token }}` - kind `delete`; body type `none`; path
  fields `team_token`; required record fields `team_token`; accepted fields `team_token`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; approval required.
- `create_cost_alert`: POST `/v2/cost_alerts` - kind `create`; body type `json`; required record
  fields `title`; accepted fields `filter`, `threshold`, `title`, `workspace_token`; risk: external
  mutation; approval required.
- `update_cost_alert`: PUT `/v2/cost_alerts/{{ record.cost_alert_token }}` - kind `update`; body
  type `json`; path fields `cost_alert_token`; required record fields `cost_alert_token`; accepted
  fields `cost_alert_token`, `title`; risk: external mutation; approval required.
- `delete_cost_alert`: DELETE `/v2/cost_alerts/{{ record.cost_alert_token }}` - kind `delete`; body
  type `none`; path fields `cost_alert_token`; required record fields `cost_alert_token`; accepted
  fields `cost_alert_token`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: destructive external mutation; approval required.
- `create_budget_alert`: POST `/v2/budget_alerts` - kind `create`; body type `json`; required record
  fields `budget_token`, `user_token`, `threshold`; accepted fields `budget_token`, `threshold`,
  `user_token`; risk: external mutation; approval required.
- `delete_budget_alert`: DELETE `/v2/budget_alerts/{{ record.budget_alert_token }}` - kind `delete`;
  body type `none`; path fields `budget_alert_token`; required record fields `budget_alert_token`;
  accepted fields `budget_alert_token`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external mutation; approval required.
- `create_business_metric`: POST `/v2/business_metrics` - kind `create`; body type `json`; required
  record fields `title`; accepted fields `title`, `workspace_token`; risk: external mutation;
  approval required.
- `update_business_metric`: PUT `/v2/business_metrics/{{ record.business_metric_token }}` - kind
  `update`; body type `json`; path fields `business_metric_token`; required record fields
  `business_metric_token`; accepted fields `business_metric_token`, `title`; risk: external
  mutation; approval required.
- `delete_business_metric`: DELETE `/v2/business_metrics/{{ record.business_metric_token }}` - kind
  `delete`; body type `none`; path fields `business_metric_token`; required record fields
  `business_metric_token`; accepted fields `business_metric_token`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation;
  approval required.
- `delete_virtual_tag_config`: DELETE `/v2/virtual_tag_configs/{{ record.token }}` - kind `delete`;
  body type `none`; path fields `token`; required record fields `token`; accepted fields `token`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; approval required.
- `update_saved_filter`: PUT `/v2/saved_filters/{{ record.saved_filter_token }}` - kind `update`;
  body type `json`; path fields `saved_filter_token`; required record fields `saved_filter_token`;
  accepted fields `filter`, `saved_filter_token`, `title`; risk: external mutation; approval
  required.
- `update_team`: PUT `/v2/teams/{{ record.team_token }}` - kind `update`; body type `json`; path
  fields `team_token`; required record fields `team_token`; accepted fields `name`, `team_token`;
  risk: external mutation; approval required.
- `update_workspace`: PUT `/v2/workspaces/{{ record.workspace_token }}` - kind `update`; body type
  `json`; path fields `workspace_token`; required record fields `workspace_token`; accepted fields
  `name`, `workspace_token`; risk: external mutation; approval required.
- `update_anomaly_alert`: PUT `/v2/anomaly_alerts/{{ record.anomaly_alert_token }}` - kind `update`;
  body type `json`; path fields `anomaly_alert_token`; required record fields `anomaly_alert_token`;
  accepted fields `anomaly_alert_token`, `title`; risk: external mutation; approval required.
- `update_budget_alert`: PUT `/v2/budget_alerts/{{ record.budget_alert_token }}` - kind `update`;
  body type `json`; path fields `budget_alert_token`; required record fields `budget_alert_token`;
  accepted fields `budget_alert_token`, `threshold`; risk: external mutation; approval required.
- `update_virtual_tag_config`: PUT `/v2/virtual_tag_configs/{{ record.token }}` - kind `update`;
  body type `json`; path fields `token`; required record fields `token`; accepted fields
  `backfill_until`, `token`; risk: external mutation; approval required.
- `create_segment`: POST `/v2/segments` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `description`, `filter`, `parent_segment_token`, `priority`, `title`,
  `track_unallocated`, `workspace_token`; risk: external mutation; approval required.
- `update_segment`: PUT `/v2/segments/{{ record.segment_token }}` - kind `update`; body type `json`;
  path fields `segment_token`; required record fields `segment_token`; accepted fields
  `description`, `filter`, `priority`, `segment_token`, `title`; risk: external mutation; approval
  required.
- `delete_segment`: DELETE `/v2/segments/{{ record.segment_token }}` - kind `delete`; body type
  `none`; path fields `segment_token`; required record fields `segment_token`; accepted fields
  `segment_token`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: destructive external mutation; approval required.
- `create_report_notification`: POST `/v2/report_notifications` - kind `create`; body type `json`;
  required record fields `title`, `cost_report_token`, `frequency`, `change`; accepted fields
  `change`, `cost_report_token`, `frequency`, `recipient_channels`, `title`, `user_tokens`,
  `workspace_token`; risk: external mutation; approval required.
- `update_report_notification`: PUT `/v2/report_notifications/{{ record.report_notification_token
  }}` - kind `update`; body type `json`; path fields `report_notification_token`; required record
  fields `report_notification_token`; accepted fields `change`, `frequency`,
  `report_notification_token`, `title`; risk: external mutation; approval required.
- `delete_report_notification`: DELETE `/v2/report_notifications/{{ record.report_notification_token
  }}` - kind `delete`; body type `none`; path fields `report_notification_token`; required record
  fields `report_notification_token`; accepted fields `report_notification_token`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; approval required.
- `create_recommendation_view`: POST `/v2/recommendation_views` - kind `create`; body type `json`;
  required record fields `title`, `workspace_token`; accepted fields `min_savings`, `provider_ids`,
  `regions`, `title`, `types`, `workspace_token`; risk: external mutation; approval required.
- `update_recommendation_view`: PUT `/v2/recommendation_views/{{ record.recommendation_view_token
  }}` - kind `update`; body type `json`; path fields `recommendation_view_token`; required record
  fields `recommendation_view_token`; accepted fields `min_savings`, `recommendation_view_token`,
  `title`; risk: external mutation; approval required.
- `delete_recommendation_view`: DELETE `/v2/recommendation_views/{{ record.recommendation_view_token
  }}` - kind `delete`; body type `none`; path fields `recommendation_view_token`; required record
  fields `recommendation_view_token`; accepted fields `recommendation_view_token`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; approval required.
- `create_network_flow_report`: POST `/v2/network_flow_reports` - kind `create`; body type `json`;
  required record fields `workspace_token`, `title`; accepted fields `date_interval`, `end_date`,
  `filter`, `flow_direction`, `flow_weight`, `groupings`, `start_date`, `title`, `workspace_token`;
  risk: external mutation; approval required.
- `update_network_flow_report`: PUT `/v2/network_flow_reports/{{ record.network_flow_report_token
  }}` - kind `update`; body type `json`; path fields `network_flow_report_token`; required record
  fields `network_flow_report_token`; accepted fields `date_interval`, `network_flow_report_token`,
  `title`; risk: external mutation; approval required.
- `delete_network_flow_report`: DELETE `/v2/network_flow_reports/{{ record.network_flow_report_token
  }}` - kind `delete`; body type `none`; path fields `network_flow_report_token`; required record
  fields `network_flow_report_token`; accepted fields `network_flow_report_token`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; approval required.
- `create_kubernetes_efficiency_report`: POST `/v2/kubernetes_efficiency_reports` - kind `create`;
  body type `json`; required record fields `workspace_token`, `title`; accepted fields
  `aggregated_by`, `date_bucket`, `date_interval`, `filter`, `groupings`, `title`,
  `workspace_token`; risk: external mutation; approval required.
- `update_kubernetes_efficiency_report`: PUT `/v2/kubernetes_efficiency_reports/{{
  record.kubernetes_efficiency_report_token }}` - kind `update`; body type `json`; path fields
  `kubernetes_efficiency_report_token`; required record fields `kubernetes_efficiency_report_token`;
  accepted fields `date_bucket`, `kubernetes_efficiency_report_token`, `title`; risk: external
  mutation; approval required.
- `delete_kubernetes_efficiency_report`: DELETE `/v2/kubernetes_efficiency_reports/{{
  record.kubernetes_efficiency_report_token }}` - kind `delete`; body type `none`; path fields
  `kubernetes_efficiency_report_token`; required record fields `kubernetes_efficiency_report_token`;
  accepted fields `kubernetes_efficiency_report_token`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; approval required.
- `create_anomaly_notification`: POST `/v2/anomaly_notifications` - kind `create`; body type `json`;
  required record fields `cost_report_token`; accepted fields `cost_report_token`,
  `recipient_channels`, `threshold`, `user_tokens`; risk: external mutation; approval required.
- `update_anomaly_notification`: PUT `/v2/anomaly_notifications/{{ record.anomaly_notification_token
  }}` - kind `update`; body type `json`; path fields `anomaly_notification_token`; required record
  fields `anomaly_notification_token`; accepted fields `anomaly_notification_token`, `threshold`;
  risk: external mutation; approval required.
- `delete_anomaly_notification`: DELETE `/v2/anomaly_notifications/{{
  record.anomaly_notification_token }}` - kind `delete`; body type `none`; path fields
  `anomaly_notification_token`; required record fields `anomaly_notification_token`; accepted fields
  `anomaly_notification_token`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: destructive external mutation; approval required.
- `create_canvas`: POST `/v2/canvases` - kind `create`; body type `json`; required record fields
  `title`, `prompt`; accepted fields `prompt`, `title`, `workspace_token`; risk: external mutation;
  approval required.
- `update_canvas`: PUT `/v2/canvases/{{ record.canvas_token }}` - kind `update`; body type `json`;
  path fields `canvas_token`; required record fields `canvas_token`; accepted fields `canvas_token`,
  `prompt`, `title`; risk: external mutation; approval required.
- `delete_canvas`: DELETE `/v2/canvases/{{ record.canvas_token }}` - kind `delete`; body type
  `none`; path fields `canvas_token`; required record fields `canvas_token`; accepted fields
  `canvas_token`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: destructive external mutation; approval required.
- `create_resource_report`: POST `/v2/resource_reports` - kind `create`; body type `json`; required
  record fields `workspace_token`; accepted fields `columns`, `filter`, `folder_token`, `title`,
  `workspace_token`; risk: external mutation; approval required.
- `update_resource_report`: PUT `/v2/resource_reports/{{ record.resource_report_token }}` - kind
  `update`; body type `json`; path fields `resource_report_token`; required record fields
  `resource_report_token`; accepted fields `columns`, `filter`, `folder_token`,
  `resource_report_token`, `title`; risk: external mutation; approval required.
- `delete_resource_report`: DELETE `/v2/resource_reports/{{ record.resource_report_token }}` - kind
  `delete`; body type `none`; path fields `resource_report_token`; required record fields
  `resource_report_token`; accepted fields `resource_report_token`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation;
  approval required.
- `create_virtual_tag_config`: POST `/v2/virtual_tag_configs` - kind `create`; body type `json`;
  required record fields `key`, `overridable`, `values`; accepted fields `backfill_until`, `key`,
  `overridable`, `values`; risk: external mutation; approval required.

## Known limits

- API coverage includes 27 stream-backed endpoint group(s), 58 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, duplicate_of=34, non_data_endpoint=28, out_of_scope=13,
  requires_elevated_scope=30.
