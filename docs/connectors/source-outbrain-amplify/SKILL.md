---
name: pm-source-outbrain-amplify
description: Outbrain Amplify connector knowledge and safe action guide.
---

# pm-source-outbrain-amplify

## Purpose

Outbrain Amplify catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
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

- family: custom_go_port
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Outbrain Amplify documentation: https://developer.outbrain.com/home-page/amplify-api/documentation/

## Configuration

- conversion_count (string): The definition of conversion count in reports. See <a href="https://amplifyv01.docs.apiary.io/#reference/performance-reporting/periodic/retrieve-performance-statistics-for-all-m...
- credentials (object) required: Credentials for making authenticated requests requires either username/password or access_token.
- end_date (string): Date in the format YYYY-MM-DD.
- geo_location_breakdown (string): The granularity used for geo location data in reports.
- report_granularity (string): The granularity used for periodic data in reports. See <a href="https://amplifyv01.docs.apiary.io/#reference/performance-reporting/periodic/retrieve-performance-statistics-for-a...
- start_date (string) required: Date in the format YYYY-MM-DD eg. 2017-01-25. Any data before this date will not be replicated.
- secret fields: credentials.access_token, credentials.password

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
pm connectors inspect source-outbrain-amplify
```

### Inspect as JSON

```bash
pm connectors inspect source-outbrain-amplify --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Outbrain Amplify documentation](https://developer.outbrain.com/home-page/amplify-api/documentation/)
