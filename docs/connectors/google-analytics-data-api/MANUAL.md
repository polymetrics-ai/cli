# pm connectors inspect google-analytics-data-api

```text
NAME
  pm connectors inspect google-analytics-data-api - Google Analytics 4 (GA4) connector manual

SYNOPSIS
  pm connectors inspect google-analytics-data-api
  pm connectors inspect google-analytics-data-api --json
  pm credentials add <name> --connector google-analytics-data-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Analytics 4 report presets plus bounded v1beta/v1alpha metadata resources from the Google Analytics Data API. The provider-derived inventory contains 24 semantic operations (20 reads, 4 writes): 11 are executable read operations and 13 remain planned behind the shared provider-query or closed reverse-ETL foundations; no raw API access is exposed.

ICON
  asset: icons/google-analytics.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  oauth2_bearer: OAuth2 bearer token with Google Analytics Data API read access.
    config: property_ids
    secrets: access_token, credentials
    supports: read=true write=false

CONFIGURATION
  property_ids (required): Comma, space, or newline separated GA4 numeric property IDs; native reads use the first property ID per read call.
  property_id: Optional single GA4 numeric property ID for direct metadata commands; defaults to the first property_ids value.
  audience_export_id: Audience export ID used by the get audience export direct command.
  audience_list_id: Audience list ID used by the v1alpha audience-list get direct command.
  recurring_audience_list_id: Recurring audience list ID used by the v1alpha recurring-audience-list get direct command.
  report_task_id: Report task ID used by the v1alpha report-task get direct command.
  base_url default=https://analyticsdata.googleapis.com: Analytics Data API base URL override for local fixture tests only.
  date_ranges_start_date default=30daysAgo: GA4 report start date, either YYYY-MM-DD or a GA4 relative token such as 30daysAgo.
  date_ranges_end_date default=today: GA4 report end date, either YYYY-MM-DD or a GA4 relative token such as today or yesterday.
  page_size default=10000: Native runReport page size; must be between 1 and 250000.
  max_pages: Native runReport page cap. Use a positive integer, 0, all, or unlimited for unbounded reads.
  mode: Set to fixture for credential-free connector-owned tests; do not use for live provider validation.
  keep_empty_rows: Legacy compatibility flag retained for credential compatibility; native preset reads currently send false.
  convert_conversions_event: Legacy compatibility flag retained for credential compatibility.
  custom_reports_array: Legacy custom report JSON string. Custom reports remain outside this documented parity slice.
  lookback_window: Legacy lookback-window setting retained for credential compatibility.
  subscription_tier: Informational GA4 property tier for quota planning.
  window_in_days: Legacy window setting retained for credential compatibility.
  access_token (secret): OAuth2 bearer access token with Analytics Data API read access; prefer --from-env or --value-stdin.
  credentials (secret): Legacy flattened bearer token payload; prefer access_token for new credentials.

ETL STREAMS
  daily_active_users: Active users, new users, and sessions broken down by day.
    primary key: property_id, date
    cursor: date
    fields: property_id(string), date(string), activeUsers(number), newUsers(number), sessions(number)
  website_overview: Top-line engagement metrics broken down by day.
    primary key: property_id, date
    cursor: date
    fields: property_id(string), date(string), activeUsers(number), newUsers(number), sessions(number), screenPageViews(number), averageSessionDuration(number), bounceRate(number)
  traffic_sources: Sessions and users by acquisition source / medium per day.
    primary key: property_id, date, sessionSource, sessionMedium
    cursor: date
    fields: property_id(string), date(string), sessionSource(string), sessionMedium(string), sessions(number), activeUsers(number), newUsers(number), engagedSessions(number)
  devices: Users and sessions by device category, OS, and browser per day.
    primary key: property_id, date, deviceCategory, operatingSystem, browser
    cursor: date
    fields: property_id(string), date(string), deviceCategory(string), operatingSystem(string), browser(string), activeUsers(number), sessions(number), screenPageViews(number)
  pages: Page views and engagement by page path and title per day.
    primary key: property_id, date, pagePath, pageTitle
    cursor: date
    fields: property_id(string), date(string), pagePath(string), pageTitle(string), screenPageViews(number), activeUsers(number), averageSessionDuration(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
  Source modes: full_refresh, incremental

PAGINATION
  type: offset_limit
  page size field: page_size
  page limit field: max_pages
  default limit: 10000

SECURITY
  read risk: external Google Analytics Data API reads for configured properties; 10 direct reads are fixed-target, bounded, and JSON-redacted, while runReport is available through five bounded presets
  write risk: unsupported; four provider-side creates are typed as planned operations but have no executable reverse-ETL action
  mutation risk: none
  approval: none for read-only operations; future provider-side creates require plan, preview, explicit approval, execute, redaction, and operation/idempotency handling before being advertised
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read GA4 report presets and bounded v1beta/v1alpha Analytics Data API metadata without exposing raw API access.
  Usage: pm google-analytics-data-api <reports|metadata|audience-exports|property-quotas|audience-lists|recurring-audience-lists|report-tasks> <command> [flags]
  Source CLI: Google Analytics Data API (v1beta + v1alpha discovery revision 20260803, retrieved 2026-08-05)
  Global flags:
    --credential (string): Credential profile name; never pass secret values as flags.: maps_to=config.credential
    --limit (integer): Maximum records emitted by report stream commands.: maps_to=query.limit
    --json (boolean): Render machine-readable output when supported.
  Report preset streams
    reports daily-active-users - Read the daily_active_users GA4 report preset. [intent=etl availability=implemented stream=daily_active_users]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
    reports website-overview - Read the website_overview GA4 report preset. [intent=etl availability=implemented stream=website_overview]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
    reports traffic-sources - Read the traffic_sources GA4 report preset. [intent=etl availability=implemented stream=traffic_sources]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
    reports devices - Read the devices GA4 report preset. [intent=etl availability=implemented stream=devices]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
    reports pages - Read the pages GA4 report preset. [intent=etl availability=implemented stream=pages]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
    reports run-funnel - Planned typed v1alpha funnel-report operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.run_funnel_report]; approval: none: planned read-only POST query; risk: medium; notes: Blocked by shared provider-query foundation #2985; no generic raw method, path, or body flags are exposed.; flags: --property-id
    reports run-realtime - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.run_realtime_report]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
    reports run-pivot - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.run_pivot_report]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
    reports batch-run - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.batch_run_reports]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
    reports batch-run-pivot - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.batch_run_pivot_reports]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
    reports check-compatibility - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.check_compatibility]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  Metadata direct reads
    metadata get - Get GA4 property metadata. [intent=direct_read availability=implemented operation=google-analytics-data-api.get_metadata]; approval: none; risk: low; notes: Fixed GET metadata endpoint; no raw method/path/body flags.; flags: --property-id
  Audience export metadata and planned user export operations
    audience-exports list - List GA4 audience export metadata for one property. [intent=direct_read availability=implemented operation=google-analytics-data-api.list_audience_exports]; approval: none; risk: medium; notes: Fixed GET list endpoint; output is bounded and JSON-redacted.; flags: --property-id, --page-size, --page-token
    audience-exports get - Get one GA4 audience export metadata resource. [intent=direct_read availability=implemented operation=google-analytics-data-api.get_audience_export]; approval: none; risk: medium; notes: Fixed GET metadata endpoint for an audience export; does not query user rows.; flags: --property-id, --audience-export-id
    audience-exports query - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.query_audience_export]; approval: none; risk: high; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id, --audience-export-id
    audience-exports create - Planned typed GA4 operation; not executable in this slice. [intent=reverse_etl availability=planned operation=google-analytics-data-api.create_audience_export]; approval: planned reverse ETL support would require plan, preview, explicit approval, execute; risk: high; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  v1alpha property quota snapshot
    property-quotas get - Get the v1alpha GA4 property quota snapshot. [intent=direct_read availability=implemented operation=google-analytics-data-api.get_property_quotas_snapshot]; approval: none; risk: low; notes: Fixed v1alpha GET quota endpoint; no raw method, path, or body flags.; flags: --property-id
  v1alpha audience-list metadata
    audience-lists list - List v1alpha GA4 audience-list metadata for one property. [intent=direct_read availability=implemented operation=google-analytics-data-api.list_audience_lists]; approval: none; risk: medium; notes: Fixed v1alpha GET list endpoint; it returns metadata only, not audience user rows.; flags: --property-id, --page-size, --page-token
    audience-lists get - Get one v1alpha GA4 audience-list metadata resource. [intent=direct_read availability=implemented operation=google-analytics-data-api.get_audience_list]; approval: none; risk: medium; notes: Fixed v1alpha GET metadata endpoint; it does not query audience user rows.; flags: --property-id, --audience-list-id
    audience-lists create - Planned typed audience-list creation; not executable in this slice. [intent=reverse_etl availability=planned operation=google-analytics-data-api.create_audience_list]; approval: planned reverse ETL support would require plan, preview, explicit approval, execute, redaction, and operation/idempotency evidence; risk: high; notes: Blocked planned write; no generic raw method, path, or body flags are exposed.; flags: --property-id
    audience-lists query - Planned typed audience-list query; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.query_audience_list]; approval: none: planned read-only POST query; risk: high; notes: Blocked by shared provider-query/redaction foundation #2985; no generic raw method, path, or body flags are exposed.; flags: --property-id, --audience-list-id
  v1alpha recurring-audience-list metadata
    recurring-audience-lists list - List v1alpha GA4 recurring-audience-list metadata for one property. [intent=direct_read availability=implemented operation=google-analytics-data-api.list_recurring_audience_lists]; approval: none; risk: medium; notes: Fixed v1alpha GET list endpoint; it returns metadata only.; flags: --property-id, --page-size, --page-token
    recurring-audience-lists get - Get one v1alpha GA4 recurring-audience-list metadata resource. [intent=direct_read availability=implemented operation=google-analytics-data-api.get_recurring_audience_list]; approval: none; risk: medium; notes: Fixed v1alpha GET metadata endpoint; it does not query audience user rows.; flags: --property-id, --recurring-audience-list-id
    recurring-audience-lists create - Planned typed recurring audience-list creation; not executable in this slice. [intent=reverse_etl availability=planned operation=google-analytics-data-api.create_recurring_audience_list]; approval: planned reverse ETL support would require plan, preview, explicit approval, execute, redaction, and operation/idempotency evidence; risk: high; notes: Blocked planned write; no generic raw method, path, or body flags are exposed.; flags: --property-id
  v1alpha report-task metadata
    report-tasks list - List v1alpha GA4 report-task metadata for one property. [intent=direct_read availability=implemented operation=google-analytics-data-api.list_report_tasks]; approval: none; risk: medium; notes: Fixed v1alpha GET list endpoint; it returns task metadata, not report task content.; flags: --property-id, --page-size, --page-token
    report-tasks get - Get one v1alpha GA4 report-task metadata resource. [intent=direct_read availability=implemented operation=google-analytics-data-api.get_report_task]; approval: none; risk: medium; notes: Fixed v1alpha GET metadata endpoint; it does not query report task content.; flags: --property-id, --report-task-id
    report-tasks create - Planned typed report-task creation; not executable in this slice. [intent=reverse_etl availability=planned operation=google-analytics-data-api.create_report_task]; approval: planned reverse ETL support would require plan, preview, explicit approval, execute, redaction, and operation/idempotency evidence; risk: high; notes: Blocked planned write; no generic raw method, path, or body flags are exposed.; flags: --property-id
    report-tasks query - Planned typed report-task query; not executable in this slice. [intent=direct_read availability=planned operation=google-analytics-data-api.query_report_task]; approval: none: planned read-only POST query; risk: medium; notes: Blocked by shared provider-query foundation #2985; no generic raw method, path, or body flags are exposed.; flags: --property-id, --report-task-id
  Help topics:
    auth - Use OAuth2 bearer credentials from env/stdin-backed credential storage; never paste tokens into chat or shell history.
    limits - Report streams are page-bounded by page_size/max_pages config; direct reads are byte-bounded and JSON-redacted.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-analytics-data-api

  # Inspect as structured JSON
  pm connectors inspect google-analytics-data-api --json

AGENT WORKFLOW
  - Run pm connectors inspect google-analytics-data-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
