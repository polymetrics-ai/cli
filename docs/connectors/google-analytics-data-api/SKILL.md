---
name: pm-google-analytics-data-api
description: Google Analytics 4 (GA4) connector knowledge and safe action guide.
---

# pm-google-analytics-data-api

## Purpose

Reads fixed Google Analytics 4 reports from the Analytics Data API runReport endpoint through declared response-header projection.

## Icon

- id: google-analytics
- asset: icons/google-analytics.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- end_date
- property_ids (required)
- start_date
- access_token (secret) (required)

## ETL Streams

- daily_active_users:
  - primary key: property_id, date
  - cursor: date
  - fields: activeUsers(number), date(string), newUsers(number), property_id(string), sessions(number)
- website_overview:
  - primary key: property_id, date
  - cursor: date
  - fields: activeUsers(number), averageSessionDuration(number), bounceRate(number), date(string), newUsers(number), property_id(string), screenPageViews(number), sessions(number)
- traffic_sources:
  - primary key: property_id, date, sessionSource, sessionMedium
  - cursor: date
  - fields: activeUsers(number), date(string), engagedSessions(number), newUsers(number), property_id(string), sessionMedium(string), sessionSource(string), sessions(number)
- devices:
  - primary key: property_id, date, deviceCategory, operatingSystem, browser
  - cursor: date
  - fields: activeUsers(number), browser(string), date(string), deviceCategory(string), operatingSystem(string), property_id(string), screenPageViews(number), sessions(number)
- pages:
  - primary key: property_id, date, pagePath, pageTitle
  - cursor: date
  - fields: activeUsers(number), averageSessionDuration(number), date(string), pagePath(string), pageTitle(string), property_id(string), screenPageViews(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: Bounded GA4 reports use a fixed provider route and bearer token; response headers may map only source-declared positional values.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Declared google-analytics-data-api API commands.
- Usage: pm google-analytics-data-api <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Other Commands
  - operations delete-v1-name - Declared direct write: DELETE /v1/{+name}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /v1/{+name}.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-name - Declared direct read: GET /v1/{+name}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1-name-cancel - Declared direct write: POST /v1/{+name}:cancel. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1/{+name}:cancel.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1alpha-name - Declared direct read: GET /v1alpha/{+name}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1alpha-name-query - Declared direct write: POST /v1alpha/{+name}:query. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1alpha/{+name}:query.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1alpha-parent-audience-lists - Declared direct read: GET /v1alpha/{+parent}/audienceLists. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1alpha-parent-audience-lists - Declared direct write: POST /v1alpha/{+parent}/audienceLists. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1alpha/{+parent}/audienceLists.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1alpha-parent-recurring-audience-lists - Declared direct read: GET /v1alpha/{+parent}/recurringAudienceLists. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1alpha-parent-recurring-audience-lists - Declared direct write: POST /v1alpha/{+parent}/recurringAudienceLists. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1alpha/{+parent}/recurringAudienceLists.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1alpha-parent-report-tasks - Declared direct read: GET /v1alpha/{+parent}/reportTasks. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1alpha-parent-report-tasks - Declared direct write: POST /v1alpha/{+parent}/reportTasks. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1alpha/{+parent}/reportTasks.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1alpha-property-run-funnel-report - Declared direct write: POST /v1alpha/{+property}:runFunnelReport. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1alpha/{+property}:runFunnelReport.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1alpha-property-run-report - Declared direct write: POST /v1alpha/{+property}:runReport. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1alpha/{+property}:runReport.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1beta-name - Declared direct read: GET /v1beta/{+name}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1beta-name-query - Declared direct write: POST /v1beta/{+name}:query. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+name}:query.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1beta-parent-audience-exports - Declared direct read: GET /v1beta/{+parent}/audienceExports. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1beta-parent-audience-exports - Declared direct write: POST /v1beta/{+parent}/audienceExports. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+parent}/audienceExports.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1beta-property-batch-run-pivot-reports - Declared direct write: POST /v1beta/{+property}:batchRunPivotReports. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+property}:batchRunPivotReports.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1beta-property-batch-run-reports - Declared direct write: POST /v1beta/{+property}:batchRunReports. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+property}:batchRunReports.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1beta-property-check-compatibility - Declared direct write: POST /v1beta/{+property}:checkCompatibility. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+property}:checkCompatibility.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1beta-property-run-pivot-report - Declared direct write: POST /v1beta/{+property}:runPivotReport. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+property}:runPivotReport.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1beta-property-run-realtime-report - Declared direct write: POST /v1beta/{+property}:runRealtimeReport. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+property}:runRealtimeReport.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1beta-property-run-report - Declared direct write: POST /v1beta/{+property}:runReport. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /v1beta/{+property}:runReport.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations hook-internal-connectors-google-analytics-data-api-daily-active-users-read-path - Declared etl: HOOK internal/connectors/google-analytics-data-api daily_active_users read path. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations hook-internal-connectors-google-analytics-data-api-website-overview-read-path - Declared etl: HOOK internal/connectors/google-analytics-data-api website_overview read path. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations hook-internal-connectors-google-analytics-data-api-traffic-sources-read-path - Declared etl: HOOK internal/connectors/google-analytics-data-api traffic_sources read path. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations hook-internal-connectors-google-analytics-data-api-devices-read-path - Declared etl: HOOK internal/connectors/google-analytics-data-api devices read path. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations hook-internal-connectors-google-analytics-data-api-pages-read-path - Declared etl: HOOK internal/connectors/google-analytics-data-api pages read path. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.

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
