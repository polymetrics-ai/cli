# pm connectors inspect google-analytics-data-api

```text
NAME
  pm connectors inspect google-analytics-data-api - Google Analytics 4 (GA4) connector manual

SYNOPSIS
  pm connectors inspect google-analytics-data-api
  pm connectors inspect google-analytics-data-api --json
  pm credentials add <name> --connector google-analytics-data-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Analytics 4 reports (active users, traffic sources, devices, pages) from the Analytics Data API runReport endpoint. Read-only. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/google-analytics.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  convert_conversions_event
  credentials
  custom_reports_array
  date_ranges_end_date
  date_ranges_start_date
  keep_empty_rows
  lookback_window
  mode
  property_ids
  subscription_tier
  window_in_days

ETL STREAMS
  daily_active_users:
    primary key: property_id, date
    cursor: date
    fields: activeUsers(), date(), newUsers(), property_id(), sessions()
  website_overview:
    primary key: property_id, date
    cursor: date
    fields: activeUsers(), averageSessionDuration(), bounceRate(), date(), newUsers(), property_id(), screenPageViews(), sessions()
  traffic_sources:
    primary key: property_id, date, sessionSource, sessionMedium
    cursor: date
    fields: activeUsers(), date(), engagedSessions(), newUsers(), property_id(), sessionMedium(), sessionSource(), sessions()
  devices:
    primary key: property_id, date, deviceCategory, operatingSystem, browser
    cursor: date
    fields: activeUsers(), browser(), date(), deviceCategory(), operatingSystem(), property_id(), screenPageViews(), sessions()
  pages:
    primary key: property_id, date, pagePath, pageTitle
    cursor: date
    fields: activeUsers(), averageSessionDuration(), date(), pagePath(), pageTitle(), property_id(), screenPageViews()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Google Analytics 4 (GA4) API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
