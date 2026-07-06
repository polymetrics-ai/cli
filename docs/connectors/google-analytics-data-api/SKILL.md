---
name: pm-google-analytics-data-api
description: Google Analytics 4 (GA4) connector knowledge and safe action guide.
---

# pm-google-analytics-data-api

## Purpose

Reads Google Analytics 4 reports (active users, traffic sources, devices, pages) from the Analytics Data API runReport endpoint. Read-only. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/google-analytics.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- convert_conversions_event
- credentials
- custom_reports_array
- date_ranges_end_date
- date_ranges_start_date
- keep_empty_rows
- lookback_window
- mode
- property_ids
- subscription_tier
- window_in_days

## ETL Streams

- daily_active_users:
  - primary key: property_id, date
  - cursor: date
  - fields: activeUsers(), date(), newUsers(), property_id(), sessions()
- website_overview:
  - primary key: property_id, date
  - cursor: date
  - fields: activeUsers(), averageSessionDuration(), bounceRate(), date(), newUsers(), property_id(), screenPageViews(), sessions()
- traffic_sources:
  - primary key: property_id, date, sessionSource, sessionMedium
  - cursor: date
  - fields: activeUsers(), date(), engagedSessions(), newUsers(), property_id(), sessionMedium(), sessionSource(), sessions()
- devices:
  - primary key: property_id, date, deviceCategory, operatingSystem, browser
  - cursor: date
  - fields: activeUsers(), browser(), date(), deviceCategory(), operatingSystem(), property_id(), screenPageViews(), sessions()
- pages:
  - primary key: property_id, date, pagePath, pageTitle
  - cursor: date
  - fields: activeUsers(), averageSessionDuration(), date(), pagePath(), pageTitle(), property_id(), screenPageViews()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Google Analytics 4 (GA4) API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-analytics-data-api
```

### Inspect as structured JSON

```bash
pm connectors inspect google-analytics-data-api --json
```

## Agent Rules

- Run pm connectors inspect google-analytics-data-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
