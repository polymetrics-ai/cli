---
name: pm-google-analytics-data-api
description: Google Analytics 4 (GA4) connector knowledge and safe action guide.
---

# pm-google-analytics-data-api

## Purpose

Reads Google Analytics 4 report presets and bounded metadata/audience-export resources from the Google Analytics Data API v1beta. Read-only; POST report/query operations that require a shared provider-query foundation remain planned rather than exposed as raw API calls.

## Icon

- asset: icons/google-analytics.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- oauth2_bearer: OAuth2 bearer token with Google Analytics Data API read access.
  - config: property_ids
  - secrets: access_token, credentials
  - supports: read=true write=false

## Configuration

- property_ids (required): Comma, space, or newline separated GA4 numeric property IDs; native reads use the first property ID per read call.
- property_id: Optional single GA4 numeric property ID for direct metadata/audience-export commands; defaults to the first property_ids value.
- audience_export_id: Audience export ID used by the get audience export direct command.
- base_url default=https://analyticsdata.googleapis.com: Analytics Data API base URL override for local fixture tests only.
- date_ranges_start_date default=30daysAgo: GA4 report start date, either YYYY-MM-DD or a GA4 relative token such as 30daysAgo.
- date_ranges_end_date default=today: GA4 report end date, either YYYY-MM-DD or a GA4 relative token such as today or yesterday.
- page_size default=10000: Native runReport page size; must be between 1 and 250000.
- max_pages: Native runReport page cap. Use a positive integer, 0, all, or unlimited for unbounded reads.
- mode: Set to fixture for credential-free connector-owned tests; do not use for live provider validation.
- keep_empty_rows: Legacy compatibility flag retained for credential compatibility; native preset reads currently send false.
- convert_conversions_event: Legacy compatibility flag retained for credential compatibility.
- custom_reports_array: Legacy custom report JSON string. Custom reports remain outside this documented parity slice.
- lookback_window: Legacy lookback-window setting retained for credential compatibility.
- subscription_tier: Informational GA4 property tier for quota planning.
- window_in_days: Legacy window setting retained for credential compatibility.
- access_token (secret): OAuth2 bearer access token with Analytics Data API read access; prefer --from-env or --value-stdin.
- credentials (secret): Legacy flattened bearer token payload; prefer access_token for new credentials.

## ETL Streams

- daily_active_users: Active users, new users, and sessions broken down by day.
  - primary key: property_id, date
  - cursor: date
  - fields: property_id(string), date(string), activeUsers(number), newUsers(number), sessions(number)
- website_overview: Top-line engagement metrics broken down by day.
  - primary key: property_id, date
  - cursor: date
  - fields: property_id(string), date(string), activeUsers(number), newUsers(number), sessions(number), screenPageViews(number), averageSessionDuration(number), bounceRate(number)
- traffic_sources: Sessions and users by acquisition source / medium per day.
  - primary key: property_id, date, sessionSource, sessionMedium
  - cursor: date
  - fields: property_id(string), date(string), sessionSource(string), sessionMedium(string), sessions(number), activeUsers(number), newUsers(number), engagedSessions(number)
- devices: Users and sessions by device category, OS, and browser per day.
  - primary key: property_id, date, deviceCategory, operatingSystem, browser
  - cursor: date
  - fields: property_id(string), date(string), deviceCategory(string), operatingSystem(string), browser(string), activeUsers(number), sessions(number), screenPageViews(number)
- pages: Page views and engagement by page path and title per day.
  - primary key: property_id, date, pagePath, pageTitle
  - cursor: date
  - fields: property_id(string), date(string), pagePath(string), pageTitle(string), screenPageViews(number), activeUsers(number), averageSessionDuration(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
- Source modes: full_refresh, incremental

## Pagination

- type: offset_limit
- page size field: page_size
- page limit field: max_pages
- default limit: 10000

## Security

- read risk: external Google Analytics Data API reads for configured properties; direct reads are fixed-target, bounded, and JSON-redacted
- write risk: unsupported
- mutation risk: none
- approval: none for read-only operations; future audience-export creation would require plan, preview, explicit approval, and execute before being advertised
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read GA4 report presets and bounded Analytics Data API metadata without exposing raw API access.
- Usage: pm google-analytics-data-api <reports|metadata|audience-exports> <command> [flags]
- Source CLI: Google Analytics Data API (v1beta discovery revision 20260729)
- Global flags:
  - --credential (string): Credential profile name; never pass secret values as flags.: maps_to=config.credential
  - --limit (integer): Maximum records emitted by report stream commands.: maps_to=query.limit
  - --json (boolean): Render machine-readable output when supported.
- Report preset streams
  - reports daily-active-users - Read the daily_active_users GA4 report preset. [intent=etl availability=implemented stream=daily_active_users]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
  - reports website-overview - Read the website_overview GA4 report preset. [intent=etl availability=implemented stream=website_overview]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
  - reports traffic-sources - Read the traffic_sources GA4 report preset. [intent=etl availability=implemented stream=traffic_sources]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
  - reports devices - Read the devices GA4 report preset. [intent=etl availability=implemented stream=devices]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
  - reports pages - Read the pages GA4 report preset. [intent=etl availability=implemented stream=pages]; approval: none; risk: low; notes: Native runReport preset; the fixed command emits stream records and does not expose raw method, URL, or body flags.
  - reports run-realtime - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  - reports run-pivot - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  - reports batch-run - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  - reports batch-run-pivot - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  - reports check-compatibility - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned]; approval: none; risk: medium; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
- Metadata direct reads
  - metadata get - Get GA4 property metadata. [intent=direct_read availability=implemented]; approval: none; risk: low; notes: Fixed GET metadata endpoint; no raw method/path/body flags.; flags: --property-id
- Audience export metadata and planned user export operations
  - audience-exports list - List GA4 audience export metadata for one property. [intent=direct_read availability=implemented]; approval: none; risk: medium; notes: Fixed GET list endpoint; output is bounded and JSON-redacted.; flags: --property-id, --page-size, --page-token
  - audience-exports get - Get one GA4 audience export metadata resource. [intent=direct_read availability=implemented]; approval: none; risk: medium; notes: Fixed GET metadata endpoint for an audience export; does not query user rows.; flags: --property-id, --audience-export-id
  - audience-exports query - Planned typed GA4 operation; not executable in this slice. [intent=direct_read availability=planned]; approval: none; risk: high; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
  - audience-exports create - Planned typed GA4 operation; not executable in this slice. [intent=reverse_etl availability=planned]; approval: planned reverse ETL support would require plan, preview, explicit approval, execute; risk: high; notes: Blocked/planned fixed operation metadata; no generic raw API escape hatch.; flags: --property-id
- Help topics:
  - auth - Use OAuth2 bearer credentials from env/stdin-backed credential storage; never paste tokens into chat or shell history.
  - limits - Report streams are page-bounded by page_size/max_pages config; direct reads are byte-bounded and JSON-redacted.

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
