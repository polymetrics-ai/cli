---
name: pm-source-google-analytics-data-api
description: Google Analytics 4 (GA4) connector knowledge and safe action guide.
---

# pm-source-google-analytics-data-api

## Purpose

Google Analytics 4 (GA4) catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/google-analytics.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

## Runtime Capabilities

- metadata=true
- check=false
- catalog=false
- read=false
- write=false
- query=false
- etl=false
- reverse_etl=false
- unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

## Native Port Plan

- family: declarative_http_source
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Data API changelog: https://developers.google.com/analytics/devguides/reporting/data/v1/changelog

## Configuration

- convert_conversions_event (boolean): Enables conversion of `conversions:*` event metrics from integers to floats. This is beneficial for preventing data rounding when the API returns float values for any `conversio...
- credentials (object): Credentials for the service
- custom_reports_array (array): You can add your Custom Analytics report by creating one.
- date_ranges_end_date (string): The end date from which to replicate report data in the format YYYY-MM-DD. Data generated after this date will not be included in the report. Not applied to custom Cohort report...
- date_ranges_start_date (string): The start date from which to replicate report data in the format YYYY-MM-DD. Data generated before this date will not be included in the report. Not applied to custom Cohort rep...
- keep_empty_rows (boolean): If false, each row with all metrics equal to 0 will not be returned. If true, these rows will be returned if they are not separately removed by a filter. More information is ava...
- lookback_window (integer): Since attribution changes after the event date, and Google Analytics has a data processing latency, we should specify how many days in the past we should refresh the data in eve...
- property_ids (array) required: A list of your Property IDs. The Property ID is a unique number assigned to each property in Google Analytics, found in your GA4 property URL. This ID allows the connector to tr...
- subscription_tier (string): Quota tier of the Google Analytics 4 properties being queried. Determines the per-property rate-limit policy applied locally once the tier-aware rate-limit budget is activated. ...
- window_in_days (integer): manual intervention needed
- secret fields: credentials.access_token, credentials.client_secret, credentials.credentials_json, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-google-analytics-data-api
```

### Inspect as JSON

```bash
pm connectors inspect source-google-analytics-data-api --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Data API changelog](https://developers.google.com/analytics/devguides/reporting/data/v1/changelog)
