---
name: pm-source-google-ads
description: Google Ads connector knowledge and safe action guide.
---

# pm-source-google-ads

## Purpose

Google Ads catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/google-adwords.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/google-ads/api/docs/release-notes

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Release notes: https://developers.google.com/google-ads/api/docs/release-notes
- Developer blog: https://ads-developers.googleblog.com/

## Configuration

- conversion_window_days (integer): A conversion window is the number of days after an ad interaction (such as an ad click or video view) during which a conversion, such as a purchase, is recorded in Google Ads. F...
- credentials (object) required
- custom_queries_array (array)
- customer_id (string): manual intervention needed
- customer_status_filter (array): A list of customer statuses to filter on. For detailed info about what each status mean refer to Google Ads <a href="https://developers.google.com/google-ads/api/reference/rpc/v...
- end_date (string): UTC date in the format YYYY-MM-DD. Any data after this date will not be replicated. (Default value of today is used if not set)
- num_workers (integer): The number of concurrent threads to use for syncing. Increasing this value may speed up syncs for accounts with many customers or streams. Adjust based on your API usage and rat...
- start_date (string): UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated. (Default value of two years ago is used if not set)
- secret fields: credentials.access_token, credentials.client_secret, credentials.developer_token, credentials.refresh_token

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
pm connectors inspect source-google-ads
```

### Inspect as JSON

```bash
pm connectors inspect source-google-ads --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Release notes](https://developers.google.com/google-ads/api/docs/release-notes)
- [Developer blog](https://ads-developers.googleblog.com/)
