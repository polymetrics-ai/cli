# pm connectors inspect vantage

```text
NAME
  pm connectors inspect vantage - Vantage connector manual

SYNOPSIS
  pm connectors inspect vantage
  pm connectors inspect vantage --json
  pm credentials add <name> --connector vantage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads cost, budget, resource-management, segment, notification, and integration data from the Vantage API and writes budgets, folders, dashboards, cost reports, resource reports, saved filters, workspaces, teams, cost alerts, budget alerts, anomaly alerts, business metrics, virtual tag configs, segments, report notifications, recommendation views, network flow reports, Kubernetes efficiency reports, anomaly notifications, and canvases.

ICON
  id: vantage
  asset: icons/vantage.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.vantage.sh/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_token (secret) (required)

ETL STREAMS
  costs:
    primary key: id
    fields: amount(string), date(string), id(string), service(string)
  cost_reports:
    primary key: token
    fields: created_at(string), date_interval(string), filter(string), folder_token(string), title(string), token(string), updated_at(string), workspace_token(string)
  budgets:
    primary key: token
    fields: created_at(string), name(string), period_amount(string), period_duration(string), token(string), updated_at(string), workspace_token(string)
  folders:
    primary key: token
    fields: created_at(string), parent_folder_token(string), title(string), token(string), updated_at(string), workspace_token(string)
  dashboards:
    primary key: token
    fields: created_at(string), date_interval(string), title(string), token(string), updated_at(string), workspace_token(string)
  business_metrics:
    primary key: token
    fields: created_at(string), title(string), token(string), updated_at(string), workspace_token(string)
  resource_reports:
    primary key: token
    fields: created_at(string), filter(string), title(string), token(string), updated_at(string), workspace_token(string)
  recommendations:
    primary key: token
    fields: created_at(string), monthly_savings(string), state(string), token(string), type(string), updated_at(string)
  teams:
    primary key: token
    fields: created_at(string), name(string), token(string), updated_at(string)
  saved_filters:
    primary key: token
    fields: created_at(string), filter(string), title(string), token(string), updated_at(string), workspace_token(string)
  workspaces:
    primary key: token
    fields: created_at(string), name(string), token(string), updated_at(string)
  virtual_tag_configs:
    primary key: token
    fields: backfill_until(string), created_at(string), key(string), token(string), updated_at(string)
  tags:
    primary key: key
    fields: key(string), values_count(integer)
  cost_alerts:
    primary key: token
    fields: created_at(string), filter(string), threshold(string), title(string), token(string), updated_at(string), workspace_token(string)
  budget_alerts:
    primary key: token
    fields: budget_token(string), created_at(string), threshold(string), token(string), updated_at(string), user_token(string)
  anomaly_alerts:
    primary key: token
    fields: created_at(string), title(string), token(string), updated_at(string), workspace_token(string)
  managed_accounts:
    primary key: token
    fields: created_at(string), name(string), token(string), updated_at(string)
  financial_commitments:
    primary key: token
    fields: commitment_type(string), created_at(string), provider(string), token(string), updated_at(string)
  segments:
    primary key: token
    fields: created_at(string), description(string), filter(string), parent_segment_token(string), priority(integer), title(string), token(string), track_unallocated(boolean), updated_at(string), workspace_token(string)
  report_notifications:
    primary key: token
    fields: change(string), cost_report_token(string), created_at(string), frequency(string), recipient_channels(array), title(string), token(string), updated_at(string), user_tokens(array)
  recommendation_views:
    primary key: token
    fields: account_ids(array), billing_account_ids(array), created_at(string), end_date(string), min_savings(number), provider_ids(array), regions(array), start_date(string), tag_key(string), tag_value(string), title(string), token(string), types(array), updated_at(string), workspace_token(string)
  network_flow_reports:
    primary key: token
    fields: created_at(string), date_interval(string), end_date(string), filter(string), flow_direction(string), flow_weight(string), groupings(array), start_date(string), title(string), token(string), updated_at(string), workspace_token(string)
  kubernetes_efficiency_reports:
    primary key: token
    fields: aggregated_by(string), created_at(string), date_bucket(string), date_interval(string), end_date(string), filter(string), groupings(array), start_date(string), title(string), token(string), updated_at(string), workspace_token(string)
  anomaly_notifications:
    primary key: token
    fields: cost_report_token(string), created_at(string), recipient_channels(array), threshold(integer), token(string), updated_at(string), user_tokens(array)
  canvases:
    primary key: token
    fields: created_at(string), prompt(string), title(string), token(string), updated_at(string), workspace_token(string)
  invoices:
    primary key: token
    fields: account_token(string), billing_period_end(string), billing_period_start(string), created_at(string), status(string), token(string), total(string), updated_at(string)
  integrations:
    primary key: token
    fields: account_identifier(string), created_at(string), provider(string), token(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_budget:
    endpoint: POST /v2/budgets
    required fields: name, period_amount
    risk: external mutation; approval required
  update_budget:
    endpoint: PUT /v2/budgets/{{ record.budget_token }}
    required fields: budget_token
    risk: external mutation; approval required
  delete_budget:
    endpoint: DELETE /v2/budgets/{{ record.budget_token }}
    required fields: budget_token
    risk: destructive external mutation; approval required
  create_folder:
    endpoint: POST /v2/folders
    required fields: title
    risk: external mutation; approval required
  update_folder:
    endpoint: PUT /v2/folders/{{ record.folder_token }}
    required fields: folder_token
    risk: external mutation; approval required
  delete_folder:
    endpoint: DELETE /v2/folders/{{ record.folder_token }}
    required fields: folder_token
    risk: destructive external mutation; approval required
  create_dashboard:
    endpoint: POST /v2/dashboards
    required fields: title
    risk: external mutation; approval required
  update_dashboard:
    endpoint: PUT /v2/dashboards/{{ record.dashboard_token }}
    required fields: dashboard_token
    risk: external mutation; approval required
  delete_dashboard:
    endpoint: DELETE /v2/dashboards/{{ record.dashboard_token }}
    required fields: dashboard_token
    risk: destructive external mutation; approval required
  create_cost_report:
    endpoint: POST /v2/cost_reports
    required fields: title
    risk: external mutation; approval required
  update_cost_report:
    endpoint: PUT /v2/cost_reports/{{ record.cost_report_token }}
    required fields: cost_report_token
    risk: external mutation; approval required
  delete_cost_report:
    endpoint: DELETE /v2/cost_reports/{{ record.cost_report_token }}
    required fields: cost_report_token
    risk: destructive external mutation; approval required
  create_saved_filter:
    endpoint: POST /v2/saved_filters
    required fields: title, filter
    risk: external mutation; approval required
  delete_saved_filter:
    endpoint: DELETE /v2/saved_filters/{{ record.saved_filter_token }}
    required fields: saved_filter_token
    risk: destructive external mutation; approval required
  create_workspace:
    endpoint: POST /v2/workspaces
    required fields: name
    risk: external mutation; approval required
  delete_workspace:
    endpoint: DELETE /v2/workspaces/{{ record.workspace_token }}
    required fields: workspace_token
    risk: destructive external mutation; approval required
  create_team:
    endpoint: POST /v2/teams
    required fields: name
    risk: external mutation; approval required
  delete_team:
    endpoint: DELETE /v2/teams/{{ record.team_token }}
    required fields: team_token
    risk: destructive external mutation; approval required
  create_cost_alert:
    endpoint: POST /v2/cost_alerts
    required fields: title
    risk: external mutation; approval required
  update_cost_alert:
    endpoint: PUT /v2/cost_alerts/{{ record.cost_alert_token }}
    required fields: cost_alert_token
    risk: external mutation; approval required
  delete_cost_alert:
    endpoint: DELETE /v2/cost_alerts/{{ record.cost_alert_token }}
    required fields: cost_alert_token
    risk: destructive external mutation; approval required
  create_budget_alert:
    endpoint: POST /v2/budget_alerts
    required fields: budget_token, user_token, threshold
    risk: external mutation; approval required
  delete_budget_alert:
    endpoint: DELETE /v2/budget_alerts/{{ record.budget_alert_token }}
    required fields: budget_alert_token
    risk: destructive external mutation; approval required
  create_business_metric:
    endpoint: POST /v2/business_metrics
    required fields: title
    risk: external mutation; approval required
  update_business_metric:
    endpoint: PUT /v2/business_metrics/{{ record.business_metric_token }}
    required fields: business_metric_token
    risk: external mutation; approval required
  delete_business_metric:
    endpoint: DELETE /v2/business_metrics/{{ record.business_metric_token }}
    required fields: business_metric_token
    risk: destructive external mutation; approval required
  delete_virtual_tag_config:
    endpoint: DELETE /v2/virtual_tag_configs/{{ record.token }}
    required fields: token
    risk: destructive external mutation; approval required
  update_saved_filter:
    endpoint: PUT /v2/saved_filters/{{ record.saved_filter_token }}
    required fields: saved_filter_token
    risk: external mutation; approval required
  update_team:
    endpoint: PUT /v2/teams/{{ record.team_token }}
    required fields: team_token
    risk: external mutation; approval required
  update_workspace:
    endpoint: PUT /v2/workspaces/{{ record.workspace_token }}
    required fields: workspace_token
    risk: external mutation; approval required
  update_anomaly_alert:
    endpoint: PUT /v2/anomaly_alerts/{{ record.anomaly_alert_token }}
    required fields: anomaly_alert_token
    risk: external mutation; approval required
  update_budget_alert:
    endpoint: PUT /v2/budget_alerts/{{ record.budget_alert_token }}
    required fields: budget_alert_token
    risk: external mutation; approval required
  update_virtual_tag_config:
    endpoint: PUT /v2/virtual_tag_configs/{{ record.token }}
    required fields: token
    risk: external mutation; approval required
  create_segment:
    endpoint: POST /v2/segments
    required fields: title
    risk: external mutation; approval required
  update_segment:
    endpoint: PUT /v2/segments/{{ record.segment_token }}
    required fields: segment_token
    risk: external mutation; approval required
  delete_segment:
    endpoint: DELETE /v2/segments/{{ record.segment_token }}
    required fields: segment_token
    risk: destructive external mutation; approval required
  create_report_notification:
    endpoint: POST /v2/report_notifications
    required fields: title, cost_report_token, frequency, change
    risk: external mutation; approval required
  update_report_notification:
    endpoint: PUT /v2/report_notifications/{{ record.report_notification_token }}
    required fields: report_notification_token
    risk: external mutation; approval required
  delete_report_notification:
    endpoint: DELETE /v2/report_notifications/{{ record.report_notification_token }}
    required fields: report_notification_token
    risk: destructive external mutation; approval required
  create_recommendation_view:
    endpoint: POST /v2/recommendation_views
    required fields: title, workspace_token
    risk: external mutation; approval required
  update_recommendation_view:
    endpoint: PUT /v2/recommendation_views/{{ record.recommendation_view_token }}
    required fields: recommendation_view_token
    risk: external mutation; approval required
  delete_recommendation_view:
    endpoint: DELETE /v2/recommendation_views/{{ record.recommendation_view_token }}
    required fields: recommendation_view_token
    risk: destructive external mutation; approval required
  create_network_flow_report:
    endpoint: POST /v2/network_flow_reports
    required fields: workspace_token, title
    risk: external mutation; approval required
  update_network_flow_report:
    endpoint: PUT /v2/network_flow_reports/{{ record.network_flow_report_token }}
    required fields: network_flow_report_token
    risk: external mutation; approval required
  delete_network_flow_report:
    endpoint: DELETE /v2/network_flow_reports/{{ record.network_flow_report_token }}
    required fields: network_flow_report_token
    risk: destructive external mutation; approval required
  create_kubernetes_efficiency_report:
    endpoint: POST /v2/kubernetes_efficiency_reports
    required fields: workspace_token, title
    risk: external mutation; approval required
  update_kubernetes_efficiency_report:
    endpoint: PUT /v2/kubernetes_efficiency_reports/{{ record.kubernetes_efficiency_report_token }}
    required fields: kubernetes_efficiency_report_token
    risk: external mutation; approval required
  delete_kubernetes_efficiency_report:
    endpoint: DELETE /v2/kubernetes_efficiency_reports/{{ record.kubernetes_efficiency_report_token }}
    required fields: kubernetes_efficiency_report_token
    risk: destructive external mutation; approval required
  create_anomaly_notification:
    endpoint: POST /v2/anomaly_notifications
    required fields: cost_report_token
    risk: external mutation; approval required
  update_anomaly_notification:
    endpoint: PUT /v2/anomaly_notifications/{{ record.anomaly_notification_token }}
    required fields: anomaly_notification_token
    risk: external mutation; approval required
  delete_anomaly_notification:
    endpoint: DELETE /v2/anomaly_notifications/{{ record.anomaly_notification_token }}
    required fields: anomaly_notification_token
    risk: destructive external mutation; approval required
  create_canvas:
    endpoint: POST /v2/canvases
    required fields: title, prompt
    risk: external mutation; approval required
  update_canvas:
    endpoint: PUT /v2/canvases/{{ record.canvas_token }}
    required fields: canvas_token
    risk: external mutation; approval required
  delete_canvas:
    endpoint: DELETE /v2/canvases/{{ record.canvas_token }}
    required fields: canvas_token
    risk: destructive external mutation; approval required
  create_resource_report:
    endpoint: POST /v2/resource_reports
    required fields: workspace_token
    risk: external mutation; approval required
  update_resource_report:
    endpoint: PUT /v2/resource_reports/{{ record.resource_report_token }}
    required fields: resource_report_token
    risk: external mutation; approval required
  delete_resource_report:
    endpoint: DELETE /v2/resource_reports/{{ record.resource_report_token }}
    required fields: resource_report_token
    risk: destructive external mutation; approval required
  create_virtual_tag_config:
    endpoint: POST /v2/virtual_tag_configs
    required fields: key, overridable, values
    risk: external mutation; approval required

SECURITY
  read risk: external Vantage API read of cloud cost, budget, resource-management, segment, notification, and integration data
  write risk: external mutation of Vantage cost-management configuration (budgets, folders, dashboards, cost reports, resource reports, saved filters, workspaces, teams, alerts, business metrics, virtual tag configs, segments, report notifications, recommendation views, network flow reports, Kubernetes efficiency reports, anomaly notifications, canvases); approval required
  approval: read: none; write: required for every action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect vantage

  # Inspect as structured JSON
  pm connectors inspect vantage --json

AGENT WORKFLOW
  - Run pm connectors inspect vantage before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
